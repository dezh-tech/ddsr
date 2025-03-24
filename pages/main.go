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
	"github.com/fiatjaf/khatru/policies"
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
	log.SetPrefix("pages ")
	log.Printf("Running %s\n", StringVersion())

	relay = khatru.NewRelay()

	LoadConfig()

	config.DiscoveryRelays = []string{
		"nos.lol", "purplepag.es",
		"jellyfish.land", "relay.primal.net", "nostr.mom", "nostr.wine", "nostr.land",
	}
	config.WorkingDirectory = "pages_wd/"
	config.RelayIcon = "https://file.nostrmedia.com/f/bd4ae3e67e29964d494172261dc45395c89f6bd2e774642e366127171dfb81f5/517788c462ab4f1ea8405ca09247ad35e5fdb462282aacc18e258896ded12bc4.png"

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
	relay.RejectEvent = append(relay.RejectEvent, func(_ context.Context, event *nostr.Event) (reject bool, msg string) {
		management.Lock()
		defer management.Unlock()

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

	relay.RejectEvent = append(relay.RejectEvent, func(ctx context.Context, event *nostr.Event) (reject bool, msg string) {
		return policies.ValidateKind(ctx, event)
	})

	relay.RejectEvent = append(relay.RejectEvent, policies.RestrictToSpecifiedKinds(false, 0, 3, 5, 10002))

	// blossom
	bl := blossom.New(relay, config.RelayURL)
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
		management.Lock()
		defer management.Unlock()

		_, banned := management.BannedPubkeys[auth.PubKey]
		if banned {
			return true, "blocked: you are blacklisted", http.StatusUnauthorized
		}

		return false, "", http.StatusOK
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
	http.ListenAndServe(":3334", relay)
	// http.ListenAndServe(config.RelayPort, relay)

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
