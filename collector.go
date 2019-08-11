package main

import (
	"io/ioutil"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

const (
	NetnsPath         = "/run/netns/"
	InterfaceStatPath = "/sys/devices/virtual/net/"
	ProcStatPath      = "/proc/"

	collectorNamespace = "netns"
	collectorSubsystem = "network"
	netnsLabel         = "netns"
	deviceLabel        = "device"
)

type Collector struct {
	logger      logrus.FieldLogger
	config      *NetnsExporterConfig
	metricDescs map[string]*prometheus.Desc
	mu          sync.Mutex
}

func NewCollector(config *NetnsExporterConfig, logger *logrus.Logger) *Collector {
	return &Collector{
		logger:      logger.WithField("component", "collector"),
		config:      config,
		metricDescs: map[string]*prometheus.Desc{},
	}
}

func (c *Collector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.metricDescs {
		ch <- desc
	}
}

func (c *Collector) Collect(ch chan<- prometheus.Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if err := c.getMetricsFromAllNamespaces(ch); err != nil {
		c.logger.Errorf("Get metrics from all of namespaces failed: %s", err)
	}
}

func (c *Collector) getMetricsFromAllNamespaces(ch chan<- prometheus.Metric) error {
	// Lock the OS Thread so we don't accidentally switch namespaces
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Get namespases files
	nsFiles, err := ioutil.ReadDir(NetnsPath)
	if err != nil {
		c.logger.Errorf("Reading list of network nemaspaces failed: %s", err)
		return err
	}

	// Get metrics from all of namespaces
	for _, ns := range nsFiles {
		c.logger.Debugf("Start getting statistics for namespace %s", ns.Name())
		if err := c.getMetricsFromNamespace(ns.Name(), ch); err != nil {
			c.logger.Errorf("Getting metrics for namespace %s failed: %s", ns.Name(), err)
		}
	}
	return nil
}

func (c *Collector) getMetricsFromNamespace(namespace string, ch chan<- prometheus.Metric) error {
	// Save current namespace
	curNs, err := netns.Get()
	if err != nil {
		c.logger.Errorf("Get current namespace failed: %s", err)
		return err
	}
	defer curNs.Close()
	defer netns.Set(curNs) //nolint:errcheck

	// Switch namespace
	ns, err := netns.GetFromName(namespace)
	if err != nil {
		c.logger.Errorf("Get net namespace by name %s failed: %s", namespace, err)
		return err
	}
	if err := netns.Set(ns); err != nil {
		c.logger.Errorf("Change net namespace failed: %s", err)
		return err
	}
	defer ns.Close()

	// Say to the kernel that we will use separate  context
	if err := syscall.Unshare(syscall.CLONE_NEWNS); err != nil {
		c.logger.Errorf("Syscall unshare failed: %s", err)
		return err
	}

	// Don't let any mounts propagate back to the parent
	// See: https://github.com/shemminger/iproute2/blob/6754e1d9783458550dce8d309efb4091ec8089a5/lib/namespace.c#L77
	// and: https://www.kernel.org/doc/Documentation/filesystems/sharedsubtree.txt
	if err := syscall.Mount("", "/", "none", syscall.MS_SLAVE|syscall.MS_REC, ""); err != nil {
		c.logger.Errorf("Mount root with rslave option failed: %s", err)
		return err
	}

	// Mount sysfs from net nemaspace
	if err := syscall.Mount(namespace, "/sys", "sysfs", 0, "ro"); err != nil {
		c.logger.Errorf("Mount /sys from the namespace failed: %s", err)
		return err
	}
	defer syscall.Unmount("/sys", syscall.MNT_DETACH) //nolint:errcheck

	// Parse interfaces statistics
	ifFiles, err := ioutil.ReadDir(InterfaceStatPath)
	if err != nil {
		c.logger.Errorf("Reading sysfs directory for interface failed: %s", err)
		return err
	}
	for _, ifFile := range ifFiles {
		// We don't need to get stat for lo interface
		if ifFile.Name() == "lo" {
			continue
		}
		c.logger.Debugf("Start getting statistics for interface %s", ifFile.Name())
		for _, metric := range c.config.InterfaceMetrics {
			key := namespace + ifFile.Name() + metric
			desc, ok := c.metricDescs[key]
			if !ok {
				// Create a new metric
				desc = prometheus.NewDesc(
					prometheus.BuildFQName(collectorNamespace, collectorSubsystem, metric+"_total"),
					"Interface statistics in the network namespace",
					[]string{netnsLabel, deviceLabel},
					nil,
				)
				c.metricDescs[key] = desc
			}
			value := c.getMetricFromFile(InterfaceStatPath + ifFile.Name() + "/statistics/" + metric)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, namespace, ifFile.Name())
		}
	}

	// Parse of /proc statistics
	for metricName, metric := range c.config.ProcMetrics {
		key := namespace + metricName
		desc, ok := c.metricDescs[key]
		if !ok {
			// Create a new metric
			desc = prometheus.NewDesc(
				prometheus.BuildFQName(collectorNamespace, collectorSubsystem, metricName+"_total"),
				"Statistics from /proc filesystem in the network namespace",
				[]string{netnsLabel},
				nil,
			)
			c.metricDescs[key] = desc
		}
		value := c.getMetricFromFile(ProcStatPath + metric.FileName)
		ch <- prometheus.MustNewConstMetric(desc, prometheus.CounterValue, value, namespace)
	}

	return nil
}

func (c *Collector) getMetricFromFile(file string) float64 {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		c.logger.Errorf("Error while reading statistic file %s: %s", file, err)
		return -1
	}
	stat, err := strconv.ParseFloat(strings.TrimSpace(string(data)), 64)
	if err != nil {
		c.logger.Printf("Error while parsing data from file %s: %s", file, err)
		return -1
	}
	return stat
}
