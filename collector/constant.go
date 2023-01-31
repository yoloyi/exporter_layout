package collector

import (
	"github.com/prometheus/client_golang/prometheus"
)

const namespace = ""

const (
	Success float64 = 1
	Failed  float64 = 0
)

const ()

func BuildDesc(fqName, help string, variableLabels []string) DescFunc {
	return func() *prometheus.Desc {
		return prometheus.NewDesc(
			fqName,
			help,
			variableLabels,
			nil,
		)
	}
}

func buildFqName(subsystem, name string) string {
	return prometheus.BuildFQName(namespace, subsystem, name)
}

func buildGaugeMetric(desc DescFunc, value float64, labelValues ...string) prometheus.Metric {
	return prometheus.MustNewConstMetric(desc(), prometheus.GaugeValue, value, labelValues...)
}
