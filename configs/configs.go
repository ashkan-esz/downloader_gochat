package configs

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type ConfigStruct struct {
	DbUrl                     string
	AccessTokenSecret         string
	RefreshTokenSecret        string
	MigrateOnStart            bool
	DefaultProfileImage       string
	AccessTokenExpireHour     int
	RefreshTokenExpireDay     int
	ActiveSessionsLimit       int
	WaitForRedisConnectionSec int
	RedisUrl                  string
	RedisPassword             string
	MongodbDatabaseUrl        string
	MongodbDatabaseName       string
	AgendaJobsCollection      string
	MainServerAddress         string
	RabbitMqUrl               string
	CorsAllowedOrigins        []string
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
	configs.MongodbDatabaseUrl = os.Getenv("MONGODB_DATABASE_URL")
	configs.MongodbDatabaseName = os.Getenv("MONGODB_DATABASE_NAME")
	configs.AgendaJobsCollection = os.Getenv("AGENDA_JOBS_COLLECTION")
	configs.MainServerAddress = os.Getenv("MAIN_SERVER_ADDRESS")
	configs.RabbitMqUrl = os.Getenv("RABBITMQ_URL")
	sessionLimit, err := strconv.Atoi(os.Getenv("ACTIVE_SESSIONS_LIMIT"))
	if err != nil || sessionLimit == 0 {
		configs.ActiveSessionsLimit = 5
	} else {
		configs.ActiveSessionsLimit = sessionLimit
	}
	configs.WaitForRedisConnectionSec, err = strconv.Atoi(os.Getenv("WAIT_REDIS_CONNECTION_SEC"))
	configs.CorsAllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), "---")
	for i := range configs.CorsAllowedOrigins {
		configs.CorsAllowedOrigins[i] = strings.TrimSpace(configs.CorsAllowedOrigins[i])
	}
	configs.AccessTokenExpireHour = 1
	configs.RefreshTokenExpireDay = 180
}
