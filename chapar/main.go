package main

import (
	"context"
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

var (
	relay  *khatru.Relay
	config Config

	//go:embed static/index.html
	landingTempl []byte
)

func main() {
	log.SetPrefix("Chapar ")
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
	relay.Info.AddSupportedNIPs([]int{17, 59, 44})

	mainDB := &badger.BadgerBackend{
		Path: path.Join(config.WorkingDirectory, "database"),
	}
	mainDB.Init()

	relay.StoreEvent = append(relay.StoreEvent, mainDB.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, mainDB.QueryEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, mainDB.DeleteEvent)

	relay.RejectEvent = append(relay.RejectEvent, func(_ context.Context, event *nostr.Event) (reject bool, msg string) {
		if event.Kind != nostr.KindGiftWrap {
			return true, "blocked: only kind 1059 is accepted"
		}

		return false, ""
	})

	relay.RejectFilter = append(relay.RejectFilter, func(ctx context.Context, filter nostr.Filter) (reject bool, msg string) {
		authed := khatru.GetAuthed(ctx)
		if authed == "" {
			return true, "auth-required: you must authenticate first"
		}

		for _, p := range filter.Tags["p"] {
			if p != authed {
				return true, "blocked: you can only read your own messages"
			}
		}

		return false, ""
	})

	mux := relay.Router()
	mux.HandleFunc("GET /{$}", StaticViewHandler)

	log.Println("Relay running on port: " + config.RelayPort)
	http.ListenAndServe(config.RelayPort, relay)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	log.Print("Received signal: Initiating graceful shutdown", "signal", sig.String())
	mainDB.Close()
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
