// Package ticketoffice implements ts-ticketoffice-service from the original TrainTicket application
package ticketoffice

import (
	"context"
	"encoding/json"
	"os"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
)

var TABLE_QUERY = "CREATE TABLE IF NOT EXISTS office (name VARCHAR(255), city VARCHAR(255), province VARCHAR(255),region VARCHAR(255), address VARCHAR(255), workTime VARCHAR(32), windowNum INT(10));"

type TicketOfficeService interface {
	GetAll(ctx context.Context) ([]Office, error)
	GetRegionList(ctx context.Context) ([]Region, error)
	GetSpecificOffices(ctx context.Context, province string, city string, region string) ([]Office, error)
	AddOffice(ctx context.Context, office Office, province string, region string, city string) error
	DeleteOffice(ctx context.Context, province string, city string, region string, officeName string) error
	UpdateOffice(ctx context.Context, oldName string, province string, region string, city string, office Office) error
}

type TicketOfficeServiceImpl struct {
	db backend.RelationalDB
}

func NewTicketOfficeServiceImpl(ctx context.Context, db backend.RelationalDB) (*TicketOfficeServiceImpl, error) {
	t := &TicketOfficeServiceImpl{db}
	_, err := t.db.Exec(ctx, TABLE_QUERY)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (t *TicketOfficeServiceImpl) GetAll(ctx context.Context) ([]Office, error) {
	var offices []Office
	findQuery := "SELECT * FROM office"
	res, err := t.db.Query(ctx, findQuery)
	if err != nil {
		return offices, err
	}

	for {
		if res.Next() {
			var office Office
			var province, city, region string
			res.Scan(&office.OfficeName, &city, &province, &region, &office.Address, &office.WorkTime, &office.WindowNum)
			offices = append(offices, office)
		} else {
			break
		}
	}

	return offices, nil
}

func (t *TicketOfficeServiceImpl) GetSpecificOffices(ctx context.Context, province string, city string, region string) ([]Office, error) {
	var offices []Office
	selQuery := "SELECT * FROM office WHERE province = ? AND city = ? AND region = ?"
	res, err := t.db.Query(ctx, selQuery, province, city, region)
	if err != nil {
		return offices, err
	}
	for {
		if res.Next() {
			var office Office
			var region, province, city string
			res.Scan(&office.OfficeName, &city, &province, &region, &office.Address, &office.WorkTime, &office.WindowNum)
			offices = append(offices, office)
		} else {
			break
		}
	}
	return offices, nil
}

func (t *TicketOfficeServiceImpl) AddOffice(ctx context.Context, office Office, province string, region string, city string) error {
	query := "INSERT INTO office (name,city,province,region,address,workTime,windowNum) VALUES(?, ?, ?, ?, ?, ?, ?);"
	_, err := t.db.Exec(ctx, query, office.OfficeName, city, province, region, office.Address, office.WorkTime, office.WindowNum)
	if err != nil {
		return err
	}
	return nil
}

func (t *TicketOfficeServiceImpl) DeleteOffice(ctx context.Context, province string, city string, region string, officeName string) error {
	query := "DELETE FROM office WHERE name = ? AND province = ? AND city = ? AND region = ?"
	_, err := t.db.Exec(ctx, query, officeName, province, city, region)
	if err != nil {
		return err
	}
	return nil
}

func (t *TicketOfficeServiceImpl) UpdateOffice(ctx context.Context, oldName string, province string, region string, city string, office Office) error {
	query := "UPDATE office SET name = ?, address = ?, workTime = ?, windowNum = ? WHERE name = ? AND province = ? AND city = ? AND region = ?"
	_, err := t.db.Exec(ctx, query, office.OfficeName, office.Address, office.WorkTime, office.WindowNum, oldName, province, city, region)
	if err != nil {
		return err
	}
	return nil
}

func GetRegionList(ctx context.Context) ([]Region, error) {
	var regions []Region
	configFile, err := os.Open("region.json")
	if err != nil {
		return regions, nil
	}

	jsonParser := json.NewDecoder(configFile)
	err = jsonParser.Decode(&regions)
	return regions, err
}
