package configs

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ConfigStruct struct {
	DbUrl                 string
	AccessTokenSecret     string
	RefreshTokenSecret    string
	MigrateOnStart        bool
	DefaultProfileImage   string
	AccessTokenExpireHour int
	RefreshTokenExpireDay int
	ActiveSessionsLimit   int
	RedisUrl              string
	RedisPassword         string
}

var configs = ConfigStruct{}

func GetConfigs() ConfigStruct {
	return configs
}

func LoadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	configs.DbUrl = os.Getenv("DB_URL")
	configs.AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	configs.RefreshTokenSecret = os.Getenv("REFRESH_TOKEN_SECRET")
	configs.MigrateOnStart = os.Getenv("MIGRATE_ON_START") == "true"
	configs.DefaultProfileImage = os.Getenv("DEFAULT_PROFILE_IMAGE")
	configs.RedisUrl = os.Getenv("REDIS_URL")
	configs.RedisPassword = os.Getenv("REDIS_PASSWORD")
	sessionLimit, err := strconv.Atoi(os.Getenv("ACTIVE_SESSIONS_LIMIT"))
	if err != nil || sessionLimit == 0 {
		configs.ActiveSessionsLimit = 5
	} else {
		configs.ActiveSessionsLimit = sessionLimit
	}
	configs.AccessTokenExpireHour = 1
	configs.RefreshTokenExpireDay = 180
}
