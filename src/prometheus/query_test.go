package prometheus

import (
	"testing"

	"io"
	"net/http"
	"net/http/httptest"

	"github.com/stretchr/testify/assert"

	"os"

	"github.com/golang/protobuf/proto"

	prometheus "github.com/prometheus/client_model/go"
)

type ksm struct {
	nodeIP string
}

func (c *ksm) Do(method, path string) (*http.Response, error) {
	f, err := os.Open("protobuf/metrics")
	if err != nil {
		return nil, err
	}
	defer f.Close() // nolint: errcheck

	req := httptest.NewRequest("GET", "http://example.com/foo", nil)
	w := httptest.NewRecorder()
	mockResponseHandler(f)(w, req)
	return w.Result(), nil

}
func (c *ksm) NodeIP() string {
	return c.nodeIP
}

func TestDo(t *testing.T) {
	// TODO create or use an agnostic test sample.
	var c = ksm{
		nodeIP: "1.2.3.4",
	}

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

	m, err := Do(&c, queries)
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

func TestQueryMatch_CustomName(t *testing.T) {
	q := Query{
		CustomName: "custom_name",
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
		Name: q.CustomName,
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
