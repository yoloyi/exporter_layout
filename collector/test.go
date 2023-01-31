package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpMonitorDesc = BuildDesc(buildFqName("test", "test_metrics"),
		"test_metrics",
		[]string{"name"},
	)
)

type TestCollector struct {
}

func (n TestCollector) Update(ch chan<- prometheus.Metric) error {
	ch <- buildGaugeMetric(httpMonitorDesc, 1, "name:test")
	return nil
}
