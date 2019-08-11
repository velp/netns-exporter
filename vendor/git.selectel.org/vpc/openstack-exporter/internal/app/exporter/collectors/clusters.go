package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/magnum"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/gophercloud/gophercloud"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// COEClusterCollector displays information about each OpenStack Magnum cluster.
type COEClusterCollector struct {
	// COEClusterStatus contains OpenStack COE cluster representation with
	// status code used as value.
	COEClusterStatus *prometheus.GaugeVec

	// COEClusterCount contains OpenStack clusters count in a single
	// project.
	COEClusterCount *prometheus.GaugeVec
}

// NewCOEClusterCollector creates an instance of the COEClusterCollector and
// instantiates the individual metrics that show information about the
// cluster.
func NewCOEClusterCollector() *COEClusterCollector {
	clusterLabels := []string{
		idLabel,
		projectIDLabel,
		regionLabel,
	}
	clusterCountLabels := []string{
		projectIDLabel,
		regionLabel,
		accountLabel,
	}

	return &COEClusterCollector{
		COEClusterStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "coe_cluster_status",
				Help:      "COE cluster status",
			},
			clusterLabels,
		),
		COEClusterCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "coe_cluster_count",
				Help:      "COE cluster count",
			},
			clusterCountLabels,
		),
	}
}

func (c *COEClusterCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{c.COEClusterStatus, c.COEClusterCount}
}

func (c *COEClusterCollector) collectCOEClusterStatuses() error {
	clients, err := magnumClients()
	if err != nil {
		return err
	}

	clusters := polling.GetCOEClusters(clients)
	for _, cluster := range clusters {
		if cluster == nil {
			continue
		}
		clusterLabels := prometheus.Labels{
			idLabel:        cluster.ID,
			projectIDLabel: cluster.ProjectID,
			regionLabel:    cluster.Region,
		}
		c.COEClusterStatus.With(clusterLabels).Set(float64(cluster.StatusCode))
	}

	return nil
}

func (c *COEClusterCollector) collectCOEClusterCount() error {
	clients, err := magnumClients()
	if err != nil {
		return err
	}

	accountsClusters, err := polling.GetAccountsCOEClusters(clients)
	if err != nil {
		return errors.Wrap(err, "unable to calculate COE clusters count")
	}
	for _, accountClusters := range accountsClusters {
		if accountClusters == nil {
			continue
		}
		countLabels := prometheus.Labels{
			projectIDLabel: accountClusters.ProjectID,
			regionLabel:    accountClusters.Region,
			accountLabel:   accountClusters.AccountName,
		}
		c.COEClusterCount.With(countLabels).Set(float64(accountClusters.Count))
	}

	return nil
}

// Describe sends the descriptors of each COEClusterCollector related metrics
// to the provided Prometheus channel.
func (c *COEClusterCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *COEClusterCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.COEClusterStatus.Reset()
	c.COEClusterCount.Reset()

	if err := c.collectCOEClusterStatuses(); err != nil {
		log.Errorf("failed to collect COE cluster statuses: %v", err)
	}
	if err := c.collectCOEClusterCount(); err != nil {
		log.Errorf("failed to collect COE clusters counts: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}

func magnumClients() (map[string]gophercloud.ServiceClient, error) {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		return nil, err
	}

	// Initialize Magnum client for every region.
	clients, err := magnum.NewMagnumV1Clients(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Magnum.RequestTimeout),
	})
	if err != nil {
		return nil, err
	}

	return clients, nil
}
