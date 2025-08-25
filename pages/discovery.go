package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

var (
	filters = []nostr.Filter{
		{
			Kinds: kindsToDiscoverAndAccept,
		},
	}

	relays = make([]*nostr.Relay, 0)

	failedMapMutex sync.RWMutex
	failedMap      = make(map[string]int)
)

func runDiscovery() {
	for _, relayURL := range config.DiscoveryRelays {
		if relayURL == "" {
			log.Println("empty relay URL found in config, skipping")
			continue
		}

		r, err := nostr.RelayConnect(context.Background(), relayURL)
		if err != nil {
			log.Printf("can't connect to %s\nerror: %s\n", relayURL, err.Error())
			continue
		}

		failedMapMutex.Lock()
		failedMap[r.URL] = 0
		failedMapMutex.Unlock()

		log.Printf("connected to: %s\n", relayURL)
		relays = append(relays, r)
	}

	if len(relays) == 0 {
		log.Fatal("failed to connect to any relays!")
	}

	for _, r := range relays {
		go collect(r)
	}
}

func reconnect(r *nostr.Relay) {
	if r == nil {
		log.Println("attempted to reconnect nil relay")
		return
	}

	failedMapMutex.RLock()
	failures := failedMap[r.URL]
	failedMapMutex.RUnlock()

	if failures < 10 {
		log.Printf("failed to connect to %s for %dth time, skipping...\n", r.URL, failures)
		return
	}

	time.Sleep(time.Second * 10)
	log.Printf("reconnect to %s for %dth time\n", r.URL, failures)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := r.Connect(ctx); err != nil {
		log.Printf("failed to reconnect to %s: %v\n", r.URL, err)
		return
	}
	go collect(r)
}

func collect(r *nostr.Relay) {
	if r == nil {
		log.Println("attempted to collect from nil relay")
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sub, err := r.Subscribe(ctx, filters)
	if err != nil {
		log.Printf("error on subscribing to: %s\nerror: %s\n", r.URL, err.Error())
		failedMapMutex.Lock()
		failedMap[r.URL]++
		failedMapMutex.Unlock()
		reconnect(r)
		return
	}

	go func() {
		for rs := range sub.ClosedReason {
			log.Printf("Connection to %s closed\nReason: %s\n", r.URL, rs)
			failedMapMutex.Lock()
			failedMap[r.URL]++
			failedMapMutex.Unlock()
			reconnect(r)
		}
	}()

	for ev := range sub.Events {
		if ev == nil {
			continue
		}

		reject := false
		for _, r := range relay.RejectEvent {
			if r == nil {
				continue
			}
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
				if d == nil {
					continue
				}
				if err := d(context.Background(), ev); err != nil {
					log.Printf("error in delete handler: %v\n", err)
				}
			}
		}

		for _, s := range relay.StoreEvent {
			if s == nil {
				continue
			}
			if err := s(context.Background(), ev); err != nil {
				log.Printf("error in store handler: %v\n", err)
			}
		}
	}
}
