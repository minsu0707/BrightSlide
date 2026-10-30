package main

import (
	"sync/atomic"
	"time"
)

const (
	verifyAttempts = 3
	settleSteps    = 6
	settleStep     = 20 * time.Millisecond
)

// writer pushes the newest slider value to the hardware, one call in flight at a time.
// DDC/CI monitors silently drop commands that arrive in quick succession, so once the
// slider goes idle it reads the value back and re-sends until the panel reports it.
type writer struct {
	d      *display
	target atomic.Int32
	wake   chan struct{}
	quit   chan struct{}
	done   chan struct{}
}

func newWriter(d *display) *writer {
	w := &writer{d: d, wake: make(chan struct{}, 1), quit: make(chan struct{}), done: make(chan struct{})}
	w.target.Store(int32(d.value))
	go w.loop(d.value)
	return w
}

func (w *writer) set(v int) {
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

func (w *writer) loop(last int) {
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
			if t != last {
				last = t
				w.d.ctl.Set(t)
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
