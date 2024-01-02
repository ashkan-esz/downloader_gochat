package main

import (
	"downloader_gochat/configs"
	"downloader_gochat/db"
	"log"
)

func main() {
	configs.LoadEnvVariables()
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	dbConn.AutoMigrate()
}
