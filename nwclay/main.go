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
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/policies"
	"github.com/nbd-wtf/go-nostr"
)

var (
	relay  *khatru.Relay
	config Config

	//go:embed static/index.html
	landingTempl []byte
)

func main() {
	log.SetPrefix("NWClay ")
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
	relay.Info.SupportedNIPs = []any{1, 46}

	mainDB := &badger.BadgerBackend{
		Path:     path.Join(config.WorkingDirectory, "database"),
		MaxLimit: 100,
	}
	mainDB.Init()

	relay.RejectCountFilter = append(relay.RejectCountFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		return true, "blocked: we don't accept count filters"
	})

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		if len(filter.Kinds) == 0 {
			return true, "blocked: please add kind 13194 or 23194 or 23195 or 23196"
		}

		if len(filter.Authors) == 0 && len(filter.Tags["p"]) == 0 {
			return true, "blocked: please add authors or #p"
		}

    for _, v := range filter.Kinds {
			if v != 13194 && v != 23194 && v != 23195 && v != 23196 {
				return true, "blocked: we only keep kind 13194 or 23194 or 23195 or 23196"
			}
		}

		return false, ""
	})

	relay.RejectEvent = append(relay.RejectEvent, policies.RestrictToSpecifiedKinds(true, 13194, 23194, 23195, 23196))

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		if !IsInTimeWindow(event.CreatedAt.Time().Unix(), config.AcceptEventsInRange) {
			return true, fmt.Sprintf("invalid: we only accept event on %d minute time frame", config.AcceptEventsInRange)
		}

		return false, ""
	})

	relay.OnEphemeralEvent = append(relay.OnEphemeralEvent, func(ctx context.Context, event *nostr.Event) {
		if err := mainDB.SaveEvent(ctx, event); err != nil {
			log.Printf("can't store event: %s\nerror: %s\n", event.String(), err.Error())
		}
	})

	relay.QueryEvents = append(relay.QueryEvents, mainDB.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, mainDB.DeleteEvent)

	go cleanDatabase()

	mux := relay.Router()
	mux.HandleFunc("GET /{$}", StaticViewHandler)

	log.Println("Relay running on port: " + config.RelayPort)
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
