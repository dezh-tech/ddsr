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
	"unicode/utf8"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
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
	log.SetPrefix("zapoli ")
	log.Printf("Running %s\n", StringVersion())

	relay = khatru.NewRelay()

	LoadConfig()

	relay.Info.Name = config.RelayName
	relay.Info.Software = "dezh.tech/ddsr/210maxi"
	relay.Info.Version = StringVersion()
	relay.Info.PubKey = config.RelayPubkey
	relay.Info.Description = config.RelayDescription
	relay.Info.Icon = config.RelayIcon
	relay.Info.Contact = config.RelayContact
	relay.Info.URL = config.RelayURL
	// relay.Info.Banner = config.RelayBanner
	relay.Info.SupportedNIPs = append(relay.Info.SupportedNIPs, "B1", 50, 45)

	mainDB := &badger.BadgerBackend{
		Path: path.Join(config.WorkingDirectory, "database"),
	}
	mainDB.Init()

	searchDB := &bluge.BlugeBackend{
		Path:          path.Join(config.WorkingDirectory, "search_database"),
		RawEventStore: mainDB,
	}
	searchDB.Init()

	relay.StoreEvent = append(relay.StoreEvent, mainDB.SaveEvent, searchDB.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, mainDB.QueryEvents, searchDB.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, mainDB.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, mainDB.DeleteEvent, searchDB.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, mainDB.ReplaceEvent, searchDB.ReplaceEvent)

	relay.RejectEvent = append(relay.RejectEvent,
		policies.RestrictToSpecifiedKinds(false, 25, 1111, 7, 5, 9734, 9735, 0, 3))

	relay.RejectEvent = append(relay.RejectEvent, func(_ context.Context, event *nostr.Event) (reject bool, msg string) {
		if event.Kind == 25 || event.Kind == 1111 {
			if utf8.RuneCountInString(event.Content) > 210 {
				return true, "invalid: .content must be <= 210 char."
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
