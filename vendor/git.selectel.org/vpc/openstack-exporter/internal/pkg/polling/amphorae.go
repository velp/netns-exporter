package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/octavia/amphorae"
	"github.com/gophercloud/gophercloud"
)

// GetAmphorae returns amphorae statuses from the Octavia API.
// It will make calls with every provided Octavia V2 service client.
func GetAmphorae(clients map[string]gophercloud.ServiceClient) []*amphorae.Amphora {
	allAmphorae := []*amphorae.Amphora{}

	for region := range clients {
		client := clients[region]
		regionAmphorae, err := amphorae.GetAmphorae(&common.GetOpts{
			Region:   region,
			Client:   &client,
			Attempts: config.Config.OpenStack.Octavia.RequestAttempts,
			Interval: time.Second * time.Duration(config.Config.OpenStack.Octavia.RequestInterval),
		})
		if err != nil {
			log.Errorf("error getting Octavia amphorae from '%s': %v",
				client.ResourceBase, err)
			continue
		}

		log.Debugf("Got %d amphorae from %s region API", len(regionAmphorae), region)
		allAmphorae = append(allAmphorae, regionAmphorae...)
	}

	return allAmphorae
}
