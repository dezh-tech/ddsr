package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"path"
	"slices"
	"sync"

	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var management *Management = &Management{
	Mutex: *new(sync.Mutex),
}

type Management struct {
	AllowedPubkeys []string `json:"allowed_keys"`

	sync.Mutex
}

func AllowPubkey(_ context.Context, pubkey, _ string) error {
	management.Lock()
	defer management.Unlock()

	if slices.Contains(management.AllowedPubkeys, pubkey) {
		return fmt.Errorf("pubkey %s is already allowed", pubkey)
	}

	management.AllowedPubkeys = append(management.AllowedPubkeys, pubkey)

	UpdateManagement()

	return nil
}

func ListAllowedPubKeys(_ context.Context) ([]nip86.PubKeyReason, error) {
	management.Lock()
	defer management.Unlock()

	// todo:: support reason.
	res := []nip86.PubKeyReason{}
	for _, p := range management.AllowedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: p,
		})
	}

	return res, nil
}

func BanPubkey(_ context.Context, pubkey, _ string) error {
	management.Lock()
	defer management.Unlock()

	if !slices.Contains(management.AllowedPubkeys, pubkey) {
		return fmt.Errorf("pubkey %s is already not allowed", pubkey)
	}

	for _, q := range relay.QueryEvents {
		ech, err := q(context.Background(), nostr.Filter{
			Authors: []string{pubkey},
		})
		if err != nil {
			return err
		}

		for evt := range ech {
			for _, dl := range relay.DeleteEvent {
				if err := dl(context.Background(), evt); err != nil {
					return err
				}
			}
		}
	}

	management.AllowedPubkeys = slices.DeleteFunc(management.AllowedPubkeys, func(key string) bool {
		return key == pubkey
	})

	UpdateManagement()

	return nil
}

func LoadManagement() {
	if !PathExists(path.Join(config.WorkingDirectory, "/management.json")) {
		data, err := json.Marshal(new(Management))
		if err != nil {
			log.Fatalf("can't make management.json: %s", err.Error())
		}

		if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
			log.Fatalf("can't make management.json: %s", err.Error())
		}
	}

	data, err := ReadFile(path.Join(config.WorkingDirectory, "/management.json"))
	if err != nil {
		log.Fatalf("can't read management.json: %s", err.Error())
	}

	if err := json.Unmarshal(data, management); err != nil {
		log.Fatalf("can't read management.json: %s", err.Error())
	}
}

func UpdateManagement() {
	data, err := json.Marshal(management)
	if err != nil {
		log.Fatalf("can't update management.json: %s", err.Error())
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
		log.Fatalf("can't update management.json: %s", err.Error())
	}
}
