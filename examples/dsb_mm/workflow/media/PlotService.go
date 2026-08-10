package media

import (
	"context"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type Plot struct {
	PlotID  int
	PlotStr string
}

type PlotService interface {
	ReadPlot(ctx context.Context, reqID int, plotID int) (string, error)
	WritePlot(ctx context.Context, reqID int, plotID int, plot string) error
}

type PlotServiceImpl struct {
	plotCache backend.Cache
	plotDB    backend.NoSQLDatabase
}

func NewPlotServiceImpl(plotCache backend.Cache, plotDB backend.NoSQLDatabase) (PlotService, error) {
	return &PlotServiceImpl{plotCache: plotCache, plotDB: plotDB}, nil
}

func (p *PlotServiceImpl) ReadPlot(ctx context.Context, reqID int, plotID int) (string, error) {
	key := strconv.Itoa(plotID)
	var plot string
	found, err := p.plotCache.Get(ctx, key, &plot)
	if err != nil {
		return "", err
	}
	if found {
		return plot, nil
	}
	collection, err := p.plotDB.GetCollection(ctx, "plot", "plot")
	if err != nil {
		return "", err
	}
	result, err := collection.FindOne(ctx, bson.D{{"plotid", plotID}})
	if err != nil {
		return "", err
	}
	var stored Plot
	found, err = result.One(ctx, &stored)
	if err != nil || !found {
		return "", err
	}
	if err := p.plotCache.Put(ctx, key, stored.PlotStr); err != nil {
		return "", err
	}
	return stored.PlotStr, nil
}

func (p *PlotServiceImpl) WritePlot(ctx context.Context, reqID int, plotID int, plot string) error {
	collection, err := p.plotDB.GetCollection(ctx, "plot", "plot")
	if err != nil {
		return err
	}
	stored := Plot{PlotID: plotID, PlotStr: plot}
	if err := collection.InsertOne(ctx, stored); err != nil {
		return err
	}
	return p.plotCache.Put(ctx, strconv.Itoa(plotID), plot)
}
