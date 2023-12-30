package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type ConfigStruct struct {
	DbUrl            string
	SigningSecretKey string
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
	configs.SigningSecretKey = os.Getenv("SIGNING_SECRET_KEY")
}
