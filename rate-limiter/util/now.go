package util

import "time"

// set value on file load
var startTime time.Time = time.Now()

func Now() time.Duration {
	return time.Since(startTime)
}
