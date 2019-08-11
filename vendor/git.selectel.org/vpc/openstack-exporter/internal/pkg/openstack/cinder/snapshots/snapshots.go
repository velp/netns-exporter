package snapshots

import (
	"encoding/json"

	"github.com/pkg/errors"
)

const (
	statusAvailable     = "available"
	statusError         = "error"
	statusErrorDeleting = "error_deleting"
)

// SnapshotStatus represents existing OpenStack Cinder snapshot status.
type SnapshotStatus string

const (
	SnapshotStatusAvailable     SnapshotStatus = statusAvailable
	SnapshotStatusError         SnapshotStatus = statusError
	SnapshotStatusErrorDeleting SnapshotStatus = statusErrorDeleting
)

// SnapshotStatusCode represents existing OpenStack Cinder snapshot status code.
type SnapshotStatusCode int

const (
	SnapshotStatusCodeOk SnapshotStatusCode = iota
	SnapshotStatusCodeError
	SnapshotStatusCodeTransition
)

// Snapshot represents a single Cinder snapshot.
type Snapshot struct {
	ID          string             `json:"id"`
	VolumeID    string             `json:"volume_id"`
	ProjectID   string             `json:"os-extended-snapshot-attributes:project_id"`
	AccountName string             `json:"-"`
	Region      string             `json:"region"`
	CreatedAt   string             `json:"created_at"`
	Status      string             `json:"status"`
	StatusCode  SnapshotStatusCode `json:"-"`
}

// GetSnapshots retrieves raw byteslice of OpenStack Cinder snapshots and builds
// a slice of Snapshot structures with populated StatusCode field.
// It will set provided region to every snapshot structure.
func GetSnapshots(raw []byte, region string) ([]*Snapshot, error) {
	snapshots := []*Snapshot{}
	if err := json.Unmarshal(raw, &snapshots); err != nil {
		return nil, errors.Wrap(err, "error unmarshalling raw snapshots data")
	}

	for _, snapshot := range snapshots {
		// Set provided region.
		snapshot.Region = region

		// Set snapshot status code.
		switch snapshot.Status {
		case
			string(SnapshotStatusAvailable):
			snapshot.StatusCode = SnapshotStatusCodeOk
		case
			string(SnapshotStatusError),
			string(SnapshotStatusErrorDeleting):
			snapshot.StatusCode = SnapshotStatusCodeError
		default:
			snapshot.StatusCode = SnapshotStatusCodeTransition
		}
	}

	return snapshots, nil
}
