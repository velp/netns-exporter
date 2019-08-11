package polling

import (
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/nova/instances"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
)

const (
	// instancesObjects contains name of objects representing Nova instances.
	instancesObjects = "instances"
)

// GetInstances retrieves Nova instances from the raw cached data and from Nova
// API.
// It will skip cached regions that configured as API regions in OpenStack
// Exporter configuration file.
func GetInstances(opts *GetObjectsOpts) []*instances.Instance {
	poller, err := NewCachePoller(NewCachePollerOpts{
		objectsPrefix: novaPrefix,
		objectsType:   instancesObjects,
		client:        opts.RedisClient,
	})
	if err != nil {
		log.Errorf("unable to initialize Nova cache poller: %s", err)
	}

	cachedRegions, err := GetCachedRegions(poller)
	if err != nil {
		log.Errorf("can't read Nova regions from cache: %s", err)
	}

	for clientsRegion := range opts.CloudClients {
		// Remove region that will be populated from the API instead of cache
		// from the cachedRegions map.
		delete(cachedRegions, clientsRegion)
	}

	allInstances := []*instances.Instance{}

	// Parse instances from cache first.
	for regionName, regionKey := range cachedRegions {
		rawData, err := GetRawCachedData(poller, regionKey)
		if err != nil {
			log.Errorf("can't read data from %s", regionKey)
			continue
		}

		log.Debugf(cache.RedisKeyMsgFmt, instancesObjects, regionKey)

		parsedData, err := instances.GetInstancesFromCache(rawData, regionName)
		if err != nil {
			log.Errorf("error parsing raw instances data: %v", err)
			continue
		}

		log.Debugf(
			cache.ObjectsCountInRegionMsgFmt,
			len(parsedData),
			instancesObjects,
			regionKey,
		)
		allInstances = append(allInstances, parsedData...)
	}

	log.Debugf(cache.ObjectsCountMsgFmt, len(allInstances), instancesObjects)

	// Retrieve instances from API.
	for region := range opts.CloudClients {
		client := opts.CloudClients[region]
		regionInstances, err := instances.GetInstancesFromAPI(&common.GetOpts{
			Region:   region,
			Client:   &client,
			Attempts: config.Config.OpenStack.Nova.RequestAttempts,
			Interval: time.Second * time.Duration(config.Config.OpenStack.Nova.RequestInterval),
		})
		if err != nil {
			log.Errorf("error getting Nova instances from '%s': %v",
				client.ResourceBase, err)
			continue
		}

		log.Debugf("Got %d instances from %s region API", len(regionInstances), region)
		allInstances = append(allInstances, regionInstances...)
	}

	log.Debugf("Got %d %s from cache and API", len(allInstances), instancesObjects)

	return allInstances
}

// GetAccountsInstances gets instances count per project in all accounts.
func GetAccountsInstances(opts *GetObjectsOpts) []*AccountObjects {
	projectsInstances := make(map[string]map[string]int)

	// Populate mapping between projects and instances count in each region.
	for _, instance := range GetInstances(opts) {
		// Instantiate empty project instances count map if it doesn't exist yet.
		if _, ok := projectsInstances[instance.ProjectID]; !ok {
			projectsInstances[instance.ProjectID] = map[string]int{
				instance.Region: 0,
			}
		} else if _, ok := projectsInstances[instance.ProjectID][instance.Region]; !ok {
			// Instantiate only region map in existing projects instance map in
			// case that it already contains counts for some other regions.
			projectsInstances[instance.ProjectID][instance.Region] = 0
		}

		// Safely increment project instances count in a single region.
		projectsInstances[instance.ProjectID][instance.Region]++
	}

	return getAccountObjects(projectsInstances)
}
