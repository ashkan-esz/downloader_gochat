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

func InitRouter(userHandler *handler.UserHandler, wsHandler *handler.WsHandler) {
	router = fiber.New()

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
	}

	router.Get("/ws/addClient", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.AddClient)
	router.Post("/ws/createRoom", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.CreateRoom)
	router.Get("/ws/joinRoom/:roomId", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.JoinRoom)
	router.Get("/ws/getRooms", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetRooms)
	router.Get("/ws/getClients/:roomId", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetClients)

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
//	@Tags			root
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
