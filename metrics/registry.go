package metrics

import (
	"fmt"
	"regexp"
	"sync"

	"github.com/mstoykov/atlas"
)

// Registry is what can create metrics
type Registry struct {
	metrics map[string]*Metric
	l       sync.RWMutex

	rootTagSet *atlas.Node
}

// NewRegistry returns a new registry
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]*Metric),
		// All the new TagSts must branch out from this root, otherwise
		// comparing them and using their Equals() method won't work correctly.
		rootTagSet: atlas.New(),
	}
}

const nameRegexString = "^[\\p{L}\\p{N}\\._ !\\?/&#\\(\\)<>%-]{1,128}$"

var compileNameRegex = regexp.MustCompile(nameRegexString)

func checkName(name string) bool {
	return compileNameRegex.Match([]byte(name))
}

// NewMetric returns new metric registered to this registry
// TODO have multiple versions returning specific metric types when we have such things
func (r *Registry) NewMetric(name string, typ MetricType, t ...ValueType) (*Metric, error) {
	r.l.Lock()
	defer r.l.Unlock()

	if !checkName(name) {
		return nil, fmt.Errorf("Invalid metric name: '%s'", name) //nolint:golint,stylecheck
	}
	oldMetric, ok := r.metrics[name]

	if !ok {
		m := r.newMetric(name, typ, t...)
		r.metrics[name] = m
		return m, nil
	}
	if oldMetric.Type != typ {
		return nil, fmt.Errorf("metric '%s' already exists but with type %s, instead of %s", name, oldMetric.Type, typ)
	}
	if len(t) > 0 {
		if t[0] != oldMetric.Contains {
			return nil, fmt.Errorf("metric '%s' already exists but with a value type %s, instead of %s",
				name, oldMetric.Contains, t[0])
		}
	}
	return oldMetric, nil
}

// MustNewMetric is like NewMetric, but will panic if there is an error
func (r *Registry) MustNewMetric(name string, typ MetricType, t ...ValueType) *Metric {
	m, err := r.NewMetric(name, typ, t...)
	if err != nil {
		panic(err)
	}
	return m
}

func (r *Registry) newMetric(name string, mt MetricType, vt ...ValueType) *Metric {
	valueType := Default
	if len(vt) > 0 {
		valueType = vt[0]
	}

	return &Metric{
		registry: r,
		Name:     name,
		Type:     mt,
		Contains: valueType,
	}
}

// Get returns the Metric with the given name. If that metric doesn't exist,
// Get() will return a nil value.
func (r *Registry) Get(name string) *Metric {
	return r.metrics[name]
}

// RootTagSet is the empty root set that all other TagSets must originate from.
func (r *Registry) RootTagSet() *TagSet {
	return (*TagSet)(r.rootTagSet)
}
