package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/gnocchi"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/gnocchi/status"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"github.com/gophercloud/gophercloud"
)

// GetGnocchiStatuses retrieve Gnocchi statuses from the Gnocchi API.
// It will make call with every provided Gnocchi V1 service client.
func GetGnocchiStatuses(clients map[string]gophercloud.ServiceClient) []*status.GnocchiStatus {
	allStatuses := []*status.GnocchiStatus{}

	for region := range clients {
		client := clients[region]
		regionStatus, err := status.GetGnocchiStatus(&gnocchi.GetOpts{
			Region:   region,
			Client:   &client,
			Attempts: config.Config.Gnocchi.RequestAttempts,
			Interval: time.Second * time.Duration(config.Config.Gnocchi.RequestInterval),
		})
		if err != nil {
			log.Errorf("error getting Gnocchi status from '%s': %v",
				client.ResourceBase, err)
			continue
		}

		log.Debugf("Got %+v Gnocchi status from %s region API", regionStatus, region)
		allStatuses = append(allStatuses, regionStatus)
	}

	return allStatuses
}
