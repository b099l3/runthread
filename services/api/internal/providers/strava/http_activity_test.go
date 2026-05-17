package strava

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestHTTPActivityFetcherListsBackfillActivities(t *testing.T) {
	var authHeader string
	var rawQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/athlete/activities" {
			t.Fatalf("path = %q, want activities", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		rawQuery = r.URL.RawQuery
		_ = json.NewEncoder(w).Encode([]map[string]any{
			{"id": 12345},
			{"id": 67890},
		})
	}))
	defer server.Close()

	summaries, err := HTTPActivityFetcher{
		BaseURL: server.URL,
		Tokens:  staticAccessTokenProvider{token: "access-token"},
	}.ListBackfillActivities(context.Background(), BackfillListRequest{
		Connection: repository.ProviderConnection{ID: "connection-1"},
		Since:      time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC),
		Until:      time.Date(2026, time.June, 30, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("ListBackfillActivities returned error: %v", err)
	}

	if authHeader != "Bearer access-token" {
		t.Fatalf("authorization = %q, want bearer token", authHeader)
	}
	if !strings.Contains(rawQuery, "per_page=200") {
		t.Fatalf("query = %q, want per page", rawQuery)
	}
	if len(summaries) != 2 || summaries[0].ActivityID != "12345" {
		t.Fatalf("summaries = %#v, want ids", summaries)
	}
}

func TestHTTPActivityFetcherFetchesActivityDetail(t *testing.T) {
	startedAt := time.Date(2026, time.June, 10, 7, 30, 0, 0, time.UTC)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/activities/12345" {
			t.Fatalf("path = %q, want activity detail", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id":                12345,
			"athlete":           map[string]any{"id": 999},
			"sport_type":        "Run",
			"name":              "Morning Run",
			"start_date":        startedAt.Format(time.RFC3339),
			"elapsed_time":      1800,
			"moving_time":       1700,
			"distance":          5000.5,
			"average_heartrate": 151.6,
		})
	}))
	defer server.Close()

	payload, err := HTTPActivityFetcher{
		BaseURL: server.URL,
		Tokens:  staticAccessTokenProvider{token: "access-token"},
	}.FetchActivityDetail(context.Background(), ActivityDetailRequest{
		Connection: repository.ProviderConnection{ID: "connection-1", AthleteID: "athlete-1"},
		ActivityID: "12345",
	})
	if err != nil {
		t.Fatalf("FetchActivityDetail returned error: %v", err)
	}

	if payload.ActivityID != "12345" {
		t.Fatalf("activity id = %q, want 12345", payload.ActivityID)
	}
	if payload.AthleteID != "athlete-1" {
		t.Fatalf("athlete id = %q, want runthread athlete", payload.AthleteID)
	}
	if payload.AverageHeartRate != 152 {
		t.Fatalf("average heart rate = %d, want rounded 152", payload.AverageHeartRate)
	}
}

func TestHTTPActivityFetcherMapsRateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer server.Close()

	_, err := HTTPActivityFetcher{
		BaseURL: server.URL,
		Tokens:  staticAccessTokenProvider{token: "access-token"},
	}.ListBackfillActivities(context.Background(), BackfillListRequest{
		Connection: repository.ProviderConnection{ID: "connection-1"},
	})
	if err != ErrRateLimited {
		t.Fatalf("error = %v, want ErrRateLimited", err)
	}
}

type staticAccessTokenProvider struct {
	token string
	err   error
}

func (p staticAccessTokenProvider) AccessToken(ctx context.Context, connection repository.ProviderConnection) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if p.err != nil {
		return "", p.err
	}
	return p.token, nil
}
