package cache

import (
	"strconv"
	"strings"
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/config"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/log"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/common"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/keystone"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/openstack/keystone/projects"
	"git.selectel.org/vpc/openstack-exporter/internal/pkg/redis"
	"github.com/pkg/errors"
)

// GetProjectsHierarchy retrieves projects hierarchy from cache.
func GetProjectsHierarchy() map[string]string {
	// Initialize Redis client.
	client := redis.NewRedisClient(redis.ClientOpts{
		Address: strings.Join([]string{
			config.Config.Redis.RedisAddress,
			strconv.Itoa(config.Config.Redis.RedisPort),
		}, ":"),
		Password: config.Config.Redis.RedisPassword,
		DB:       config.Config.Redis.RedisDatabase,
	})

	// Get cached projects keys.
	cachePrefix := config.Config.OpenStack.CacheKeyPrefix
	keysPattern := cachePrefix + "*"
	keys := GetKeys(client, keysPattern)

	// Populate projects hierarchy.
	hierarchy := make(map[string]string, len(keys))
	for _, key := range keys {
		topLevelProjectName, err := redis.Get(client, key)
		if err != nil {
			log.Errorf("error getting cached data for %s: %s", key, err)
			continue
		}
		if topLevelProjectName == "" {
			// Skip empty top-level project name.
			continue
		}
		hierarchy[strings.TrimPrefix(key, cachePrefix)] = topLevelProjectName
	}

	return hierarchy
}

// RefreshProjectsHierarchy populates top level and basic projects hierarchy
// mapping in cache.
func RefreshProjectsHierarchy() error {
	hierarchy, err := getHierarchy()
	if err != nil {
		return errors.Wrap(err, "error getting projects data from OpenStack")
	}

	cacheHierarchy(hierarchy)

	return nil
}

func getHierarchy() (map[string]string, error) {
	// Authenticate OpenStack provider.
	provider, err := openstack.NewProviderClient(&openstack.ClientOpts{
		AuthURL:     config.Config.OpenStack.Keystone.AuthURL,
		Username:    config.Config.OpenStack.Keystone.Username,
		Password:    config.Config.OpenStack.Keystone.Password,
		ProjectName: config.Config.OpenStack.Keystone.ProjectName,
		DomainName:  config.Config.OpenStack.Keystone.DomainName,
		Region:      config.Config.OpenStack.Keystone.Region,
	})
	if err != nil {
		return nil, errors.Wrap(err, "error authenticating in OpenStack")
	}

	// Initialize Keystone V3 client to retrieve projects data from API.
	keystoneClient, err := keystone.NewKeystoneV3Client(&common.NewClientOpts{
		Provider:     provider,
		EndpointType: config.Config.OpenStack.EndpointType,
		Timeout:      time.Second * time.Duration(config.Config.OpenStack.Keystone.RequestTimeout),
	})
	if err != nil {
		return nil, errors.Wrap(err, "error initializing Keystone client")
	}

	keystoneGetOpts := &common.GetOpts{
		Client:   keystoneClient,
		Region:   config.Config.OpenStack.Keystone.Region,
		Attempts: config.Config.OpenStack.Keystone.RequestAttempts,
		Interval: time.Second * time.Duration(config.Config.OpenStack.Keystone.RequestInterval),
	}

	// Get basic projects data.
	basicProjects, err := projects.GetProjects(keystoneGetOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error getting Keystone projects")
	}

	// Get top-level projects data.
	rawTopLevelProjects, err := projects.GetTopLevelProjects(keystoneGetOpts)
	if err != nil {
		return nil, errors.Wrap(err, "error getting Keystone top-level projects")
	}

	return projects.GetProjectsHierarchy(
		basicProjects,
		projects.FlattenTopLevelProjects(rawTopLevelProjects),
	), nil
}

func cacheHierarchy(hierarchy map[string]string) {
	// Initialize Redis client.
	client := redis.NewRedisClient(redis.ClientOpts{
		Address: strings.Join([]string{
			config.Config.Redis.RedisAddress,
			strconv.Itoa(config.Config.Redis.RedisPort),
		}, ":"),
		Password: config.Config.Redis.RedisPassword,
		DB:       config.Config.Redis.RedisDatabase,
	})

	// Cache projects hierarchy.
	for projectID, topLevelProjectName := range hierarchy {
		err := redis.Set(
			client,
			config.Config.OpenStack.CacheKeyPrefix+projectID,
			topLevelProjectName,
			time.Second*time.Duration(config.Config.OpenStack.CacheExpiration),
		)
		if err != nil {
			log.Errorf(
				"error caching account name for project %s: %s",
				projectID,
				err,
			)
		}
	}
}
