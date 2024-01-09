// package config implements ts-config-service from the train ticket application
package config

import (
	"context"
	"errors"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"go.mongodb.org/mongo-driver/bson"
)

// Config Service manages Config variables in the application
type ConfigService interface {
	// Creates a new config variable
	Create(ctx context.Context, conf Config) error
	// Updates an existing `conf` config variable
	Update(ctx context.Context, conf Config) (bool, error)
	// Find a config variable using its `name`
	Find(ctx context.Context, name string) (Config, error)
	// Deletes an existing config variable using its `name`
	Delete(ctx context.Context, name string) error
	// Find all config variables
	FindAll(ctx context.Context) ([]Config, error)
}

// Implementation of Config Service
type ConfigServiceImpl struct {
	db backend.NoSQLDatabase
}

// Creates a new ConfigService object
func NewConfigServiceImpl(ctx context.Context, db backend.NoSQLDatabase) (*ConfigServiceImpl, error) {
	return &ConfigServiceImpl{db: db}, nil
}

func (c *ConfigServiceImpl) Create(ctx context.Context, conf Config) error {
	coll, err := c.db.GetCollection(ctx, "config", "config")
	if err != nil {
		return err
	}
	query := bson.D{{"name", conf.Name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return err
	}
	var saved_conf Config
	exists, err := res.One(ctx, &saved_conf)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("Config with name " + conf.Name + " already exists")
	}
	return coll.InsertOne(ctx, conf)
}

func (c *ConfigServiceImpl) Update(ctx context.Context, conf Config) (bool, error) {
	coll, err := c.db.GetCollection(ctx, "config", "config")
	if err != nil {
		return false, err
	}
	query := bson.D{{"name", conf.Name}}
	return coll.Upsert(ctx, query, conf)
}

func (c *ConfigServiceImpl) Find(ctx context.Context, name string) (Config, error) {
	coll, err := c.db.GetCollection(ctx, "config", "config")
	if err != nil {
		return Config{}, err
	}
	query := bson.D{{"name", name}}
	res, err := coll.FindOne(ctx, query)
	if err != nil {
		return Config{}, err
	}
	var conf Config
	exists, err := res.One(ctx, &conf)
	if err != nil {
		return Config{}, err
	}
	if !exists {
		return Config{}, errors.New("Config with name " + name + " does not exist")
	}
	return conf, nil
}

func (c *ConfigServiceImpl) Delete(ctx context.Context, name string) error {
	coll, err := c.db.GetCollection(ctx, "config", "config")
	if err != nil {
		return err
	}
	query := bson.D{{"name", name}}
	return coll.DeleteOne(ctx, query)
}

func (c *ConfigServiceImpl) FindAll(ctx context.Context) ([]Config, error) {
	coll, err := c.db.GetCollection(ctx, "config", "config")
	if err != nil {
		return []Config{}, err
	}
	var configs []Config
	res, err := coll.FindMany(ctx, bson.D{})
	if err != nil {
		return configs, err
	}
	err = res.All(ctx, &configs)
	if err != nil {
		return configs, err
	}
	return configs, nil
}
