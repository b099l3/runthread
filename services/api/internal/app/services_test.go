package app

import (
	"testing"

	"github.com/runthread/runthread/services/api/internal/repository"
)

func TestNewServicesBuildsCoreLoopServiceFromStore(t *testing.T) {
	store := repository.NewInMemoryStore()

	services, err := NewServices(store)
	if err != nil {
		t.Fatalf("NewServices returned error: %v", err)
	}

	if services.CoreLoop.Runner == nil {
		t.Fatal("CoreLoop.Runner is nil")
	}
	if services.CoreLoop.Store != store {
		t.Fatalf("CoreLoop.Store = %T, want provided store", services.CoreLoop.Store)
	}
	if services.CurrentPlanWeek.Store != store {
		t.Fatalf("CurrentPlanWeek.Store = %T, want provided store", services.CurrentPlanWeek.Store)
	}
	if services.ProviderConnect.Store != store {
		t.Fatalf("ProviderConnect.Store = %T, want provided store", services.ProviderConnect.Store)
	}
}

func TestNewServicesRequiresStore(t *testing.T) {
	_, err := NewServices(nil)
	if err == nil {
		t.Fatal("NewServices returned nil error, want missing store error")
	}
}
