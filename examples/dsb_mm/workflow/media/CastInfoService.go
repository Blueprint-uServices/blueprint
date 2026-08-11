package media

import (
	"context"
	"strconv"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

type CastInfoService interface {
	WriteCastInfo(ctx context.Context, reqID int, castInfoID int, name string, gender bool, intro string) error
	ReadCastInfo(ctx context.Context, reqID int, castInfoIDs []int) ([]CastInfo, error)
}

type CastInfoServiceImpl struct {
	castInfoCache backend.Cache
	castInfoDB    backend.NoSQLDatabase
}

func NewCastInfoService(ctx context.Context, castInfoCache backend.Cache, castInfoDB backend.NoSQLDatabase) (CastInfoService, error) {
	return &CastInfoServiceImpl{castInfoCache: castInfoCache, castInfoDB: castInfoDB}, nil
}

func (c *CastInfoServiceImpl) WriteCastInfo(ctx context.Context, reqID int, castInfoID int, name string, gender bool, intro string) error {
	collection, err := c.castInfoDB.GetCollection(ctx, "cast-info", "cast-info")
	if err != nil {
		return err
	}
	info := CastInfo{CastInfoId: castInfoID, Name: name, Gender: gender, Intro: intro}
	if err := collection.InsertOne(ctx, info); err != nil {
		return err
	}
	return c.castInfoCache.Put(ctx, strconv.Itoa(castInfoID), info)
}

func (c *CastInfoServiceImpl) ReadCastInfo(ctx context.Context, reqID int, castInfoIDs []int) ([]CastInfo, error) {
	infos := make([]CastInfo, 0, len(castInfoIDs))

	for _, id := range castInfoIDs {
		var info CastInfo
		found, err := c.castInfoCache.Get(ctx, strconv.Itoa(id), &info)
		if err != nil {
			return nil, err
		}
		if !found {
			collection, err := c.castInfoDB.GetCollection(ctx, "cast-info", "cast-info")
			if err != nil {
				return nil, err
			}
			result, err := collection.FindOne(ctx, bson.D{{"castinfoid", id}})
			if err != nil {
				return nil, err
			}
			found, err = result.One(ctx, &info)
			if err != nil {
				return nil, err
			}
			if !found {
				continue
			}
			if err := c.castInfoCache.Put(ctx, strconv.Itoa(id), info); err != nil {
				return nil, err
			}
		}
		infos = append(infos, info)
	}
	return infos, nil
}
