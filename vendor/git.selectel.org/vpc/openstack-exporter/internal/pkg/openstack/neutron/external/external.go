package external

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/utils"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/networks"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
)

// Network contains Neutron V2 network with the external attribute.
type Network struct {
	networks.Network
	external.NetworkExternalExt
}

// GetExternalNetworks retrieves all external networks in current region.
func GetExternalNetworks(opts *common.GetOpts) ([]*Network, error) {
	var externalNetworks []*Network
	externalAttr := true

	externalNetworkPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return networks.List(opts.Client, external.ListOptsExt{
			ListOptsBuilder: networks.ListOpts{},
			External:        &externalAttr,
		}).AllPages()
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get external networks")
	}
	err = networks.ExtractNetworksInto(externalNetworkPages, &externalNetworks)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read external networks response body")
	}

	return externalNetworks, nil
}
