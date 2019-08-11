package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/neutron/netips"
	"github.com/gophercloud/gophercloud"
)

// GetExternalNetworksIPs returns IP availability information for all external
// networks from the Neutron V2 API.
//  It will make calls with every provided Neutron V2 service client.
func GetExternalNetworksIPs(clients map[string]gophercloud.ServiceClient) []*netips.NetworkIPs {
	externalNetworkIPs := []*netips.NetworkIPs{}

	for region := range clients {
		client := clients[region]
		regionExternalNetworkIPs, err := netips.GetExternalNetworkIPs(&common.GetOpts{
			Region:   region,
			Client:   &client,
			Attempts: config.Config.OpenStack.Neutron.RequestAttempts,
			Interval: time.Second * time.Duration(config.Config.OpenStack.Neutron.RequestInterval),
		})
		if err != nil {
			log.Errorf("error getting IP availability for external networks from '%s': %v",
				client.ResourceBase, err)
			continue
		}
		externalNetworkIPs = append(externalNetworkIPs, regionExternalNetworkIPs...)
	}

	return externalNetworkIPs
}
