package clusters

import (
	"fmt"
	"strings"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/utils"
	"github.com/gophercloud/gophercloud/openstack/containerinfra/v1/clusters"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
)

const (
	// Actual statuses: https://github.com/openstack/magnum/blob/master/magnum/objects/fields.py#L18
	// We're checking last part of status.
	statusComplete   = "COMPLETE"
	statusFailed     = "FAILED"
	statusInProgress = "PROGRESS"

	errInvalidStatusFmt = "got invalid cluster status: %s"
)

// ClusterStatusCode represents existing OpenStack Magnum cluster status code.
type ClusterStatusCode int

const (
	StatusCodeComplete ClusterStatusCode = iota
	StatusCodeFailed
	StatusCodeInProgress
)

// Cluster represents simplified Magnum cluster with status represented as int
// value.
type Cluster struct {
	ID         string            `json:"id"`
	ProjectID  string            `json:"project_id"`
	Region     string            `json:"-"`
	StatusCode ClusterStatusCode `json:"-"`
}

// GetClusters retrieves OpenStack Magnum clusters in specified region and builds
// a slice of simplified cluster structures.
func GetClusters(opts *common.GetOpts) ([]*Cluster, error) {
	clusterPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return clusters.ListDetail(opts.Client, clusters.ListOpts{}).AllPages()
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get clusters")
	}
	allClusters, err := clusters.ExtractClusters(clusterPages)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read clusters response body")
	}

	return flattenClusters(allClusters, opts.Region), nil
}

func flattenClusters(extendedClusters []clusters.Cluster, region string) []*Cluster {
	flattenedClusters := []*Cluster{}

	for _, cluster := range extendedClusters {
		statusCode, err := getClusterStatusCode(cluster.Status)
		if err != nil {
			log.Errorf("skipping cluster %s: %s", cluster.UUID, err)
			continue
		}

		flattenedClusters = append(flattenedClusters, &Cluster{
			ID:         cluster.UUID,
			ProjectID:  cluster.ProjectID,
			Region:     region,
			StatusCode: statusCode,
		})
	}

	return flattenedClusters
}

func getClusterStatusCode(status string) (ClusterStatusCode, error) {
	statusParts := strings.Split(status, "_")

	// Some strange status.
	if len(statusParts) < 2 {
		return 0, fmt.Errorf(errInvalidStatusFmt, status)
	}

	// Complete/failed status.
	if len(statusParts) == 2 {
		if statusParts[1] == statusComplete {
			return StatusCodeComplete, nil
		}
		if statusParts[1] == statusFailed {
			return StatusCodeFailed, nil
		}
	}

	// In progress status.
	if len(statusParts) == 3 {
		if statusParts[2] == statusInProgress {
			return StatusCodeInProgress, nil
		}
	}

	// Unknown status.
	return 0, fmt.Errorf(errInvalidStatusFmt, status)
}
