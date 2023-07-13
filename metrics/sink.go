package metrics

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"time"
)

var (
	_ Sink = &CounterSink{}
	_ Sink = &GaugeSink{}
	_ Sink = NewTrendSink()
	_ Sink = &RateSink{}
)

type Sink interface {
	Add(s Sample)                              // Add a sample to the sink.
	Format(t time.Duration) map[string]float64 // Data for thresholds.
	IsEmpty() bool                             // Check if the Sink is empty.

	Drain() ([]byte, error) // Drain encodes the current sink values and clears them.
	Merge([]byte) error     // Merge decoeds the given values and merges them with the values in the current sink.
}

// NewSink creates the related Sink for
// the provided MetricType.
func NewSink(mt MetricType) Sink {
	var sink Sink
	switch mt {
	case Counter:
		sink = &CounterSink{}
	case Gauge:
		sink = &GaugeSink{}
	case Trend:
		sink = NewTrendSink()
	case Rate:
		sink = &RateSink{}
	default:
		// Should not be possible to create
		// an invalid metric type except for specific
		// and controlled tests
		panic(fmt.Sprintf("MetricType %q is not supported", mt))
	}
	return sink
}

type CounterSink struct {
	Value float64
	First time.Time
}

func (c *CounterSink) Add(s Sample) {
	c.Value += s.Value
	if c.First.IsZero() {
		c.First = s.Time
	}
}

// IsEmpty indicates whether the CounterSink is empty.
func (c *CounterSink) IsEmpty() bool { return c.First.IsZero() }

func (c *CounterSink) Format(t time.Duration) map[string]float64 {
	return map[string]float64{
		"count": c.Value,
		"rate":  c.Value / (float64(t) / float64(time.Second)),
	}
}

// Drain encodes the current sink values and clears them.
// TODO: something more robust and efficient
func (c *CounterSink) Drain() ([]byte, error) {
	res := []byte(fmt.Sprintf("%d %b", c.First.UnixMilli(), c.Value))
	c.Value = 0
	return res, nil
}

// Merge decoeds the given values and merges them with the values in the current sink.
func (c *CounterSink) Merge(from []byte) error {
	var firstMs int64
	var val float64
	_, err := fmt.Sscanf(string(from), "%d %b", &firstMs, &val)
	if err != nil {
		return err
	}

	c.Value += val
	if first := time.UnixMilli(firstMs); c.First.After(first) {
		c.First = first
	}

	return nil
}

type GaugeSink struct {
	Last     time.Time
	Value    float64
	Max, Min float64
	minSet   bool
}

// IsEmpty indicates whether the GaugeSink is empty.
func (g *GaugeSink) IsEmpty() bool { return !g.minSet }

func (g *GaugeSink) Add(s Sample) {
	g.Last = s.Time
	g.Value = s.Value
	if s.Value > g.Max {
		g.Max = s.Value
	}
	if s.Value < g.Min || !g.minSet {
		g.Min = s.Value
		g.minSet = true
	}
}

func (g *GaugeSink) Format(t time.Duration) map[string]float64 {
	return map[string]float64{"value": g.Value}
}

// Drain encodes the current sink values and clears them.
//
// TODO: something more robust and efficient
func (g *GaugeSink) Drain() ([]byte, error) {
	res := []byte(fmt.Sprintf("%d %b %b %b", g.Last.UnixMilli(), g.Value, g.Min, g.Max))

	g.Last = time.Time{}
	g.Value = 0

	return res, nil
}

// Merge decoeds the given values and merges them with the values in the current sink.
func (g *GaugeSink) Merge(from []byte) error {
	var lastMms int64
	var val, min, max float64
	_, err := fmt.Sscanf(string(from), "%d %b %b %b", &lastMms, &val, &min, &max)
	if err != nil {
		return err
	}

	last := time.UnixMilli(lastMms)
	if last.After(g.Last) {
		g.Last = last
		g.Value = val
	}

	if max > g.Max {
		g.Max = max
	}
	if min < g.Min || !g.minSet {
		g.Min = min
		g.minSet = true
	}

	return nil
}

// NewTrendSink makes a Trend sink with the OpenHistogram circllhist histogram.
func NewTrendSink() *TrendSink {
	return &TrendSink{}
}

type TrendSink struct {
	values []float64
	sorted bool

	count    uint64
	min, max float64
	sum      float64
}

// IsEmpty indicates whether the TrendSink is empty.
func (t *TrendSink) IsEmpty() bool { return t.count == 0 }

func (t *TrendSink) Add(s Sample) {
	if t.count == 0 {
		t.max, t.min = s.Value, s.Value
	} else {
		if s.Value > t.max {
			t.max = s.Value
		}
		if s.Value < t.min {
			t.min = s.Value
		}
	}

	t.values = append(t.values, s.Value)
	t.sorted = false
	t.count++
	t.sum += s.Value
}

// P calculates the given percentile from sink values.
func (t *TrendSink) P(pct float64) float64 {
	switch t.count {
	case 0:
		return 0
	case 1:
		return t.values[0]
	default:
		if !t.sorted {
			sort.Float64s(t.values)
			t.sorted = true
		}

		// If percentile falls on a value in Values slice, we return that value.
		// If percentile does not fall on a value in Values slice, we calculate (linear interpolation)
		// the value that would fall at percentile, given the values above and below that percentile.
		i := pct * (float64(t.count) - 1.0)
		j := t.values[int(math.Floor(i))]
		k := t.values[int(math.Ceil(i))]
		f := i - math.Floor(i)
		return j + (k-j)*f
	}
}

// Min returns the minimum value.
func (t *TrendSink) Min() float64 {
	return t.min
}

// Max returns the maximum value.
func (t *TrendSink) Max() float64 {
	return t.max
}

// Count returns the number of recorded values.
func (t *TrendSink) Count() uint64 {
	return t.count
}

// Avg returns the average (i.e. mean) value.
func (t *TrendSink) Avg() float64 {
	if t.count > 0 {
		return t.sum / float64(t.count)
	}
	return 0
}

// Total returns the total (i.e. "sum") value for all measurements.
func (t *TrendSink) Total() float64 {
	return t.sum
}

func (t *TrendSink) Format(tt time.Duration) map[string]float64 {
	// TODO: respect the summaryTrendStats for REST API
	return map[string]float64{
		"min":   t.Min(),
		"max":   t.Max(),
		"avg":   t.Avg(),
		"med":   t.P(0.5),
		"p(90)": t.P(0.90),
		"p(95)": t.P(0.95),
	}
}

// Drain encodes the current sink values and clears them.
//
// TODO: obviously use something more efficient (e.g. protobuf)
func (t *TrendSink) Drain() ([]byte, error) {
	res, err := json.Marshal(t.values)
	*t = TrendSink{}
	return res, err
}

// Merge decoeds the given values and merges them with the values in the current sink.
func (t *TrendSink) Merge(from []byte) error {
	// TODO: obviously use something more efficient (e.g. protobuf), this is
	// just for demo purposes
	var values []float64
	if err := json.Unmarshal(from, &values); err != nil {
		return err
	}
	for _, v := range values {
		t.Add(Sample{Value: v})
	}
	return nil
}

type RateSink struct {
	Trues int64
	Total int64
}

// IsEmpty indicates whether the RateSink is empty.
func (r *RateSink) IsEmpty() bool { return r.Total == 0 }

func (r *RateSink) Add(s Sample) {
	r.Total += 1
	if s.Value != 0 {
		r.Trues += 1
	}
}

func (r RateSink) Format(t time.Duration) map[string]float64 {
	var rate float64
	if r.Total > 0 {
		rate = float64(r.Trues) / float64(r.Total)
	}

	return map[string]float64{"rate": rate}
}

// Drain encodes the current sink values and clears them.
//
// TODO: something more robust and efficient
func (r *RateSink) Drain() ([]byte, error) {
	res := []byte(fmt.Sprintf("%d %d", r.Trues, r.Total))
	r.Trues = 0
	r.Total = 0
	return res, nil
}

// Merge decoeds the given values and merges them with the values in the current sink.
func (r *RateSink) Merge(from []byte) error {
	var trues, total int64
	_, err := fmt.Sscanf(string(from), "%d %d", &trues, &total)
	if err != nil {
		return err
	}

	r.Trues += trues
	r.Total += total
	return nil
}
