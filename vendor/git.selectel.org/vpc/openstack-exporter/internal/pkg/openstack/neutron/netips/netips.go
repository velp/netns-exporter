package netips

import (
	"fmt"
	"strconv"
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/neutron/external"
	"github.com/gophercloud/gophercloud/openstack/networking/v2/extensions/networkipavailabilities"
	"github.com/pkg/errors"
)

// NetworkIPs contains IP availability information for a single network.
type NetworkIPs struct {
	ID           string
	ProjectID    string
	Region       string
	TotalIPs     float64
	UsedIPs      float64
	AvailableIPs float64
}

// GetExternalNetworkIPs retrieves IP availability information for every
// external network.
func GetExternalNetworkIPs(opts *common.GetOpts) ([]*NetworkIPs, error) {
	// Get all external networks.
	externalNetworks, err := external.GetExternalNetworks(opts)
	if err != nil {
		return nil, err
	}

	// Retrieve IP availabilities for each external network.
	allIPAvailabilities := make([]*networkipavailabilities.NetworkIPAvailability, len(externalNetworks))
	for externalNetworkIdx, externalNetwork := range externalNetworks {
		ipAvailability, err := getIPAvailabilities(&getIPAvailabilitiesOpts{
			opts,
			externalNetwork.ID,
		})
		if err != nil {
			return nil, errors.Wrapf(err, "error getting IP availability for network '%s", externalNetwork.ID)
		}
		allIPAvailabilities[externalNetworkIdx] = ipAvailability
	}

	// Flatten Neutron IP availabilities into simplified structures.
	return flattenIPAvailabilities(allIPAvailabilities, opts.Region)
}

// getIPAvailabilitiesOpts contains options for the getIPAvailabilities.
type getIPAvailabilitiesOpts struct {
	*common.GetOpts
	id string
}

// getIPAvailabilities is a helper function to retrieve IP availability
// information for a single network.
func getIPAvailabilities(opts *getIPAvailabilitiesOpts) (*networkipavailabilities.NetworkIPAvailability, error) {
	var (
		ipAvailability *networkipavailabilities.NetworkIPAvailability
		err            error
	)

	for i := 0; i < opts.Attempts; i++ {
		ipAvailability, err = networkipavailabilities.Get(opts.Client, opts.id).Extract()
		if err == nil {
			return ipAvailability, nil
		}
		time.Sleep(opts.Interval)
		log.Debugf("retrying after error: %s", err)
	}

	return nil, fmt.Errorf("after %d attempts, last error was: %s", opts.Attempts, err)
}

// flattenIPAvailabilities is a helper function to get set of simplified IP
// data from the Neutron V2 IP availabilities.
func flattenIPAvailabilities(availabilities []*networkipavailabilities.NetworkIPAvailability, region string) ([]*NetworkIPs, error) {
	networkIPs := make([]*NetworkIPs, len(availabilities))

	for availabilityIdx, availabilityData := range availabilities {
		// Convert totalIPs and usedIPs strings into float64.
		totalIPs, err := strconv.ParseFloat(availabilityData.TotalIPs, 64)
		if err != nil {
			return nil, fmt.Errorf("got invalid total IPs for network '%s': %s",
				availabilityData.NetworkID,
				availabilityData.TotalIPs,
			)
		}
		usedIPs, err := strconv.ParseFloat(availabilityData.UsedIPs, 64)
		if err != nil {
			return nil, fmt.Errorf("got invalid used IPs for network '%s': %s",
				availabilityData.NetworkID,
				availabilityData.UsedIPs,
			)
		}
		// Calculate availableIPs from totalIPs and usedIPs.
		availableIPs := totalIPs - usedIPs

		// Add retrieved availability data into networkIPs.
		networkIPs[availabilityIdx] = &NetworkIPs{
			ID:           availabilityData.NetworkID,
			ProjectID:    availabilityData.ProjectID,
			Region:       region,
			TotalIPs:     totalIPs,
			UsedIPs:      usedIPs,
			AvailableIPs: availableIPs,
		}
	}

	return networkIPs, nil
}
