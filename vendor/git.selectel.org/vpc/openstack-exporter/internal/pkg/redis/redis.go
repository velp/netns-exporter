package redis

import (
	"time"

	"github.com/go-redis/redis"
)

// ClientOpts represents parameters for the Redis client.
type ClientOpts struct {
	Address    string
	Password   string
	DB         int
	PoolSize   int
	MaxConnAge int
}

// NewRedisClient initializes new Redis client.
func NewRedisClient(opts ClientOpts) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:       opts.Address,
		Password:   opts.Password,
		DB:         opts.DB,
		PoolSize:   opts.PoolSize,
		MaxConnAge: time.Second * time.Duration(opts.MaxConnAge),
	})
}

// Get retrieves string value by a key.
// It returns an empty value without error if specified key doesn't exist.
func Get(client *redis.Client, key string) (string, error) {
	value, err := client.Get(key).Result()
	if err == redis.Nil {
		err = nil
	}

	return value, err
}

// GetKeys returns keys by provided pattern.
func GetKeys(client *redis.Client, pattern string) []string {
	return client.Keys(pattern).Val()
}

// Set writes a value to a key and sets expiration time.
func Set(client *redis.Client, key, value string, exp time.Duration) error {
	return client.Set(key, value, exp).Err()
}
