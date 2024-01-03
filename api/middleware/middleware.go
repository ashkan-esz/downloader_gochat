package middleware

import (
	"downloader_gochat/pkg/response"
	"downloader_gochat/util"

	"github.com/gofiber/fiber/v2"
)

func AuthMiddleware(c *fiber.Ctx) error {
	err := util.TokenValid(c)
	if err != nil {
		return response.ResponseError(c, err.Error(), fiber.StatusUnauthorized)
		//c.Abort()
		//return
	}

	return c.Next()
}

func CORSMiddleware(c *fiber.Ctx) error {
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Credentials", "true")
	c.Set("Access-Control-Allow-Headers", "Content-Type,Content-Length")
	c.Set("Access-Control-Allow-Method", "POST, GET, DELETE, PUT")

	if c.Method() == "OPTIONS" {
		//c.AbortWithStatus(204)
		return c.SendStatus(204)
		//return
	}

	return c.Next()
}
