package db

import (
	"database/sql"
	_ "github.com/lib/pq"
)

//todo : handle errors
//todo : check sqlc, Bun

type Database struct {
	db *sql.DB
}

func NewDatabase() (*Database, error) {
	//todo : use env for db url
	db, err := sql.Open("postgres", "postgres://root:mysecretpassword@localhost:5432/go-chat?sslmode=disable")
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
