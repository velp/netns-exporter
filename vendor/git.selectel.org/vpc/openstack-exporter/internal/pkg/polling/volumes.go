package polling

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/cinder/volumes"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
	"github.com/pkg/errors"
)

const (
	// volumesObjects contains name of objects representing Cinder volumes.
	volumesObjects = "volumes"
)

// GetVolumes returns unmarshalled Volume structures from the raw volumes data.
func GetVolumes(opts *GetObjectsOpts) ([]*volumes.Volume, error) {
	poller, err := NewCachePoller(NewCachePollerOpts{
		objectsPrefix: cinderPrefix,
		objectsType:   volumesObjects,
		client:        opts.RedisClient,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize poller")
	}

	regions, err := GetCachedRegions(poller)
	if err != nil {
		return nil, errors.Wrap(err, "error reading regions")
	}

	parsedVolumes := []*volumes.Volume{}
	for regionName, regionKey := range regions {
		rawData, err := GetRawCachedData(poller, regionKey)
		if err != nil {
			log.Errorf("can't read data from %s", regionKey)
			continue
		}

		log.Debugf(cache.RedisKeyMsgFmt, volumesObjects, regionKey)

		parsedData, err := volumes.GetVolumes(rawData, regionName)
		if err != nil {
			log.Errorf("error parsing raw volumes data: %v", err)
			continue
		}

		log.Debugf(
			cache.ObjectsCountInRegionMsgFmt,
			len(parsedData),
			volumesObjects,
			regionKey,
		)
		parsedVolumes = append(parsedVolumes, parsedData...)
	}

	log.Debugf(cache.ObjectsCountMsgFmt, len(parsedVolumes), volumesObjects)

	return parsedVolumes, nil
}

// GetAccountsVolumes gets volumes count per project in all accounts.
func GetAccountsVolumes(opts *GetObjectsOpts) ([]*AccountObjects, error) {
	volumesData, err := GetVolumes(opts)
	if err != nil {
		return nil, err
	}

	projectsVolumes := make(map[string]map[string]int)

	// Populate mapping between projects and volumes count in each region.
	for _, volume := range volumesData {
		// Instantiate empty project volumes count map if it doesn't exist yet.
		if _, ok := projectsVolumes[volume.ProjectID]; !ok {
			projectsVolumes[volume.ProjectID] = map[string]int{
				volume.Region: 0,
			}
		} else if _, ok := projectsVolumes[volume.ProjectID][volume.Region]; !ok {
			// Instantiate only region map in existing projects volume map in
			// case that it already contains counts for some other regions.
			projectsVolumes[volume.ProjectID][volume.Region] = 0
		}

		// Safely increment project volumes count in a single region.
		projectsVolumes[volume.ProjectID][volume.Region]++
	}

	return getAccountObjects(projectsVolumes), nil
}
