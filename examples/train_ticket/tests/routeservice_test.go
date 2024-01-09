package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/route"
	"github.com/blueprint-uservices/blueprint/runtime/core/registry"
	"github.com/blueprint-uservices/blueprint/runtime/plugins/simplenosqldb"
	"github.com/stretchr/testify/require"
)

var routeServiceRegistry = registry.NewServiceRegistry[route.RouteService]("route_service")

func init() {
	routeServiceRegistry.Register("local", func(ctx context.Context) (route.RouteService, error) {
		db, err := simplenosqldb.NewSimpleNoSQLDB(ctx)
		if err != nil {
			return nil, err
		}
		return route.NewRouteServiceImpl(ctx, db)
	})
}

func genTestRouteData() ([]route.Route, []route.RouteInfo) {
	routes := []route.Route{}
	routeInfos := []route.RouteInfo{}
	for i := 0; i < 10; i++ {
		stations := []string{}
		distances := []int64{}
		distance_strs := []string{}
		for j := i; j <= i+5; j++ {
			stations = append(stations, fmt.Sprintf("Station%d", j))
			distances = append(distances, int64(j*100+100))
			distance_strs = append(distance_strs, fmt.Sprintf("%d", j*100+100))
		}
		r := route.Route{
			ID:           fmt.Sprintf("Route%d", i),
			StartStation: stations[0],
			EndStation:   stations[len(stations)-1],
			Stations:     stations,
			Distances:    distances,
		}
		rinfo := route.RouteInfo{
			ID:           r.ID,
			StartStation: stations[0],
			EndStation:   stations[len(stations)-1],
			StationList:  strings.Join(stations, ","),
			DistanceList: strings.Join(distance_strs, ","),
		}

		routes = append(routes, r)
		routeInfos = append(routeInfos, rinfo)
	}
	return routes, routeInfos
}

func TestRouteService(t *testing.T) {
	ctx := context.Background()
	service, err := routeServiceRegistry.Get(ctx)
	require.NoError(t, err)

	routes, infos := genTestRouteData()
	// Test Create
	for idx, i := range infos {
		r, err := service.CreateAndModify(ctx, i)
		routes[idx].ID = r.ID
		require.NoError(t, err)
		requireRoute(t, routes[idx], r)
	}

	// Test GetAll Routes
	all_routes, err := service.GetAllRoutes(ctx)
	require.NoError(t, err)
	require.Len(t, all_routes, len(routes))

	// Test GetRouteById
	for _, d := range routes {
		r, err := service.GetRouteById(ctx, d.ID)
		require.NoError(t, err)
		require.Equal(t, d, r)
	}

	// Test GetRoutebyIds
	all_ids := []string{}
	for _, d := range routes {
		all_ids = append(all_ids, d.ID)
	}
	all_routes, err = service.GetRouteByIds(ctx, all_ids)
	require.NoError(t, err)
	require.Len(t, all_routes, len(all_ids))

	// Test GetRouteByStartAndEnd
	for _, d := range routes {
		r, err := service.GetRouteByStartAndEnd(ctx, d.StartStation, d.EndStation)
		require.NoError(t, err)
		require.Equal(t, d, r)
	}

	// Test Delete Route
	for _, d := range routes {
		err = service.DeleteRoute(ctx, d.ID)
	}
}

func requireRoute(t *testing.T, expected route.Route, actual route.Route) {
	require.Equal(t, expected.ID, actual.ID)
	require.Equal(t, expected.StartStation, actual.StartStation)
	require.Equal(t, expected.EndStation, actual.EndStation)
	for idx := 0; idx < len(expected.Stations); idx++ {
		require.Equal(t, expected.Stations[idx], actual.Stations[idx])
		require.Equal(t, expected.Distances[idx], actual.Distances[idx])
	}
}
