package main

import (
	"io/ioutil"
	"runtime"

	yaml "gopkg.in/yaml.v2"
)

type NetnsExporterConfig struct {
	APIServer        APIServerConfig       `yaml:"api_server"`
	InterfaceMetrics []string              `yaml:"interface_metrics"`
	ProcMetrics      map[string]ProcMetric `yaml:"proc_metrics"`
	Threads          int                   `yaml:"threads"`
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
