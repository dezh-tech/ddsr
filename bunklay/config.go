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

	RelayPort string `envconfig:"RELAY_PORT"`
}

func LoadConfig() {
	if err := envconfig.Process("", &config); err != nil {
		log.Fatalf("failed to read from env: %s", err)
		return
	}
}
