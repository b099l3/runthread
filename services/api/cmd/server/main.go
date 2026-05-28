package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/config"
	"github.com/runthread/runthread/services/api/internal/domain"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/postgres"
	"github.com/runthread/runthread/services/api/internal/providerimport"
	"github.com/runthread/runthread/services/api/internal/providers/strava"
	"github.com/runthread/runthread/services/api/internal/repository"
	rpchandler "github.com/runthread/runthread/services/api/internal/rpc/handler"
	"github.com/runthread/runthread/services/api/internal/rpc/runthread/v1/runthreadv1connect"
	"github.com/runthread/runthread/services/api/internal/startup"
)

const stravaWebhookRetryBatchLimit = 25
const stravaInitialBackfillTimeout = 2 * time.Minute

func main() {
	cfg := config.Load()
	storage, err := startup.ComposeStorage(cfg)
	if err != nil {
		log.Printf("api server storage setup failed error=%v", err)
		os.Exit(1)
	}
	defer func() {
		if err := storage.Cleanup(); err != nil {
			log.Printf("api server storage cleanup failed error=%v", err)
		}
	}()
	services, err := app.NewServices(storage.Store)
	if err != nil {
		log.Printf("api server app setup failed error=%v", err)
		os.Exit(1)
	}
	stravaRuntime, err := composeStravaRuntime(cfg, storage)
	if err != nil {
		log.Printf("api server Strava setup failed error=%v", err)
		os.Exit(1)
	}
	if stravaRuntime != nil {
		services.ProviderConnect.ProviderStarters = map[string]app.ProviderConnectionStarter{
			app.ProviderStrava: strava.ConnectionStarter{
				OAuth:       stravaRuntime.OAuth,
				RedirectURI: cfg.StravaOAuthRedirectURI,
				Scopes:      []string{"activity:read_all"},
			},
		}
	}

	mux := newMux(services, muxOptions{
		strava: stravaRuntime,
	})

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if stravaRuntime != nil {
		startStravaWebhookRetryWorker(ctx, stravaRuntime, time.Duration(cfg.StravaWebhookRetryIntervalSeconds)*time.Second)
	}

	go func() {
		log.Printf("api server starting addr=%s storage=%s database_configured=%t", server.Addr, storage.Kind, cfg.DatabaseConfigured())
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("api server failed error=%v", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("api server shutdown failed error=%v", err)
		os.Exit(1)
	}

	log.Print("api server stopped")
}

type muxOptions struct {
	strava *stravaRuntime
}

func newMux(services app.Services, options ...muxOptions) *http.ServeMux {
	var option muxOptions
	if len(options) > 0 {
		option = options[0]
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}` + "\n"))
	})
	runthreadPath, runthreadHandler := runthreadv1connect.NewRunthreadServiceHandler(
		rpchandler.NewRunthreadService(services),
	)
	mux.Handle(runthreadPath, runthreadHandler)
	if option.strava != nil {
		mux.HandleFunc("/providers/strava/oauth/callback", stravaOAuthCallbackHandler(option.strava))
		mux.HandleFunc("/providers/strava/webhook", stravaWebhookHandler(option.strava))
		mux.HandleFunc("/providers/strava/retry-imports", stravaRetryImportsHandler(option.strava))
	}
	return mux
}

type stravaRuntime struct {
	OAuth              *strava.OAuthService
	Backfill           strava.BackfillService
	Webhook            strava.WebhookService
	WebhookRetries     strava.WebhookService
	Store              repository.Store
	RedirectURI        string
	WebhookVerifyToken string
}

func composeStravaRuntime(cfg config.Config, storage startup.Storage) (*stravaRuntime, error) {
	if !cfg.StravaOAuthConfigured() {
		return nil, nil
	}
	store := storage.Store
	providerStore, ok := store.(repository.ProviderStore)
	if !ok {
		return nil, nil
	}
	baseURL := strings.TrimRight(cfg.StravaAPIBaseURL, "/")
	if baseURL == "" {
		baseURL = config.DefaultStravaAPIBaseURL
	}
	tokenStore, err := composeStravaTokenStore(cfg, storage)
	if err != nil {
		return nil, err
	}
	exchanger := strava.HTTPCodeExchanger{
		ClientID:     cfg.StravaClientID,
		ClientSecret: cfg.StravaClientSecret,
		TokenURL:     baseURL + "/oauth/token",
	}
	oauth := &strava.OAuthService{
		Store:        providerStore,
		Exchanger:    exchanger,
		Tokens:       tokenStore,
		ClientID:     cfg.StravaClientID,
		AuthorizeURL: baseURL + "/oauth/authorize",
	}
	importer, err := providerimport.NewService(providerStore, store)
	if err != nil {
		return nil, fmt.Errorf("compose Strava provider import: %w", err)
	}
	tokenManager := strava.TokenManager{
		Store:     tokenStore,
		Refresher: exchanger,
	}
	webhookDeduper, err := composeStravaWebhookDeduper(storage)
	if err != nil {
		return nil, err
	}
	fetcher := strava.HTTPActivityFetcher{
		BaseURL: baseURL,
		Tokens:  tokenManager,
	}
	webhookService := strava.WebhookService{
		Providers: providerStore,
		Importer:  importer,
		Fetcher:   fetcher,
		Verifier: strava.SignatureVerifier{
			SigningSecret: cfg.StravaClientSecret,
		},
		Deduper: webhookDeduper,
	}
	return &stravaRuntime{
		OAuth: oauth,
		Backfill: strava.BackfillService{
			Providers: providerStore,
			Importer:  importer,
			Fetcher:   fetcher,
		},
		Webhook:            webhookService,
		WebhookRetries:     webhookService,
		Store:              store,
		RedirectURI:        cfg.StravaOAuthRedirectURI,
		WebhookVerifyToken: cfg.StravaWebhookVerifyToken,
	}, nil
}

func composeStravaWebhookDeduper(storage startup.Storage) (strava.WebhookDeduper, error) {
	if storage.DB == nil {
		return strava.NewInMemoryWebhookDeduper(), nil
	}
	deduper, err := postgres.NewProviderWebhookDeduper(storage.DB, strava.ProviderName)
	if err != nil {
		return nil, fmt.Errorf("compose Strava webhook deduper: %w", err)
	}
	return deduper, nil
}

func composeStravaTokenStore(cfg config.Config, storage startup.Storage) (interface {
	strava.TokenStore
	strava.TokenRepository
}, error) {
	if storage.DB == nil {
		return strava.NewInMemoryTokenStore(), nil
	}
	store, err := postgres.NewStravaTokenStore(storage.DB, cfg.ProviderTokenKey)
	if err != nil {
		return nil, fmt.Errorf("compose Strava token store: %w", err)
	}
	return store, nil
}

func stravaOAuthCallbackHandler(runtime *stravaRuntime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		query := r.URL.Query()
		if providerError := query.Get("error"); providerError != "" {
			http.Error(w, "strava authorization failed", http.StatusBadRequest)
			return
		}
		response, err := runtime.OAuth.CompleteOAuthCallback(r.Context(), strava.CompleteOAuthCallbackRequest{
			State:       query.Get("state"),
			Code:        query.Get("code"),
			RedirectURI: runtime.RedirectURI,
		})
		if err != nil {
			http.Error(w, "strava authorization callback failed", http.StatusBadRequest)
			return
		}
		runStravaInitialBackfill(runtime, response.Connection.ID)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(stravaOAuthCallbackPage("runthread://provider/strava/connected")))
	}
}

func stravaOAuthCallbackPage(returnURL string) string {
	return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Strava connected</title>
</head>
<body>
  <h1>Strava connected</h1>
  <p>You can return to Runthread.</p>
  <p><a href="` + returnURL + `">Open Runthread</a></p>
</body>
</html>
`
}

func runStravaInitialBackfill(runtime *stravaRuntime, providerConnectionID string) {
	backfillCtx, cancel := context.WithTimeout(context.Background(), stravaInitialBackfillTimeout)
	defer cancel()

	backfillResult, err := runtime.Backfill.RunInitialBackfill(backfillCtx, strava.RunBackfillRequest{
		ProviderConnectionID: providerConnectionID,
	})
	if err != nil && backfillResult.Status != strava.BackfillStatusDeferred && backfillResult.Status != strava.BackfillStatusPartial {
		log.Printf("strava initial backfill failed provider_connection_id=%s status=%s error=%v", providerConnectionID, backfillResult.Status, err)
	}
	if err := completeStravaBackfillImports(backfillCtx, runtime.Store, backfillResult); err != nil {
		log.Printf("strava initial backfill completion failed provider_connection_id=%s error=%v", providerConnectionID, err)
	}
}

func stravaWebhookHandler(runtime *stravaRuntime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleStravaWebhookValidation(w, r, runtime)
		case http.MethodPost:
			handleStravaWebhookEvent(w, r, runtime)
		default:
			w.Header().Set("Allow", http.MethodGet+", "+http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

func handleStravaWebhookValidation(w http.ResponseWriter, r *http.Request, runtime *stravaRuntime) {
	query := r.URL.Query()
	if query.Get("hub.mode") != "subscribe" || query.Get("hub.verify_token") == "" || query.Get("hub.verify_token") != runtime.WebhookVerifyToken {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	challenge := query.Get("hub.challenge")
	if challenge == "" {
		http.Error(w, "missing challenge", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"hub.challenge": challenge})
}

func handleStravaWebhookEvent(w http.ResponseWriter, r *http.Request, runtime *stravaRuntime) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		http.Error(w, "read Strava webhook body failed", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	result, err := runtime.Webhook.HandleWebhook(r.Context(), strava.HandleWebhookRequest{
		Body:      body,
		Signature: r.Header.Get("X-Strava-Signature"),
	})
	if err != nil {
		log.Printf("strava webhook failed action=%s event_id=%s object_type=%s object_id=%s error=%v", result.Action, result.Event.EventID, result.Event.ObjectType, result.Event.ObjectID, err)
		status := http.StatusInternalServerError
		if strings.Contains(err.Error(), "verify Strava webhook") {
			status = http.StatusUnauthorized
		} else if result.Action == "" || strings.Contains(err.Error(), "decode Strava webhook") || strings.Contains(err.Error(), "unsupported Strava webhook") || strings.Contains(err.Error(), "strava webhook event") || strings.Contains(err.Error(), "webhook body") {
			status = http.StatusBadRequest
		} else if strava.IsRetryableWebhookError(err) {
			status = http.StatusServiceUnavailable
			w.Header().Set("Retry-After", stravaWebhookRetryAfter(err))
		}
		http.Error(w, "strava webhook failed", status)
		return
	}
	if result.Import != nil && result.Import.ImportedActivity != nil {
		backfill := strava.BackfillResult{Imports: []providerimport.ImportResult{*result.Import}}
		if err := completeStravaBackfillImports(r.Context(), runtime.Store, backfill); err != nil {
			log.Printf("strava webhook completion failed event_id=%s imported_activity_id=%s error=%v", result.Event.EventID, result.Import.ImportedActivity.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": string(result.Action)})
}

func stravaWebhookRetryAfter(err error) string {
	if errors.Is(err, strava.ErrRateLimited) {
		return "900"
	}
	return "60"
}

type retryStravaImportsHTTPBody struct {
	AthleteID      string `json:"athleteId"`
	GoalID         string `json:"goalId"`
	TargetWeekDate string `json:"targetWeekDate"`
	Limit          int    `json:"limit"`
}

func stravaRetryImportsHandler(runtime *stravaRuntime) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		defer r.Body.Close()

		var body retryStravaImportsHTTPBody
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&body); err != nil {
			http.Error(w, "decode retry imports request failed", http.StatusBadRequest)
			return
		}
		var targetWeekDate time.Time
		if body.TargetWeekDate != "" {
			parsed, err := time.Parse("2006-01-02", body.TargetWeekDate)
			if err != nil {
				http.Error(w, "targetWeekDate must be YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			targetWeekDate = parsed
		}

		result, err := app.ImportedActivityRetryService{Store: runtime.Store}.RetryImportedActivities(r.Context(), app.RetryImportedActivitiesRequest{
			AthleteID:      body.AthleteID,
			GoalID:         body.GoalID,
			TargetWeekDate: targetWeekDate,
			Limit:          body.Limit,
		})
		if err != nil {
			http.Error(w, "retry Strava imports failed", http.StatusInternalServerError)
			log.Printf("strava retry imports failed athlete_id=%s error=%v", body.AthleteID, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(result)
	}
}

func startStravaWebhookRetryWorker(ctx context.Context, runtime *stravaRuntime, interval time.Duration) {
	if runtime == nil || interval <= 0 {
		return
	}
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Printf("strava webhook retry worker started interval=%s", interval)
		for {
			select {
			case <-ctx.Done():
				log.Print("strava webhook retry worker stopped")
				return
			case <-ticker.C:
				result, err := retryStravaWebhookImports(ctx, runtime)
				if err != nil {
					log.Printf("strava webhook retry worker failed error=%v", err)
					continue
				}
				if result.Attempted > 0 || result.Skipped > 0 || result.Failed > 0 {
					log.Printf("strava webhook retry worker completed attempted=%d succeeded=%d failed=%d skipped=%d", result.Attempted, result.Succeeded, result.Failed, result.Skipped)
				}
			}
		}
	}()
}

func retryStravaWebhookImports(ctx context.Context, runtime *stravaRuntime) (strava.RetryWebhookImportsResult, error) {
	if runtime == nil {
		return strava.RetryWebhookImportsResult{}, fmt.Errorf("strava runtime is required")
	}
	return runtime.WebhookRetries.RetryFailedWebhookImports(ctx, strava.RetryWebhookImportsRequest{
		Limit: stravaWebhookRetryBatchLimit,
	})
}

func completeStravaBackfillImports(ctx context.Context, store repository.Store, backfill strava.BackfillResult) error {
	if store == nil {
		return fmt.Errorf("repository store is required")
	}
	currentPlan := app.CurrentPlanWeekService{
		Store:   store,
		Planner: planning.NewWeeklyPlanner(),
	}
	matcher := app.ProviderActivityMatchService{Store: store}
	completer := app.ProviderActivityCompletionService{Store: store}

	for _, importResult := range backfill.Imports {
		if importResult.ImportedActivity == nil {
			continue
		}
		activity := *importResult.ImportedActivity
		goal, err := currentGoalForActivity(ctx, store, activity)
		if err != nil {
			return err
		}
		plan, err := currentPlan.GetCurrentPlanWeek(ctx, app.GetCurrentPlanWeekRequest{
			AthleteID:      activity.AthleteID,
			GoalID:         goal.ID,
			TargetWeekDate: activity.StartedAt,
		})
		if err != nil {
			return fmt.Errorf("load current plan week for imported activity %q: %w", activity.ID, err)
		}
		match, err := matcher.MatchProviderActivity(ctx, app.MatchProviderActivityRequest{
			ImportedActivity: activity,
			PlanWeek:         plan.PlanWeek,
		})
		if err != nil {
			return fmt.Errorf("match imported activity %q: %w", activity.ID, err)
		}
		if match.WorkoutMatch.Status != domain.WorkoutMatchStatusMatched {
			continue
		}
		if _, err := completer.CompleteMatchedProviderActivity(ctx, app.CompleteMatchedProviderActivityRequest{
			WorkoutMatch:     match.WorkoutMatch,
			ImportedActivity: activity,
			PlanWeek:         match.PlanWeek,
			PlannedWorkout:   match.PlannedWorkout,
		}); err != nil {
			return fmt.Errorf("complete imported activity %q: %w", activity.ID, err)
		}
	}
	return nil
}

func currentGoalForActivity(ctx context.Context, store repository.Store, activity domain.ImportedActivity) (domain.TrainingGoal, error) {
	goal, err := store.GetCurrentTrainingGoal(ctx, activity.AthleteID)
	if err != nil {
		return domain.TrainingGoal{}, fmt.Errorf("get current training goal for athlete %q: %w", activity.AthleteID, err)
	}
	return goal, nil
}
