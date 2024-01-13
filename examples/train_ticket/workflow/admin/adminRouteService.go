package admin

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
)

type AdminRouteService interface {
	GetAllRoutes(ctx context.Context) ([]route.Route, error)
	AddRoute(ctx context.Context, r route.RouteInfo) (route.Route, error)
	DeleteRoute(ctx context.Context, id string) error
}

type AdminRouteServiceImpl struct {
	routeService route.RouteService
}

func NewAdminRouteServiceImpl(ctx context.Context, routeService route.RouteService) (*AdminRouteServiceImpl, error) {
	return &AdminRouteServiceImpl{routeService}, nil
}

func (a *AdminRouteServiceImpl) GetAllRoutes(ctx context.Context) ([]route.Route, error) {
	return a.routeService.GetAllRoutes(ctx)
}

func (a *AdminRouteServiceImpl) AddRoute(ctx context.Context, r route.RouteInfo) (route.Route, error) {
	return a.routeService.CreateAndModify(ctx, r)
}

func (a *AdminRouteServiceImpl) DeleteRoute(ctx context.Context, id string) error {
	return a.routeService.DeleteRoute(ctx, id)
}
