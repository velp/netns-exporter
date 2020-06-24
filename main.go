package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sirupsen/logrus"
)

const defaultConfigFile = "/etc/netns-exporter/config.yaml"

var (
	cfgPath     string
	logFilePath string
	logLevel    string
	threads     int
)

func init() { //nolint:gochecknoinits
	flag.StringVar(&cfgPath, "config", defaultConfigFile, fmt.Sprintf("Path to config file (default: %s)", defaultConfigFile))
	flag.StringVar(&logFilePath, "log-file", "", "Write logs to file (default: send logs to stdout)")
	flag.StringVar(&logLevel, "log-level", "info", "Logging level (default: info)")
	flag.IntVar(&threads, "threads", 0, "Number of threads for collecting data (default: number of CPU cors)")
}

func main() {
	flag.Parse()
	// Init logger
	logger := logrus.New()
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		log.Fatalf("Parsing log level failed: %s", err)
	}
	logger.SetLevel(level)
	if logFilePath != "" {
		logFile, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0755)
		if err != nil {
			log.Fatalf("Failed to initialize log file %s", err)
		}
		logger.SetOutput(logFile)
	} else {
		logger.SetOutput(os.Stdout)
	}
	// Load config
	config, err := LoadConfig(cfgPath)
	if err != nil {
		logger.Fatalf("Loading config failed: %s", err)
	}
	if threads > 0 {
		config.Threads = threads
	}
	// Run exporter
	apiServer, err := NewAPIServer(config, logger)
	if err != nil {
		logger.Fatalf("Creating API server failed: %s", err)
	}
	if err := apiServer.Start(); err != nil {
		logger.Fatalf("Starting API server failed: %s", err)
	}
}
