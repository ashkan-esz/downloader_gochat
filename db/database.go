package db

import (
	"database/sql"
	"downloader_gochat/configs"

	_ "github.com/lib/pq"
)

//todo : handle errors
//todo : check sqlc, Bun

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	db, err := sql.Open("postgres", configs.DbUrl)
	if err != nil {
		return nil, err
	}
	return &Database{db: db}, nil
}

func (d *Database) Close() {
	d.db.Close()
}

func (d *Database) GetDB() *sql.DB {
	return d.db
}
