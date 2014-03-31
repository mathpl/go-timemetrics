package timemetrics

import (
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
}

type timeValueTuple struct {
	v int64
	t time.Time
}

// NewMeter constructs a new StandardMeter and launches a goroutine.
func NewMeter(t time.Time) Meter {
	m := &StandardMeter{
		0,
		NewEWMA1(t),
		NewEWMA5(t),
		NewEWMA15(t),
		t,
		t,
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

// arbiter receives inputs and sends outputs.  It counts each input and updates
// the various moving averages and the mean rate of events.  It sends a copy of
// the meterV as output.
//func (m *StandardMeter) arbiter() {
//	snapshot := &MeterSnapshot{}
//	a1 := NewEWMA1(m.lastEWMAUpdate, m.interval)
//	a5 := NewEWMA5(m.lastEWMAUpdate, m.interval)
//	a15 := NewEWMA15(m.lastEWMAUpdate, m.interval)

//	for {
//		select {
//		case n := <-m.in:
//			if n.t.After(m.lastUpdate) {
//				m.lastUpdate = n.t
//			}

//			if n.t.After(m.lastEWMAUpdate) {
//				m.lastEWMAUpdate = n.t
//			}

//			snapshot.count += n.v
//			a1.Update(n.v)
//			a5.Update(n.v)
//			a15.Update(n.v)
//			snapshot.rate1 = a1.Rate()
//			snapshot.rate5 = a5.Rate()
//			snapshot.rate15 = a15.Rate()

//			snapshot.lastUpdate = m.lastUpdate
//			snapshot.lastEWMAUpdate = m.lastEWMAUpdate
//		case m.out <- snapshot:
//		case n := <-m.crunch:
//			if n.t.After(m.lastEWMAUpdate) {
//				m.lastEWMAUpdate = n.t
//			}

//			a1.Tick(n.t)
//			a5.Tick(n.t)
//			a15.Tick(n.t)
//			snapshot.rate1 = a1.Rate()
//			snapshot.rate5 = a5.Rate()
//			snapshot.rate15 = a15.Rate()
//			snapshot.lastUpdate = m.lastUpdate
//			snapshot.lastEWMAUpdate = m.lastEWMAUpdate
//		}
//	}
//}
