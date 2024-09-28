package api

import (
	"context"
	"downloader_gochat/api/middleware"
	"downloader_gochat/configs"
	_ "downloader_gochat/docs"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/repository"
	"downloader_gochat/pkg/response"
	"errors"
	"slices"
	"time"

	"github.com/gofiber/contrib/fibersentry"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/swagger"
)

var router *fiber.App

type Handlers struct {
	UserHandler  *handler.UserHandler
	WsHandler    *handler.WsHandler
	NotifHandler *handler.NotificationHandler
	MediaHandler *handler.MediaHandler
	AdminHandler *handler.AdminHandler
	UserRepo     *repository.UserRepository
}

func InitRouter(handlers *Handlers) {
	var defaultErrorHandler = func(c *fiber.Ctx, err error) error {
		// Status code defaults to 500
		code := fiber.StatusInternalServerError

		// Retrieve the custom status code if it's a *fiber.Error
		var e *fiber.Error
		if errors.As(err, &e) {
			code = e.Code
		}

		// Set Content-Type: text/plain; charset=utf-8
		c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

		// Return status code with error message
		//return c.Status(code).SendString(err.Error())
		//fmt.Println(err.Error())
		return response.ResponseError(c, "Internal Error", code)
	}

	router = fiber.New(fiber.Config{
		BodyLimit:    100 * 1024 * 1024,
		ErrorHandler: defaultErrorHandler,
	})

	router.Use(helmet.New())
	router.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			return middleware.LocalhostRegex.MatchString(origin) ||
				slices.Index(configs.GetConfigs().CorsAllowedOrigins, origin) != -1
		},
		AllowCredentials: true,
	}))
	router.Use(timeoutMiddleware(time.Second * 2))
	router.Use(recover.New())
	// router.Use(logger.New())
	router.Use(compress.New())

	limiterMiddleware := limiter.New(limiter.Config{
		Max:        6,
		Expiration: 60 * time.Second,
		LimitReached: func(c *fiber.Ctx) error {
			return response.ResponseError(c, "Wait for 1 min before retry", fiber.StatusTooManyRequests)
		},
	})

	router.Use(fibersentry.New(fibersentry.Config{
		Repanic:         true,
		WaitForDelivery: false,
	}))

	userRoutes := router.Group("v1/user")
	{
		userRoutes.Post("/signup", handlers.UserHandler.RegisterUser)
		userRoutes.Post("/login", handlers.UserHandler.Login)
		userRoutes.Put("/getToken", middleware.IsAuthRefreshToken, handlers.UserHandler.GetToken)
		userRoutes.Put("/logout", middleware.AuthMiddleware, handlers.UserHandler.LogOut)
		userRoutes.Put("/setNotifToken/:notifToken", middleware.AuthMiddleware, handlers.UserHandler.SetNotifToken)
		userRoutes.Post("/follow/:followId", middleware.AuthMiddleware, handlers.UserHandler.FollowUser)
		userRoutes.Delete("/unfollow/:followId", middleware.AuthMiddleware, handlers.UserHandler.UnFollowUser)
		userRoutes.Get("/followers/:userId/:skip/:limit", middleware.AuthMiddleware, handlers.UserHandler.GetUserFollowers)
		userRoutes.Get("/followings/:userId/:skip/:limit", middleware.AuthMiddleware, handlers.UserHandler.GetUserFollowings)
		userRoutes.Get("/userSettings/:settingName", middleware.AuthMiddleware, handlers.UserHandler.GetUserSettings)
		userRoutes.Put("/updateUserSettings/:settingName", middleware.AuthMiddleware, handlers.UserHandler.UpdateUserSettings)
		userRoutes.Put("/updateFavoriteGenres/:genres", middleware.AuthMiddleware, handlers.UserHandler.UpdateUserFavoriteGenres)
		userRoutes.Get("/activeSessions", middleware.AuthMiddleware, handlers.UserHandler.GetActiveSessions)
		userRoutes.Get("/profile", middleware.AuthMiddleware, handlers.UserHandler.GetUserProfile)
		userRoutes.Get("/roles_and_permissions", middleware.AuthMiddleware, handlers.UserHandler.GetUserRolePermission)
		userRoutes.Post("/editProfile", middleware.AuthMiddleware, handlers.UserHandler.EditUserProfile)
		userRoutes.Put("/updatePassword", middleware.AuthMiddleware, handlers.UserHandler.UpdateUserPassword)
		userRoutes.Get("/sendVerifyEmail", limiterMiddleware, middleware.AuthMiddleware, handlers.UserHandler.SendVerifyEmail)
		userRoutes.Get("/verifyEmail/:userId/:token", limiterMiddleware, handlers.UserHandler.VerifyEmail)
		userRoutes.Delete("/deleteAccount", limiterMiddleware, middleware.AuthMiddleware, handlers.UserHandler.SendDeleteAccount)
		userRoutes.Get("/deleteAccount/:userId/:token", limiterMiddleware, handlers.UserHandler.DeleteUserAccount)
		userRoutes.Post("/uploadProfileImage", middleware.AuthMiddleware, handlers.UserHandler.UploadProfileImage)
		userRoutes.Delete("/removeProfileImage/:fileName", middleware.AuthMiddleware, handlers.UserHandler.RemoveProfileImage)
		userRoutes.Put("/forceLogout/:deviceId", middleware.AuthMiddleware, handlers.UserHandler.ForceLogoutDevice)
		userRoutes.Put("/forceLogoutAll", middleware.AuthMiddleware, handlers.UserHandler.ForceLogoutAll)
		userRoutes.Get("/notifications/:skip/:limit", middleware.AuthMiddleware, handlers.NotifHandler.GetUserNotifications)
		userRoutes.Put("/notifications/batchUpdateStatus/:id/:entityTypeId/:status", middleware.AuthMiddleware, handlers.NotifHandler.BatchUpdateUserNotificationStatus)
		userRoutes.Post("/media/upload", middleware.AuthMiddleware, handlers.MediaHandler.UploadFile)
	}

	router.Get("/ws/addClient/:deviceId", middleware.AuthMiddleware, handlers.WsHandler.AddClient)
	router.Get("/ws/singleChat/messages", middleware.AuthMiddleware, handlers.WsHandler.GetSingleChatMessages)
	router.Get("/ws/singleChat/list", middleware.AuthMiddleware, handlers.WsHandler.GetSingleChatList)

	adminRoutes := router.Group("v1/admin")
	{
		adminRoutes.Get("/status", middleware.AuthMiddleware, middleware.CheckUserPermission(handlers.UserRepo, "admin_get_server_status"), handlers.AdminHandler.GetServerStatus)
	}

	router.Get("/", middleware.CheckUserPermission(handlers.UserRepo, "admin_get_server_status"), HealthCheck)
	router.Get("/metrics", middleware.AuthMiddleware, middleware.CheckUserPermission(handlers.UserRepo, "admin_get_server_status"), monitor.New())

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
