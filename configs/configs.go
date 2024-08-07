package configs

import (
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type ConfigStruct struct {
	Port                         string
	DbUrl                        string
	AccessTokenSecret            string
	RefreshTokenSecret           string
	MigrateOnStart               bool
	DefaultProfileImage          string
	AccessTokenExpireHour        int
	RefreshTokenExpireDay        int
	ActiveSessionsLimit          int
	WaitForRedisConnectionSec    int
	BlurHashConsumerCount        int
	RedisUrl                     string
	RedisPassword                string
	MongodbDatabaseUrl           string
	MongodbDatabaseName          string
	AgendaJobsCollection         string
	ServerAddress                string
	MainServerAddress            string
	RabbitMqUrl                  string
	FirebaseAuthKey              string
	CorsAllowedOrigins           []string
	CloudStorageEndpoint         string
	CloudStorageWebsiteEndpoint  string
	CloudStorageAccessKey        string
	CloudStorageSecretAccessKey  string
	CloudStorageBucketNamePrefix string
	SentryDns                    string
	SentryRelease                string
	PrintErrors                  bool
	Domain                       string
	AppDeepLink                  string
}

var configs = ConfigStruct{}

func GetConfigs() ConfigStruct {
	return configs
}

func LoadEnvVariables() {
	if err := godotenv.Load(); err != nil {
		log.Printf("Error loading .env file: %v", err)
	}

	configs.Port = os.Getenv("PORT")
	configs.DbUrl = os.Getenv("POSTGRES_DATABASE_URL")
	configs.AccessTokenSecret = os.Getenv("ACCESS_TOKEN_SECRET")
	configs.RefreshTokenSecret = os.Getenv("REFRESH_TOKEN_SECRET")
	configs.MigrateOnStart = os.Getenv("MIGRATE_ON_START") == "true"
	configs.DefaultProfileImage = os.Getenv("DEFAULT_PROFILE_IMAGE")
	configs.RedisUrl = os.Getenv("REDIS_URL")
	configs.RedisPassword = os.Getenv("REDIS_PASSWORD")
	configs.MongodbDatabaseUrl = os.Getenv("MONGODB_DATABASE_URL")
	configs.MongodbDatabaseName = os.Getenv("MONGODB_DATABASE_NAME")
	configs.AgendaJobsCollection = os.Getenv("AGENDA_JOBS_COLLECTION")
	configs.ServerAddress = os.Getenv("SERVER_ADDRESS")
	configs.MainServerAddress = os.Getenv("MAIN_SERVER_ADDRESS")
	configs.RabbitMqUrl = os.Getenv("RABBITMQ_URL")
	configs.FirebaseAuthKey = os.Getenv("FIREBASE_AUTH_KEY")
	configs.CloudStorageEndpoint = os.Getenv("CLOUAD_STORAGE_ENDPOINT")
	configs.CloudStorageWebsiteEndpoint = os.Getenv("CLOUAD_STORAGE_WEBSITE_ENDPOINT")
	configs.CloudStorageAccessKey = os.Getenv("CLOUAD_STORAGE_ACCESS_KEY")
	configs.CloudStorageSecretAccessKey = os.Getenv("CLOUAD_STORAGE_SECRET_ACCESS_KEY")
	configs.CloudStorageBucketNamePrefix = os.Getenv("BUCKET_NAME_PREFIX")
	sessionLimit, err := strconv.Atoi(os.Getenv("ACTIVE_SESSIONS_LIMIT"))
	if err != nil || sessionLimit == 0 {
		configs.ActiveSessionsLimit = 5
	} else {
		configs.ActiveSessionsLimit = sessionLimit
	}
	blurHashConsumerCount, err := strconv.Atoi(os.Getenv("BLURHASH_CONSUMER_COUNT"))
	if err != nil || sessionLimit == 0 {
		configs.BlurHashConsumerCount = 1
	} else {
		configs.BlurHashConsumerCount = blurHashConsumerCount
	}
	configs.WaitForRedisConnectionSec, err = strconv.Atoi(os.Getenv("WAIT_REDIS_CONNECTION_SEC"))
	configs.CorsAllowedOrigins = strings.Split(os.Getenv("CORS_ALLOWED_ORIGINS"), "---")
	for i := range configs.CorsAllowedOrigins {
		configs.CorsAllowedOrigins[i] = strings.TrimSpace(configs.CorsAllowedOrigins[i])
	}
	configs.AccessTokenExpireHour, err = strconv.Atoi(os.Getenv("ACCESS_TOKEN_EXPIRE_HOUR"))
	configs.RefreshTokenExpireDay, err = strconv.Atoi(os.Getenv("REFRESH_TOKEN_EXPIRE_DAY"))
	configs.SentryDns = os.Getenv("SENTRY_DNS")
	configs.SentryRelease = os.Getenv("SENTRY_RELEASE")
	configs.PrintErrors = os.Getenv("PRINT_ERRORS") == "true"
	configs.Domain = os.Getenv("DOMAIN")
	configs.AppDeepLink = os.Getenv("APP_DEEP_LINK")
}
