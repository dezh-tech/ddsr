package main

import (
	"log"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	RelayName        string `envconfig:"RELAY_NAME"`
	RelayPubkey      string `envconfig:"RELAY_PUBKEY"`
	RelayDescription string `envconfig:"RELAY_DESCRIPTION"`
	RelayURL         string `envconfig:"RELAY_URL"`
	RelayContact     string `envconfig:"RELAY_CONTACT"`
	RelayIcon        string `envconfig:"RELAY_ICON"`
	RelayBanner      string `envconfig:"RELAY_BANNER"`

	WorkingDirectory string `envconfig:"WORKING_DIR"`

	RelayPort string `envconfig:"RELAY_PORT"`

	Admins              []string `envconfig:"ADMIN_PUBKEYS"`
	Moderators          []string `envconfig:"MODERATOR_PUBKEYS"`
	ReadFallbackRelays  []string `envconfig:"READ_FALLBACK_RELAYS"`
	WriteFallbackRelays []string `envconfig:"WRITE_FALLBACK_RELAYS"`
	PubkeyWhiteListed   bool     `envconfig:"PUBKEY_WHITE_LISTED"`
}

func LoadConfig() {
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("failed to read from env: %s", err)
		return
	}
}
