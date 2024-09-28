package main

import (
	"context"
	"downloader_gochat/api"
	"downloader_gochat/cloudStorage"
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
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/h2non/bimg"
)

// @title						Go Chat Server
// @version					2.0
// @description				Chat service of the downloader_api project.
// @termsOfService				http://swagger.io/terms/
// @contact.name				API Support
// @contact.url				http://www.swagger.io/support
// @contact.email				support@swagger.io
// @license.name				Apache 2.0
// @license.url				http://www.apache.org/licenses/LICENSE-2.0.html
// @host						chat.movieTracker.site
// @BasePath					/
// @schemes					https
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
// @tokenUrl					chat.movieTracker.site/v1/user/getToken
// @Accept						json
// @Produce					json
func main() {
	configs.LoadEnvVariables()
	bimg.VipsCacheSetMax(0)
	bimg.VipsCacheSetMaxMem(0)

	err := sentry.Init(sentry.ClientOptions{
		Dsn:     configs.GetConfigs().SentryDns,
		Release: configs.GetConfigs().SentryRelease,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: 1,
		EnableTracing:    true,
		Debug:            true,
		AttachStacktrace: true,
		//BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
		//	if hint.Context != nil {
		//		if c, ok := hint.Context.Value(sentry.RequestContextKey).(*fiber.Ctx); ok {
		//			// You have access to the original Context if it panicked
		//			fmt.Println(utils.ImmutableString(c.Hostname()))
		//		}
		//	}
		//	fmt.Println(event)
		//	return event
		//},
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)

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
	go configs.LoadDbConfigs(mongoDB.GetDB())

	ctx, cancel := context.WithCancel(context.Background())
	rabbit := rabbitmq.Start(ctx)
	defer cancel()

	pushNotifSvc := service.NewPushNotificationService()
	telegramMessageSvc := service.NewTelegramMessageService()
	cloudStorageSvc := cloudStorage.StartS3StorageService()

	userRep := repository.NewUserRepository(dbConn.GetDB(), mongoDB.GetDB())
	userSvc := service.NewUserService(userRep, rabbit, cloudStorageSvc)
	userHandler := handler.NewUserHandler(userSvc)

	wsRep := repository.NewWsRepository(dbConn.GetDB(), mongoDB.GetDB())
	wsSvc := service.NewWsService(wsRep, userRep, rabbit)
	wsHandler := handler.NewWsHandler(wsSvc)

	movieRep := repository.NewMovieRepository(dbConn.GetDB(), mongoDB.GetDB())

	notifRep := repository.NewNotificationRepository(dbConn.GetDB(), mongoDB.GetDB())
	notifSvc := service.NewNotificationService(notifRep, userRep, movieRep, rabbit, pushNotifSvc, telegramMessageSvc)
	notifHandler := handler.NewNotificationHandler(notifSvc)

	mediaRep := repository.NewMediaRepository(dbConn.GetDB(), mongoDB.GetDB())
	mediaSvc := service.NewMediaService(mediaRep, userRep, wsRep, rabbit, cloudStorageSvc)
	mediaHandler := handler.NewMediaHandler(mediaSvc)

	castRep := repository.NewCastRepository(dbConn.GetDB(), mongoDB.GetDB())
	_ = service.NewBlurHashService(movieRep, castRep, rabbit)

	api.InitRouter(userHandler, wsHandler, notifHandler, mediaHandler)
	api.Start("0.0.0.0:" + configs.GetConfigs().Port)
}
