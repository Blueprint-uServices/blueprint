// Package catalogue implements the SockShop catalogue microservice
package catalogue

import (
	"context"
	"errors"
	"strings"

	"gitlab.mpi-sws.org/cld/blueprint/runtime/core/backend"
)

type (
	// The SockShop CatalogueService stores an inventory of Socks being sold by the shop.
	CatalogueService interface {
		List(ctx context.Context, tags []string, order string, pageNum, pageSize int) ([]Sock, error)
		Count(ctx context.Context, tags []string) (int, error)
		Get(ctx context.Context, id string) (Sock, error)
		Tags(ctx context.Context) ([]string, error)
	}

	// Sock describes the things on offer in the catalogue.
	Sock struct {
		ID          string   `json:"id" db:"id"`
		Name        string   `json:"name" db:"name"`
		Description string   `json:"description" db:"description"`
		ImageURL    []string `json:"imageUrl" db:"-"`
		ImageURL_1  string   `json:"-" db:"image_url_1"`
		ImageURL_2  string   `json:"-" db:"image_url_2"`
		Price       float32  `json:"price" db:"price"`
		Count       int      `json:"count" db:"count"`
		Tags        []string `json:"tag" db:"-"`
		TagString   string   `json:"-" db:"tag_name"`
	}
)

// Implementation of [CatalogueService]
type catalogueImpl struct {
	db backend.RelationalDB
}

// Creates a [CatalogueService] instance that stores the item catalogue in the provided relational database
func NewCatalogueService(ctx context.Context, db backend.RelationalDB) (CatalogueService, error) {
	c := &catalogueImpl{db: db}
	return c, c.init(ctx)
}

// Creates database tables if they don't already exist
func (c *catalogueImpl) init(ctx context.Context) error {
	_, err := c.db.Exec(ctx, createTables)
	return err
}

// ErrNotFound is returned when there is no sock for a given ID.
var ErrNotFound = errors.New("not found")

// ErrDBConnection is returned when connection with the database fails.
var ErrDBConnection = errors.New("database connection error")

var baseQuery = `SELECT sock.sock_id AS id, 
						sock.name, 
						sock.description, 
						sock.price, 
						sock.count, 
						sock.image_url_1, 
						sock.image_url_2, 
						GROUP_CONCAT(tag.name) AS tag_name 
				FROM sock 
				JOIN sock_tag ON sock.sock_id=sock_tag.sock_id 
				JOIN tag ON sock_tag.tag_id=tag.tag_id`

// List implements CatalogueService.
func (s *catalogueImpl) List(ctx context.Context, tags []string, order string, pageNum int, pageSize int) ([]Sock, error) {
	var socks []Sock
	query := baseQuery

	var args []interface{}

	for i, t := range tags {
		if i == 0 {
			query += " WHERE tag.name=?"
			args = append(args, t)
		} else {
			query += " OR tag.name=?"
			args = append(args, t)
		}
	}

	query += " GROUP BY id"

	if order != "" {
		query += " ORDER BY ?"
		args = append(args, order)
	}

	query += ";"

	err := s.db.Select(ctx, &socks, query, args...)
	if err != nil {
		return []Sock{}, ErrDBConnection
	}
	for i, s := range socks {
		socks[i].ImageURL = []string{s.ImageURL_1, s.ImageURL_2}
		socks[i].Tags = strings.Split(s.TagString, ",")
	}

	socks = cut(socks, pageNum, pageSize)

	return socks, nil
}

// Count implements CatalogueService.
func (s *catalogueImpl) Count(ctx context.Context, tags []string) (int, error) {
	query := "SELECT COUNT(DISTINCT sock.sock_id) FROM sock JOIN sock_tag ON sock.sock_id=sock_tag.sock_id JOIN tag ON sock_tag.tag_id=tag.tag_id"

	var args []interface{}

	for i, t := range tags {
		if i == 0 {
			query += " WHERE tag.name=?"
			args = append(args, t)
		} else {
			query += " OR tag.name=?"
			args = append(args, t)
		}
	}

	query += ";"

	sel, err := s.db.Prepare(ctx, query)

	if err != nil {
		return 0, ErrDBConnection
	}
	defer sel.Close()

	var count int
	err = sel.QueryRowContext(ctx, args...).Scan(&count)

	if err != nil {
		return 0, ErrDBConnection
	}

	return count, nil
}

// Get implements CatalogueService.
func (s *catalogueImpl) Get(ctx context.Context, id string) (Sock, error) {
	query := baseQuery + " WHERE sock.sock_id =? GROUP BY sock.sock_id;"

	var sock Sock
	err := s.db.Get(ctx, &sock, query, id)
	if err != nil {
		return Sock{}, ErrNotFound
	}

	sock.ImageURL = []string{sock.ImageURL_1, sock.ImageURL_2}
	sock.Tags = strings.Split(sock.TagString, ",")

	return sock, nil
}

// Tags implements CatalogueService.
func (s *catalogueImpl) Tags(ctx context.Context) ([]string, error) {
	var tags []string
	query := "SELECT name FROM tag;"
	rows, err := s.db.Query(ctx, query)
	if err != nil {
		return []string{}, ErrDBConnection
	}
	var tag string
	for rows.Next() {
		err = rows.Scan(&tag)
		if err != nil {
			continue
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func cut(socks []Sock, pageNum, pageSize int) []Sock {
	if pageNum == 0 || pageSize == 0 {
		return []Sock{} // pageNum is 1-indexed
	}
	start := (pageNum * pageSize) - pageSize
	if start > len(socks) {
		return []Sock{}
	}
	end := (pageNum * pageSize)
	if end > len(socks) {
		end = len(socks)
	}
	return socks[start:end]
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

var createTables = `CREATE TABLE IF NOT EXISTS sock (
	sock_id varchar(40) NOT NULL, 
	name varchar(20), 
	description varchar(200), 
	price float, 
	count int, 
	image_url_1 varchar(40), 
	image_url_2 varchar(40), 
	PRIMARY KEY(sock_id)
);

CREATE TABLE IF NOT EXISTS tag (
	tag_id MEDIUMINT NOT NULL AUTO_INCREMENT, 
	name varchar(20), 
	PRIMARY KEY(tag_id)
);

CREATE TABLE IF NOT EXISTS sock_tag (
	sock_id varchar(40), 
	tag_id MEDIUMINT NOT NULL, 
	FOREIGN KEY (sock_id) 
		REFERENCES sock(sock_id), 
	FOREIGN KEY(tag_id)
		REFERENCES tag(tag_id)
);`
