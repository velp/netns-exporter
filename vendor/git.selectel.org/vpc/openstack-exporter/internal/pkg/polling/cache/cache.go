package cache

import (
	"fmt"
	"strings"
	"time"

	"git.selectel.org/vpc/openstack-exporter/internal/pkg/redis"
	goredis "github.com/go-redis/redis"
	"github.com/pkg/errors"
)

const (
	cacheTimestampFormat = "2006-01-02T15:04:05"
	validKeyPartsCount   = 5
	regionPartIndex      = 2
	errCacheKeyUsage     = "error using cache key"

	// RedisKeyMsgFmt represents message format that can be used for debugging
	// selected Redis key.
	RedisKeyMsgFmt = "Using the following Redis key for %s: '%s'"

	// ObjectsCountInRegionMsgFmt represents message format that can be used for
	// debugging count of parsed objects from a single key.
	ObjectsCountInRegionMsgFmt = "Got %d %s from Redis key '%s'"

	// ObjectsCountMsgFmt represents message format that can be used for
	// debugging total count of parsed objects.
	ObjectsCountMsgFmt = "Got %d %s from Redis"
)

// GetObjects retrieves raw byteslice representation of needed objects from
// Redis specified by key.
func GetObjects(client *goredis.Client, key string) ([]byte, error) {
	objects, err := redis.Get(client, key)
	if err != nil {
		return nil, errors.Wrap(err, "error getting data from cache")
	}

	return []byte(objects), nil
}

// GetKeys retrieves available keys from Redis specified by pattern.
func GetKeys(client *goredis.Client, pattern string) []string {
	return redis.GetKeys(client, pattern)
}

// GetAvailableRegions gets keys in common pattern and retrieves available
// region names.
func GetAvailableRegions(keys []string) ([]string, error) {
	foundRegions := make(map[string]bool)

	// Build map with unique region names.
	for _, key := range keys {
		keyParts := strings.Split(key, "/")
		if err := checkKeyLength(keyParts); err != nil {
			return nil, errors.Wrap(err, errCacheKeyUsage)
		}
		region := keyParts[regionPartIndex]
		if _, ok := foundRegions[region]; !ok {
			foundRegions[region] = true
		}
	}

	// Convert regions map to a slice.
	regions := make([]string, len(foundRegions))
	var i int
	for region := range foundRegions {
		regions[i] = region
		i++
	}

	return regions, nil
}

// GetKeyTimestamps retrieves available timestamps for a single Redis key.
func GetKeyTimestamps(client *goredis.Client, pattern string) ([]string, error) {
	keys := redis.GetKeys(client, pattern)

	timestamps := make([]string, len(keys))
	for i, key := range keys {
		keyParts := strings.Split(key, "/")
		if err := checkKeyLength(keyParts); err != nil {
			return nil, errors.Wrap(err, errCacheKeyUsage)
		}
		timestamps[i] = keyParts[4]
	}

	return timestamps, nil
}

// GetParsedTimestamps returns slice of timestamps in Go's Time format from
// their string representation.
func GetParsedTimestamps(timestamps []string) ([]time.Time, error) {
	if len(timestamps) == 0 {
		return nil, errors.New("provided timestamps slice is empty")
	}

	parsedTimestamps := make([]time.Time, len(timestamps))

	for i, timestamp := range timestamps {
		parsedTimestamp, err := time.Parse(cacheTimestampFormat, timestamp)
		if err != nil {
			return nil, errors.Wrap(err, "error parsing timestamp")
		}
		parsedTimestamps[i] = parsedTimestamp
	}

	return parsedTimestamps, nil
}

// GetLatestTimestamp gets slice of timestamps in Go's Time format and returns
// latest in it's string representation.
func GetLatestTimestamp(parsedTimestamps []time.Time) (string, error) {
	if len(parsedTimestamps) == 0 {
		return "", errors.New("provided parsedTimestamps slice is empty")
	}

	latestParsedTimeStamp := parsedTimestamps[0]

	for _, parsedTimestamp := range parsedTimestamps {
		if parsedTimestamp.After(latestParsedTimeStamp) {
			latestParsedTimeStamp = parsedTimestamp
		}
	}

	latestTimestamp := latestParsedTimeStamp.Format(cacheTimestampFormat)
	return latestTimestamp, nil
}

func checkKeyLength(keyParts []string) error {
	if len(keyParts) != validKeyPartsCount {
		return fmt.Errorf("key: %s is invalid, want %d separate parts, got %d",
			strings.Join(keyParts, "/"),
			validKeyPartsCount,
			len(keyParts),
		)
	}

	return nil
}
