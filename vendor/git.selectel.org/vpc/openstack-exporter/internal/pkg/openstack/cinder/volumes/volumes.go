package volumes

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	statusAvailable       = "available"
	statusInUse           = "in-use"
	statusManaging        = "managing"
	statusMaintenance     = "maintenance"
	statusRestoringBackup = "restoring-backup"
	statusError           = "error"
	statusErrorDeleting   = "error_deleting"
	statusErrorManaging   = "error_managing"
	statusErrorRestoring  = "error_restoring"
)

// VolumeStatus represents existing OpenStack Cinder volume status.
type VolumeStatus string

const (
	VolumeStatusAvailable       VolumeStatus = statusAvailable
	VolumeStatusInUse           VolumeStatus = statusInUse
	VolumeStatusManaging        VolumeStatus = statusManaging
	VolumeStatusMaintenance     VolumeStatus = statusMaintenance
	VolumeStatusRestoringBackup VolumeStatus = statusRestoringBackup
	VolumeStatusError           VolumeStatus = statusError
	VolumeStatusErrorDeleting   VolumeStatus = statusErrorDeleting
	VolumeStatusErrorManaging   VolumeStatus = statusErrorManaging
	VolumeStatusErrorRestoring  VolumeStatus = statusErrorRestoring
)

// VolumeStatusCode represents existing OpenStack Cinder volume status code.
type VolumeStatusCode int

const (
	// VolumeStatusCodeOk contains status code of the active Cinder volume.
	VolumeStatusCodeOk VolumeStatusCode = iota

	// VolumeStatusCodeError contains status code of the Cinder volume in error
	// state.
	VolumeStatusCodeError

	// VolumeStatusCodeTransition contains status code of the Cinder volume in
	// transition state.
	VolumeStatusCodeTransition
)

// Volume represents a single Cinder V3 volume.
type Volume struct {
	ID               string           `json:"id"`
	Status           string           `json:"status"`
	Size             int              `json:"size"`
	Region           string           `json:"-"`
	AvailabilityZone string           `json:"availability_zone"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
	VolumeType       string           `json:"volume_type"`
	SnapshotID       string           `json:"snapshot_id"`
	SourceVolID      string           `json:"source_volid"`
	UserID           string           `json:"user_id"`
	ProjectID        string           `json:"os-vol-tenant-attr:tenant_id"`
	AccountName      string           `json:"-"`
	StatusCode       VolumeStatusCode `json:"-"`
}

// GetVolumes retrieves raw byteslice of OpenStack Cinder volumes and builds
// a slice of Volume structures with populated StatusCode field.
// It will set provided region to every instance structure.
func GetVolumes(raw []byte, region string) ([]*Volume, error) {
	volumes := []*Volume{}
	err := json.Unmarshal(raw, &volumes)
	if err != nil {
		return nil, errors.Wrap(err, "error unmarshalling raw volumes data")
	}

	for _, volume := range volumes {
		// Set provided region.
		volume.Region = region

		switch volume.Status {
		case
			string(VolumeStatusAvailable),
			string(VolumeStatusInUse),
			string(VolumeStatusManaging),
			string(VolumeStatusMaintenance),
			string(VolumeStatusRestoringBackup):
			volume.StatusCode = VolumeStatusCodeOk
		case
			string(VolumeStatusError),
			string(VolumeStatusErrorDeleting),
			string(VolumeStatusErrorManaging),
			string(VolumeStatusErrorRestoring):
			volume.StatusCode = VolumeStatusCodeError
		default:
			volume.StatusCode = VolumeStatusCodeTransition
		}
	}

	return volumes, nil
}
