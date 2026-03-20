package timer

import "time"

func Reset(timer *time.Timer, duration time.Duration) {
	if !timer.Stop() {
		drainChannel(timer)
	}
	timer.Reset(duration)
}

func drainChannel(timer *time.Timer) {
	select {
	case <-timer.C:
	default:
	}
}
