package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	keystoneProjects "git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/keystone/projects"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/nova/instances"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/octavia/amphorae"
	"github.com/go-redis/redis"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack/identity/v3/projects"
)

// GetOctaviaInstancesOpts represents common options to performing GET request to retrieve Octavia instances.
type GetOctaviaInstancesOpts struct {
	OctaviaClients map[string]gophercloud.ServiceClient
	KeystoneClient *gophercloud.ServiceClient
	RedisClient    *redis.Client
}

// GetOctaviaOrphanedInstances retrieves Octavia orphaned instances. It will make calls with every provided
// service client and retrieve data from the raw cache and from the Openstack API.
func GetOctaviaOrphanedInstances(opts *GetOctaviaInstancesOpts) map[string]string {
	var (
		allAmphorae           []*amphorae.Amphora
		allInstances          []*instances.Instance
		octaviaProject        *projects.Project
		allAmphoraeComputeIDs = make(map[string]struct{})
		allOrphanedInstances  = make(map[string]string)
	)

	// Get amphorae from the Octavia API.
	allAmphorae = GetAmphorae(opts.OctaviaClients)

	// Populate map with all amphorae compute IDs.
	for _, amphora := range allAmphorae {
		allAmphoraeComputeIDs[amphora.ComputeID] = struct{}{}
	}

	// Get Nova instances from the raw cached data.
	allInstances = GetInstances(&GetObjectsOpts{
		CloudClients: map[string]gophercloud.ServiceClient{},
		RedisClient:  opts.RedisClient,
	})

	// Get Octavia service project from the OpenStack Identity V3 API.
	octaviaProject, err := keystoneProjects.GetOctaviaProject(&common.GetOpts{
		Client:   opts.KeystoneClient,
		Attempts: config.Config.OpenStack.Keystone.RequestAttempts,
		Interval: time.Second * time.Duration(config.Config.OpenStack.Keystone.RequestInterval),
	})
	if err != nil {
		log.Errorf("error getting Octavia service project: %v", err)
	}

	// Compare all Nova instances with amphorae compute IDs and populate Octavia orphaned instances map.
	for _, instance := range allInstances {
		if instance.ProjectID == octaviaProject.ID {
			if _, ok := allAmphoraeComputeIDs[instance.ID]; ok {
				continue
			} else {
				allOrphanedInstances[instance.ID] = instance.Region
			}
		}
	}

	return allOrphanedInstances
}
