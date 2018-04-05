package exoip

import (
	"time"
)

// CurrentTimeMillis represents the current timestamp
func CurrentTimeMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}
