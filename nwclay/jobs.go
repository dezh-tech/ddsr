package main

import (
	"context"
	"log"
	"time"

	"github.com/nbd-wtf/go-nostr"
)

var filter = nostr.Filter{}

func cleanDatabase() {
	ticker := time.NewTicker(time.Duration(config.KeepNotesFor) * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Running cleaner...")
		for _, qe := range relay.QueryEvents {
			ch, err := qe(context.Background(), filter)
			if err != nil {
				log.Printf("can't read from database: %s", err.Error())
			}

			for ev := range ch {
				now := time.Now()
				duration := now.Sub(ev.CreatedAt.Time())
				minutesPassed := int(duration.Minutes())

				if (ev.Kind != 13194) && (minutesPassed >= config.KeepNotesFor) {
					for _, de := range relay.DeleteEvent {
						if err := de(context.Background(), ev); err != nil {
							log.Printf("can't delete event: %s\nerror: %s\n", ev.String(), err.Error())
						}
						log.Printf("event %s has been pruned\n", ev.String())
					}
				}
			}
		}
	}
}
