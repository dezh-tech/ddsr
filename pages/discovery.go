package main

import (
	"context"
	"log"

	"github.com/nbd-wtf/go-nostr"
)

var (
	filters = []nostr.Filter{
		{
			Kinds: []int{0, 3, 5, 10002, 1984, 10063},
		},
	}

	relays = make([]*nostr.Relay, 0)
)

func runDiscovery() {
	for _, relayURL := range config.DiscoveryRelays {
		r, err := nostr.RelayConnect(context.Background(), relayURL)
		if err != nil {
			log.Printf("can't connect to %s\nerror: %s\n", relayURL, err.Error())

			continue
		}

		log.Printf("connoted to: %s\n", relayURL)

		relays = append(relays, r)
	}

	if len(relays) == 0 {
		log.Fatal("failed to connect to any relays!")
	}

	for _, r := range relays {
		go collect(r)
	}
}

func collect(r *nostr.Relay) {
	sub, err := r.Subscribe(context.Background(), filters)
	if err != nil {
		log.Printf("error on subscribing to: %s\nerror: %s\n", r.URL, err.Error())
	}

	// TODO::: retry to connect if reason is timeout.
	go func() {
		for rs := range sub.ClosedReason {
			log.Printf("Connection to %s closed\nReason: %s", r.URL, rs)
		}
	}()

	for ev := range sub.Events {
		reject := false
		for _, r := range relay.RejectEvent {
			rejectEvent, _ := r(context.Background(), ev)
			if rejectEvent {
				reject = true
			}
		}

		if reject {
			continue
		}

		log.Printf("discovered new event: %s\n", ev.String())

		if ev.Kind == 5 {
			for _, d := range relay.DeleteEvent {
				if err := d(context.Background(), ev); err != nil {
					log.Println(err)
				}
			}
		}

		for _, s := range relay.StoreEvent {
			if err := s(context.Background(), ev); err != nil {
				log.Println(err)
			}
		}
	}
}
