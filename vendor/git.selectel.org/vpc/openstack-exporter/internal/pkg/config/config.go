package config

import (
	"io/ioutil"
	"log"

	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

const (
	apiTelemetryDefaultPath  = "/metrics"
	apiRequestDefaultTimeout = 30

	redisDefaultPoolSize   = 4
	redisDefaultMaxConnAge = 120

	openstackDefaultEndpointType         = "admin"
	openstackDefaultCacheKeyPrefix       = "project_"
	openstackDefaultCacheRefreshInterval = 600
	openstackDefaultCacheExpiration      = 900
	openstackDefaultRequestTimeout       = 10
	openstackDefaultRequestAttempts      = 2
	openstackDefaultRequestInterval      = 1

	openstackDefaultOctaviaDomainName  = "Default"
	openstackDefaultOctaviaProjectName = "octavia"

	gnocchiDefaultEndpointType    = "admin"
	gnocchiDefaultRequestTimeout  = 10
	gnocchiDefaultRequestAttempts = 2
	gnocchiDefaultRequestInterval = 1
)

// Config is a global container for configuration options.
var Config *OpenStackExporterConfig

// OpenStackExporterConfig contains all application parameters.
type OpenStackExporterConfig struct {
	Log       LogConfig       `yaml:"log"`
	Redis     RedisConfig     `yaml:"redis"`
	APIServer APIServerConfig `yaml:"api_server"`
	OpenStack OpenStackConfig `yaml:"openstack"`
	Gnocchi   GnocchiConfig   `yaml:"gnocchi"`
}

// LogConfig contains logger configuration.
type LogConfig struct {
	File      string `yaml:"file"`
	Debug     bool   `yaml:"debug"`
	UseStdout bool   `yaml:"use_stdout"`
}

// RedisConfig contains configuration to work with Redis server.
type RedisConfig struct {
	RedisAddress    string `yaml:"redis_address"`
	RedisPort       int    `yaml:"redis_port"`
	RedisPassword   string `yaml:"redis_password"`
	RedisDatabase   int    `yaml:"redis_database"`
	RedisPoolSize   int    `yaml:"redis_pool_size"`
	RedisMaxConnAge int    `yaml:"redis_max_conn_age"`
}

// APIServerConfig contains configuration for the Exporter API server.
type APIServerConfig struct {
	ServerAddress  string `yaml:"server_address"`
	ServerPort     int    `yaml:"server_port"`
	RequestTimeout int    `yaml:"request_timeout"`
	TelemetryPath  string `yaml:"telemetry_path"`
}

// OpenStackConfig contains parameters to communicate with the OpenStack API and
// to cache OpenStack-specific data into Redis.
type OpenStackConfig struct {
	EndpointType string `yaml:"endpoint_type"`

	CacheKeyPrefix       string `yaml:"cache_key_prefix"`
	CacheRefreshInterval int    `yaml:"cache_refresh_interval"`
	CacheExpiration      int    `yaml:"cache_expiration"`

	Keystone KeystoneConfig `yaml:"keystone"`
	Nova     NovaConfig     `yaml:"nova"`
	Neutron  NeutronConfig  `yaml:"neutron"`
	Octavia  OctaviaConfig  `yaml:"octavia"`
	Magnum   MagnumConfig   `yaml:"magnum"`
}

// KeystoneConfig contains Keystone-specific parameters.
type KeystoneConfig struct {
	AuthURL         string `yaml:"auth_url"`
	Username        string `yaml:"username"`
	Password        string `yaml:"password"`
	ProjectName     string `yaml:"project_name"`
	DomainName      string `yaml:"domain_name"`
	Region          string `yaml:"region"`
	RequestTimeout  int    `yaml:"request_timeout"`
	RequestAttempts int    `yaml:"request_attempts"`
	RequestInterval int    `yaml:"request_interval"`
}

// NovaConfig contains Nova-specific parameters.
type NovaConfig struct {
	Regions         []string `yaml:"regions"`
	RequestTimeout  int      `yaml:"request_timeout"`
	RequestAttempts int      `yaml:"request_attempts"`
	RequestInterval int      `yaml:"request_interval"`
}

// NeutronConfig contains Neutron-specific parameters.
type NeutronConfig struct {
	Regions         []string `yaml:"regions"`
	RequestTimeout  int      `yaml:"request_timeout"`
	RequestAttempts int      `yaml:"request_attempts"`
	RequestInterval int      `yaml:"request_interval"`
}

// OctaviaConfig contains Octavia-specific parameters.
type OctaviaConfig struct {
	DomainName      string   `yaml:"domain_name"`
	ProjectName     string   `yaml:"project_name"`
	Regions         []string `yaml:"regions"`
	RequestTimeout  int      `yaml:"request_timeout"`
	RequestAttempts int      `yaml:"request_attempts"`
	RequestInterval int      `yaml:"request_interval"`
}

// MagnumConfig contains Magnum-specific parameters.
type MagnumConfig struct {
	Regions         []string `yaml:"regions"`
	RequestTimeout  int      `yaml:"request_timeout"`
	RequestAttempts int      `yaml:"request_attempts"`
	RequestInterval int      `yaml:"request_interval"`
}

// GnocchiConfig contains parameters to communicate with the Gnocchi API.
type GnocchiConfig struct {
	EndpointType    string   `yaml:"endpoint_type"`
	Regions         []string `yaml:"regions"`
	RequestTimeout  int      `yaml:"request_timeout"`
	RequestAttempts int      `yaml:"request_attempts"`
	RequestInterval int      `yaml:"request_interval"`
}

// InitFromFile reads config file and initializes global config.
func InitFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	if err := InitFromString(data); err != nil {
		return err
	}

	log.Printf("Config loaded from %v.", path)
	return nil
}

// InitFromString reads raw string and initializes global config.
func InitFromString(data []byte) error {
	cfg := OpenStackExporterConfig{}
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Set default API server settings if omitted.
	setDefaultStr(&cfg.APIServer.TelemetryPath, apiTelemetryDefaultPath)
	setDefaultInt(&cfg.APIServer.RequestTimeout, apiRequestDefaultTimeout)

	// Set default Redis settings if omitted.
	setDefaultInt(&cfg.Redis.RedisPoolSize, redisDefaultPoolSize)
	setDefaultInt(&cfg.Redis.RedisMaxConnAge, redisDefaultMaxConnAge)

	// Set default OpenStack common settings if omitted.
	setDefaultStr(&cfg.OpenStack.EndpointType, openstackDefaultEndpointType)
	setDefaultStr(&cfg.OpenStack.CacheKeyPrefix, openstackDefaultCacheKeyPrefix)
	setDefaultInt(&cfg.OpenStack.CacheExpiration, openstackDefaultCacheExpiration)
	setDefaultInt(&cfg.OpenStack.CacheRefreshInterval, openstackDefaultCacheRefreshInterval)

	// Set default Keystone API settings if omitted.
	setDefaultInt(&cfg.OpenStack.Keystone.RequestTimeout, openstackDefaultRequestTimeout)
	setDefaultInt(&cfg.OpenStack.Keystone.RequestAttempts, openstackDefaultRequestAttempts)
	setDefaultInt(&cfg.OpenStack.Keystone.RequestInterval, openstackDefaultRequestInterval)

	// Set default Nova API settings if omitted.
	setDefaultInt(&cfg.OpenStack.Nova.RequestTimeout, openstackDefaultRequestTimeout)
	setDefaultInt(&cfg.OpenStack.Nova.RequestAttempts, openstackDefaultRequestAttempts)
	setDefaultInt(&cfg.OpenStack.Nova.RequestInterval, openstackDefaultRequestInterval)

	// Set default Neutron API settings if omitted.
	setDefaultInt(&cfg.OpenStack.Neutron.RequestTimeout, openstackDefaultRequestTimeout)
	setDefaultInt(&cfg.OpenStack.Neutron.RequestAttempts, openstackDefaultRequestAttempts)
	setDefaultInt(&cfg.OpenStack.Neutron.RequestInterval, openstackDefaultRequestInterval)

	// Set default Magnum API settings if omitted.
	setDefaultInt(&cfg.OpenStack.Magnum.RequestTimeout, openstackDefaultRequestTimeout)
	setDefaultInt(&cfg.OpenStack.Magnum.RequestAttempts, openstackDefaultRequestAttempts)
	setDefaultInt(&cfg.OpenStack.Magnum.RequestInterval, openstackDefaultRequestInterval)

	// Set default Octavia API settings if omitted.
	setDefaultStr(&cfg.OpenStack.Octavia.DomainName, openstackDefaultOctaviaDomainName)
	setDefaultStr(&cfg.OpenStack.Octavia.ProjectName, openstackDefaultOctaviaProjectName)
	setDefaultInt(&cfg.OpenStack.Octavia.RequestTimeout, openstackDefaultRequestTimeout)
	setDefaultInt(&cfg.OpenStack.Octavia.RequestAttempts, openstackDefaultRequestAttempts)
	setDefaultInt(&cfg.OpenStack.Octavia.RequestInterval, openstackDefaultRequestInterval)

	// Set default Gnocchi API settings if omitted.
	setDefaultStr(&cfg.Gnocchi.EndpointType, gnocchiDefaultEndpointType)
	setDefaultInt(&cfg.Gnocchi.RequestTimeout, gnocchiDefaultRequestTimeout)
	setDefaultInt(&cfg.Gnocchi.RequestAttempts, gnocchiDefaultRequestAttempts)
	setDefaultInt(&cfg.Gnocchi.RequestInterval, gnocchiDefaultRequestInterval)

	Config = &cfg
	return nil
}

// CheckConfig helps to check if global application config is ready.
func CheckConfig() error {
	if Config == nil {
		return errors.New("global configuration file is not initialized")
	}
	return nil
}

func setDefaultInt(configParameter *int, defaultValue int) {
	if *configParameter <= 0 {
		*configParameter = defaultValue
	}
}

func setDefaultStr(configParameter *string, defaultValue string) {
	if *configParameter == "" {
		*configParameter = defaultValue
	}
}
