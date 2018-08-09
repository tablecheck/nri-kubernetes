package timer

import (
	"time"

	"github.com/newrelic/infra-integrations-sdk/log"
)

// Track measures time, which elapsed since provided start time
func Track(start time.Time, name string, logger log.Logger) {
	elapsed := time.Since(start)
	logger.Debugf("%s took %s", name, elapsed)
}
