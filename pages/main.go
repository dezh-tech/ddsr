package main

import (
	"context"
	_ "embed"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"slices"
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"
	"github.com/nbd-wtf/go-nostr"
	"github.com/nbd-wtf/go-nostr/nip86"
)

var (
	relay  *khatru.Relay
	config Config

	//go:embed static/index.html
	landingTempl []byte

	kindsToDiscoverAndAccept = []int{0, 3, 5, 10002, 1984, 10063, 10000,
		10001, 10002, 10003, 10004, 10005, 10006, 10007, 10009, 10012,
		10015, 10020, 10030, 10050, 10086, 10087, 10088, 10089, 30315}
)

func main() {
	log.SetPrefix("Pages ")
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
	relay.Info.AddSupportedNIPs([]int{50, 45})

	persistStore := &badger.BadgerBackend{
		Path: path.Join(config.WorkingDirectory, "database"),
	}
	persistStore.Init()

	searchDB := &bluge.BlugeBackend{
		Path:          path.Join(config.WorkingDirectory, "search_database"),
		RawEventStore: persistStore,
	}
	searchDB.Init()

	relay.StoreEvent = append(relay.StoreEvent, persistStore.SaveEvent, searchDB.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, persistStore.QueryEvents, searchDB.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, persistStore.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, persistStore.DeleteEvent, searchDB.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, persistStore.ReplaceEvent, searchDB.ReplaceEvent)
	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		management.Lock()
		defer management.Unlock()

		if !slices.Contains(kindsToDiscoverAndAccept, event.Kind) {
			return true, fmt.Sprintf("blocked: we only accept kinds: %v", kindsToDiscoverAndAccept)
		}

		_, bannedPubkey := management.BannedPubkeys[event.PubKey]
		if bannedPubkey {
			return true, "blocked: you are blacklisted"
		}

		_, bannedEvent := management.BannedEvents[event.ID]
		if bannedEvent {
			return true, "blocked: event is blocked"
		}

		rjct, message := policies.ValidateKind(ctx, event)
		if rjct {
			return rjct, message
		}

		return false, ""
	})

	relay.StoreEvent = append(relay.StoreEvent, func(ctx context.Context, event *nostr.Event) error {
		management.Lock()
		defer management.Unlock()

		if event.Kind == nostr.KindReporting {
			for _, t := range event.Tags {
				if len(t) < 2 {
					continue
				}

				if t[0] == "e" && t[1] != "" {
					if len(t[1]) == 64 {
						management.ModerationEvents[t[1]] = event.Content
					}
				}

				if slices.Contains(config.Moderators, event.PubKey) {
					for _, d := range relay.DeleteEvent {
						if err := d(ctx, event); err != nil {
							return err
						}
					}
				}
			}
		}

		UpdateManagement()

		return nil
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

	go runDiscovery()

	log.Println("Relay running on port: ", config.RelayPort)
	http.ListenAndServe(config.RelayPort, relay)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	log.Print("Received signal: Initiating graceful shutdown", "signal", sig.String())
	persistStore.Close()
	searchDB.Close()
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
