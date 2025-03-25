package main

import (
	"time"
)

// IsInTimeWindow checks if ts (unix timestamp) is at most for tw minutes ago or future.
func IsInTimeWindow(ts int64, tw int) bool {
	timestampTime := time.Unix(ts, 0)
	now := time.Now()

	startTime := now.Add(time.Duration(-tw) * time.Minute)
	endTime := now.Add(time.Duration(tw) * time.Minute)

	return timestampTime.After(startTime) && timestampTime.Before(endTime)
}
