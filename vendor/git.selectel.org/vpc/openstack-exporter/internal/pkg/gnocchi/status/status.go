package status

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/gnocchi"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"github.com/gophercloud/utils/gnocchi/metric/v1/status"
	"github.com/pkg/errors"
)

// GnocchiStatus contains simplified Gnocchi status of measures backlog.
type GnocchiStatus struct {
	// Processors represents total number of running metricd proccessors.
	Processors int

	// Measures represents total number of measures to process.
	Measures int

	// Metrics represents total number of metric having measures to process.
	Metrics int

	// Region represents current region.
	Region string
}

// GetGnocchiStatus retrieves a Gnocchi status in current region.
func GetGnocchiStatus(opts *gnocchi.GetOpts) (*GnocchiStatus, error) {
	var (
		gnocchiStatus *status.Status
		err           error
		details       bool
	)

	for i := 0; i < opts.Attempts; i++ {
		gnocchiStatus, err = status.Get(opts.Client, status.GetOpts{
			Details: &details,
		}).Extract()
		if err == nil {
			return readGnocchiStatus(gnocchiStatus, opts.Region), nil
		}
		time.Sleep(opts.Interval)
		log.Debugf("retrying retrieve Gnocchi status after error: %s", err)
	}

	return nil, errors.Wrap(err, "unable to get Gnocchi status")
}

func readGnocchiStatus(gnocchiStatus *status.Status, region string) *GnocchiStatus {
	return &GnocchiStatus{
		Processors: len(gnocchiStatus.Metricd.Processors),
		Measures:   gnocchiStatus.Storage.Summary.Measures,
		Metrics:    gnocchiStatus.Storage.Summary.Metrics,
		Region:     region,
	}
}
