package polling

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack"
	"github.com/gophercloud/gophercloud"
	"github.com/pkg/errors"
)

// NewOpenStackProvider authenticates in the OpenStack with parameters provided
// in application config.
func NewOpenStackProvider() (*gophercloud.ProviderClient, error) {
	provider, err := openstack.NewProviderClient(&openstack.ClientOpts{
		AuthURL:     config.Config.OpenStack.Keystone.AuthURL,
		Username:    config.Config.OpenStack.Keystone.Username,
		Password:    config.Config.OpenStack.Keystone.Password,
		ProjectName: config.Config.OpenStack.Keystone.ProjectName,
		DomainName:  config.Config.OpenStack.Keystone.DomainName,
		Region:      config.Config.OpenStack.Keystone.Region,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error authenticating in OpenStack")
	}

	return provider, nil
}
