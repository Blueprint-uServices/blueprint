package admin

import (
	"context"

	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/config"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/contacts"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/price"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/station"
	"github.com/blueprint-uservices/blueprint/examples/train_ticket/workflow/train"
)

type AdminBasicService interface {
	GetAllContacts(ctx context.Context) ([]contacts.Contact, error)
	DeleteContacts(ctx context.Context, contactId string) error
	ModifyContacts(ctx context.Context, c contacts.Contact) (contacts.Contact, error)
	AddContacts(ctx context.Context, c contacts.Contact) (contacts.Contact, error)
	GetAllStations(ctx context.Context) ([]station.Station, error)
	DeleteStation(ctx context.Context, s station.Station) error
	AddStation(ctx context.Context, s station.Station) (station.Station, error)
	ModifyStation(ctx context.Context, s station.Station) (station.Station, error)
	GetAllTrains(ctx context.Context) ([]train.TrainType, error)
	DeleteTrain(ctx context.Context, trainId string) error
	ModifyTrain(ctx context.Context, t train.TrainType) (train.TrainType, error)
	AddTrain(ctx context.Context, t train.TrainType) (train.TrainType, error)
	GetAllPrices(ctx context.Context) ([]price.PriceConfig, error)
	DeletePrice(ctx context.Context, pc price.PriceConfig) error
	ModifyPrice(ctx context.Context, pc price.PriceConfig) (price.PriceConfig, error)
	AddPrice(ctx context.Context, pc price.PriceConfig) (price.PriceConfig, error)
	GetAllConfigs(ctx context.Context) ([]config.Config, error)
	DeleteConfig(ctx context.Context, name string) error
	ModifyConfig(ctx context.Context, c config.Config) (config.Config, error)
	AddConfig(ctx context.Context, c config.Config) (config.Config, error)
}

type AdminBasicServiceImpl struct {
	contactsService contacts.ContactsService
	stationService  station.StationService
	trainService    train.TrainService
	priceService    price.PriceService
	configService   config.ConfigService
}

func NewAdminBasicServiceImpl(ctx context.Context, contactsService contacts.ContactsService, stationService station.StationService, trainService train.TrainService, priceService price.PriceService, configService config.ConfigService) (*AdminBasicServiceImpl, error) {
	return &AdminBasicServiceImpl{contactsService, stationService, trainService, priceService, configService}, nil
}

func (a *AdminBasicServiceImpl) GetAllContacts(ctx context.Context) ([]contacts.Contact, error) {
	return a.contactsService.GetAllContacts(ctx)
}

func (a *AdminBasicServiceImpl) DeleteContacts(ctx context.Context, id string) error {
	c, err := a.contactsService.FindContactsById(ctx, id)
	if err != nil {
		return err
	}
	return a.contactsService.Delete(ctx, c)
}

func (a *AdminBasicServiceImpl) AddContacts(ctx context.Context, c contacts.Contact) (contacts.Contact, error) {
	return c, a.contactsService.CreateContacts(ctx, c)
}

func (a *AdminBasicServiceImpl) ModifyContacts(ctx context.Context, c contacts.Contact) (contacts.Contact, error) {
	_, err := a.contactsService.Modify(ctx, c)
	return c, err
}

func (a *AdminBasicServiceImpl) GetAllStations(ctx context.Context) ([]station.Station, error) {
	return a.stationService.AllStations(ctx)
}

func (a *AdminBasicServiceImpl) DeleteStation(ctx context.Context, s station.Station) error {
	return a.stationService.DeleteStation(ctx, s.ID)
}

func (a *AdminBasicServiceImpl) AddStation(ctx context.Context, s station.Station) (station.Station, error) {
	return s, a.stationService.CreateStation(ctx, s)
}

func (a *AdminBasicServiceImpl) ModifyStation(ctx context.Context, s station.Station) (station.Station, error) {
	_, err := a.stationService.UpdateStation(ctx, s)
	return s, err
}

func (a *AdminBasicServiceImpl) GetAllTrains(ctx context.Context) ([]train.TrainType, error) {
	return a.trainService.AllTrains(ctx)
}

func (a *AdminBasicServiceImpl) DeleteTrain(ctx context.Context, trainId string) error {
	_, err := a.trainService.Delete(ctx, trainId)
	return err
}
func (a *AdminBasicServiceImpl) ModifyTrain(ctx context.Context, t train.TrainType) (train.TrainType, error) {
	_, err := a.trainService.Update(ctx, t)
	return t, err
}

func (a *AdminBasicServiceImpl) AddTrain(ctx context.Context, t train.TrainType) (train.TrainType, error) {
	_, err := a.trainService.Create(ctx, t)
	return t, err
}

func (a *AdminBasicServiceImpl) GetAllPrices(ctx context.Context) ([]price.PriceConfig, error) {
	return a.priceService.GetAllPriceConfig(ctx)
}

func (a *AdminBasicServiceImpl) DeletePrice(ctx context.Context, pc price.PriceConfig) error {
	return a.priceService.DeletePriceConfig(ctx, pc.ID)
}

func (a *AdminBasicServiceImpl) ModifyPrice(ctx context.Context, pc price.PriceConfig) (price.PriceConfig, error) {
	_, err := a.priceService.UpdatePriceConfig(ctx, pc)
	return pc, err
}

func (a *AdminBasicServiceImpl) AddPrice(ctx context.Context, pc price.PriceConfig) (price.PriceConfig, error) {
	return pc, a.priceService.CreateNewPriceConfig(ctx, pc)
}

func (a *AdminBasicServiceImpl) GetAllConfigs(ctx context.Context) ([]config.Config, error) {
	return a.configService.FindAll(ctx)
}

func (a *AdminBasicServiceImpl) DeleteConfig(ctx context.Context, name string) error {
	return a.configService.Delete(ctx, name)
}

func (a *AdminBasicServiceImpl) ModifyConfig(ctx context.Context, conf config.Config) (config.Config, error) {
	_, err := a.configService.Update(ctx, conf)
	return conf, err
}

func (a *AdminBasicServiceImpl) AddConfig(ctx context.Context, c config.Config) (config.Config, error) {
	return c, a.configService.Create(ctx, c)
}
