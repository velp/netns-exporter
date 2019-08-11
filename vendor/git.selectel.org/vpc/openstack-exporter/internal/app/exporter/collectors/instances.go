package collectors

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/nova"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/go-redis/redis"
	"github.com/gophercloud/gophercloud"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	errOnlyCacheInstancesStatusesFmt = "will only collect instances statuses from cache: %s"
	errOnlyCacheInstancesCountFmt    = "will only collect instances count from cache: %s"
)

// InstanceCollector displays information about each OpenStack instance.
type InstanceCollector struct {
	// InstanceStatus contains OpenStack instance representation with status code
	// used as value.
	InstanceStatus *prometheus.GaugeVec

	// InstanceCount contains OpenStack instances count in a single
	// project.
	InstanceCount *prometheus.GaugeVec

	redisClient *redis.Client
}

// NewInstanceCollector creates an instance of the InstanceCollector and
// instantiates the individual metrics that show information about the instance.
func NewInstanceCollector(redisClient *redis.Client) *InstanceCollector {
	instanceLabels := []string{
		idLabel,
		projectIDLabel,
		regionLabel,
		zoneLabel,
	}
	instanceCountLabels := []string{
		projectIDLabel,
		regionLabel,
		accountLabel,
	}

	return &InstanceCollector{
		InstanceStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instance_status",
				Help:      "Instance status",
			},
			instanceLabels,
		),
		InstanceCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "instance_count",
				Help:      "Instance count",
			},
			instanceCountLabels,
		),
		redisClient: redisClient,
	}
}

func (c *InstanceCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{c.InstanceStatus, c.InstanceCount}
}

func (c *InstanceCollector) collectInstanceStatuses() {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		log.Errorf(errOnlyCacheInstancesStatusesFmt, err)
	}

	var cloudClients map[string]gophercloud.ServiceClient

	if provider != nil {
		// Initialize Nova client for every region.
		cloudClients, err = nova.NewNovaV2Clients(&common.NewClientOpts{
			Provider:     provider,
			EndpointType: config.Config.OpenStack.EndpointType,
			Timeout:      time.Second * time.Duration(config.Config.OpenStack.Nova.RequestTimeout),
		})
		if err != nil {
			log.Errorf(errOnlyCacheInstancesStatusesFmt, err)
		}
	}

	allInstances := polling.GetInstances(&polling.GetObjectsOpts{
		CloudClients: cloudClients,
		RedisClient:  c.redisClient,
	})

	for _, instance := range allInstances {
		if instance == nil {
			continue
		}
		instanceLabels := prometheus.Labels{
			idLabel:        instance.ID,
			projectIDLabel: instance.ProjectID,
			regionLabel:    instance.Region,
			zoneLabel:      instance.AvailabilityZone,
		}
		c.InstanceStatus.With(instanceLabels).Set(float64(instance.StatusCode))
	}
}

func (c *InstanceCollector) collectInstanceCount() {
	// Authenticate in OpenStack.
	provider, err := polling.NewOpenStackProvider()
	if err != nil {
		log.Errorf(errOnlyCacheInstancesCountFmt, err)
	}

	var cloudClients map[string]gophercloud.ServiceClient

	if provider != nil {
		// Initialize Nova client for every region.
		cloudClients, err = nova.NewNovaV2Clients(&common.NewClientOpts{
			Provider:     provider,
			EndpointType: config.Config.OpenStack.EndpointType,
			Timeout:      time.Second * time.Duration(config.Config.OpenStack.Nova.RequestTimeout),
		})
		if err != nil {
			log.Errorf(errOnlyCacheInstancesCountFmt, err)
		}
	}

	allInstances := polling.GetAccountsInstances(&polling.GetObjectsOpts{
		CloudClients: cloudClients,
		RedisClient:  c.redisClient,
	})

	for _, accountInstances := range allInstances {
		if accountInstances == nil {
			continue
		}
		countLabels := prometheus.Labels{
			projectIDLabel: accountInstances.ProjectID,
			regionLabel:    accountInstances.Region,
			accountLabel:   accountInstances.AccountName,
		}
		c.InstanceCount.With(countLabels).Set(float64(accountInstances.Count))
	}
}

// Describe sends the descriptors of each InstanceCollector related metrics to
// the provided Prometheus channel.
func (c *InstanceCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *InstanceCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.InstanceStatus.Reset()
	c.InstanceCount.Reset()

	// Collect metrics.
	c.collectInstanceStatuses()
	c.collectInstanceCount()

	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
