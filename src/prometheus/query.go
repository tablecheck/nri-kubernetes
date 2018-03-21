package prometheus

import (
	"fmt"
	"io"
	"net/http"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/client"
	prometheus "github.com/prometheus/client_model/go"
)

const (
	metricsPath = "/metrics"
)

//TODO: See https://github.com/prometheus/prom2json/blob/master/prom2json.go#L171 for how to connect, how to parse plain text, etc

// Query represents the query object. It will run against Prometheus metrics.
type Query struct {
	CustomName string
	MetricName string
	Labels     Labels
	Value      Value // TODO Only supported Counter and Gauge
}

// Execute runs the query.
func (q Query) Execute(promMetricFamily *prometheus.MetricFamily) (metricFamily MetricFamily) {
	if promMetricFamily.GetName() != q.MetricName {
		return
	}

	if len(promMetricFamily.Metric) <= 0 {
		// Should not happen
		return
	}
	var matches []Metric
	for _, promMetric := range promMetricFamily.Metric {
		if len(q.Labels) > 0 {
			// Match by labels
			if !q.Labels.AreIn(promMetric.Label) {
				continue
			}
		}

		value := valueFromPrometheus(promMetricFamily.GetType(), promMetric)

		if q.Value != nil {
			// Match by value
			if q.Value.String() != value.String() {
				continue
			}
		}

		m := Metric{
			Labels: labelsFromPrometheus(promMetric.Label),
			Value:  value,
		}

		matches = append(matches, m)
	}

	var name string
	if q.CustomName != "" {
		name = q.CustomName
	} else {
		name = promMetricFamily.GetName()
	}

	metricFamily = MetricFamily{
		Name:    name,
		Type:    promMetricFamily.GetType().String(),
		Metrics: matches,
	}

	return
}

func valueFromPrometheus(metricType prometheus.MetricType, metric *prometheus.Metric) Value {
	switch metricType {
	case prometheus.MetricType_COUNTER:
		return CounterValue(metric.Counter.GetValue())
	case prometheus.MetricType_GAUGE:
		return GaugeValue(metric.Gauge.GetValue())
	case prometheus.MetricType_HISTOGRAM:
		// Not supported yet
		fallthrough
	case prometheus.MetricType_SUMMARY:
		// Not supported yet
		fallthrough
	case prometheus.MetricType_UNTYPED:
		// Not supported yet
		fallthrough
	default:
		return EmptyValue
	}
}

// Do is the main entry point. It runs queries against the Prometheus metrics provided by the endpoint.
func Do(c client.HTTPClient, queries []Query) ([]MetricFamily, error) {
	resp, err := c.Do(http.MethodGet, metricsPath)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close() // nolint: errcheck

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error calling kube-state-metrics endpoint. Got status code: %d", resp.StatusCode)
	}

	metrics := make([]MetricFamily, 0)
	for {
		promMetricFamily := prometheus.MetricFamily{}
		_, err = pbutil.ReadDelimited(resp.Body, &promMetricFamily)

		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		for _, q := range queries {
			f := q.Execute(&promMetricFamily)
			if f.Valid() {
				metrics = append(metrics, q.Execute(&promMetricFamily))

			}
		}
	}

	return metrics, nil
}

func labelsFromPrometheus(pairs []*prometheus.LabelPair) Labels {
	labels := make(Labels)
	for _, p := range pairs {
		labels[p.GetName()] = p.GetValue()
	}

	return labels
}
