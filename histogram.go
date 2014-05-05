package timemetrics

import (
	"fmt"
	"time"
)

// Histograms calculate distribution statistics from a series of int64 values.
type Histogram interface {
	Clear(time.Time)
	Count() int64
	Max() int64
	Mean() float64
	Min() int64
	Percentile(float64) float64
	Percentiles([]float64) []float64
	Sample() Sample
	StdDev() float64
	Update(time.Time, int64)
	Variance() float64
	GetMaxTime() time.Time
	GetKeys(time.Time, string) []string
	NbKeys() int
	Stale(time.Time) bool
	PushKeysTime(time.Time) bool
	ZeroOut()
}

// StandardHistogram is the standard implementation of a Histogram and uses a
// Sample to bound its memory use.
type StandardHistogram struct {
	sample         Sample
	lastUpdate     time.Time
	staleThreshold int
}

// NewHistogram constructs a new StandardHistogram from a Sample.
func NewHistogram(s Sample, staleThreshold int) Histogram {
	return &StandardHistogram{sample: s, staleThreshold: staleThreshold}
}

// Clear clears the histogram and its sample.
func (h *StandardHistogram) Clear(t time.Time) { h.sample.Clear(t) }

// Count returns the number of samples recorded since the histogram was last
// cleared.
func (h *StandardHistogram) Count() int64 { return h.sample.Count() }

// Max returns the maximum value in the sample.
func (h *StandardHistogram) Max() int64 { return h.sample.Max() }

// Mean returns the mean of the values in the sample.
func (h *StandardHistogram) Mean() float64 { return h.sample.Mean() }

// Min returns the minimum value in the sample.
func (h *StandardHistogram) Min() int64 { return h.sample.Min() }

// Percentile returns an arbitrary percentile of the values in the sample.
func (h *StandardHistogram) Percentile(p float64) float64 {
	return h.sample.Percentile(p)
}

// Percentiles returns a slice of arbitrary percentiles of the values in the
// sample.
func (h *StandardHistogram) Percentiles(ps []float64) []float64 {
	return h.sample.Percentiles(ps)
}

// Sample returns the Sample underlying the histogram.
func (h *StandardHistogram) Sample() Sample { return h.sample }

// StdDev returns the standard deviation of the values in the sample.
func (h *StandardHistogram) StdDev() float64 { return h.sample.StdDev() }

// Update samples a new value.
func (h *StandardHistogram) Update(t time.Time, v int64) {
	if t.After(h.lastUpdate) {
		h.lastUpdate = t
	}
	h.sample.Update(t, v)
}

// Variance returns the variance of the values in the sample.
func (h *StandardHistogram) Variance() float64 { return h.sample.Variance() }

func (h *StandardHistogram) GetMaxTime() time.Time { return h.lastUpdate }

func (h *StandardHistogram) GetKeys(ct time.Time, name string) []string {
	t := int(h.GetMaxTime().Unix())
	ps := h.Percentiles([]float64{0.5, 0.75, 0.95, 0.99, 0.999})

	keys := make([]string, 10)

	keys[0] = fmt.Sprintf(name, "min", t, fmt.Sprintf("%d", h.Min()))
	keys[1] = fmt.Sprintf(name, "max", t, fmt.Sprintf("%d", h.Max()))
	keys[2] = fmt.Sprintf(name, "mean", t, fmt.Sprintf("%.4f", h.Mean()))
	keys[3] = fmt.Sprintf(name, "std-dev", t, fmt.Sprintf("%.4f", h.StdDev()))
	keys[4] = fmt.Sprintf(name, "p50", t, fmt.Sprintf("%d", int64(ps[0])))
	keys[5] = fmt.Sprintf(name, "p75", t, fmt.Sprintf("%d", int64(ps[1])))
	keys[6] = fmt.Sprintf(name, "p95", t, fmt.Sprintf("%d", int64(ps[2])))
	keys[7] = fmt.Sprintf(name, "p99", t, fmt.Sprintf("%d", int64(ps[3])))
	keys[8] = fmt.Sprintf(name, "p999", t, fmt.Sprintf("%d", int64(ps[4])))
	keys[9] = fmt.Sprintf(name, "sample_size", t, fmt.Sprintf("%d", h.Sample().Size()))

	return keys
}

func (h *StandardHistogram) NbKeys() int {
	return 10
}

func (h *StandardHistogram) Stale(t time.Time) bool {
	return t.Sub(h.GetMaxTime()) > time.Duration(h.staleThreshold)*time.Minute
}

func (h *StandardHistogram) PushKeysTime(t time.Time) bool {
	return h.lastUpdate.After(t)
}

func (h *StandardHistogram) ZeroOut() {
	h.sample.ZeroOut()
}
