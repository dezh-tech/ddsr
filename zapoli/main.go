package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/fiatjaf/eventstore/badger"
	"github.com/fiatjaf/eventstore/bluge"
	"github.com/fiatjaf/khatru"
	"github.com/nbd-wtf/go-nostr"
)

func main() {
	if err := godotenv.Load(); err != nil {
		return nil, Error{
			reason: err.Error(),
		}
	}

	relay := khatru.NewRelay()

	relay.Info.Name = "my relay"
	relay.Info.Software = "dezh.tech/ddsr/zapoli"
	relay.Info.Version = ""
	relay.Info. = ""
	relay.Info.PubKey = "79be667ef9dcbbac55a06295ce870b07029bfcdb2dce28d959f2815b16f81798"
	relay.Info.Description = "this is my custom relay"
	relay.Info.Icon = "https://external-content.duckduckgo.com/iu/?u=https%3A%2F%2Fliquipedia.net%2Fcommons%2Fimages%2F3%2F35%2FSCProbe.jpg&f=1&nofb=1&ipt=0cbbfef25bce41da63d910e86c3c343e6c3b9d63194ca9755351bb7c2efa3359&ipo=images"
	relay.Info.AddSupportedNIPs(50,82)
	

	persistStore :=  &badger.BadgerBackend{
		Path: "db",

	}

	persistStore.Init()
	defer persistStore.Close()


	store :=  &bluge.BlugeBackend{
		Path: "db",
		RawEventStore: persistStore,
	}
	store.Init()
	defer store.Close()

	relay.StoreEvent = append(relay.StoreEvent, store.SaveEvent)
	relay.QueryEvents = append(relay.QueryEvents, store.QueryEvents)
	relay.CountEvents = append(relay.CountEvents, store.CountEvents)
	relay.DeleteEvent = append(relay.DeleteEvent, store.DeleteEvent)
	relay.ReplaceEvent = append(relay.ReplaceEvent, store.ReplaceEvent)

	// there are many other configurable things you can set
	relay.RejectEvent = append(relay.RejectEvent,policies.ValidateKind)

	mux := relay.Router()
	// set up other http handlers
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("content-type", "text/html")
		fmt.Fprintf(w, `<b>welcome</b> to my relay!`)
	})

	// start the server
	fmt.Println("running on :3334")
	http.ListenAndServe(":3334", relay)
}