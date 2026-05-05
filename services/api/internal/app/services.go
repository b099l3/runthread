package app

import (
	"fmt"

	"github.com/runthread/runthread/services/api/internal/coreloop"
	"github.com/runthread/runthread/services/api/internal/planning"
	"github.com/runthread/runthread/services/api/internal/repository"
)

type Services struct {
	CoreLoop        CoreLoopService
	CurrentPlanWeek CurrentPlanWeekService
}

func NewServices(store repository.Store) (Services, error) {
	if store == nil {
		return Services{}, fmt.Errorf("repository store is required")
	}

	return Services{
		CoreLoop: CoreLoopService{
			Runner: coreloop.NewService(),
			Store:  store,
		},
		CurrentPlanWeek: CurrentPlanWeekService{
			Store:   store,
			Planner: planning.NewWeeklyPlanner(),
		},
	}, nil
}
