package collectors

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

// SnapshotCollector displays information about each OpenStack volume.
type SnapshotCollector struct {
	// SnapshotStatus contains OpenStack volume snapshot representation with
	// status code used as value.
	SnapshotStatus *prometheus.GaugeVec

	// SnapshotCount contains OpenStack snapshots count in a single project.
	SnapshotCount *prometheus.GaugeVec

	redisClient *redis.Client
}

// NewSnapshotCollector creates an instance of the SnapshotCollector and
// instantiates the individual metrics that show information about the snapshot.
func NewSnapshotCollector(redisClient *redis.Client) *SnapshotCollector {
	snapshotLabels := []string{
		idLabel,
		projectIDLabel,
		regionLabel,
	}
	snapshotCountLabels := []string{
		projectIDLabel,
		regionLabel,
		accountLabel,
	}

	return &SnapshotCollector{
		SnapshotStatus: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "snapshot_status",
				Help:      "Snapshot status",
			},
			snapshotLabels,
		),
		SnapshotCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Namespace: namespace,
				Name:      "snapshot_count",
				Help:      "Snapshot count",
			},
			snapshotCountLabels,
		),
		redisClient: redisClient,
	}
}

func (c *SnapshotCollector) collectorList() []prometheus.Collector {
	return []prometheus.Collector{c.SnapshotStatus, c.SnapshotCount}
}

func (c *SnapshotCollector) collectSnaphotStatuses() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	snapshot, err := polling.GetSnapshots(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to poll snapshots data")
	}
	for _, snapshot := range snapshot {
		if snapshot == nil {
			continue
		}
		snapshotLabels := prometheus.Labels{
			idLabel:        snapshot.ID,
			projectIDLabel: snapshot.ProjectID,
			regionLabel:    snapshot.Region,
		}
		c.SnapshotStatus.With(snapshotLabels).Set(float64(snapshot.StatusCode))
	}

	return nil
}

func (c *SnapshotCollector) collectSnaphotCount() error {
	pollingOpts := &polling.GetObjectsOpts{RedisClient: c.redisClient}
	accountsSnapshots, err := polling.GetAccountsSnapshots(pollingOpts)
	if err != nil {
		return errors.Wrap(err, "unable to calculate snapshots count")
	}
	for _, accountSnapshots := range accountsSnapshots {
		if accountSnapshots == nil {
			continue
		}
		countLabels := prometheus.Labels{
			projectIDLabel: accountSnapshots.ProjectID,
			regionLabel:    accountSnapshots.Region,
			accountLabel:   accountSnapshots.AccountName,
		}
		c.SnapshotCount.With(countLabels).Set(float64(accountSnapshots.Count))
	}

	return nil
}

// Describe sends the descriptors of each SnapshotCollector related metrics to
// the provided Prometheus channel.
func (c *SnapshotCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range c.collectorList() {
		metric.Describe(ch)
	}
}

// Collect sends all the collected metrics to the provided Prometheus channel.
// It requires the caller to handle synchronization.
func (c *SnapshotCollector) Collect(ch chan<- prometheus.Metric) {
	// Reset current labels.
	c.SnapshotStatus.Reset()
	c.SnapshotCount.Reset()

	if err := c.collectSnaphotStatuses(); err != nil {
		log.Errorf("failed to collect snapshots metrics: %v", err)
	}
	if err := c.collectSnaphotCount(); err != nil {
		log.Errorf("failed to collect snapshots counts: %v", err)
	}
	for _, metric := range c.collectorList() {
		metric.Collect(ch)
	}
}
