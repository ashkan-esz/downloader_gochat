package main

import (
	"downloader_gochat/api"
	"downloader_gochat/configs"
	"downloader_gochat/db"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/repository"
	"downloader_gochat/internal/service"
	"downloader_gochat/internal/ws"
	"log"
)

func main() {
	configs.LoadEnvVariables()
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	dbConn.AutoMigrate()

	userRep := repository.NewUserRepository(dbConn.GetDB())
	userSvc := service.NewUserService(userRep)
	userHandler := handler.NewUserHandler(userSvc)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)
	go hub.Run()

	api.InitRouter(userHandler, wsHandler)
	api.Start("0.0.0.0:8080")
}
