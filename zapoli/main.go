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
	"slices"
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/blossom"
	"github.com/kehiy/blobstore/disk"
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
	log.SetPrefix("zapoli ")
	log.Printf("Running %s\n", StringVersion())

	relay = khatru.NewRelay()

	LoadConfig()

	relay.Info.Name = config.RelayName
	relay.Info.Software = "dezh.tech/ddsr/zapoli"
	relay.Info.Version = StringVersion()
	relay.Info.PubKey = config.RelayPubkey
	relay.Info.Description = config.RelayDescription
	relay.Info.Icon = config.RelayIcon
	relay.Info.Contact = config.RelayContact
	relay.Info.URL = config.RelayURL
	// relay.Info.Banner = config.RelayBanner
	relay.Info.AddSupportedNIPs([]int{50, 82})

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
	relay.RejectEvent = append(relay.RejectEvent, func(_ context.Context, event *nostr.Event) (reject bool, msg string) {
		if !slices.Contains(management.AllowedPubkeys, event.PubKey) {
			return true, "restricted: you are not in the whitelist"
		}

		return false, ""
	})

	// blossom
	bl := blossom.New(relay, "http://0.0.0.0"+config.BlossomPort)
	bl.Store = blossom.EventStoreBlobIndexWrapper{Store: persistStore, ServiceURL: bl.ServiceURL}

	if !PathExists(path.Join(config.WorkingDirectory, "blossom")) {
		if err := Mkdir(path.Join(config.WorkingDirectory, "blossom")); err != nil {
			log.Fatalf("can't make blossom directory: %s", err.Error())
		}
	}
	blobStorage := disk.New(path.Join(config.WorkingDirectory, "blossom"))

	bl.StoreBlob = append(bl.StoreBlob, blobStorage.Store)
	bl.LoadBlob = append(bl.LoadBlob, blobStorage.Load)
	bl.DeleteBlob = append(bl.DeleteBlob, blobStorage.Delete)
	bl.RejectUpload = append(bl.RejectUpload, func(_ context.Context, auth *nostr.Event, _ int, _ string) (bool, string, int) {
		if !slices.Contains(management.AllowedPubkeys, auth.PubKey) {
			return true, "restricted: you are not in the whitelist", http.StatusUnauthorized
		}

		return false, "", http.StatusOK
	})

	LoadManagement()

	relay.ManagementAPI.AllowPubKey = AllowPubkey
	relay.ManagementAPI.BanPubKey = BanPubkey
	relay.ManagementAPI.RejectAPICall = append(relay.ManagementAPI.RejectAPICall,
		func(ctx context.Context, mp nip86.MethodParams) (reject bool, msg string) {
			auth := khatru.GetAuthed(ctx)
			if !slices.Contains(config.Admins, auth) {
				return true, "your are not an admin"
			}

			return false, ""
		})

	mux := relay.Router()
	mux.HandleFunc("GET /{$}", StaticViewHandler)

	log.Println("Relay running on port: ", config.RelayPort)
	log.Println("Blossom server running on port: ", config.BlossomPort)
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
