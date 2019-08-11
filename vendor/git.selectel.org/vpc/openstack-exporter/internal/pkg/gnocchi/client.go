package gnocchi

import (
	"net/http"
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/utils/gnocchi"
	"github.com/pkg/errors"
)

// GetOpts represents common options for performing GET request against Gnocchi API.
type GetOpts struct {
	Client   *gophercloud.ServiceClient
	Region   string
	Attempts int
	Interval time.Duration
}

// ClientOpts represents options to initialize Gnocchi API clients.
type ClientOpts struct {
	Provider     *gophercloud.ProviderClient
	EndpointType string
	Timeout      time.Duration
}

// NewGnocchiV1Clients returns reference to instances of the Gnocchi V1 service
// clients that will be initialized for every region specified in the config.
// Clients will be saved into map with region names as keys.
func NewGnocchiV1Clients(opts *ClientOpts) (map[string]gophercloud.ServiceClient, error) {
	regions := config.Config.Gnocchi.Regions
	clients := make(map[string]gophercloud.ServiceClient, len(regions))

	for _, region := range regions {
		client, err := gnocchi.NewGnocchiV1(opts.Provider, gophercloud.EndpointOpts{
			Region:       region,
			Availability: gophercloud.Availability(opts.EndpointType),
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error initializing Gnocchi client in %s region", region)
		}
		client.HTTPClient = http.Client{
			Timeout: opts.Timeout,
		}
		clients[region] = *client
	}

	return clients, nil
}
