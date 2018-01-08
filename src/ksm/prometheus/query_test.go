package prometheus

import (
	"testing"

	"io"
	"net/http"

	"github.com/stretchr/testify/assert"

	"net/http/httptest"
	"os"

	"github.com/golang/protobuf/proto"
	prometheus "github.com/prometheus/client_model/go"
)

func TestFoo(t *testing.T) {
	// TODO create or use an agnostic test sample.
	f, err := os.Open("protobuf/metrics")
	assert.NoError(t, err)

	defer f.Close() // nolint: errcheck

	s := httptest.NewServer(mockResponseHandler(f))
	c := http.DefaultClient

	queryMetricName := "kube_pod_status_phase"
	queryLabels := Labels{
		"namespace": "default",
		"pod":       "smoya-ghtop-6878dbdcc4-x2c5f",
	}

	queries := []Query{
		{
			MetricName: queryMetricName,
			Labels:     queryLabels,
			Value:      GaugeValue(1),
		},
	}

	expectedLabels := queryLabels
	expectedLabels["phase"] = "Running"
	expectedMetrics := []MetricFamily{
		{
			Name: queryMetricName,
			Type: "GAUGE",
			Metrics: []Metric{
				{
					Labels: expectedLabels,
					Value:  GaugeValue(1),
				},
			},
		},
	}

	m, err := Do(s.URL, queries, c)
	assert.NoError(t, err)

	assert.Equal(t, expectedMetrics, m)
}

func TestLabelsAreIn(t *testing.T) {
	expectedLabels := Labels{
		"namespace": "default",
		"pod":       "nr-123456789",
	}

	l := []*prometheus.LabelPair{
		{
			Name:  proto.String("condition"),
			Value: proto.String("false"),
		},
		{
			Name:  proto.String("namespace"),
			Value: proto.String("default"),
		},
		{
			Name:  proto.String("pod"),
			Value: proto.String("nr-123456789"),
		},
	}

	assert.True(t, expectedLabels.AreIn(l))
}

func TestQueryMatch(t *testing.T) {
	q := Query{
		MetricName: "kube_pod_status_phase",
		Labels: Labels{
			"namespace": "default",
			"pod":       "nr-123456789",
		},
		Value: GaugeValue(1),
	}

	metrictType := prometheus.MetricType_GAUGE
	r := prometheus.MetricFamily{
		Name: proto.String(q.MetricName),
		Type: &metrictType,
		Metric: []*prometheus.Metric{
			{
				Gauge: &prometheus.Gauge{
					Value: proto.Float64(1),
				},
				Label: []*prometheus.LabelPair{
					{
						Name:  proto.String("namespace"),
						Value: proto.String("default"),
					},
					{
						Name:  proto.String("pod"),
						Value: proto.String("nr-123456789"),
					},
				},
			},
			{
				Gauge: &prometheus.Gauge{
					Value: proto.Float64(0),
				},
				Label: []*prometheus.LabelPair{
					{
						Name:  proto.String("namespace"),
						Value: proto.String("default"),
					},
					{
						Name:  proto.String("pod"),
						Value: proto.String("nr-123456789"),
					},
				},
			},
		},
	}

	expectedMetrics := MetricFamily{
		Name: q.MetricName,
		Type: "GAUGE",
		Metrics: []Metric{
			{
				Labels: q.Labels,
				Value:  GaugeValue(1),
			},
		},
	}

	assert.Equal(t, expectedMetrics, q.Execute(&r))
}

func mockResponseHandler(mockResponse io.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, mockResponse) // nolint: errcheck
	}
}
