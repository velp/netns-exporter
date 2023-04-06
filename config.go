package main

import (
	"io/ioutil"
	"regexp"
	"runtime"

	yaml "gopkg.in/yaml.v2"
)

type NetnsExporterConfig struct {
	APIServer        APIServerConfig       `yaml:"api_server"`
	InterfaceMetrics []string              `yaml:"interface_metrics"`
	ProcMetrics      map[string]ProcMetric `yaml:"proc_metrics"`
	Threads          int                   `yaml:"threads"`
	NamespacesFilter NamespacesFilter      `yaml:"namespaces_filter"`
	DeviceFilter 	DeviceFilter 		   `yaml:"device_filter"`
}

type ProcMetric struct {
	FileName string `yaml:"file"`
}

type APIServerConfig struct {
	ServerAddress  string `yaml:"server_address"`
	ServerPort     int    `yaml:"server_port"`
	RequestTimeout int    `yaml:"request_timeout"`
	TelemetryPath  string `yaml:"telemetry_path"`
}

type NamespacesFilter struct {
	BlacklistPattern string `yaml:"blacklist_pattern"`
	WhitelistPattern string `yaml:"whitelist_pattern"`

	BlacklistRegexp *regexp.Regexp
	WhitelistRegexp *regexp.Regexp
}

type DeviceFilter struct {
	BlacklistPattern string `yaml:"blacklist_pattern"`
	WhitelistPattern string `yaml:"whitelist_pattern"`

	BlacklistRegexp *regexp.Regexp
	WhitelistRegexp *regexp.Regexp
}

func LoadConfig(path string) (*NetnsExporterConfig, error) {
	cfg := NetnsExporterConfig{}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	cfg.Threads = runtime.NumCPU()

	return &cfg, nil
}

func (nsFilter *NamespacesFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain NamespacesFilter

	err := unmarshal((*plain)(nsFilter))
	if err != nil {
		return err
	}

	nsFilter.BlacklistRegexp, err = regexp.Compile(nsFilter.BlacklistPattern)
	if err != nil {
		return err
	}

	nsFilter.WhitelistRegexp, err = regexp.Compile(nsFilter.WhitelistPattern)
	if err != nil {
		return err
	}

	return nil
}


func (devFilter *DeviceFilter) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain DeviceFilter

	err := unmarshal((*plain)(devFilter))
	if err != nil {
		return err
	}

	devFilter.BlacklistRegexp, err = regexp.Compile(devFilter.BlacklistPattern)
	if err != nil {
		return err
	}

	devFilter.WhitelistRegexp, err = regexp.Compile(devFilter.WhitelistPattern)
	if err != nil {
		return err
	}

	return nil
}

