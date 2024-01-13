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

func IsAuthRefreshToken(c *fiber.Ctx) error {
	refreshToken := c.Cookies("refreshToken", "")
	if refreshToken == "" {
		refreshToken = c.Get("refreshtoken", "")
		if refreshToken == "" {
			refreshToken = c.Get("refreshToken", "")
		}
	}

	if refreshToken == "" {
		return response.ResponseError(c, "Unauthorized, refreshToken not provided", fiber.StatusUnauthorized)
	}

	token, claims, err := util.VerifyRefreshToken(refreshToken)
	if err != nil {
		return response.ResponseError(c, "Unauthorized, Invalid refreshToken", fiber.StatusUnauthorized)
	}
	if token == nil || claims == nil {
		return response.ResponseError(c, "Unauthorized, Invalid refreshToken metaData", fiber.StatusUnauthorized)
	}

	//todo : check redis blacklist

	c.Locals("refreshToken", refreshToken)
	c.Locals("jwtUserData", claims)
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
