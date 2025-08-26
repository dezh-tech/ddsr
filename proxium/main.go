package main

import (
	"context"
	_ "embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"

	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var (
	relay  *khatru.Relay
	config Config

	//go:embed static/index.html
	landingTempl []byte
)

func main() {
	log.SetPrefix("Proxium ")
	log.Printf("Running %s\n", StringVersion())

	relay = khatru.NewRelay()

	LoadConfig()

	relay.Info.Name = config.RelayName
	relay.Info.Software = "https://github.com/dezh-tech/ddsr"
	relay.Info.Version = StringVersion()
	relay.Info.PubKey = config.RelayPubkey
	relay.Info.Description = config.RelayDescription
	relay.Info.Icon = config.RelayIcon
	relay.Info.Contact = config.RelayContact
	relay.Info.URL = config.RelayURL
	relay.Info.Banner = config.RelayBanner
	relay.Info.AddSupportedNIPs([]int{50})

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		management.Lock()
		defer management.Unlock()

		if config.PubkeyWhiteListed {
			_, allowedPubkey := management.BannedPubkeys[event.PubKey]
			if !allowedPubkey {
				return true, "blocked: you are not whitelisted"
			}
		}

		_, bannedPubkey := management.BannedPubkeys[event.PubKey]
		if bannedPubkey {
			return true, "blocked: you are blacklisted"
		}

		_, bannedEvent := management.BannedEvents[event.ID]
		if bannedEvent {
			return true, "blocked: event is blocked"
		}

		return false, ""
	})

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		management.Lock()
		defer management.Unlock()

		pubkey := khatru.GetAuthed(ctx)
		if pubkey == "" {
			return true, "auth-required: we can't serve event to users without AUTH"
		}

		if config.PubkeyWhiteListed {
			_, allowedPubkey := management.BannedPubkeys[pubkey]
			if !allowedPubkey {
				return true, "blocked: you are not whitelisted"
			}
		}

		_, bannedPubkey := management.BannedPubkeys[pubkey]
		if bannedPubkey {
			return true, "blocked: you are blacklisted"
		}

		return false, ""
	})

	relay.StoreEvent = append(relay.StoreEvent, func(ctx context.Context, event *nostr.Event) error {
		management.Lock()
		defer management.Unlock()

		go broadcast(event)

		if event.Kind == nostr.KindReporting {
			for _, t := range event.Tags {
				if len(t) < 2 {
					continue
				}

				if t[0] == "e" && t[1] != "" {
					if len(t[1]) == 64 {
						management.ModerationEvents[t[1]] = event.Content
					}

					if slices.Contains(config.Moderators, event.PubKey) {
						management.BannedEvents[t[1]] = event.Content
					}
				}

				if t[0] == "p" && t[1] != "" {
					if slices.Contains(config.Moderators, event.PubKey) {
						management.BannedPubkeys[t[1]] = event.Content
					}
				}
			}

			UpdateManagement()
		}

		return nil
	})

	relay.DeleteEvent = append(relay.DeleteEvent, func(ctx context.Context, event *nostr.Event) error {
		broadcast(event)
		return nil
	})

	relay.QueryEvents = append(relay.QueryEvents, func(ctx context.Context,
		filter nostr.Filter) (chan *nostr.Event, error) {
		ch := make(chan *nostr.Event)

		pubkey := khatru.GetAuthed(ctx)
		if pubkey == "" {
			return nil, errors.New("auth-required: we can't serve event to users without AUTH")
		}

		events := aggregate(&nostr.Filters{filter}, pubkey)

		go func() {
			defer close(ch)

			for _, event := range events {
				ch <- event
			}
		}()

		return ch, nil
	})

	LoadManagement()

	for _, admin := range config.Admins {
		_, isAdmin := management.Admins[admin]
		if isAdmin {
			continue
		}

		management.Admins[admin] = []string{"*"}

		UpdateManagement()
	}

	relay.ManagementAPI.AllowPubKey = AllowPubkey
	relay.ManagementAPI.BanPubKey = BanPubkey
	relay.ManagementAPI.BanEvent = BanEvent
	relay.ManagementAPI.ListBannedEvents = ListBannedEvents
	relay.ManagementAPI.ListEventsNeedingModeration = ListEventsNeedingModeration
	relay.ManagementAPI.BlockIP = BlockIP
	relay.ManagementAPI.UnblockIP = UnblockIP
	relay.ManagementAPI.ListBlockedIPs = ListBlockedIPs
	relay.ManagementAPI.GrantAdmin = GrantAdmin
	relay.ManagementAPI.RevokeAdmin = RevokeAdmin
	relay.ManagementAPI.ListBannedEvents = ListBannedEvents
	relay.ManagementAPI.RejectAPICall = append(relay.ManagementAPI.RejectAPICall,
		func(ctx context.Context, mp nip86.MethodParams) (reject bool, msg string) {
			auth := khatru.GetAuthed(ctx)
			methods, isAdmin := management.Admins[auth]
			if !isAdmin {
				return true, "your are not an admin"
			}

			if !slices.Contains(methods, "*") {
				if !slices.Contains(methods, mp.MethodName()) {
					return true, "you don't have access to this method"
				}
			}

			return false, ""
		})

	mux := relay.Router()
	mux.HandleFunc("GET /{$}", StaticViewHandler)

	log.Println("Relay running on port: ", config.RelayPort)
	http.ListenAndServe(config.RelayPort, relay)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	log.Print("Received signal: Initiating graceful shutdown", "signal", sig.String())
}

func StaticViewHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	t := template.New("webpage")
	t, err := t.Parse(string(landingTempl))
	if err != nil {
		http.Error(w, "Error parsing template", http.StatusInternalServerError)

		return
	}

	err = t.Execute(w, relay.Info)
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)

		return
	}
}
