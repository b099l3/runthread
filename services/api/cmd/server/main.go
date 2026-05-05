package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/runthread/runthread/services/api/internal/app"
	"github.com/runthread/runthread/services/api/internal/config"
	rpchandler "github.com/runthread/runthread/services/api/internal/rpc/handler"
	"github.com/runthread/runthread/services/api/internal/rpc/runthread/v1/runthreadv1connect"
	"github.com/runthread/runthread/services/api/internal/startup"
)

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

	mux := newMux(services)

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

func newMux(services app.Services) *http.ServeMux {
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
	return mux
}
