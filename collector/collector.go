package collector

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"sync"
	"time"
)

type DescFunc func() *prometheus.Desc

type Collector interface {
	Update(ch chan<- prometheus.Metric) error
}

var (
	scrapeDurationDesc = BuildDesc(buildFqName("scrape", "collector_duration_seconds"),
		"exporter: Duration of a collector scrape.",
		[]string{"collector"},
	)

	scrapeSuccessDesc = BuildDesc(buildFqName("scrape", "collector_success"),
		"exporter: Whether a collector succeeded.",
		[]string{"collector"},
	)
)

// ErrNoData indicates the collector found no data to collect, but had no other error.
var ErrNoData = errors.New("collector returned no data")

func IsNoDataError(err error) bool {
	return err == ErrNoData
}

type SCollector struct {
	Collectors map[string]Collector
}

func NewSCollector() SCollector {
	collector := map[string]Collector{
		"test": TestCollector{},
	}
	return SCollector{
		Collectors: collector,
	}
}

func (m SCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- scrapeDurationDesc()
	ch <- scrapeSuccessDesc()
}

func (m SCollector) Collect(ch chan<- prometheus.Metric) {
	wg := sync.WaitGroup{}
	wg.Add(len(m.Collectors))
	for name, c := range m.Collectors {
		go func(name string, c Collector) {
			execute(name, c, ch)
			wg.Done()
		}(name, c)
	}
	wg.Wait()
}

func execute(name string, c Collector, ch chan<- prometheus.Metric) {
	begin := time.Now()
	err := c.Update(ch)
	duration := time.Since(begin)
	var success = Failed
	if err != nil {
		if IsNoDataError(err) {
			log.Debugln("msg", "collector returned no data", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		} else {
			log.Errorln("msg", "collector failed", "name", name, "duration_seconds", duration.Seconds(), "err", err)
		}

	} else {
		log.Debugln("msg", "collector succeeded", "name", name, "duration_seconds", duration.Seconds())
		success = Success
	}

	ch <- buildGaugeMetric(scrapeDurationDesc, duration.Seconds(), name)
	ch <- buildGaugeMetric(scrapeSuccessDesc, success, name)
}
