// Package engine contains the internal metrics engine responsible for
// aggregating metrics during the test and evaluating thresholds against them.
package engine

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
	"go.k6.io/k6/errext"
	"go.k6.io/k6/errext/exitcodes"
	"go.k6.io/k6/lib"
	"go.k6.io/k6/metrics"
	"go.k6.io/k6/output"
	"gopkg.in/guregu/null.v3"
)

const thresholdsRate = 2 * time.Second

// MetricsEngine is the internal metrics engine that k6 uses to keep track of
// aggregated metric sample values. They are used to generate the end-of-test
// summary and to evaluate the test thresholds.
type MetricsEngine struct {
	es     *lib.ExecutionState
	logger logrus.FieldLogger

	outputIngester *outputIngester

	// These can be both top-level metrics or sub-metrics
	metricsWithThresholds []*metrics.Metric

	breachedThresholdsCount uint32

	sinks map[*metrics.Metric]metrics.Sink

	// TODO: completely refactor:
	//   - make these private, add a method to export the raw data
	//   - do not use an unnecessary map for the observed metrics
	//   - have one lock per metric instead of a a global one, when
	//     the metrics are decoupled from their types
	metricsLock     sync.Mutex
	observedMetrics map[string]*metrics.Metric
}

// NewMetricsEngine creates a new metrics Engine with the given parameters.
func NewMetricsEngine(es *lib.ExecutionState) (*MetricsEngine, error) {
	me := &MetricsEngine{
		es:              es,
		logger:          es.Test.Logger.WithField("component", "metrics-engine"),
		observedMetrics: make(map[string]*metrics.Metric),
		sinks:           make(map[*metrics.Metric]metrics.Sink),
	}

	if !(me.es.Test.RuntimeOptions.NoSummary.Bool && me.es.Test.RuntimeOptions.NoThresholds.Bool) {
		err := me.initSubMetricsAndThresholds()
		if err != nil {
			return nil, err
		}
	}

	return me, nil
}

// CreateIngester returns a pseudo-Output that uses the given metric samples to
// update the engine's inner state.
func (me *MetricsEngine) CreateIngester() output.Output {
	me.outputIngester = &outputIngester{
		logger:        me.logger.WithField("component", "metrics-engine-ingester"),
		metricsEngine: me,
	}
	return me.outputIngester
}

func (me *MetricsEngine) getThresholdMetricOrSubmetric(name string) (*metrics.Metric, error) {
	// TODO: replace with strings.Cut after Go 1.18
	nameParts := strings.SplitN(name, "{", 2)

	metric := me.es.Test.Registry.Get(nameParts[0])
	if metric == nil {
		return nil, fmt.Errorf("metric '%s' does not exist in the script", nameParts[0])
	}
	if len(nameParts) == 1 { // no sub-metric
		return metric, nil
	}

	submetricDefinition := nameParts[1]
	if submetricDefinition[len(submetricDefinition)-1] != '}' {
		return nil, fmt.Errorf("missing ending bracket, sub-metric format needs to be 'metric{key:value}'")
	}
	sm, err := metric.AddSubmetric(submetricDefinition[:len(submetricDefinition)-1])
	if err != nil {
		return nil, err
	}

	if sm.Metric.Observed {
		// Do not repeat warnings for the same sub-metrics
		return sm.Metric, nil
	}

	if _, ok := sm.Tags.Get("vu"); ok {
		me.logger.Warnf(
			"The high-cardinality 'vu' metric tag was made non-indexable in k6 v0.41.0, so thresholds"+
				" like '%s' that are based on it won't work correctly.",
			name,
		)
	}

	if _, ok := sm.Tags.Get("iter"); ok {
		me.logger.Warnf(
			"The high-cardinality 'iter' metric tag was made non-indexable in k6 v0.41.0, so thresholds"+
				" like '%s' that are based on it won't work correctly.",
			name,
		)
	}

	return sm.Metric, nil
}

func (me *MetricsEngine) markObserved(metric *metrics.Metric) {
	if !metric.Observed {
		metric.Observed = true
		me.observedMetrics[metric.Name] = metric
	}
}

func (me *MetricsEngine) initSubMetricsAndThresholds() error {
	for metricName, thresholds := range me.es.Test.Options.Thresholds {
		metric, err := me.getThresholdMetricOrSubmetric(metricName)

		if me.es.Test.RuntimeOptions.NoThresholds.Bool {
			if err != nil {
				me.logger.WithError(err).Warnf("Invalid metric '%s' in threshold definitions", metricName)
			}
			continue
		}

		if err != nil {
			return fmt.Errorf("invalid metric '%s' in threshold definitions: %w", metricName, err)
		}

		metric.Thresholds = thresholds
		me.metricsWithThresholds = append(me.metricsWithThresholds, metric)

		swm := metrics.NewSinkWithMetric(metric) // maybe better just SinkByType?
		me.sinks[metric] = swm.Sink

		// Mark the metric (and the parent metric, if we're dealing with a
		// submetric) as observed, so they are shown in the end-of-test summary,
		// even if they don't have any metric samples during the test run
		me.markObserved(metric)
		if metric.Sub != nil {
			subswm := metrics.NewSinkWithMetric(metric.Sub.Parent)
			me.sinks[subswm.Metric] = subswm.Sink
			me.markObserved(metric.Sub.Parent)
		}
	}

	// TODO: refactor out of here when https://github.com/grafana/k6/issues/1321
	// lands and there is a better way to enable a metric with tag
	if me.es.Test.Options.SystemTags.Has(metrics.TagExpectedResponse) {
		_, err := me.getThresholdMetricOrSubmetric("http_req_duration{expected_response:true}")
		if err != nil {
			return err // shouldn't happen, but ¯\_(ツ)_/¯
		}
	}

	return nil
}

// StartThresholdCalculations spins up a new goroutine to crunch thresholds and
// returns a callback that will stop the goroutine and finalizes calculations.
func (me *MetricsEngine) StartThresholdCalculations(abortRun func(error)) (
	finalize func() (breached []string),
) {
	stop := make(chan struct{})
	done := make(chan struct{})

	go func() {
		defer close(done)
		ticker := time.NewTicker(thresholdsRate)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				breached, shouldAbort := me.evaluateThresholds(true)
				if shouldAbort {
					err := fmt.Errorf(
						"thresholds on metrics '%s' were breached; at least one has abortOnFail enabled, stopping test prematurely",
						strings.Join(breached, ", "),
					)
					me.logger.Debug(err.Error())
					err = errext.WithAbortReasonIfNone(
						errext.WithExitCodeIfNone(err, exitcodes.ThresholdsHaveFailed), errext.AbortedByThreshold,
					)
					abortRun(err)
				}
			case <-stop:
				return
			}
		}
	}()

	return func() []string {
		if me.outputIngester != nil {
			// Stop the ingester so we don't get any more metrics
			err := me.outputIngester.Stop()
			if err != nil {
				me.logger.WithError(err).Warnf("There was a problem stopping the output ingester.")
			}
		}
		close(stop)
		<-done

		breached, _ := me.evaluateThresholds(false)
		return breached
	}
}

// evaluateThresholds processes all of the thresholds.
//
// TODO: refactor, optimize
func (me *MetricsEngine) evaluateThresholds(ignoreEmptySinks bool) (breachedThersholds []string, shouldAbort bool) {
	me.metricsLock.Lock()
	defer me.metricsLock.Unlock()

	t := me.es.GetCurrentTestRunDuration()

	me.logger.Debugf("Running thresholds on %d metrics...", len(me.metricsWithThresholds))
	for _, m := range me.metricsWithThresholds {
		sink := me.sinks[m]
		// If either the metric has no thresholds defined, or its sinks
		// are empty, let's ignore its thresholds execution at this point.
		if len(m.Thresholds.Thresholds) == 0 || (ignoreEmptySinks && sink.IsEmpty()) {
			continue
		}
		m.Tainted = null.BoolFrom(false)

		succ, err := m.Thresholds.Run(sink, t)
		if err != nil {
			me.logger.WithField("metric_name", m.Name).WithError(err).Error("Threshold error")
			continue
		}
		if succ {
			continue // threshold passed
		}
		breachedThersholds = append(breachedThersholds, m.Name)
		m.Tainted = null.BoolFrom(true)
		if m.Thresholds.Abort {
			shouldAbort = true
		}
	}
	if len(breachedThersholds) > 0 {
		sort.Strings(breachedThersholds)
	}
	me.logger.Debugf("Thresholds on %d metrics breached: %v", len(breachedThersholds), breachedThersholds)
	atomic.StoreUint32(&me.breachedThresholdsCount, uint32(len(breachedThersholds)))
	return breachedThersholds, shouldAbort
}

// GetMetricsWithBreachedThresholdsCount returns the number of metrics for which
// the thresholds were breached (failed) during the last processing phase. This
// API is safe to use concurrently.
func (me *MetricsEngine) GetMetricsWithBreachedThresholdsCount() uint32 {
	return atomic.LoadUint32(&me.breachedThresholdsCount)
}

// TODO: remove the metric, it used for passing the submetric
// but it is a mess because it will be different frome the Sample's metric
func (me *MetricsEngine) AddSample(m *metrics.Metric, s metrics.Sample) {
	sink, ok := me.sinks[m]
	if !ok {
		sink = metrics.NewSinkWithMetric(m).Sink
		me.sinks[m] = sink
	}

	sink.Add(s)
	me.markObserved(s.Metric)
}

func (me *MetricsEngine) ObservedMetrics() map[string]metrics.SinkWithMetric {
	me.metricsLock.Lock()
	defer me.metricsLock.Unlock()

	swm := make(map[string]metrics.SinkWithMetric) //], 0, len(me.observedMetrics))
	for id, om := range me.observedMetrics {
		swm[id] = metrics.SinkWithMetric{
			Sink:   me.sinks[om],
			Metric: om,
		}
	}
	return swm
}

func (me *MetricsEngine) ObservedMetricByID(id string) (metrics.SinkWithMetric, bool) {
	me.metricsLock.Lock()
	defer me.metricsLock.Unlock()
	m, ok := me.observedMetrics[id]
	if !ok {
		return metrics.SinkWithMetric{}, false
	}
	return metrics.SinkWithMetric{
		Sink: me.sinks[m], Metric: m,
	}, true
}
