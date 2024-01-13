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
	UserId      int64  `json:"userId"`
	Username    string `json:"username"`
	Role        string `json:"role"`
	GeneratedAt int64  `json:"generatedAt"`
	ExpiresAt   int64  `json:"expiresAt"`
	jwt.RegisteredClaims
}

type TokenDetail struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    int64
}

func CreateJwtToken(id int64, username string, role string) (*TokenDetail, error) {
	myConfigs := configs.GetConfigs()
	accessExpire := jwt.NewNumericDate(time.Now().Add(time.Duration(myConfigs.AccessTokenExpireHour) * time.Hour))
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, MyJwtClaims{
		UserId:      id,
		Username:    username,
		Role:        role,
		GeneratedAt: time.Now().UnixMilli(),
		ExpiresAt:   accessExpire.UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.FormatInt(id, 10),
			ExpiresAt: accessExpire,
			ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		},
	})

	refreshExpire := jwt.NewNumericDate(time.Now().Add(time.Duration(myConfigs.RefreshTokenExpireDay) * 24 * time.Hour))
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, MyJwtClaims{
		UserId:      id,
		Username:    username,
		Role:        role,
		GeneratedAt: time.Now().UnixMilli(),
		ExpiresAt:   refreshExpire.UnixMilli(),
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    strconv.FormatInt(id, 10),
			ExpiresAt: refreshExpire,
			ID:        strconv.FormatInt(time.Now().UnixNano(), 10),
		},
	})

	accToken, err := token.SignedString([]byte(myConfigs.AccessTokenSecret))
	if err != nil {
		return nil, err
	}
	refToken, err := refreshToken.SignedString([]byte(myConfigs.RefreshTokenSecret))
	if err != nil {
		return nil, err
	}
	return &TokenDetail{AccessToken: accToken, ExpiresAt: accessExpire.UnixMilli(), RefreshToken: refToken}, nil
}

func TokenValid(c *fiber.Ctx) error {
	tokenString := ExtractToken(c)
	token, err := VerifyToken(tokenString)
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return err
	}

	return nil
}

func ExtractTokenMetadata(c *fiber.Ctx) (*MyJwtClaims, error) {
	tokenString := ExtractToken(c)
	token, err := VerifyToken(tokenString)
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

		uid, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil, err
		}

		return &MyJwtClaims{
			UserId:   uid,
			Username: username,
		}, nil
	}

	return nil, err
}

func VerifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signature method")
		}
		return []byte(configs.GetConfigs().AccessTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

func VerifyRefreshToken(tokenString string) (*jwt.Token, *MyJwtClaims, error) {
	claims := MyJwtClaims{}
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("wrong signature method")
		}
		return []byte(configs.GetConfigs().RefreshTokenSecret), nil
	})

	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	return token, &claims, nil
}

func ExtractToken(c *fiber.Ctx) string {
	token := c.Get("Authorization", "")
	strArr := strings.Split(token, " ")
	if len(strArr) == 2 {
		return strArr[1]
	} else if len(strArr) == 1 && len(token) > 30 {
		return token
	}
	return ""
}
