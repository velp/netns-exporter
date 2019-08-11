package polling

import (
	"fmt"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
	"github.com/go-redis/redis"
	goredis "github.com/go-redis/redis"
	"github.com/gophercloud/gophercloud"
	"github.com/pkg/errors"
)

const (
	rootPrefix = "metrics"

	// cinderPrefix contains top-level key prefix for all Cinder metrics.
	cinderPrefix = rootPrefix + "/cinder"

	// novaPrefix contains top-level key prefix for all Nova metrics.
	novaPrefix = rootPrefix + "/nova"

	// octaviaPrefix contains top-level key prefix for all Octavia metrics.
	octaviaPrefix = rootPrefix + "/octavia"
)

// GetObjectsOpts contains common options to work with all polling methods.
type GetObjectsOpts struct {
	CloudClients map[string]gophercloud.ServiceClient
	RedisClient  *redis.Client
}

// CachePoller represents special struct that can be used to poll data from
// the cache.
type CachePoller struct {
	// allObjectsCmd is used to retrieve all selected objects keys from the cache.
	allObjectsCmd string

	// objectsTimestampsCmd is used to retrieve all timestamps of the selected
	// object from the cache.
	objectsTimestampsCmd string

	// objectsCmd is used to retrieve to retrieve raw data for the object with a
	// specific timestamp.
	objectsCmd string

	// client contains initialized client to work with cache.
	client *goredis.Client

	// objectsType contains plural name of cached objects.
	objectsType string
}

// NewCachePollerOpts contains options to instantiate a new CachePoller.
type NewCachePollerOpts struct {
	// objectsPrefix contains key prefix for the selected cached objects.
	// Examples: "/metrics/cinder", "/metrics/nova", etc.
	objectsPrefix string

	// objectsType represents plural name of cached objects.
	// Examples: "volumes", "instances", "loadbalancers", etc.
	objectsType string

	// client contains initialized client to work with cache.
	client *goredis.Client
}

// NewCachePoller instantiates a new CachePoller and returns its reference.
func NewCachePoller(opts NewCachePollerOpts) (*CachePoller, error) {
	if err := config.CheckConfig(); err != nil {
		return nil, errors.Wrap(err, "error reading config")
	}

	cachePoller := &CachePoller{
		allObjectsCmd:        opts.objectsPrefix + "/*/" + opts.objectsType + "/*",
		objectsTimestampsCmd: opts.objectsPrefix + "/%s/" + opts.objectsType + "/*",
		objectsCmd:           opts.objectsPrefix + "/%s/" + opts.objectsType + "/%s",
		client:               opts.client,
		objectsType:          opts.objectsType,
	}

	return cachePoller, nil
}

// GetCachedRegions retrieves region keys for cached objects.
// It will use latest timestamp if there are many listings for some region.
func GetCachedRegions(poller *CachePoller) (map[string]string, error) {
	allKeys := cache.GetKeys(poller.client, poller.allObjectsCmd)
	if len(allKeys) == 0 {
		return nil, fmt.Errorf("can't get any %s data from cache", poller.objectsType)
	}

	regionsNames, err := cache.GetAvailableRegions(allKeys)
	if err != nil {
		return nil, errors.Wrap(err, "error reading regions")
	}

	regions := make(map[string]string, len(regionsNames))

	for _, regionName := range regionsNames {
		// Retrieve available timestamps for a single region.
		regionObjectsTimestamps, err := cache.GetKeyTimestamps(
			poller.client,
			fmt.Sprintf(poller.objectsTimestampsCmd, regionName),
		)
		if err != nil {
			return nil, errors.Wrap(err, "error reading timestamps")
		}

		// Parse available timestamps and get latest.
		regionObjectsParsedTimestamps, err := cache.GetParsedTimestamps(regionObjectsTimestamps)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing timestamps")
		}
		latestTimestamp, err := cache.GetLatestTimestamp(regionObjectsParsedTimestamps)
		if err != nil {
			return nil, errors.Wrap(err, "error getting latest timestamp")
		}

		// Add full cache key that can be used for getting raw data.
		regions[regionName] = fmt.Sprintf(poller.objectsCmd, regionName, latestTimestamp)
	}

	return regions, nil
}

// GetRawCachedData will get raw listings of objects from a single cache key.
func GetRawCachedData(poller *CachePoller, key string) ([]byte, error) {
	data, err := cache.GetObjects(poller.client, key)
	if err != nil {
		return nil, errors.Wrapf(err, "can't retrieve objects from %s", key)
	}

	return data, nil
}
