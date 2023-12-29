package main

import (
	"downloader_gochat/api"
	"downloader_gochat/db"
	"downloader_gochat/internal/user"
	"downloader_gochat/internal/ws"
	"log"
)

func main() {
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}

	userRep := user.NewRepository(dbConn.GetDB())
	userSvc := user.NewService(userRep)
	userHandler := user.NewHandler(userSvc)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)
	go hub.Run()

	api.InitApi(userHandler, wsHandler)
	api.Start("0.0.0.0:8080")
}
