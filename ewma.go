package timemetrics

import (
	//"fmt"
	"math"
	"time"
)

// EWMAs continuously calculate an exponentially-weighted moving average
// based on an outside source of clock ticks.
type EWMA interface {
	Rate() float64
	Snapshot() EWMA
	Tick(time.Time)
	Update(int64)
}

// NewEWMA constructs a new EWMA with the given alpha.
func NewEWMA(t time.Time, alpha float64) EWMA {
	if UseNilMetrics {
		return NilEWMA{}
	}
	return &StandardEWMA{alpha: alpha, lastUpdate: t}
}

// NewEWMA1 constructs a new EWMA for a one-minute moving average.
func NewEWMA1(t time.Time, interval int) EWMA {
	return NewEWMA(t, 1-math.Exp(float64(-interval)/60/1))
}

// NewEWMA5 constructs a new EWMA for a five-minute moving average.
func NewEWMA5(t time.Time, interval int) EWMA {
	return NewEWMA(t, 1-math.Exp(float64(-interval)/60/5))
}

// NewEWMA15 constructs a new EWMA for a fifteen-minute moving average.
func NewEWMA15(t time.Time, interval int) EWMA {
	return NewEWMA(t, 1-math.Exp(float64(-interval)/60/15))
}

// EWMASnapshot is a read-only copy of another EWMA.
type EWMASnapshot float64

// Rate returns the rate of events per second at the time the snapshot was
// taken.
func (a EWMASnapshot) Rate() float64 { return float64(a) }

// Snapshot returns the snapshot.
func (a EWMASnapshot) Snapshot() EWMA { return a }

// Tick panics.
func (EWMASnapshot) Tick(time.Time) {
	panic("Tick called on an EWMASnapshot")
}

// Update panics.
func (EWMASnapshot) Update(int64) {
	panic("Update called on an EWMASnapshot")
}

// NilEWMA is a no-op EWMA.
type NilEWMA struct{}

// Rate is a no-op.
func (NilEWMA) Rate() float64 { return 0.0 }

// Snapshot is a no-op.
func (NilEWMA) Snapshot() EWMA { return NilEWMA{} }

// Tick is a no-op.
func (NilEWMA) Tick(time.Time) {}

// Update is a no-op.
func (NilEWMA) Update(n int64) {}

// StandardEWMA is the standard implementation of an EWMA and tracks the number
// of uncounted events and processes them on each tick.  It uses the
// sync/atomic package to manage uncounted events.
type StandardEWMA struct {
	uncounted  int64 // /!\ this should be the first member to ensure 64-bit alignment
	alpha      float64
	rate       float64
	init       bool
	lastUpdate time.Time
}

// Rate returns the moving average rate of events per minute.
func (a *StandardEWMA) Rate() float64 {
	return a.rate
}

// Snapshot returns a read-only copy of the EWMA.
func (a *StandardEWMA) Snapshot() EWMA {
	return EWMASnapshot(a.Rate())
}

// Tick ticks the clock to update the moving average.  It assumes it is called
// every five seconds.
// FIXME: tick only when it's time.
func (a *StandardEWMA) Tick(t time.Time) {
	if a.uncounted != 0 && t.Sub(a.lastUpdate) != 0 {
		instantRate := float64(1e9*a.uncounted) / float64(t.Sub(a.lastUpdate))
		if a.init {
			a.rate += a.alpha * (instantRate - a.rate)
		} else {
			a.init = true
			a.rate = instantRate
		}
		a.uncounted = 0
		a.lastUpdate = t
	}
}

// Update adds n uncounted events.
func (a *StandardEWMA) Update(n int64) {
	a.uncounted += n
}
