package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/gnocchi"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/prometheus/client_golang/prometheus"
)

// GnocchiStatusCollector displays information about Gnocchi status.
type GnocchiStatusCollector struct {
	// MetricdProcessors contains total running Gnocchi metricd processors.
	MetricdProcessors *prometheus.GaugeVec

	// MetricdMeasures contains total Gnocchi measures to process.
	MetricdMeasures *prometheus.GaugeVec

	// MetricdMetrics contains total Gnocchi metrics having measures to process.
	MetricdMetrics *prometheus.GaugeVec
}

// NewGnocchiStatusCollector creates an instance of the GnocchiStatusCollector and
// instantiates the individual metrics that show information about Gnocchi status.
func NewGnocchiStatusCollector() *GnocchiStatusCollector {
	metricdLabels := []string{
		regionLabel,
	}

	return &GnocchiStatusCollector{
		MetricdProcessors: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "metricd_processors",
				Help:      "Metricd processors",
			},
			metricdLabels,
		),
		MetricdMeasures: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "metricd_measures",
				Help:      "Measures to proccess",
			},
			metricdLabels,
		),
		MetricdMetrics: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "metricd_metrics",
				Help:      "Metrics to proccess",
			},
			metricdLabels,
		),
	}
}

func (c *GnocchiStatusCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		c.MetricdProcessors,
		c.MetricdMeasures,
		c.MetricdMetrics,
	}
}

func (c *GnocchiStatusCollector) collectGnocchiStatuses() error {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		return err
	}

	// Initialize Gnocchi client for every region.
	clients, err := gnocchi.NewGnocchiV1Clients(&gnocchi.ClientOpts{
		Provider:     provider,
		EndpointType: config.Config.Gnocchi.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.Gnocchi.RequestTimeout),
	})
	if err != nil {
		return err
	}

	// Get Gnocchi statuses and populate metrics for Gnocchi measures backlog.
	for _, gnocchiStatus := range polling.GetGnocchiStatuses(clients) {
		if gnocchiStatus == nil {
			continue
		}
		labels := prometheus.Labels{
			regionLabel: gnocchiStatus.Region,
		}
		c.MetricdProcessors.With(labels).Set(float64(gnocchiStatus.Processors))
		c.MetricdMeasures.With(labels).Set(float64(gnocchiStatus.Measures))
		c.MetricdMetrics.With(labels).Set(float64(gnocchiStatus.Metrics))
	}

	return nil
}

// Describe sends the descriptors of each GnocchiStatusCollector related metrics to
// the provided Prometheus channel.
func (c *GnocchiStatusCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *GnocchiStatusCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.MetricdProcessors.Reset()
	c.MetricdMeasures.Reset()
	c.MetricdMetrics.Reset()

	if err := c.collectGnocchiStatuses(); err != nil {
		log.Errorf("failed to collect gnocchi status metrics: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
