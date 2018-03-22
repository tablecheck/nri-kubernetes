package kubelet

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"fmt"

	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric"
	"github.com/newrelic/infra-integrations-beta/integrations/kubernetes/src/kubelet/metric/testdata"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type testClient struct {
	handler http.HandlerFunc
}

func (c *testClient) Do(method, path string) (*http.Response, error) {
	s := httptest.NewServer(c.handler)

	req, _ := http.NewRequest(method, fmt.Sprintf("%s%s", s.URL, path), nil)

	return s.Client().Do(req)
}

func (c *testClient) NodeIP() string {
	// nothing to do
	return ""
}

func rawGroupsHandlerFunc(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case metric.KubeletPodsPath:
		f, err := os.Open("metric/testdata/kubelet_pods_payload.json") // TODO move fetch and testdata to just kubelet package.
		if err != nil {
			panic(err)
		}

		defer f.Close() // nolint: errcheck

		io.Copy(w, f)
	case metric.StatsSummaryPath:
		f, err := os.Open("metric/testdata/kubelet_stats_summary_payload.json") // TODO move fetch and testdata to just kubelet package.
		if err != nil {
			panic(err)
		}

		defer f.Close() // nolint: errcheck

		io.Copy(w, f)
	}
}

func TestGroup(t *testing.T) {
	c := testClient{
		handler: rawGroupsHandlerFunc,
	}

	grouper := NewGrouper(&c, logrus.StandardLogger(), metric.PodsFetchFunc(&c))
	r, errGroup := grouper.Group(nil) // TODO pass definition once this feature is developed. See IHOST-332.
	assert.Nil(t, errGroup)

	assert.Equal(t, testdata.ExpectedGroupData, r)
}
