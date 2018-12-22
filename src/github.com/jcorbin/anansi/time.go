package anansi

import "time"

// Timer implements a timer useful for timing frame draws. Its channel should
// be used in a terminal app main loop select. Its Start and Stop methods may
// be called during an update after leaving such select.
type Timer struct {
	C        <-chan time.Time
	deadline time.Time
	timer    *time.Timer
}

// Request the timer to expire at most d time in the future, maybe sooner.
// Should not be called concurrently with Stop.
func (t *Timer) Request(d time.Duration) {
	now := time.Now()
	if dd := t.deadline.Sub(now); dd > 0 && dd < d {
		return
	}
	t.deadline = now.Add(d)
	if t.timer == nil {
		t.timer = time.NewTimer(d)
		t.C = t.timer.C
	} else {
		t.timer.Reset(d)
	}
}

// Cancel the timer. Should not be called concurrently with Request or
// receiving on Timer.C.
func (t *Timer) Cancel() {
	t.deadline = time.Time{}
	if !t.timer.Stop() {
		<-t.C
	}
}
