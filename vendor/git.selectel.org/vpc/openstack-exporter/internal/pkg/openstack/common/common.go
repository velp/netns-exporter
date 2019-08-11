package common

import (
	"time"

	"github.com/gophercloud/gophercloud"
)

// GetOpts represents common options for performing GET request against
// OpenStack API.
type GetOpts struct {
	Client   *gophercloud.ServiceClient
	Region   string
	Attempts int
	Interval time.Duration
}

// NewClientOpts represents common options to initialize OpenStack API clients.
type NewClientOpts struct {
	Provider     *gophercloud.ProviderClient
	EndpointType string
	Timeout      time.Duration
}
