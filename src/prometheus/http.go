package prometheus

import (
	"fmt"
	"io"
	"net/http"

	"github.com/prometheus/common/expfmt"
)

// ProtobufferAcceptHeader is the required accept header for asking for a protobuffer resource instead of a plain txt.
const ProtobufferAcceptHeader = `application/vnd.google.protobuf;proto=io.prometheus.client.MetricFamily;encoding=delimited;q=0.7,text/plain;version=0.0.4;q=0.3`

// NewRequest returns a new Request given a method, URL, setting the required header for accepting protobuf.
func NewRequest(method, url string) (*http.Request, error) {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("Accept", ProtobufferAcceptHeader)

	return r, nil
}

// TextToProtoHandleFunc is a http.HandlerFunc that serves protobuf metrics from plain txt.
// Useful for testing purposes
func TextToProtoHandleFunc(r io.Reader) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		p := expfmt.TextParser{}
		mf, err := p.TextToMetricFamilies(r)
		if err != nil {
			http.Error(w, fmt.Sprintf("error parsing metric families: %s", err), http.StatusInternalServerError)
			return
		}

		enc := expfmt.NewEncoder(w, expfmt.FmtProtoDelim)
		for _, s := range mf {
			err := enc.Encode(s)
			if err != nil {
				http.Error(w, fmt.Sprintf("error encoding metric families: %s", err), http.StatusInternalServerError)
				return
			}
		}
	}
}
