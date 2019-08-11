package openstack

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/openstack/clientconfig"
)

// ClientOpts represents parameters for the OpenStack provider client.
type ClientOpts struct {
	AuthURL     string
	Username    string
	Password    string
	ProjectName string
	DomainName  string
	Region      string
}

// NewProviderClient returns reference to an instance of the OpenStack provider
// client.
// It can be used to authenticate clients for different OpenStack services.
func NewProviderClient(opts *ClientOpts) (*gophercloud.ProviderClient, error) {
	return clientconfig.AuthenticatedClient(&clientconfig.ClientOpts{
		AuthInfo: &clientconfig.AuthInfo{
			AuthURL:     opts.AuthURL,
			Username:    opts.Username,
			Password:    opts.Password,
			ProjectName: opts.ProjectName,
			DomainName:  opts.DomainName,
		},
		RegionName: opts.Region,
	})
}
