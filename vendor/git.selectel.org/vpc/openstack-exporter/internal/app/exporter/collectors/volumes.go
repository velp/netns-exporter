package collectors

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// Volume-specific metrics labels.
	volumeTypeLabel = "volume_type"

	// Common volume reading error.
	readVolumeErr = "unable to read volumes from cache"
)

// VolumeCollector displays information about each OpenStack volume.
type VolumeCollector struct {
	// VolumeStatus contains OpenStack volume representation with status code
	// used as value.
	VolumeStatus *prometheus.GaugeVec

	// VolumeSize contains OpenStack volume representation with its size used as
	// value.
	VolumeSize *prometheus.GaugeVec

	// VolumeCount contains OpenStack volumes count in a single project.
	VolumeCount *prometheus.GaugeVec

	redisClient *redis.Client
}

// NewVolumeCollector creates an instance of the VolumeCollector and
// instantiates the individual metrics that show information about the volume.
func NewVolumeCollector(redisClient *redis.Client) *VolumeCollector {
	volumeLabels := []string{
		idLabel,
		projectIDLabel,
		volumeTypeLabel,
		regionLabel,
		zoneLabel,
	}
	volumeCountLabels := []string{
		projectIDLabel,
		regionLabel,
		accountLabel,
	}

	return &VolumeCollector{
		VolumeStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "volume_status",
				Help:      "Volume status",
			},
			volumeLabels,
		),
		VolumeSize: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "volume_size",
				Help:      "Volume size",
			},
			volumeLabels,
		),
		VolumeCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "volume_count",
				Help:      "Volume count",
			},
			volumeCountLabels,
		),
		redisClient: redisClient,
	}
}

func (c *VolumeCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{c.VolumeStatus, c.VolumeSize, c.VolumeCount}
}

func (c *VolumeCollector) collectVolumeStatuses() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	volumes, err := polling.GetVolumes(pollingOpts)
	if err != nil {
		return errors.Wrap(err, readVolumeErr)
	}
	for _, volume := range volumes {
		if volume == nil {
			continue
		}
		volumeLabels := prometheus.Labels{
			idLabel:         volume.ID,
			projectIDLabel:  volume.ProjectID,
			volumeTypeLabel: volume.VolumeType,
			regionLabel:     volume.Region,
			zoneLabel:       volume.AvailabilityZone,
		}
		c.VolumeStatus.With(volumeLabels).Set(float64(volume.StatusCode))
	}

	return nil
}

func (c *VolumeCollector) collectVolumeSizes() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	volumes, err := polling.GetVolumes(pollingOpts)
	if err != nil {
		return errors.Wrap(err, readVolumeErr)
	}
	for _, volume := range volumes {
		if volume == nil {
			continue
		}
		volumeLabels := prometheus.Labels{
			idLabel:         volume.ID,
			projectIDLabel:  volume.ProjectID,
			volumeTypeLabel: volume.VolumeType,
			regionLabel:     volume.Region,
			zoneLabel:       volume.AvailabilityZone,
		}
		c.VolumeSize.With(volumeLabels).Set(float64(volume.Size))
	}

	return nil
}

func (c *VolumeCollector) collectVolumeCount() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	accountsVolumes, err := polling.GetAccountsVolumes(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to calculate volumes count")
	}
	for _, accountVolumes := range accountsVolumes {
		if accountVolumes == nil {
			continue
		}
		countLabels := prometheus.Labels{
			projectIDLabel: accountVolumes.ProjectID,
			regionLabel:    accountVolumes.Region,
			accountLabel:   accountVolumes.AccountName,
		}
		c.VolumeCount.With(countLabels).Set(float64(accountVolumes.Count))
	}

	return nil
}

// Describe sends the descriptors of each VolumeCollector related metrics to
// the provided Prometheus channel.
func (c *VolumeCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *VolumeCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.VolumeStatus.Reset()
	c.VolumeSize.Reset()
	c.VolumeCount.Reset()

	if err := c.collectVolumeStatuses(); err != nil {
		log.Errorf("failed to collect volumes statuses: %v", err)
	}
	if err := c.collectVolumeSizes(); err != nil {
		log.Errorf("failed to collect volumes sizes: %v", err)
	}
	if err := c.collectVolumeCount(); err != nil {
		log.Errorf("failed to collect volumes counts: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
