package timemetrics

import (
	"fmt"
	"sync/atomic"
	"time"
)

// Counters hold an int64 value that can be incremented and decremented.
type Counter interface {
	Clear(time.Time)
	Count() int64
	Dec(time.Time, int64)
	Inc(time.Time, int64)
	Update(time.Time, int64)
	GetMaxTime() time.Time
	GetKeys(time.Time, string) []string
	NbKeys() int
	Stale(time.Time) bool
	PushKeysTime(time.Time) bool
}

// NewCounter constructs a new StandardCounter.
func NewCounter(t time.Time, staleThreshold int) Counter {
	return &StandardCounter{0, t, staleThreshold}
}

// StandardCounter is the standard implementation of a Counter and uses the
// sync/atomic package to manage a single int64 value.
type StandardCounter struct {
	count          int64
	lastUpdate     time.Time
	staleThreshold int
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

func (c *StandardCounter) Update(t time.Time, i int64) {
	c.Inc(t, i)
}

func (c *StandardCounter) GetMaxTime() time.Time {
	return c.lastUpdate
}

func (c *StandardCounter) GetKeys(ct time.Time, name string) []string {
	t := int(c.GetMaxTime().Unix())

	keys := make([]string, 1)
	keys[0] = fmt.Sprintf(name, "count", t, fmt.Sprintf("%d", c.Count()))

	return keys
}

func (c *StandardCounter) NbKeys() int {
	return 1
}

func (c *StandardCounter) Stale(t time.Time) bool {
	return t.Sub(c.GetMaxTime()) > time.Duration(c.staleThreshold)*time.Minute
}

func (c *StandardCounter) PushKeysTime(t time.Time) bool {
	return c.lastUpdate.After(t)
}
