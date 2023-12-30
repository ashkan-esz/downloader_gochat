package main

import (
	"downloader_gochat/configs"
	"downloader_gochat/db"
	"log"
)

func main() {
	configs.LoadEnvVariables()
	_, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
}
