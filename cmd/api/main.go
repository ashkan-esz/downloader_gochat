package main

import (
	"downloader_gochat/api"
	"downloader_gochat/configs"
	"downloader_gochat/db"
	"downloader_gochat/db/redis"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/repository"
	"downloader_gochat/internal/service"
	"downloader_gochat/internal/ws"
	"downloader_gochat/pkg/geoip"
	"log"
)

// @title						Fiber Swagger Example API
// @version					2.0
// @description				This is a sample server server.
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.url				http://www.swagger.io/support
// @contact.email				support@swagger.io
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @host						localhost:8080
// @BasePath					/
// @schemes					http
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
// @tokenUrl					http://localhost:8080/v1/user/token
// @Accept						json
// @Produce					json
func main() {
	configs.LoadEnvVariables()
	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	dbConn.AutoMigrate()

	go redis.ConnectRedis()
	go geoip.Load()

	userRep := repository.NewUserRepository(dbConn.GetDB())
	userSvc := service.NewUserService(userRep)
	userHandler := handler.NewUserHandler(userSvc)

	hub := ws.NewHub()
	wsHandler := ws.NewHandler(hub)
	go hub.Run()

	api.InitRouter(userHandler, wsHandler)
	api.Start("0.0.0.0:8080")
}
