package prometheus

import (
	"io"
	"log"
	"net/http"

	"time"

	"github.com/matttproud/golang_protobuf_extensions/pbutil"
	prometheus "github.com/prometheus/client_model/go"
)

const acceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

//TODO: See https://github.com/prometheus/prom2json/blob/master/prom2json.go#L171 for how to connect, how to parse plain text, etc

// Query represents the query object. It will run against Prometheus metrics.
type Query struct {
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

	metricFamily = MetricFamily{
		Name:    promMetricFamily.GetName(),
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

// Do is the main entry point. It runs queries agains the Prometheus metrics provided by the endpoint.
func Do(endpoint string, queries []Query) ([]MetricFamily, error) {

	r, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Accept", acceptHeader)

	c := http.DefaultClient
	c.Timeout = 5 * time.Second

	resp, err := c.Do(r)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	metrics := make([]MetricFamily, 0)
	for {
		promMetricFamily := prometheus.MetricFamily{}
		_, err = pbutil.ReadDelimited(resp.Body, &promMetricFamily)

		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatal(err)
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
