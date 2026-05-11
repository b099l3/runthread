package handler

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"github.com/runthread/runthread/services/api/internal/app"
	rpcv1 "github.com/runthread/runthread/services/api/internal/rpc/runthread/v1"
)

type RunthreadService struct {
	coreLoop        app.CoreLoopService
	currentPlanWeek app.CurrentPlanWeekService
	providerConnect app.ProviderConnectionService
}

func NewRunthreadService(services app.Services) RunthreadService {
	return RunthreadService{
		coreLoop:        services.CoreLoop,
		currentPlanWeek: services.CurrentPlanWeek,
		providerConnect: services.ProviderConnect,
	}
}

func (s RunthreadService) GetCurrentPlanWeek(ctx context.Context, req *connect.Request[rpcv1.GetCurrentPlanWeekRequest]) (*connect.Response[rpcv1.GetCurrentPlanWeekResponse], error) {
	appRequest, err := getCurrentPlanWeekRequestToApp(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response, err := s.currentPlanWeek.GetCurrentPlanWeek(ctx, appRequest)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get current plan week: %w", err))
	}

	return connect.NewResponse(getCurrentPlanWeekResponseFromApp(response)), nil
}

func (s RunthreadService) GetProviderConnectionStatus(ctx context.Context, req *connect.Request[rpcv1.GetProviderConnectionStatusRequest]) (*connect.Response[rpcv1.GetProviderConnectionStatusResponse], error) {
	appRequest, err := getProviderConnectionStatusRequestToApp(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response, err := s.providerConnect.GetProviderConnectionStatus(ctx, appRequest)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("get provider connection status: %w", err))
	}

	return connect.NewResponse(getProviderConnectionStatusResponseFromApp(response)), nil
}

func (s RunthreadService) StartProviderConnection(ctx context.Context, req *connect.Request[rpcv1.StartProviderConnectionRequest]) (*connect.Response[rpcv1.StartProviderConnectionResponse], error) {
	appRequest, err := startProviderConnectionRequestToApp(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response, err := s.providerConnect.StartProviderConnection(ctx, appRequest)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("start provider connection: %w", err))
	}

	return connect.NewResponse(startProviderConnectionResponseFromApp(response)), nil
}

func (s RunthreadService) CompleteImportedActivity(ctx context.Context, req *connect.Request[rpcv1.CompleteImportedActivityRequest]) (*connect.Response[rpcv1.CompleteImportedActivityResponse], error) {
	appRequest, err := completeImportedActivityRequestToApp(req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	response, err := s.coreLoop.CompleteImportedActivity(ctx, appRequest)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("complete imported activity: %w", err))
	}

	return connect.NewResponse(completeImportedActivityResponseFromApp(response)), nil
}
