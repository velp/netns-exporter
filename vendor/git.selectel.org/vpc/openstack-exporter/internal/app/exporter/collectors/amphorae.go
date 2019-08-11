package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/octavia"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/prometheus/client_golang/prometheus"
)

const loadbalancerIDLabel = "loadbalancer_id"

// AmphoraCollector displays information about each OpenStack amphora.
type AmphoraCollector struct {
	// AmphoraStatus contains OpenStack amphora representation with
	// status code used as value.
	AmphoraStatus *prometheus.GaugeVec
}

// NewAmphoraCollector creates an instance of the AmphoraCollector and
// instantiates the individual metrics that show information about the
// amphora.
func NewAmphoraCollector() *AmphoraCollector {
	amphoraLabels := []string{
		idLabel,
		loadbalancerIDLabel,
		regionLabel,
	}

	return &AmphoraCollector{
		AmphoraStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "amphora_status",
				Help:      "Amphora status",
			},
			amphoraLabels,
		),
	}
}

func (c *AmphoraCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{c.AmphoraStatus}
}

func (c *AmphoraCollector) collectAmphoraStatuses() error {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		return err
	}

	// Initialize Octavia client for every region.
	clients, err := octavia.NewOctaviaV2Clients(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Octavia.RequestTimeout),
	})
	if err != nil {
		return err
	}

	amphorae := polling.GetAmphorae(clients)
	for _, amphora := range amphorae {
		if amphora == nil {
			continue
		}
		amphoraLabels := prometheus.Labels{
			idLabel:             amphora.ID,
			loadbalancerIDLabel: amphora.LoadbalancerID,
			regionLabel:         amphora.Region,
		}
		c.AmphoraStatus.With(amphoraLabels).Set(float64(amphora.StatusCode))
	}

	return nil
}

// Describe sends the descriptors of each AmphoraCollector related metrics
// to the provided Prometheus channel.
func (c *AmphoraCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *AmphoraCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.AmphoraStatus.Reset()

	if err := c.collectAmphoraStatuses(); err != nil {
		log.Errorf("failed to collect amphorae statuses: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
