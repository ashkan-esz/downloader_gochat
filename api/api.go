package api

import (
	"downloader_gochat/api/middleware"
	"downloader_gochat/internal/handler"
	"downloader_gochat/internal/ws"

	"github.com/gin-gonic/gin"
)

var router *gin.Engine

//todo : add api : check new message
//todo : add api : get messages
//todo : add api : set message as read

func InitRouter(userHandler *handler.UserHandler, wsHandler *ws.Handler) {
	router = gin.Default()

	userRoutes := router.Group("v1/user")
	{
		userRoutes.POST("/signup", middleware.CORSMiddleware(), userHandler.RegisterUser)
		userRoutes.POST("/login", middleware.CORSMiddleware(), userHandler.Login)
		userRoutes.GET("/logout", middleware.AuthMiddleware(), middleware.CORSMiddleware(), userHandler.LogOut)
		userRoutes.GET("/", middleware.AuthMiddleware(), middleware.CORSMiddleware(), userHandler.GetAllUser)
		userRoutes.GET("/:user_id", middleware.AuthMiddleware(), middleware.CORSMiddleware(), userHandler.GetDetailUser)
	}

	router.POST("/ws/createRoom", middleware.AuthMiddleware(), middleware.CORSMiddleware(), wsHandler.CreateRoom)
	router.GET("/ws/joinRoom/:roomId", middleware.AuthMiddleware(), middleware.CORSMiddleware(), wsHandler.JoinRoom)
	router.GET("/ws/getRooms", middleware.AuthMiddleware(), middleware.CORSMiddleware(), wsHandler.GetRooms)
	router.GET("/ws/getClients/:roomId", middleware.AuthMiddleware(), middleware.CORSMiddleware(), wsHandler.GetClients)
}

func Start(addr string) error {
	return router.Run(addr)
}
