package api

import (
	"context"
	"downloader_gochat/api/middleware"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/ws"
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/monitor"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

var router *fiber.App

//todo : add api : check new message
//todo : add api : get messages
//todo : add api : set message as read

func InitRouter(userHandler *handler.UserHandler, wsHandler *ws.Handler) {
	router = fiber.New()

	router.Use(helmet.New())
	router.Use(timeoutMiddleware(time.Second * 2))
	router.Use(recover.New())
	//router.Use(logger.New())
	router.Use(compress.New())

	userRoutes := router.Group("v1/user")
	{
		userRoutes.Post("/signup", middleware.CORSMiddleware, userHandler.RegisterUser)
		userRoutes.Post("/login", middleware.CORSMiddleware, userHandler.Login)
		userRoutes.Get("/logout", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.LogOut)
		userRoutes.Get("/", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetAllUser)
		userRoutes.Get("/:user_id", middleware.CORSMiddleware, middleware.AuthMiddleware, userHandler.GetDetailUser)
	}

	router.Post("/ws/createRoom", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.CreateRoom)
	router.Get("/ws/joinRoom/:roomId", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.JoinRoom)
	router.Get("/ws/getRooms", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetRooms)
	router.Get("/ws/getClients/:roomId", middleware.CORSMiddleware, middleware.AuthMiddleware, wsHandler.GetClients)

	router.Get("/metrics", monitor.New())
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
