package polling

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/cinder/snapshots"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
	"github.com/pkg/errors"
)

const (
	// snapshotsObjects contains name of objects representing Cinder snapshots.
	snapshotsObjects = "snapshots"
)

// GetSnapshots returns unmarshalled Snapshot structures from the raw snapshots data.
func GetSnapshots(opts *GetObjectsOpts) ([]*snapshots.Snapshot, error) {
	poller, err := NewCachePoller(NewCachePollerOpts{
		objectsPrefix: cinderPrefix,
		objectsType:   snapshotsObjects,
		client:        opts.RedisClient,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize poller")
	}

	regions, err := GetCachedRegions(poller)
	if err != nil {
		return nil, errors.Wrap(err, "error reading regions")
	}

	parsedSnapshots := []*snapshots.Snapshot{}
	for regionName, regionKey := range regions {
		rawData, err := GetRawCachedData(poller, regionKey)
		if err != nil {
			log.Errorf("can't read data from %s", regionKey)
			continue
		}

		log.Debugf(cache.RedisKeyMsgFmt, snapshotsObjects, regionKey)

		parsedData, err := snapshots.GetSnapshots(rawData, regionName)
		if err != nil {
			log.Errorf("error parsing raw snapshots data: %v", err)
			continue
		}

		log.Debugf(
			cache.ObjectsCountInRegionMsgFmt,
			len(parsedData),
			snapshotsObjects,
			regionKey,
		)
		parsedSnapshots = append(parsedSnapshots, parsedData...)
	}

	log.Debugf(cache.ObjectsCountMsgFmt, len(parsedSnapshots), snapshotsObjects)

	return parsedSnapshots, nil
}

// GetAccountsSnapshots gets snapshots count per project in all accounts.
func GetAccountsSnapshots(opts *GetObjectsOpts) ([]*AccountObjects, error) {
	snapshotsData, err := GetSnapshots(opts)
	if err != nil {
		return nil, err
	}

	projectsSnapshots := make(map[string]map[string]int)

	// Populate mapping between projects and snapshots count in each region.
	for _, snapshot := range snapshotsData {
		// Instantiate empty project snapshots count map if it doesn't exist yet.
		if _, ok := projectsSnapshots[snapshot.ProjectID]; !ok {
			projectsSnapshots[snapshot.ProjectID] = map[string]int{
				snapshot.Region: 0,
			}
		} else if _, ok := projectsSnapshots[snapshot.ProjectID][snapshot.Region]; !ok {
			// Instantiate only region map in existing projects snapshot map in
			// case that it already contains counts for some other regions.
			projectsSnapshots[snapshot.ProjectID][snapshot.Region] = 0
		}

		// Safely increment project snapshots count in a single region.
		projectsSnapshots[snapshot.ProjectID][snapshot.Region]++
	}

	return getAccountObjects(projectsSnapshots), nil
}
