package timemetrics

import (
	"fmt"
	"time"
)

// Meters count events to produce exponentially-weighted moving average rates
// at one-, five-, and fifteen-minutes and a mean rate.
type Meter interface {
	Count() int64
	Mark(time.Time, int64)
	CrunchEWMA(time.Time)
	Rate1() float64
	Rate5() float64
	Rate15() float64
	GetMaxTime() time.Time
	GetMaxEWMATime() time.Time
	Update(time.Time, int64)
	GetKeys(time.Time, string) []string
	NbKeys() int
	Stale(time.Time) bool
	PushKeysTime(t time.Time) bool
}

type timeValueTuple struct {
	v int64
	t time.Time
}

// NewMeter constructs a new StandardMeter and launches a goroutine.
func NewMeter(t time.Time, interval int, staleThreshold int) Meter {
	m := &StandardMeter{
		0,
		NewEWMA1(t),
		NewEWMA5(t),
		NewEWMA15(t),
		t,
		t,
		interval,
		staleThreshold,
	}

	return m
}

// StandardMeter is the standard implementation of a Meter and uses a
// goroutine to synchronize its calculations and a time.Ticker to pass time.
type StandardMeter struct {
	count          int64
	a1             EWMA
	a5             EWMA
	a15            EWMA
	lastUpdate     time.Time
	lastEWMAUpdate time.Time
	ewmaInterval   int
	staleThreshold int
}

// Count returns the number of events recorded.
func (m *StandardMeter) Count() int64 {
	return m.count
}

// Mark records the occurance of n events.
func (m *StandardMeter) Mark(t time.Time, n int64) {
	m.a1.Update(n)
	m.a5.Update(n)
	m.a15.Update(n)

	m.count++
	m.lastUpdate = t
}

func (m *StandardMeter) Update(t time.Time, i int64) {
	m.Mark(t, i)
}

// Rate1 returns the one-minute moving average rate of events per minute.
func (m *StandardMeter) Rate1() float64 {
	return m.a1.Rate()
}

// Rate5 returns the five-minute moving average rate of events per minute.
func (m *StandardMeter) Rate5() float64 {
	return m.a5.Rate()
}

// Rate15 returns the fifteen-minute moving average rate of events per minute.
func (m *StandardMeter) Rate15() float64 {
	return m.a15.Rate()
}

func (m *StandardMeter) GetMaxTime() time.Time {
	return m.lastUpdate
}

func (m *StandardMeter) GetMaxEWMATime() time.Time {
	return m.lastEWMAUpdate
}

func (m *StandardMeter) CrunchEWMA(t time.Time) {
	m.a1.Tick(t)
	m.a5.Tick(t)
	m.a15.Tick(t)

	m.lastEWMAUpdate = t
}

func (m *StandardMeter) GetKeys(ct time.Time, name string) []string {
	t := int(m.GetMaxTime().Unix())

	var keys []string
	if ct.Sub(m.lastEWMAUpdate) >= time.Duration(m.ewmaInterval)*time.Second {
		//fmt.Printf("%s - %s = %s > %s\n", ct, m.GetMaxEWMATime(), ct.Sub(m.GetMaxEWMATime()), time.Duration(m.ewmaInterval)*time.Second)
		//fmt.Printf("CRUNCH TIME: %s > %s\n", ct, time.Duration(m.ewmaInterval))
		m.CrunchEWMA(ct)
		keys = make([]string, 4)

		keys[1] = fmt.Sprintf(name, "rate._1min", t, fmt.Sprintf("%.4f", m.Rate1()))
		keys[2] = fmt.Sprintf(name, "rate._5min", t, fmt.Sprintf("%.4f", m.Rate5()))
		keys[3] = fmt.Sprintf(name, "rate._15min", t, fmt.Sprintf("%.4f", m.Rate15()))
	} else {
		keys = make([]string, 1)
	}

	keys[0] = fmt.Sprintf(name, "count", t, fmt.Sprintf("%d", m.Count()))

	return keys
}

func (m *StandardMeter) NbKeys() int {
	return 4
}

func (m *StandardMeter) Stale(t time.Time) bool {
	return t.Sub(m.GetMaxTime()) > time.Duration(m.staleThreshold)*time.Minute
}

func (m *StandardMeter) PushKeysTime(t time.Time) bool {
	return m.lastUpdate.After(t) || t.Sub(m.lastEWMAUpdate) > time.Duration(m.ewmaInterval)*time.Second
}
