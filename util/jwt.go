package util

import (
	"downloader_gochat/configs"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MyJwtClaims struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type TokenDetail struct {
	AccessToken string
	ExpireAt    int64
}

func CreateJwtToken(id int64, username string) (*TokenDetail, error) {
	ExpireAt := jwt.NewNumericDate(time.Now().Add(24 * time.Hour))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyJwtClaims{
		ID:       strconv.FormatInt(id, 10),
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.FormatInt(id, 10),
			ExpiresAt: ExpireAt,
		},
	})

	ss, err := token.SignedString([]byte(configs.GetConfigs().SigningSecretKey))
	if err != nil {
		return nil, err
	}
	return &TokenDetail{AccessToken: ss, ExpireAt: ExpireAt.Unix()}, nil
}

func TokenValid(c *fiber.Ctx) error {
	token, err := VerifyToken(c)
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}

	return nil
}

func ExtractTokenMetadata(c *fiber.Ctx) (*MyJwtClaims, error) {
	token, err := VerifyToken(c)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return nil, err
		}

		id, ok := claims["id"].(string)
		if !ok {
			return nil, err
		}

		return &MyJwtClaims{
			ID:       id,
			Username: username,
		}, nil
	}

	return nil, err
}

func VerifyToken(c *fiber.Ctx) (*jwt.Token, error) {
	tokenString := ExtractToken(c)
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signature method")
		}
		return []byte(configs.GetConfigs().SigningSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func ExtractToken(c *fiber.Ctx) string {
	token := c.Get("Authorization", "")
	strArr := strings.Split(token, " ")
	if len(strArr) == 2 {
		return strArr[1]
	}
	return ""
}
