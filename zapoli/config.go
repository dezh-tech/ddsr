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

	DBPath          string
	BlobStoragePath string

	RelayPort   string
	BlossomPort string
}

func LoadConfig() Config {
	godotenv.Load(".env")

	if os.Getenv("RELAY_ICON") == "" {
		os.Setenv("RELAY_ICON", "https://file.nostrmedia.com/f/bd4ae3e67e29964d494172261dc45395c89f6bd2e774642e366127171dfb81f5/634c632c4672e4d7e3cc94f0789f440842d1f435d92277aa2e233cfd0055ffa7.png")
	}

	if os.Getenv("RELAY_BANNER") == "" {
		os.Setenv("RELAY_BANNER", "https://file.nostrmedia.com/f/bd4ae3e67e29964d494172261dc45395c89f6bd2e774642e366127171dfb81f5/0255481dd1eaa1632928a9258e1faa727eeaf45f3b6d8b8ec0a959e13d092298.png")
	}

	if os.Getenv("RELAY_CONTACT") == "" {
		os.Setenv("RELAY_CONTACT", getEnv("RELAY_PUBKEY"))
	}

	config := Config{
		RelayName:        getEnv("RELAY_NAME"),
		RelayPubkey:      getEnv("RELAY_PUBKEY"),
		RelayDescription: getEnv("RELAY_DESCRIPTION"),
		RelayURL:         getEnv("RELAY_URL"),
		RelayContact:     getEnv("RELAY_CONTACT"),
		RelayIcon:        getEnv("RELAY_ICON"),
		RelayBanner:      getEnv("RELAY_BANNER"),

		DBPath:          getEnv("DB_PATH"),
		BlobStoragePath: getEnv("BLOB_STORAGE_PATH"),

		RelayPort:   getEnv("RELAY_PORT"),
		BlossomPort: getEnv("BLOSSOM_PORT"),
	}

	return config
}

func getEnv(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Fatalf("Environment variable %s not set", key)
	}
	return value
}
