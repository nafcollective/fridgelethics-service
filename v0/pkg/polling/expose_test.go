package polling

import "time"

func SetAfter(fn func(d time.Duration) <-chan time.Time) {
	after = fn
}

func SetDone(fn func()) {
	done = fn
}
