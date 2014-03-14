package timemetrics

import (
	"sync/atomic"
	"time"
)

// Counters hold an int64 value that can be incremented and decremented.
type Counter interface {
	Clear(time.Time)
	Count() int64
	Dec(time.Time, int64)
	Inc(time.Time, int64)
	Snapshot() Counter
	GetMaxTime() time.Time
}

// NewCounter constructs a new StandardCounter.
func NewCounter(t time.Time) Counter {
	if UseNilMetrics {
		return NilCounter{}
	}
	return &StandardCounter{0, t}
}

// CounterSnapshot is a read-only copy of another Counter.
type CounterSnapshot struct {
	count      int64
	lastUpdate time.Time
}

// Clear panics.
func (CounterSnapshot) Clear(time.Time) {
	panic("Clear called on a CounterSnapshot")
}

// Count returns the count at the time the snapshot was taken.
func (c CounterSnapshot) Count() int64 { return int64(c.count) }

// Dec panics.
func (CounterSnapshot) Dec(time.Time, int64) {
	panic("Dec called on a CounterSnapshot")
}

// Inc panics.
func (CounterSnapshot) Inc(time.Time, int64) {
	panic("Inc called on a CounterSnapshot")
}

func (c CounterSnapshot) GetMaxTime() time.Time {
	return c.lastUpdate
}

// Snapshot returns the snapshot.
func (c CounterSnapshot) Snapshot() Counter { return c }

// NilCounter is a no-op Counter.
type NilCounter struct{}

// Clear is a no-op.
func (NilCounter) Clear(time.Time) {}

// Count is a no-op.
func (NilCounter) Count() int64 { return 0 }

// Dec is a no-op.
func (NilCounter) Dec(t time.Time, i int64) {}

// Inc is a no-op.
func (NilCounter) Inc(t time.Time, i int64) {}

// Snapshot is a no-op.
func (NilCounter) Snapshot() Counter { return NilCounter{} }

func (NilCounter) GetMaxTime() time.Time { return time.Now() }

// StandardCounter is the standard implementation of a Counter and uses the
// sync/atomic package to manage a single int64 value.
type StandardCounter struct {
	count      int64
	lastUpdate time.Time
}

// Clear sets the counter to zero.
func (c *StandardCounter) Clear(t time.Time) {
	atomic.StoreInt64(&c.count, 0)
	c.lastUpdate = t
}

// Count returns the current count.
func (c *StandardCounter) Count() int64 {
	return atomic.LoadInt64(&c.count)
}

// Dec decrements the counter by the given amount.
func (c *StandardCounter) Dec(t time.Time, i int64) {
	atomic.AddInt64(&c.count, -i)
	if t.After(c.lastUpdate) {
		c.lastUpdate = t
	}
}

// Inc increments the counter by the given amount.
func (c *StandardCounter) Inc(t time.Time, i int64) {
	atomic.AddInt64(&c.count, i)
	if t.After(c.lastUpdate) {
		c.lastUpdate = t
	}
}

// Snapshot returns a read-only copy of the counter.
func (c *StandardCounter) Snapshot() Counter {
	return CounterSnapshot{count: c.count, lastUpdate: c.lastUpdate}
}

func (c *StandardCounter) GetMaxTime() time.Time {
	return c.lastUpdate
}
