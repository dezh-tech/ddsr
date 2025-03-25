package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
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
	BannedPubkeys    map[string]string   `json:"banned_keys"`
	BlockedIPs       map[string]string   `json:"blocked_ips"`
	BannedEvents     map[string]string   `json:"banned_events"`
	ModerationEvents map[string]string   `json:"moderation_events"`
	Admins           map[string][]string `json:"admins"`

	sync.Mutex
}

func BanPubkey(_ context.Context, pubkey, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBanned := management.BannedPubkeys[pubkey]
	if alreadyBanned {
		return fmt.Errorf("pubkey %s is already banned", pubkey)
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

	management.BannedPubkeys[pubkey] = reason

	UpdateManagement()

	return nil
}

func BlockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBlocked := management.BlockedIPs[ip.String()]
	if alreadyBlocked {
		return fmt.Errorf("ip %s is already blocked", ip.String())
	}

	management.BlockedIPs[ip.String()] = reason

	UpdateManagement()

	return nil
}

func UnblockIP(_ context.Context, ip net.IP, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, blocked := management.BlockedIPs[ip.String()]
	if !blocked {
		return fmt.Errorf("ip %s is not blocked", ip.String())
	}

	delete(management.BlockedIPs, ip.String())

	UpdateManagement()

	return nil
}

func BanEvent(_ context.Context, id string, reason string) error {
	management.Lock()
	defer management.Unlock()

	_, alreadyBanned := management.BannedEvents[id]
	if alreadyBanned {
		return fmt.Errorf("event %s is already banned", id)
	}

	for _, q := range relay.QueryEvents {
		ech, err := q(context.Background(), nostr.Filter{
			IDs: []string{id},
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

	management.BannedEvents[id] = reason

	UpdateManagement()

	return nil
}

func ListBannedEvents(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IDReason{}
	for id, reason := range management.BannedEvents {
		res = append(res, nip86.IDReason{
			ID:     id,
			Reason: reason,
		})
	}

	return res, nil
}

func ListBannedPubKeys(_ context.Context) ([]nip86.PubKeyReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.PubKeyReason{}
	for pubkey, reason := range management.BannedPubkeys {
		res = append(res, nip86.PubKeyReason{
			PubKey: pubkey,
			Reason: reason,
		})
	}

	return res, nil
}

func ListBlockedIPs(ctx context.Context) ([]nip86.IPReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IPReason{}
	for ip, reason := range management.BlockedIPs {
		res = append(res, nip86.IPReason{
			IP:     ip,
			Reason: reason,
		})
	}

	return res, nil
}

func ListEventsNeedingModeration(_ context.Context) ([]nip86.IDReason, error) {
	management.Lock()
	defer management.Unlock()

	res := []nip86.IDReason{}
	for id, reason := range management.ModerationEvents {
		res = append(res, nip86.IDReason{
			ID:     id,
			Reason: reason,
		})
	}

	return res, nil
}

func GrantAdmin(ctx context.Context, pubkey string, methods []string) error {
	management.Lock()
	defer management.Unlock()

	if len(methods) == 0 {
		return errors.New("Methods can't be 0")
	}

	management.Admins[pubkey] = methods

	UpdateManagement()

	return nil
}

func RevokeAdmin(ctx context.Context, pubkey string, methods []string) error {
	management.Lock()
	defer management.Unlock()

	_, isAdmin := management.Admins[pubkey]
	if !isAdmin {
		return fmt.Errorf("Pubkey %s is already not in admins list", pubkey)
	}

	management.Admins[pubkey] = slices.DeleteFunc(management.Admins[pubkey], func(m string) bool {
		return slices.Contains(methods, m)
	})

	allowedMethods, _ := management.Admins[pubkey]
	if len(allowedMethods) == 0 {
		delete(management.Admins, pubkey)
	}

	UpdateManagement()

	return nil
}

func LoadManagement() {
	if !PathExists(path.Join(config.WorkingDirectory, "/management.json")) {
		data, err := json.Marshal(Management{
			BannedPubkeys:    make(map[string]string),
			BlockedIPs:       make(map[string]string),
			BannedEvents:     make(map[string]string),
			ModerationEvents: make(map[string]string),
			Admins:           make(map[string][]string),
		})
		if err != nil {
			log.Fatalf("can't make management.json", "err", err.Error())
		}

		if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
			log.Fatalf("can't make management.json", "err", err.Error())
		}
	}

	data, err := ReadFile(path.Join(config.WorkingDirectory, "/management.json"))
	if err != nil {
		log.Fatalf("can't read management.json", "err", err.Error())
	}

	if err := json.Unmarshal(data, management); err != nil {
		log.Fatalf("can't read management.json", "err", err.Error())
	}
}

func UpdateManagement() {
	data, err := json.Marshal(management)
	if err != nil {
		log.Fatalf("can't update management.json", "err", err.Error())
	}

	if err := WriteFile(path.Join(config.WorkingDirectory, "/management.json"), data); err != nil {
		log.Fatalf("can't update management.json", "err", err.Error())
	}
}
