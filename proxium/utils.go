package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/nbd-wtf/go-nostr"
)

func getProxiedWriteList(pubkey string) []string {
	relays := make([]*nostr.Relay, 0)
	for _, r := range config.ReadFallbackRelays {
		relay, err := nostr.RelayConnect(context.Background(), r)
		if err != nil {
			log.Printf("can't connect to relay: %s", r)
		}

		relays = append(relays, relay)
	}

	listRelays := make([]string, 0)
	for _, relay := range relays {
		sub, err := relay.Subscribe(context.Background(), nostr.Filters{
			nostr.Filter{
				Kinds:   []int{10090},
				Authors: []string{pubkey},
			},
		})

		if err != nil {
			log.Printf("Can't fetch from relay: %s", relay.URL)
			continue
		}

	eventLoop:
		for {
			select {
			case ev := <-sub.Events:
				isValid, err := ev.CheckSignature()
				if err != nil {
					break eventLoop
				}

				if isValid && ev.Kind == 10090 {
					if len(ev.Tags) == 0 {
						break eventLoop
					}

					for _, t := range ev.Tags {
						if len(t) < 2 {
							continue
						}

						if len(t) == 3 && t[2] == "write" {
							listRelays = append(listRelays, t[1])
						}

						if len(t) == 2 {
							listRelays = append(listRelays, t[1])
						}
					}

					return listRelays
				}
				break eventLoop
			case <-sub.EndOfStoredEvents:
				break eventLoop
			}
		}
	}

	return config.WriteFallbackRelays
}

func getProxiedReadList(pubkey string) []string {
	relays := make([]*nostr.Relay, 0)
	for _, r := range config.ReadFallbackRelays {
		relay, err := nostr.RelayConnect(context.Background(), r)
		if err != nil {
			log.Printf("can't connect to relay: %s", r)
		}

		relays = append(relays, relay)
	}

	listRelays := make([]string, 0)
	for _, relay := range relays {
		sub, err := relay.Subscribe(context.Background(), nostr.Filters{
			nostr.Filter{
				Kinds:   []int{10090},
				Authors: []string{pubkey},
			},
		})

		if err != nil {
			log.Printf("Can't fetch from relay: %s", relay.URL)
			continue
		}

	eventLoop:
		for {
			select {
			case ev := <-sub.Events:
				isValid, err := ev.CheckSignature()
				if err != nil {
					break eventLoop
				}

				if isValid && ev.Kind == 10090 {
					if len(ev.Tags) == 0 {
						break eventLoop
					}

					for _, t := range ev.Tags {
						if len(t) < 2 {
							continue
						}

						if len(t) == 3 && t[2] == "read" {
							listRelays = append(listRelays, t[1])
						}

						if len(t) == 2 {
							listRelays = append(listRelays, t[1])
						}
					}

					return listRelays
				}
				break eventLoop
			case <-sub.EndOfStoredEvents:
				break eventLoop
			}
		}
	}

	return config.ReadFallbackRelays
}

func Mkdir(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("could not create directory %s", path)
	}

	return nil
}

func ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

func WriteFile(filename string, data []byte) error {
	if err := Mkdir(filepath.Dir(filename)); err != nil {
		return err
	}
	if err := os.WriteFile(filename, data, 0o600); err != nil {
		return fmt.Errorf("failed to write to %s: %w", filename, err)
	}

	return nil
}

func PathExists(path string) bool {
	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}

	return err == nil
}
