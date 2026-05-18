package strava

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

type AccessTokenProvider interface {
	AccessToken(ctx context.Context, connection repository.ProviderConnection) (string, error)
}

type HTTPActivityFetcher struct {
	Client  *http.Client
	BaseURL string
	Tokens  AccessTokenProvider
}

var _ ActivityFetcher = HTTPActivityFetcher{}

func (f HTTPActivityFetcher) ListBackfillActivities(ctx context.Context, req BackfillListRequest) ([]MockActivitySummary, error) {
	values := url.Values{}
	values.Set("page", "1")
	values.Set("per_page", "200")
	if !req.Since.IsZero() {
		values.Set("after", strconv.FormatInt(req.Since.Unix(), 10))
	}
	if !req.Until.IsZero() {
		values.Set("before", strconv.FormatInt(req.Until.Unix(), 10))
	}

	body, err := f.get(ctx, req.Connection, "/api/v3/athlete/activities", values)
	if err != nil {
		return nil, err
	}

	var payload []stravaActivitySummary
	if err := decodeJSON(body, &payload); err != nil {
		return nil, fmt.Errorf("decode Strava activity summaries: %w", err)
	}
	summaries := make([]MockActivitySummary, 0, len(payload))
	for _, activity := range payload {
		id := activity.ID.String()
		if id == "" {
			continue
		}
		summaries = append(summaries, MockActivitySummary{ActivityID: id})
	}
	return summaries, nil
}

func (f HTTPActivityFetcher) FetchActivityDetail(ctx context.Context, req ActivityDetailRequest) (MockActivityPayload, error) {
	if strings.TrimSpace(req.ActivityID) == "" {
		return MockActivityPayload{}, fmt.Errorf("strava activity id is required")
	}
	body, err := f.get(ctx, req.Connection, "/api/v3/activities/"+url.PathEscape(req.ActivityID), nil)
	if err != nil {
		return MockActivityPayload{}, err
	}

	var payload stravaActivityDetail
	if err := decodeJSON(body, &payload); err != nil {
		return MockActivityPayload{}, fmt.Errorf("decode Strava activity detail: %w", err)
	}
	return payload.mockPayload(req.Connection.AthleteID)
}

func (f HTTPActivityFetcher) get(ctx context.Context, connection repository.ProviderConnection, path string, values url.Values) ([]byte, error) {
	if f.Tokens == nil {
		return nil, fmt.Errorf("strava access token provider is required")
	}
	accessToken, err := f.Tokens.AccessToken(ctx, connection)
	if err != nil {
		return nil, err
	}

	endpoint := strings.TrimRight(f.baseURL(), "/") + path
	if len(values) > 0 {
		endpoint += "?" + values.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("build Strava API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := f.Client
	if client == nil {
		client = http.DefaultClient
	}
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: call Strava API: %v", ErrTemporaryFailure, err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read Strava API response: %w", err)
	}
	if response.StatusCode == http.StatusTooManyRequests {
		return nil, ErrRateLimited
	}
	if response.StatusCode >= 500 {
		return nil, fmt.Errorf("%w: Strava API returned status %d: %s", ErrTemporaryFailure, response.StatusCode, strings.TrimSpace(string(body)))
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, fmt.Errorf("Strava API returned status %d: %s", response.StatusCode, strings.TrimSpace(string(body)))
	}
	return body, nil
}

func (f HTTPActivityFetcher) baseURL() string {
	if f.BaseURL == "" {
		return "https://www.strava.com"
	}
	return f.BaseURL
}

type stravaActivitySummary struct {
	ID json.Number `json:"id"`
}

type stravaActivityDetail struct {
	ID               json.Number `json:"id"`
	SportType        string      `json:"sport_type"`
	Type             string      `json:"type"`
	Name             string      `json:"name"`
	StartDate        time.Time   `json:"start_date"`
	ElapsedTime      int         `json:"elapsed_time"`
	MovingTime       int         `json:"moving_time"`
	DistanceMeters   float64     `json:"distance"`
	AverageHeartRate float64     `json:"average_heartrate"`
}

func (d stravaActivityDetail) mockPayload(athleteID string) (MockActivityPayload, error) {
	activityID := d.ID.String()
	if activityID == "" {
		return MockActivityPayload{}, fmt.Errorf("strava activity id is required")
	}
	if athleteID == "" {
		return MockActivityPayload{}, fmt.Errorf("runthread athlete id is required")
	}
	sportType := d.SportType
	if sportType == "" {
		sportType = d.Type
	}
	return MockActivityPayload{
		ActivityID:       activityID,
		AthleteID:        athleteID,
		StravaSportType:  sportType,
		Name:             d.Name,
		StartDate:        d.StartDate,
		ElapsedTime:      d.ElapsedTime,
		MovingTime:       d.MovingTime,
		DistanceMeters:   d.DistanceMeters,
		AverageHeartRate: int(math.Round(d.AverageHeartRate)),
	}, nil
}

func decodeJSON(body []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	return decoder.Decode(target)
}
