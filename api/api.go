package api

import (
	"context"
	"downloader_gochat/api/middleware"
	_ "downloader_gochat/docs"
	"downloader_gochat/internal/handler"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

var router *fiber.App

func InitRouter(userHandler *handler.UserHandler, wsHandler *handler.WsHandler, notifHandler *handler.NotificationHandler, mediaHandler *handler.MediaHandler) {
	router = fiber.New(fiber.Config{
		BodyLimit: 100 * 1024 * 1024,
	})

	router.Use(helmet.New())
	router.Use(cors.New())
	router.Use(timeoutMiddleware(time.Second * 2))
	router.Use(recover.New())
	// router.Use(logger.New())
	router.Use(compress.New())

	userRoutes := router.Group("v1/user")
	{
		userRoutes.Post("/signup", middleware.CORSMiddleware, userHandler.RegisterUser)
		userRoutes.Post("/login", middleware.CORSMiddleware, userHandler.Login)
		userRoutes.Put("/getToken", middleware.CORSMiddleware, middleware.IsAuthRefreshToken, userHandler.GetToken)
		userRoutes.Put("/logout", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.LogOut)
		userRoutes.Put("/setNotifToken/:notifToken", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.SetNotifToken)
		userRoutes.Post("/follow/:followId", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.FollowUser)
		userRoutes.Delete("/unfollow/:followId", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.UnFollowUser)
		userRoutes.Get("/followers/:userId/:skip/:limit", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetUserFollowers)
		userRoutes.Get("/followings/:userId/:skip/:limit", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetUserFollowings)
		userRoutes.Get("/userSettings/:settingName", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetUserSettings)
		userRoutes.Put("/updateUserSettings/:settingName", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.UpdateUserSettings)
		userRoutes.Put("/updateFavoriteGenres/:genres", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.UpdateUserFavoriteGenres)
		userRoutes.Get("/activeSessions", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetActiveSessions)
		userRoutes.Get("/profile", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetUserProfile)
		userRoutes.Get("/notifications/:skip/:limit", middleware.CORSMiddleware, middleware.AuthMiddleware, notifHandler.GetUserNotifications)
		userRoutes.Put("/notifications/batchUpdateStatus/:id/:entityTypeId/:status", middleware.CORSMiddleware, middleware.AuthMiddleware, notifHandler.BatchUpdateUserNotificationStatus)
		userRoutes.Post("/media/upload", middleware.CORSMiddleware, middleware.AuthMiddleware, mediaHandler.UploadFile)
	}

	//todo :
	//router.Get("/ws/addClient", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.AddClient)
	//router.Get("/ws/singleChat/messages", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetSingleChatMessages)
	//router.Get("/ws/singleChat/list", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetSingleChatList)

	router.Get("/ws/addClient", middleware.CORSMiddleware, wsHandler.AddClient)
	router.Get("/ws/singleChat/messages", middleware.CORSMiddleware, wsHandler.GetSingleChatMessages)
	router.Get("/ws/singleChat/list", middleware.CORSMiddleware, wsHandler.GetSingleChatList)

	router.Get("/", HealthCheck)
	router.Get("/metrics", monitor.New())

	router.Get("/swagger/*", swagger.HandlerDefault) // default
}

func Start(addr string) error {
	return router.Listen(addr)
}

func timeoutMiddleware(timeout time.Duration) func(c *fiber.Ctx) error {
	return func(c *fiber.Ctx) error {

		// wrap the request context with a timeout
		ctx, cancel := context.WithTimeout(c.Context(), timeout)

		defer func() {
			// check if context timeout was reached
			if errors.Is(ctx.Err(), context.DeadlineExceeded) {

				// write response and abort the request
				//c.Writer.WriteHeader(fiber.StatusGatewayTimeout)
				c.SendStatus(fiber.StatusGatewayTimeout)
				//c.Abort()
			}

			//cancel to clear resources after finished
			cancel()
		}()

		// replace request with context wrapped request
		//c.Request = c.Request.WithContext(ctx)
		return c.Next()
	}
}

// HealthCheck godoc
//
//	@Summary		Show the status of server.
//	@Description	get the status of server.
//	@Tags			System
//	@Success		200	{object}	map[string]interface{}
//	@Router			/ [get]
func HealthCheck(c *fiber.Ctx) error {
	res := map[string]interface{}{
		"data": "Server is up and running",
	}

	if err := c.JSON(res); err != nil {
		return err
	}

	return nil
}
