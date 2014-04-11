package timemetrics

import (
	"time"
)

type Metric interface {
	Update(time.Time, int64)
	GetKeys(time.Time, string) []string
	GetMaxTime() time.Time
	NbKeys() int
	Stale(time.Time) bool
}
