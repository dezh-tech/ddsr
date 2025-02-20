package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	RelayName        string
	RelayPubkey      string
	RelayDescription string
	RelayURL         string
	RelayContact     string
	RelayIcon        string
	RelayBanner      string

	WorkingDirectory string

	RelayPort   string
	BlossomPort string
}

func LoadConfig() {
	if err := godotenv.Load(".env"); err != nil {
		log.Fatalf("Can't load .env: %s\n", err.Error())
	}

	config = Config{
		RelayName:        getEnv("RELAY_NAME"),
		RelayPubkey:      getEnv("RELAY_PUBKEY"),
		RelayDescription: getEnv("RELAY_DESCRIPTION"),
		RelayURL:         getEnv("RELAY_URL"),
		RelayContact:     getEnv("RELAY_CONTACT"),
		RelayIcon:        getEnv("RELAY_ICON"),
		RelayBanner:      getEnv("RELAY_BANNER"),
		WorkingDirectory: getEnv("WORKING_DIR"),

		RelayPort:   getEnv("RELAY_PORT"),
		BlossomPort: getEnv("BLOSSOM_PORT"),
	}
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
