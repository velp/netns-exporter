package amphorae

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/utils"
	"github.com/gophercloud/gophercloud/openstack/loadbalancer/v2/amphorae"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
)

const (
	// Actual statuses:
	// https://github.com/openstack/octavia/blob/master/octavia/common/constants.py#L122
	//
	// We treat all other statuses as "unknown/other" because we don't need to
	// save all existing Octavia statuses in our Prometheus TSDB.
	statusAllocated = "ALLOCATED"
	statusReady     = "READY"
	statusError     = "ERROR"
)

// AmphoraStatusCode represents amphora status int.
type AmphoraStatusCode int

const (
	// AmphoraStatusCodeOk contains status code of the active Octavia amphora.
	AmphoraStatusCodeOk AmphoraStatusCode = iota

	// AmphoraStatusCodeError contains status code of the failed Octavia amphora.
	AmphoraStatusCodeError

	// AmphoraStatusCodeOther contains status code of the Octavia amphora in all
	// other statuses (pending, allocated, etc.).
	AmphoraStatusCodeOther
)

// Amphora represent simplified amphora with status represented as int.
type Amphora struct {
	ID             string
	LoadbalancerID string
	ComputeID      string
	Region         string
	StatusCode     AmphoraStatusCode
}

// GetAmphorae retrieves Octavia amphorae and flattens them into simplified
// structures.
func GetAmphorae(opts *common.GetOpts) ([]*Amphora, error) {
	amphoraePages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return amphorae.List(opts.Client, amphorae.ListOpts{}).AllPages()
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get amphorae")
	}
	allAmphorae, err := amphorae.ExtractAmphorae(amphoraePages)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read amphorae response body")
	}

	return flattenAmphorae(allAmphorae, opts.Region), nil
}

func flattenAmphorae(extendedAmphorae []amphorae.Amphora, region string) []*Amphora {
	flattenedAmphorae := make([]*Amphora, len(extendedAmphorae))

	for amphoraeIdx, extendedAmphora := range extendedAmphorae {
		flattenedAmphorae[amphoraeIdx] = &Amphora{
			ID:             extendedAmphora.ID,
			LoadbalancerID: extendedAmphora.LoadbalancerID,
			ComputeID:      extendedAmphora.ComputeID,
			Region:         region,
		}

		switch extendedAmphora.Status {
		case statusReady, statusAllocated:
			flattenedAmphorae[amphoraeIdx].StatusCode = AmphoraStatusCodeOk
		case statusError:
			flattenedAmphorae[amphoraeIdx].StatusCode = AmphoraStatusCodeError
		default:
			flattenedAmphorae[amphoraeIdx].StatusCode = AmphoraStatusCodeOther
		}
	}

	return flattenedAmphorae
}
