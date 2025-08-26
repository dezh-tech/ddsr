package main

import (
	"context"
	"log"

	"github.com/nbd-wtf/go-nostr"
)

func broadcast(e *nostr.Event) {
	relayUrls := getProxiedWriteList(e.PubKey)

	for _, url := range relayUrls {
		relay, err := nostr.RelayConnect(context.Background(), url)
		if err != nil {
			log.Printf("failed to connect to %s: %v\n", url, err)
			continue
		}

		err = relay.Publish(context.Background(), *e)
		if err != nil {
			log.Printf("failed to publish to %s: %v\n", url, err)
		}
		relay.Close()
	}

}

func aggregate(f *nostr.Filters, pubkey string) map[string]*nostr.Event {
	events := make(map[string]*nostr.Event)

	relayUrls := getProxiedReadList(pubkey)

	for _, url := range relayUrls {
		relay, err := nostr.RelayConnect(context.Background(), url)
		if err != nil {
			continue
		}

		sub, err := relay.Subscribe(context.Background(), *f)
		if err != nil {
			relay.Close()
			continue
		}
		
		eventLoop:
		for {
			select {
			case ev := <-sub.Events:
				if isValid, err := ev.CheckSignature(); err == nil && isValid {
					events[ev.ID] = ev
				}
			case <-sub.EndOfStoredEvents:
				break eventLoop
			}
		}

		relay.Close()
	}

	return events
}
