package prometheus

import (
	"net/http"
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
