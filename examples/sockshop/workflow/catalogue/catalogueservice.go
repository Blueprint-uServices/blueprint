// Package catalogue implements the SockShop catalogue microservice
package catalogue

import (
	"context"
	"strings"

	"github.com/blueprint-uservices/blueprint/runtime/core/backend"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type (
	// The SockShop CatalogueService stores an inventory of Socks being sold by the shop.
	CatalogueService interface {
		// List socks that match any of the tags specified.  Sort the results in the specified order,
		// then return a subset of the results.
		List(ctx context.Context, tags []string, order string, pageNum, pageSize int) ([]Sock, error)

		// Counts the number of socks that match any of the tags specified.
		Count(ctx context.Context, tags []string) (int, error)

		// Gets details about a [Sock]
		Get(ctx context.Context, id string) (Sock, error)

		// Lists all tags
		Tags(ctx context.Context) ([]string, error)

		// New for Blueprint: adds tags to the database if they do not already exist.
		AddTags(ctx context.Context, tags []string) error

		// New for Blueprint: adds a sock to the database.
		// If sock.ID is "" then an ID is generated; otherwise the provided ID is used.
		// If the sock has tags that aren't yet in the DB, then the tags are added to the DB.
		// If the sock ID already exists in the database, then the sock is updated
		// Returns the ID of the sock
		AddSock(ctx context.Context, sock Sock) (string, error)

		// New for Blueprint: deletes a sock from the database.
		DeleteSock(ctx context.Context, id string) error
	}

	// Sock describes the things on offer in the catalogue.
	Sock struct {
		ID          string   `json:"id" db:"sock_id"`
		Name        string   `json:"name" db:"name"`
		Description string   `json:"description" db:"description"`
		ImageURL    []string `json:"imageUrl" db:"-"`
		ImageURL_1  string   `json:"-" db:"image_url_1"`
		ImageURL_2  string   `json:"-" db:"image_url_2"`
		Price       float32  `json:"price" db:"price"`
		Quantity    int      `json:"quantity" db:"quantity"`
		Tags        []string `json:"tag" db:"-"`
		TagString   string   `json:"-" db:"tag_name"`
	}

	tag struct {
		ID   int    `db:"tag_id"`
		Name string `db:"name"`
	}
)

// ErrNotFound is returned when there is no sock for a given ID.
var ErrNotFound = errors.New("not found")

// ErrDBConnection is returned when connection with the database fails.
var ErrDBConnection = errors.New("database connection error")

var baseQuery = `SELECT sock.sock_id, 
						sock.name, 
						sock.description, 
						sock.price, 
						sock.quantity, 
						sock.image_url_1, 
						sock.image_url_2, 
						GROUP_CONCAT(DISTINCT alltags.name) AS tag_name 
				FROM sock 
				JOIN sock_tag allsocktags ON sock.sock_id=allsocktags.sock_id 
				JOIN tag alltags ON allsocktags.tag_id=alltags.tag_id
				JOIN sock_tag ON sock.sock_id=sock_tag.sock_id
				JOIN tag ON tag.tag_id=sock_tag.tag_id`

// Implementation of [CatalogueService].  Method implementations are pulled directly from the original
// SockShop implementation, which was written in golang.
type catalogueImpl struct {
	db backend.RelationalDB
}

// Creates a [CatalogueService] instance that stores the item catalogue in the provided relational database
func NewCatalogueService(ctx context.Context, db backend.RelationalDB) (CatalogueService, error) {
	c := &catalogueImpl{db: db}
	return c, c.createTables(ctx)
}

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

	query += " GROUP BY sock.sock_id"

	if order != "" {
		query += " ORDER BY ?"
		args = append(args, order)
	}

	query += ";"

	err := s.db.Select(ctx, &socks, query, args...)
	if err != nil {
		return []Sock{}, errors.Wrap(err, "CatalogueService.List")
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
		return 0, err
	}
	defer sel.Close()

	var count int
	err = sel.QueryRow(args...).Scan(&count)

	if err != nil {
		return 0, err
	}

	return count, nil
}

// Get implements CatalogueService.
func (s *catalogueImpl) Get(ctx context.Context, id string) (Sock, error) {
	query := baseQuery + " WHERE sock.sock_id =? GROUP BY sock.sock_id;"

	var sock Sock
	err := s.db.Get(ctx, &sock, query, id)
	if err != nil {
		return Sock{}, errors.Wrapf(err, "CatalogueService.Get %v", id)
	}

	sock.ImageURL = []string{sock.ImageURL_1, sock.ImageURL_2}
	sock.Tags = strings.Split(sock.TagString, ",")

	return sock, nil
}

// Tags implements CatalogueService.
func (s *catalogueImpl) Tags(ctx context.Context) ([]string, error) {
	var tags []string
	err := s.db.Select(ctx, &tags, "SELECT name FROM tag")
	return tags, err
}

// AddTags implements CatalogueService.
func (s *catalogueImpl) AddTags(ctx context.Context, tags []string) error {
	_, err := s.addTags(ctx, tags...)
	return err
}

// AddSock implements CatalogueService.
func (s *catalogueImpl) AddSock(ctx context.Context, sock Sock) (string, error) {
	// Delete any existing sock with this ID
	if sock.ID != "" {
		if err := s.DeleteSock(ctx, sock.ID); err != nil {
			return "", err
		}
	} else {
		sock.ID = uuid.NewString()
	}

	// Add the sock
	_, err := s.db.Exec(ctx, "INSERT INTO sock (sock_id, name, description, price, quantity, image_url_1, image_url_2) VALUES (?, ?, ?, ?, ?, ?, ?);",
		sock.ID, sock.Name, sock.Description, sock.Price, sock.Quantity, sock.ImageURL_1, sock.ImageURL_2)
	if err != nil {
		return "", err
	}

	// Make sure the tags are in the DB
	tagIds, err := s.addTags(ctx, sock.Tags...)
	if err != nil {
		return "", err
	}

	// Add the tags to the sock
	for _, tagId := range tagIds {
		_, err = s.db.Exec(ctx, "INSERT INTO sock_tag (sock_id, tag_id) VALUES (?, ?);", sock.ID, tagId)
		if err != nil {
			return "", err
		}
	}

	return sock.ID, nil
}

// DeleteSock implements CatalogueService.
func (s *catalogueImpl) DeleteSock(ctx context.Context, id string) error {
	if id == "" {
		return nil
	}

	// Delete sock's tags
	_, err := s.db.Exec(ctx, "DELETE FROM sock_tag WHERE sock_tag.sock_id=?;", id)
	if err != nil {
		return err
	}

	// Delete existing sock
	_, err = s.db.Exec(ctx, "DELETE FROM sock WHERE sock.sock_id=?;", id)
	if err != nil {
		return err
	}

	return nil
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

// DB query to add tags
func (s *catalogueImpl) addTags(ctx context.Context, tags ...string) ([]int, error) {
	var currentTags []tag
	if err := s.db.Select(ctx, &currentTags, "SELECT * FROM tag;"); err != nil {
		return nil, err
	}

	tagLookup := make(map[string]int)
	for _, tag := range currentTags {
		tagLookup[tag.Name] = tag.ID
	}

	tagIds := []int{}
	for _, tagName := range tags {
		if _, tagAlreadyExists := tagLookup[tagName]; !tagAlreadyExists {
			// Insert the tag
			res, err := s.db.Exec(ctx, "INSERT INTO tag (name) VALUES (?);", tagName)
			if err != nil {
				return nil, err
			}
			id, err := res.LastInsertId()
			if err != nil {
				return nil, err
			}
			tagLookup[tagName] = int(id)
		}

		tagIds = append(tagIds, tagLookup[tagName])
	}

	return tagIds, nil
}

// Creates database tables if they don't already exist
func (c *catalogueImpl) createTables(ctx context.Context) (err error) {
	if _, err = c.db.Exec(ctx, createSockTable); err != nil {
		return errors.Wrap(err, "unable to create sock table")
	}
	if _, err = c.db.Exec(ctx, createTagTable); err != nil {
		if _, err = c.db.Exec(ctx, createTagTable2); err != nil {
			return errors.Wrap(err, "unable to create Tag table")
		}
	}
	if _, err = c.db.Exec(ctx, createSockTagTable); err != nil {
		return errors.Wrap(err, "unable to create socktag table")
	}
	return nil
}

var createSockTable = `CREATE TABLE IF NOT EXISTS sock (
	sock_id varchar(40) NOT NULL, 
	name varchar(20), 
	description varchar(200), 
	price float, 
	quantity int, 
	image_url_1 varchar(40), 
	image_url_2 varchar(40), 
	PRIMARY KEY(sock_id)
);`

// AUTOINCREMENT should be AUTO_INCREMENT in mysql
var createTagTable = `CREATE TABLE IF NOT EXISTS tag (
	tag_id INTEGER PRIMARY KEY AUTO_INCREMENT, 
	name varchar(20)
);`

var createTagTable2 = `CREATE TABLE IF NOT EXISTS tag (
	tag_id INTEGER PRIMARY KEY AUTOINCREMENT, 
	name varchar(20)
);`

var createSockTagTable = `CREATE TABLE IF NOT EXISTS sock_tag (
	sock_id varchar(40), 
	tag_id INTEGER, 
	FOREIGN KEY (sock_id) 
		REFERENCES sock(sock_id), 
	FOREIGN KEY(tag_id)
		REFERENCES tag(tag_id)
);`
