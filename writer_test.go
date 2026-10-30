package main

import (
	"math"
	"sync"
	"testing"
	"time"
)

// fakeDDC mimics a monitor that needs ~61 ms per write and records when each value lands.
type fakeDDC struct {
	mu      sync.Mutex
	current int
	applied []sample
}

func (f *fakeDDC) Get() (int, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.current, true
}

func (f *fakeDDC) Set(p int) bool {
	time.Sleep(61 * time.Millisecond)
	f.mu.Lock()
	f.current = p
	f.applied = append(f.applied, sample{time.Now(), p})
	f.mu.Unlock()
	return true
}

func (f *fakeDDC) Effective(p int) int { return p }
func (f *fakeDDC) Close()              {}

// drag moves a simulated slider 0 -> 100 over `duration` (events every 8 ms, like a fast mouse
// drag) and returns the mean gap between the slider and what the panel showed at each write.
func drag(t *testing.T, lead, duration time.Duration, path func(frac float64) float64) (meanGap float64, final, last int) {
	leadTime = lead
	f := &fakeDDC{}
	d := &display{ctl: f}
	w := newWriter(d)

	start := time.Now()
	slider := func(at time.Time) int {
		return int(math.Round(100 * path(math.Min(1, float64(at.Sub(start))/float64(duration)))))
	}
	for time.Since(start) < duration {
		w.set(slider(time.Now()))
		time.Sleep(8 * time.Millisecond)
	}
	w.set(100)
	time.Sleep(700 * time.Millisecond)
	w.stop()

	f.mu.Lock()
	defer f.mu.Unlock()
	var sum float64
	n := 0
	for _, a := range f.applied {
		if a.at.Sub(start) > duration {
			break
		}
		sum += math.Abs(float64(slider(a.at) - a.v))
		n++
	}
	if n == 0 {
		t.Fatal("no writes during the drag")
	}
	return sum / float64(n), f.current, f.applied[len(f.applied)-1].v
}

func TestPredictionReducesLagAndStillLandsOnFinalValue(t *testing.T) {
	defer func(old time.Duration) { leadTime = old }(leadTime)

	base, baseFinal, _ := drag(t, 0, 1200*time.Millisecond, ramp)
	pred, predFinal, predLast := drag(t, 60*time.Millisecond, 1200*time.Millisecond, ramp)
	t.Logf("mean gap between slider and panel: no prediction %.1f, with prediction %.1f", base, pred)

	if baseFinal != 100 || predFinal != 100 || predLast != 100 {
		t.Fatalf("final value not reached: base=%d predicted=%d last=%d", baseFinal, predFinal, predLast)
	}
	if pred >= base*0.8 {
		t.Fatalf("prediction did not reduce the lag enough: %.1f vs %.1f", pred, base)
	}
}

func TestStillSliderIsNeverPredicted(t *testing.T) {
	leadTime = 60 * time.Millisecond
	w := &writer{}
	w.history = []sample{{time.Now().Add(-500 * time.Millisecond), 10}, {time.Now().Add(-400 * time.Millisecond), 40}}
	if got := w.aim(40); got != 40 {
		t.Fatalf("paused slider should send its own value, got %d", got)
	}
	w.history = nil
	if got := w.aim(55); got != 55 {
		t.Fatalf("single click should send its own value, got %d", got)
	}
}

func ramp(frac float64) float64 { return frac }

// wiggle swings the slider back and forth, the worst case for look-ahead (many reversals).
func wiggle(frac float64) float64 { return 0.5 - 0.5*math.Cos(frac*2*math.Pi*3) }

func TestPredictionDoesNotHurtWhenSliderReverses(t *testing.T) {
	defer func(old time.Duration) { leadTime = old }(leadTime)

	base, _, _ := drag(t, 0, 1800*time.Millisecond, wiggle)
	pred, _, _ := drag(t, 60*time.Millisecond, 1800*time.Millisecond, wiggle)
	t.Logf("wiggle: mean gap no prediction %.1f, with prediction %.1f", base, pred)

	if pred > base*1.1 {
		t.Fatalf("prediction made reversing drags worse: %.1f vs %.1f", pred, base)
	}
}
