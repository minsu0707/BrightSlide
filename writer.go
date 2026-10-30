package main

import (
	"math"
	"sync"
	"sync/atomic"
	"time"
)

const (
	verifyAttempts = 3
	settleSteps    = 6
	settleStep     = 20 * time.Millisecond

	sampleWindow  = 80 * time.Millisecond // slider history used to estimate drag speed
	stoppedAfter  = 40 * time.Millisecond // no slider event for this long means the drag paused
	maxLeadValues = 12                    // never aim further ahead than this many percent
	maxSamples    = 8
)

// leadTime is roughly how long one DDC/CI write takes to land (~60 ms measured). While the
// slider is moving, the writer aims where the slider will be by then so the panel trails less.
var leadTime = 60 * time.Millisecond

type sample struct {
	at time.Time
	v  int
}

// writer pushes the newest slider value to the hardware, one call in flight at a time.
// DDC/CI monitors silently drop commands that arrive in quick succession, so once the
// slider goes idle it reads the value back and re-sends until the panel reports it.
type writer struct {
	d      *display
	target atomic.Int32
	wake   chan struct{}
	quit   chan struct{}
	done   chan struct{}

	mu      sync.Mutex
	history []sample
}

func newWriter(d *display) *writer {
	w := &writer{d: d, wake: make(chan struct{}, 1), quit: make(chan struct{}), done: make(chan struct{})}
	w.target.Store(int32(d.value))
	go w.loop(d.value)
	return w
}

func (w *writer) set(v int) {
	w.mu.Lock()
	if n := len(w.history); n == 0 || w.history[n-1].v != v {
		w.history = append(w.history, sample{time.Now(), v})
		if len(w.history) > maxSamples {
			w.history = w.history[1:]
		}
	}
	w.mu.Unlock()

	w.target.Store(int32(v))
	select {
	case w.wake <- struct{}{}:
	default:
	}
}

func (w *writer) stop() {
	close(w.quit)
	<-w.done
}

// aim returns the value to send for slider position t: t itself when the slider is still,
// or t pushed ahead by the current drag speed when it is moving.
func (w *writer) aim(t int) int {
	if leadTime <= 0 {
		return t
	}
	now := time.Now()
	w.mu.Lock()
	h := w.history
	var first, last sample
	ok := len(h) >= 2
	if ok {
		last = h[len(h)-1]
		for _, s := range h {
			if last.at.Sub(s.at) <= sampleWindow {
				first = s
				break
			}
		}
	}
	w.mu.Unlock()

	if !ok || last.v != t || now.Sub(last.at) > stoppedAfter || last.at.Sub(first.at) <= 0 {
		return t
	}
	velocity := float64(last.v-first.v) / last.at.Sub(first.at).Seconds()
	lead := int(math.Round(velocity * leadTime.Seconds()))
	lead = clamp(lead, -maxLeadValues, maxLeadValues)
	return clamp(t+lead, 0, 100)
}

func (w *writer) loop(sent int) {
	defer close(w.done)
	for {
		select {
		case <-w.wake:
		case <-w.quit:
			return
		}
		attempts := verifyAttempts
		for {
			t := int(w.target.Load())
			if want := w.aim(t); want != sent {
				sent = want
				w.d.ctl.Set(want)
				attempts = verifyAttempts
				continue
			}
			if attempts == 0 {
				break
			}
			idle, quit := w.waitIdle(t)
			if quit {
				return
			}
			if !idle {
				continue
			}
			if cur, ok := w.d.ctl.Get(); ok && abs(cur-w.d.ctl.Effective(t)) <= 1 {
				break
			}
			attempts--
			w.d.ctl.Set(t)
		}
	}
}

// waitIdle returns idle=false as soon as a newer slider value shows up.
func (w *writer) waitIdle(t int) (idle, quit bool) {
	for i := 0; i < settleSteps; i++ {
		select {
		case <-time.After(settleStep):
		case <-w.quit:
			return false, true
		}
		if int(w.target.Load()) != t {
			return false, false
		}
	}
	return true, false
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
