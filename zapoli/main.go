package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/fiatjaf/khatru/blossom"
	"github.com/fiatjaf/khatru/policies"
	"github.com/kehiy/blobstore/disk"
)


var (
	relay  *khatru.Relay
	config Config
)

func main() {

	log.SetPrefix("zapoli ")
	log.Printf("Running %s\n", StringVersion())

	relay = khatru.NewRelay()

	config := LoadConfig()

	relay.Info.Name = "zapoli"
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
		Path: config.DBPath,
	}

	persistStore.Init()
	defer persistStore.Close()

	store := &bluge.BlugeBackend{
		Path:          config.DBPath,
		RawEventStore: persistStore,
	}
	store.Init()
	defer store.Close()

	relay.StoreEvent = append(relay.StoreEvent, persistStore.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, persistStore.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, persistStore.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, persistStore.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, persistStore.ReplaceEvent)

	relay.RejectEvent = append(relay.RejectEvent, policies.ValidateKind)

	// blossom
	bl := blossom.New(relay, "0.0.0.0"+config.BlossomPort)
	bl.Store = blossom.EventStoreBlobIndexWrapper{Store: persistStore, ServiceURL: bl.ServiceURL}

	blobStorage := disk.New(config.BlobStoragePath)

	bl.StoreBlob = append(bl.StoreBlob, blobStorage.Store)
	bl.LoadBlob = append(bl.LoadBlob, blobStorage.Load)
	bl.DeleteBlob = append(bl.DeleteBlob, blobStorage.Delete)

	LoadManagement()

	relay.ManagementAPI.AllowPubKey = AllowPubkey
	relay.ManagementAPI.BanPubKey = BanPubkey


	mux := relay.Router()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<b>welcome</b> to zapoly!`)
	})

	fmt.Println("running on" + config.RelayPort)
	http.ListenAndServe(config.RelayPort, relay)


	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	sig := <-sigChan

	log.Print("Received signal: Initiating graceful shutdown", "signal", sig.String())
	persistStore.Close()
	store.Close()
	relay.Shutdown(context.Background())
}
