package collectors

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// LoadBalancerCollector displays information about each OpenStack loadbalancer.
type LoadBalancerCollector struct {
	// LoadBalancerStatus contains OpenStack loadbalancer representation with
	// operating status code used as value.
	LoadBalancerOperatingStatus *prometheus.GaugeVec

	// LoadBalancerProvisionStatus contains OpenStack loadbalancer
	// representation with provisioning status code used as value.
	LoadBalancerProvisioningStatus *prometheus.GaugeVec

	// LoadBalancerCount contains OpenStack loadbalancers count in a single
	// project.
	LoadBalancerCount *prometheus.GaugeVec

	redisClient *redis.Client
}

// NewLoadBalancerCollector creates an instance of the LoadBalancerCollector and
// instantiates the individual metrics that show information about the
// loadbalancer.
func NewLoadBalancerCollector(redisClient *redis.Client) *LoadBalancerCollector {
	loadbalancerLabels := []string{
		idLabel,
		projectIDLabel,
		regionLabel,
	}
	loadbalancerCountLabels := []string{
		projectIDLabel,
		regionLabel,
		accountLabel,
	}

	return &LoadBalancerCollector{
		LoadBalancerOperatingStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "loadbalancer_operating_status",
				Help:      "LoadBalancer operating status",
			},
			loadbalancerLabels,
		),
		LoadBalancerProvisioningStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "loadbalancer_provisioning_status",
				Help:      "LoadBalancer provisioning status",
			},
			loadbalancerLabels,
		),
		LoadBalancerCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "loadbalancer_count",
				Help:      "LoadBalancer count",
			},
			loadbalancerCountLabels,
		),
		redisClient: redisClient,
	}
}

func (c *LoadBalancerCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{
		c.LoadBalancerOperatingStatus,
		c.LoadBalancerProvisioningStatus,
		c.LoadBalancerCount,
	}
}

func (c *LoadBalancerCollector) collectLoadBalancerOperatingStatuses() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	loadbalancer, err := polling.GetLoadBalancers(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to poll loadbalancers data")
	}
	for _, loadbalancer := range loadbalancer {
		if loadbalancer == nil {
			continue
		}
		loadbalancerLabels := prometheus.Labels{
			idLabel:        loadbalancer.ID,
			projectIDLabel: loadbalancer.ProjectID,
			regionLabel:    loadbalancer.Region,
		}
		c.LoadBalancerOperatingStatus.With(loadbalancerLabels).Set(float64(loadbalancer.OperatingStatusCode))
	}

	return nil
}

func (c *LoadBalancerCollector) collectLoadBalancerProvisioningStatuses() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	loadbalancer, err := polling.GetLoadBalancers(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to poll loadbalancers data")
	}
	for _, loadbalancer := range loadbalancer {
		if loadbalancer == nil {
			continue
		}
		loadbalancerLabels := prometheus.Labels{
			idLabel:        loadbalancer.ID,
			projectIDLabel: loadbalancer.ProjectID,
			regionLabel:    loadbalancer.Region,
		}
		c.LoadBalancerProvisioningStatus.With(loadbalancerLabels).Set(float64(loadbalancer.ProvisioningStatusCode))
	}

	return nil
}

func (c *LoadBalancerCollector) collectLoadBalancerCount() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	accountsLoadBalancers, err := polling.GetAccountsLoadBalancers(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to calculate loadbalancers count")
	}
	for _, accountLoadBalancers := range accountsLoadBalancers {
		if accountLoadBalancers == nil {
			continue
		}
		countLabels := prometheus.Labels{
			projectIDLabel: accountLoadBalancers.ProjectID,
			regionLabel:    accountLoadBalancers.Region,
			accountLabel:   accountLoadBalancers.AccountName,
		}
		c.LoadBalancerCount.With(countLabels).Set(float64(accountLoadBalancers.Count))
	}

	return nil
}

// Describe sends the descriptors of each LoadBalancerCollector related metrics
// to the provided Prometheus channel.
func (c *LoadBalancerCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *LoadBalancerCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.LoadBalancerOperatingStatus.Reset()
	c.LoadBalancerProvisioningStatus.Reset()
	c.LoadBalancerCount.Reset()

	if err := c.collectLoadBalancerOperatingStatuses(); err != nil {
		log.Errorf("failed to collect loadbalancers operating statuses: %v", err)
	}
	if err := c.collectLoadBalancerProvisioningStatuses(); err != nil {
		log.Errorf("failed to collect loadbalancers provisioning statuses: %v", err)
	}
	if err := c.collectLoadBalancerCount(); err != nil {
		log.Errorf("failed to collect loadbalancers counts: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
