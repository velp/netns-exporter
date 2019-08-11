package loadbalancers

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	// Actual statuses:
	// https://github.com/openstack/octavia/blob/master/octavia/common/constants.py
	operatingStatusOnline  = "ONLINE"
	operatingStatusOffline = "OFFLINE"
	operatingStatusError   = "ERROR"

	provisioningStatusActive = "ACTIVE"
	provisioningStatusError  = "ERROR"
)

// LoadBalancerOperatingStatus represents existing OpenStack Octavia
// loadbalancer operating status.
type LoadBalancerOperatingStatus string

const (
	OperatingStatusOnline  LoadBalancerOperatingStatus = operatingStatusOnline
	OperatingStatusOffline LoadBalancerOperatingStatus = operatingStatusOffline
	OperatingStatusError   LoadBalancerOperatingStatus = operatingStatusError
)

// LoadBalancerProvisioningStatus represents existing OpenStack Octavia
// loadbalancer provisioning status.
type LoadBalancerProvisioningStatus string

const (
	ProvisioningStatusActive LoadBalancerProvisioningStatus = provisioningStatusActive
	ProvisioningStatusError  LoadBalancerProvisioningStatus = provisioningStatusError
)

// LoadBalancerOperatingStatusCode represents existing OpenStack Octavia
// loadbalancer operating status code.
type LoadBalancerOperatingStatusCode int

const (
	// OperatingStatusCodeOnline contains status code of the online Octavia
	// loadbalancer.
	OperatingStatusCodeOnline LoadBalancerOperatingStatusCode = iota

	// OperatingStatusCodeOffline contains status code of the offline Octavia
	// loadbalancer.
	OperatingStatusCodeOffline

	// OperatingStatusCodeError contains status code of the Octavia
	// loadbalancer in error status.
	OperatingStatusCodeError

	// OperatingStatusCodeUnknown contains status code of the unknown Octavia
	// loadbalancer.
	OperatingStatusCodeUnknown
)

// LoadBalancerProvisioningStatusCode represents existing OpenStack Octavia
// loadbalancer provisioning status code.
type LoadBalancerProvisioningStatusCode int

const (
	// ProvisioningStatusCodeActive contains status code of the online Octavia
	// loadbalancer.
	ProvisioningStatusCodeActive LoadBalancerProvisioningStatusCode = iota

	// ProvisioningStatusCodeError contains status code of the offline Octavia
	// loadbalancer.
	ProvisioningStatusCodeError

	// ProvisioningStatusCodeUnknown contains status code of the
	// unknown Octavia loadbalancer.
	ProvisioningStatusCodeUnknown
)

// LoadBalancer represents an Octavia loadbalancer.
type LoadBalancer struct {
	ID                     string                             `json:"id"`
	ProvisioningStatus     string                             `json:"provisioning_status"`
	OperatingStatus        string                             `json:"operating_status"`
	ProjectID              string                             `json:"project_id"`
	AccountName            string                             `json:"-"`
	Region                 string                             `json:"-"`
	OperatingStatusCode    LoadBalancerOperatingStatusCode    `json:"-"`
	ProvisioningStatusCode LoadBalancerProvisioningStatusCode `json:"-"`
}

// GetLoadBalancers retrieves raw byteslice of OpenStack Octavia loadbalancers
// and builds a slice of LoadBalancer structures with populated
// OperatingStatusCode field.
// It will set provided region to every loadbalancer structure.
func GetLoadBalancers(raw []byte, region string) ([]*LoadBalancer, error) {
	loadbalancers := []*LoadBalancer{}
	err := json.Unmarshal(raw, &loadbalancers)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling raw loadbalancers data")
	}

	for _, loadbalancer := range loadbalancers {
		// Set provided region.
		loadbalancer.Region = region

		// Set operating status code.
		switch loadbalancer.OperatingStatus {
		case string(OperatingStatusOnline):
			loadbalancer.OperatingStatusCode = OperatingStatusCodeOnline
		case string(OperatingStatusOffline):
			loadbalancer.OperatingStatusCode = OperatingStatusCodeOffline
		case string(OperatingStatusError):
			loadbalancer.OperatingStatusCode = OperatingStatusCodeError
		default:
			loadbalancer.OperatingStatusCode = OperatingStatusCodeUnknown
		}

		// Set provisioning status code.
		switch loadbalancer.ProvisioningStatus {
		case string(ProvisioningStatusActive):
			loadbalancer.ProvisioningStatusCode = ProvisioningStatusCodeActive
		case string(ProvisioningStatusError):
			loadbalancer.ProvisioningStatusCode = ProvisioningStatusCodeError
		default:
			loadbalancer.ProvisioningStatusCode = ProvisioningStatusCodeUnknown
		}
	}

	return loadbalancers, nil
}
