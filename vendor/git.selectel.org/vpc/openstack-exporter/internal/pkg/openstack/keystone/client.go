package keystone

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
)

// NewKeystoneV3Client returns reference to an instance of the Keystone V3 service client.
func NewKeystoneV3Client(opts *common.NewClientOpts) (*gophercloud.ServiceClient, error) {
	return openstack.NewIdentityV3(opts.Provider, gophercloud.EndpointOpts{
		Region:       config.Config.OpenStack.Keystone.Region,
		Availability: gophercloud.Availability(opts.EndpointType),
	})
}
