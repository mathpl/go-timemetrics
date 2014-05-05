package timemetrics

import (
	"time"
)

type Metric interface {
	Update(time.Time, int64)
	GetKeys(time.Time, string) []string
	GetMaxTime() time.Time
	NbKeys() int
	PushKeysTime(time.Time) bool
	Stale(time.Time) bool
	ZeroOut()
}
