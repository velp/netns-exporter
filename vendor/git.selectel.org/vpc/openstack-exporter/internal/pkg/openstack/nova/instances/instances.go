package instances

import (
	"encoding/json"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/utils"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/availabilityzones"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/extensions/extendedstatus"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"
	"github.com/gophercloud/gophercloud/pagination"
	"github.com/pkg/errors"
)

const (
	vmStateError = "error"
)

// InstanceStatusCode represents existing OpenStack Nova instance status code.
type InstanceStatusCode int

const (
	// InstanceStatusCodeOk contains status code of the active Nova instance.
	InstanceStatusCodeOk InstanceStatusCode = iota

	// InstanceStatusCodeError contains status code of the Nova instance in
	// error state.
	InstanceStatusCodeError

	// InstanceStatusCodeTransition contains status code of the Nova instance in
	// transition state.
	InstanceStatusCodeTransition
)

// Instance represents a single Nova V2 instance.
type Instance struct {
	ID               string             `json:"id"`
	ProjectID        string             `json:"tenant_id"`
	UserID           string             `json:"user_id"`
	VMState          string             `json:"OS-EXT-STS:vm_state"`
	TaskState        string             `json:"OS-EXT-STS:task_state"`
	InstanceName     string             `json:"OS-EXT-SRV-ATTR:instance_name"`
	Region           string             `json:"-"`
	AvailabilityZone string             `json:"OS-EXT-AZ:availability_zone"`
	CreatedAt        string             `json:"created"`
	Flavor           string             `json:"-"`
	RAM              int                `json:"-"`
	VCPUs            int                `json:"-"`
	LocalDisk        int                `json:"-"`
	Status           string             `json:"status"`
	StatusCode       InstanceStatusCode `json:"-"`
}

// UnmarshalJSON implements custom unmarshalling method for the Instance type.
// We need it to read flavor information from the nested JSON and populate RAM,
// VCPUs and LocalDisk fields.
func (i *Instance) UnmarshalJSON(b []byte) error {
	// Unmarshal raw JSON into custom structure.
	type instance Instance
	var expandedInstance struct {
		instance
		FlavorData map[string]interface{} `json:"flavor"`
	}
	if err := json.Unmarshal(b, &expandedInstance); err != nil {
		return err
	}

	*i = Instance(expandedInstance.instance)

	// Collect flavor data if available.
	if len(expandedInstance.FlavorData) == 0 {
		return nil
	}

	if rawFlavor, ok := expandedInstance.FlavorData["original_name"]; ok {
		if flavor, ok := rawFlavor.(string); ok {
			i.Flavor = flavor
		}
	}
	if rawRAM, ok := expandedInstance.FlavorData["ram"]; ok {
		if ram, ok := rawRAM.(float64); ok {
			i.RAM = int(ram)
		}
	}
	if rawVCPUs, ok := expandedInstance.FlavorData["vcpus"]; ok {
		if vCPUs, ok := rawVCPUs.(float64); ok {
			i.VCPUs = int(vCPUs)
		}
	}
	if rawLocalDisk, ok := expandedInstance.FlavorData["disk"]; ok {
		if localDisk, ok := rawLocalDisk.(float64); ok {
			i.LocalDisk = int(localDisk)
		}
	}

	return nil
}

// GetInstancesFromCache retrieves raw byteslice of OpenStack Nova instances
// and builds a slice of Instance structures with populated StatusCode field.
// It will set provided region to every instance structure.
func GetInstancesFromCache(raw []byte, region string) ([]*Instance, error) {
	instances := []*Instance{}
	err := json.Unmarshal(raw, &instances)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling raw instances data")
	}

	// Here we're setting status code by following rules:
	//  - set transition code if OS-EXT-STS:task_state value is not empty
	//  - set error code if OS-EXT-STS:vm_state value is error
	//  - set ok code in any other cases
	for _, instance := range instances {
		// Set provided region.
		instance.Region = region

		if instance.TaskState != "" {
			instance.StatusCode = InstanceStatusCodeTransition
		}
		if instance.VMState == vmStateError {
			instance.StatusCode = InstanceStatusCodeError
		}
	}

	return instances, nil
}

// ServerWithExtendedStatus represents Nova instance with additional
// if OS-EXT-STS and OS-EXT-AZ fields.
type ServerWithExtendedStatus struct {
	servers.Server
	extendedstatus.ServerExtendedStatusExt
	availabilityzones.ServerAvailabilityZoneExt
}

// GetInstancesFromAPI retrieves Nova instances from API and flattens them into
// simplified structures.
func GetInstancesFromAPI(opts *common.GetOpts) ([]*Instance, error) {
	var allInstances []ServerWithExtendedStatus

	instancesPages, err := utils.RetryForPager(opts.Attempts, opts.Interval, func() (pagination.Page, error) {
		return servers.List(opts.Client, servers.ListOpts{AllTenants: true}).AllPages()
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to get instances")
	}
	err = servers.ExtractServersInto(instancesPages, &allInstances)
	if err != nil {
		return nil, errors.Wrap(err, "unable to read instances response body")
	}

	return flattenInstances(allInstances, opts.Region), nil
}

func flattenInstances(extendedInstances []ServerWithExtendedStatus, region string) []*Instance {
	flattenedInstances := make([]*Instance, len(extendedInstances))

	for instanceIdx, extendedInstance := range extendedInstances {
		flattenedInstances[instanceIdx] = &Instance{
			ID:               extendedInstance.ID,
			Region:           region,
			ProjectID:        extendedInstance.TenantID,
			AvailabilityZone: extendedInstance.AvailabilityZone,
		}

		// Here we're setting status code by following rules:
		//  - set transition code if OS-EXT-STS:task_state value is not empty
		//  - set error code if OS-EXT-STS:vm_state value is error
		//  - set ok code in any other cases
		if extendedInstance.TaskState != "" {
			flattenedInstances[instanceIdx].StatusCode = InstanceStatusCodeTransition
		}
		if extendedInstance.VmState == vmStateError {
			flattenedInstances[instanceIdx].StatusCode = InstanceStatusCodeError
		}
	}

	return flattenedInstances
}
