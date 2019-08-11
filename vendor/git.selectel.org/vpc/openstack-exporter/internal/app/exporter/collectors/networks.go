package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/neutron"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/prometheus/client_golang/prometheus"
)

// NetworkCollector displays information about OpenStack networks.
type NetworkCollector struct {
	// ExternalNetworksTotalIPs total IPs information for each external network.
	ExternalNetworksTotalIPs *prometheus.GaugeVec

	// ExternalNetworksUsedIPs used IPs information for each external network.
	ExternalNetworksUsedIPs *prometheus.GaugeVec

	// ExternalNetworksAvailableIPs contains available IPs information for each external network.
	ExternalNetworksAvailableIPs *prometheus.GaugeVec
}

// NewNetworkCollector creates an instance of the NetworkCollector and
// instantiates the individual metrics that show information about the network.
func NewNetworkCollector() *NetworkCollector {
	externalNetworkIPsLabels := []string{
		idLabel,
		regionLabel,
	}

	return &NetworkCollector{
		ExternalNetworksTotalIPs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "external_net_total_ips",
				Help:      "Total IPs in external network",
			},
			externalNetworkIPsLabels,
		),
		ExternalNetworksUsedIPs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "external_net_used_ips",
				Help:      "Used IPs in external network",
			},
			externalNetworkIPsLabels,
		),
		ExternalNetworksAvailableIPs: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "external_net_available_ips",
				Help:      "Available IPs in external network",
			},
			externalNetworkIPsLabels,
		),
	}
}

func (c *NetworkCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		c.ExternalNetworksTotalIPs,
		c.ExternalNetworksUsedIPs,
		c.ExternalNetworksAvailableIPs,
	}
}

func (c *NetworkCollector) collectExternalIPs() error {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		return err
	}

	// Initialize Neutron client for every region.
	clients, err := neutron.NewNeutronV2Clients(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Neutron.RequestTimeout),
	})
	if err != nil {
		return err
	}

	// Get IP availabilities and populate total and used IP metrics for
	// external networks.
	for _, networkIPs := range polling.GetExternalNetworksIPs(clients) {
		if networkIPs == nil {
			continue
		}
		labels := prometheus.Labels{
			idLabel:     networkIPs.ID,
			regionLabel: networkIPs.Region,
		}
		c.ExternalNetworksTotalIPs.With(labels).Set(networkIPs.TotalIPs)
		c.ExternalNetworksUsedIPs.With(labels).Set(networkIPs.UsedIPs)
		c.ExternalNetworksAvailableIPs.With(labels).Set(networkIPs.AvailableIPs)
	}

	return nil
}

// Describe sends the descriptors of each NetworkCollector related metrics to
// the provided Prometheus channel.
func (c *NetworkCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *NetworkCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.ExternalNetworksTotalIPs.Reset()
	c.ExternalNetworksUsedIPs.Reset()
	c.ExternalNetworksAvailableIPs.Reset()

	if err := c.collectExternalIPs(); err != nil {
		log.Errorf("failed to collect networks metrics: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
