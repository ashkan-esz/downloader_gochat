package main

import (
	"context"
	"downloader_gochat/api"
	"downloader_gochat/configs"
	"downloader_gochat/db"
	"downloader_gochat/db/mongodb"
	"downloader_gochat/db/redis"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/repository"
	"downloader_gochat/internal/service"
	"downloader_gochat/pkg/geoip"
	"downloader_gochat/rabbitmq"
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

	go redis.ConnectRedis()
	go geoip.Load()

	dbConn, err := db.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize database connection: %s", err)
	}
	dbConn.AutoMigrate()

	mongoDB, err := mongodb.NewDatabase()
	if err != nil {
		log.Fatalf("could not initialize mongodb database connection: %s", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	rabbit := rabbitmq.Start(ctx)
	defer cancel()

	userRep := repository.NewUserRepository(dbConn.GetDB(), mongoDB.GetDB())
	userSvc := service.NewUserService(userRep, rabbit)
	userHandler := handler.NewUserHandler(userSvc)

	wsRep := repository.NewWsRepository(dbConn.GetDB(), mongoDB.GetDB())
	wsSvc := service.NewWsService(wsRep, userRep, rabbit)
	wsHandler := handler.NewWsHandler(wsSvc)

	notifRep := repository.NewNotificationRepository(dbConn.GetDB(), mongoDB.GetDB())
	notifSvc := service.NewNotificationService(notifRep, userRep, rabbit)
	notifHandler := handler.NewNotificationHandler(notifSvc)

	api.InitRouter(userHandler, wsHandler, notifHandler)
	api.Start("0.0.0.0:8080")
}
