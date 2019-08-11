package polling

import (
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/octavia/loadbalancers"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/polling/cache"
	"github.com/pkg/errors"
)

const (
	// loadbalancersObjects contains name of objects representing Octavia
	// loadbalancers.
	loadbalancersObjects = "loadbalancers"
)

// GetLoadBalancers returns unmarshalled LoadBalancer structures from the raw
// loadbalancers data.
func GetLoadBalancers(opts *GetObjectsOpts) ([]*loadbalancers.LoadBalancer, error) {
	poller, err := NewCachePoller(NewCachePollerOpts{
		objectsPrefix: octaviaPrefix,
		objectsType:   loadbalancersObjects,
		client:        opts.RedisClient,
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize poller")
	}

	regions, err := GetCachedRegions(poller)
	if err != nil {
		return nil, errors.Wrap(err, "error reading regions")
	}

	parsedLoadBalancers := []*loadbalancers.LoadBalancer{}
	for regionName, regionKey := range regions {
		rawData, err := GetRawCachedData(poller, regionKey)
		if err != nil {
			log.Errorf("can't read data from %s", regionKey)
			continue
		}

		log.Debugf(cache.RedisKeyMsgFmt, loadbalancersObjects, regionKey)

		parsedData, err := loadbalancers.GetLoadBalancers(rawData, regionName)
		if err != nil {
			log.Errorf("error parsing raw loadbalancers data: %v", err)
			continue
		}

		log.Debugf(
			cache.ObjectsCountInRegionMsgFmt,
			len(parsedData),
			loadbalancersObjects,
			regionKey,
		)
		parsedLoadBalancers = append(parsedLoadBalancers, parsedData...)
	}

	log.Debugf(cache.ObjectsCountMsgFmt, len(parsedLoadBalancers), loadbalancersObjects)

	return parsedLoadBalancers, nil
}

// GetAccountsLoadBalancers gets loadbalancers count per project in all accounts.
func GetAccountsLoadBalancers(opts *GetObjectsOpts) ([]*AccountObjects, error) {
	loadbalancersData, err := GetLoadBalancers(opts)
	if err != nil {
		return nil, err
	}

	projectsLoadBalancers := make(map[string]map[string]int)

	// Populate mapping between projects and Loadbalancers count in each region.
	for _, Loadbalancer := range loadbalancersData {
		// Instantiate empty project Loadbalancers count map if it doesn't exist yet.
		if _, ok := projectsLoadBalancers[Loadbalancer.ProjectID]; !ok {
			projectsLoadBalancers[Loadbalancer.ProjectID] = map[string]int{
				Loadbalancer.Region: 0,
			}
		} else if _, ok := projectsLoadBalancers[Loadbalancer.ProjectID][Loadbalancer.Region]; !ok {
			// Instantiate only region map in existing projects Loadbalancer map in
			// case that it already contains counts for some other regions.
			projectsLoadBalancers[Loadbalancer.ProjectID][Loadbalancer.Region] = 0
		}

		// Safely increment project Loadbalancers count in a single region.
		projectsLoadBalancers[Loadbalancer.ProjectID][Loadbalancer.Region]++
	}

	return getAccountObjects(projectsLoadBalancers), nil
}
