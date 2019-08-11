package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/magnum/clusters"
	"github.com/gophercloud/gophercloud"
)

// GetCOEClusters returns clusters statuses from the OpenStack COE API.
// It will make calls with every provided Magnum V1 service client.
func GetCOEClusters(clients map[string]gophercloud.ServiceClient) []*clusters.Cluster {
	allClusters := []*clusters.Cluster{}

	for region := range clients {
		client := clients[region]
		regionClusters, err := clusters.GetClusters(&common.GetOpts{
			Region:   region,
			Client:   &client,
			Attempts: config.Config.OpenStack.Magnum.RequestAttempts,
			Interval: time.Second * time.Duration(config.Config.OpenStack.Magnum.RequestInterval),
		})
		if err != nil {
			log.Errorf("error getting Magnum clusters from '%s': %v",
				client.ResourceBase, err)
			continue
		}

		log.Debugf("Got %d COE clusters from %s region API", len(regionClusters), region)
		allClusters = append(allClusters, regionClusters...)
	}

	return allClusters
}

// GetAccountsCOEClusters gets COE clusters count per project in all accounts.
func GetAccountsCOEClusters(clients map[string]gophercloud.ServiceClient) ([]*AccountObjects, error) {
	clustersData := GetCOEClusters(clients)

	projectsClusters := make(map[string]map[string]int)

	// Populate mapping between projects and COE clusters count in each region.
	for _, cluster := range clustersData {
		if _, ok := projectsClusters[cluster.ProjectID]; !ok {
			// Instantiate empty project clusters count map if it doesn't exist yet.
			projectsClusters[cluster.ProjectID] = map[string]int{
				cluster.Region: 0,
			}
		} else if _, ok := projectsClusters[cluster.ProjectID][cluster.Region]; !ok {
			// Instantiate only region map in existing projects cluster map in
			// case that it already contains counts for some other regions.
			projectsClusters[cluster.ProjectID][cluster.Region] = 0
		}

		// Safely increment project clusters count in a single region.
		projectsClusters[cluster.ProjectID][cluster.Region]++
	}

	return getAccountObjects(projectsClusters), nil
}
