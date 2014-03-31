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
	Tick(time.Time)
	Update(int64)
}

// NewEWMA constructs a new EWMA with the given alpha.
func NewEWMA(t time.Time, over int) EWMA {
	return &StandardEWMA{lastUpdate: t, over: over}
}

// NewEWMA1 constructs a new EWMA for a one-minute moving average.
func NewEWMA1(t time.Time) EWMA {
	return NewEWMA(t, 1)
}

// NewEWMA5 constructs a new EWMA for a five-minute moving average.
func NewEWMA5(t time.Time) EWMA {
	return NewEWMA(t, 5)
}

// NewEWMA15 constructs a new EWMA for a fifteen-minute moving average.
func NewEWMA15(t time.Time) EWMA {
	return NewEWMA(t, 15)
}

// StandardEWMA is the standard implementation of an EWMA and tracks the number
// of uncounted events and processes them on each tick.  It uses the
// sync/atomic package to manage uncounted events.
type StandardEWMA struct {
	uncounted  int64 // /!\ this should be the first member to ensure 64-bit alignment
	rate       float64
	init       bool
	over       int
	lastUpdate time.Time
}

// Rate returns the moving average rate of events per minute.
func (a *StandardEWMA) Rate() float64 {
	return a.rate * 1e9
}

// Tick ticks the clock to update the moving average.  It assumes it is called
// every five seconds.
// FIXME: tick only when it's time.
func (a *StandardEWMA) Tick(t time.Time) {
	diff_time := t.Unix() - a.lastUpdate.Unix()
	if diff_time != 0 {
		instantRate := float64(a.uncounted) / float64(diff_time*1e9)

		//Recalculate alpha
		alpha := float64(1 - math.Exp(float64(-diff_time)/60.0/float64(a.over)))

		//fmt.Printf("%d / %d / %d = %f\n", -diff_time, 60, a.over, float64(-diff_time)/60.0/float64(a.over))
		//fmt.Printf("%f * (%f - %f) = %f\n", a.alpha, instantRate, a.rate, a.alpha*(instantRate-a.rate))
		if a.init {
			a.rate += alpha * (instantRate - a.rate)
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
