package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/keystone"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/octavia"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/go-redis/redis"
	"github.com/prometheus/client_golang/prometheus"
)

// OctaviaInstanceCollector displays information about each OpenStack Octavia instance.
type OctaviaInstanceCollector struct {
	// OctaviaOrphanedInstance contains Openstack Octavia orphaned instance representation.
	OctaviaOrphanedInstance *prometheus.GaugeVec

	redisClient *redis.Client
}

// NewOctaviaInstanceCollector creates an instance of the OctaviaInstanceCollector and instantiates
// the individual metrics that show information about the Octavia instance.
func NewOctaviaInstanceCollector(redisClient *redis.Client) *OctaviaInstanceCollector {
	octaviaOrphanedInstanceLabels := []string{
		idLabel,
		regionLabel,
	}
	return &OctaviaInstanceCollector{
		OctaviaOrphanedInstance: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "octavia_orphaned_instance",
				Help:      "Octavia orphaned instance",
			},
			octaviaOrphanedInstanceLabels,
		),
		redisClient: redisClient,
	}
}

func (c *OctaviaInstanceCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		c.OctaviaOrphanedInstance,
	}
}

func (c *OctaviaInstanceCollector) collectOctaviaOrphanedInstances() error {
	// Authenticate in Openstack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		return err
	}

	// Initialize Octavia V2 service client for every region.
	octaviaClients, err := octavia.NewOctaviaV2Clients(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Octavia.RequestTimeout),
	})
	if err != nil {
		return err
	}

	// Initialize Keystone V3 service client.
	keystoneClient, err := keystone.NewKeystoneV3Client(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Keystone.RequestTimeout),
	})
	if err != nil {
		return err
	}

	allOrphanedInstances := polling.GetOctaviaOrphanedInstances(&polling.GetOctaviaInstancesOpts{
		OctaviaClients: octaviaClients,
		KeystoneClient: keystoneClient,
		RedisClient:    c.redisClient,
	})

	for instanceID, region := range allOrphanedInstances {
		labels := prometheus.Labels{
			idLabel:     instanceID,
			regionLabel: region,
		}
		c.OctaviaOrphanedInstance.With(labels).Set(float64(1))
	}

	return nil
}

// Describe sends the descriptors of each OctaviaInstanceCollector related metrics
// to the provided Prometheus channel.
func (c *OctaviaInstanceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *OctaviaInstanceCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.OctaviaOrphanedInstance.Reset()

	if err := c.collectOctaviaOrphanedInstances(); err != nil {
		log.Errorf("failed to collect octavia instances: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
