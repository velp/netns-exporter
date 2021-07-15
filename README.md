# Netns exporter

Prometheus exporter for monitoring interface statistics and /proc filesystem statistics in each Linux network namespace.

**⚠️The exporter should be run with root privileges on the host system or have capabilities to mange network namespaces.⚠️**

## Install
Download [releases](https://github.com/velp/netns-exporter/releases).

## Build
 For Linux system:

 ```shell
 make build
 ```

 For MacOS and other UNIX-like systems:
 ```shell
 make build-in-docker
 ```

Docker Image
 ```shell
 make docker-image
 ```

## Example
For example, for two Linux network namespaces like:

```
# ip netns
test2 (id: 1)
test (id: 0)

# ip netns exec test ip a
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
3: kokoko: <BROADCAST,NOARP> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether 9e:2e:82:3f:38:3a brd ff:ff:ff:ff:ff:ff

# ip netns exec test2 ip a
1: lo: <LOOPBACK> mtu 65536 qdisc noop state DOWN group default qlen 1000
    link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
4: qwerty: <BROADCAST,NOARP> mtu 1500 qdisc noop state DOWN group default qlen 1000
    link/ether ce:b3:45:ef:b3:a6 brd ff:ff:ff:ff:ff:ff
```

and with config file:

```
...
interface_metrics: ["rx_bytes", "rx_packets", "rx_dropped", "tx_bytes", "tx_packets"]
proc_metrics:
  nat_conntrack:
    file: sys/net/netfilter/nf_conntrack_count
```

the exporter will return metrics like:

```
...
# HELP netns_exporter_build_info A metric with a constant '1' value labeled by version, revision, branch, and goversion from which netns_exporter was built.
# TYPE netns_exporter_build_info gauge
netns_exporter_build_info{branch="",goversion="go1.12.7",revision="",version=""} 1
# HELP netns_network_nat_conntrack_total Statistics from /proc filesystem in the network namespace
# TYPE netns_network_nat_conntrack_total counter
netns_network_nat_conntrack_total{netns="test"} 0
netns_network_nat_conntrack_total{netns="test2"} 0
# HELP netns_network_rx_bytes_total Interface statistics in the network namespace
# TYPE netns_network_rx_bytes_total counter
netns_network_rx_bytes_total{device="kokoko",netns="test"} 0
netns_network_rx_bytes_total{device="qwerty",netns="test2"} 0
# HELP netns_network_rx_dropped_total Interface statistics in the network namespace
# TYPE netns_network_rx_dropped_total counter
netns_network_rx_dropped_total{device="kokoko",netns="test"} 0
netns_network_rx_dropped_total{device="qwerty",netns="test2"} 0
# HELP netns_network_rx_packets_total Interface statistics in the network namespace
# TYPE netns_network_rx_packets_total counter
netns_network_rx_packets_total{device="kokoko",netns="test"} 0
netns_network_rx_packets_total{device="qwerty",netns="test2"} 0
# HELP netns_network_tx_bytes_total Interface statistics in the network namespace
# TYPE netns_network_tx_bytes_total counter
netns_network_tx_bytes_total{device="kokoko",netns="test"} 0
netns_network_tx_bytes_total{device="qwerty",netns="test2"} 0
# HELP netns_network_tx_packets_total Interface statistics in the network namespace
# TYPE netns_network_tx_packets_total counter
netns_network_tx_packets_total{device="kokoko",netns="test"} 0
netns_network_tx_packets_total{device="qwerty",netns="test2"} 0
...
```
Also, netns-exporter provides an optional ability to filter namespaces by regular expression:
```
...
namespaces_filter:
  blacklist_pattern: regexp_pattern1
  whitelist_pattern: regexp_pattern2
```
With the simultaneous use of two filters, the blacklist filter has a higher priority.

## Run in Docker
```bash
docker run --name netns-exporter -d -p8080:8080 --privileged --mount type=bind,source="/run/netns,target=/run/netns,bind-propagation=slave" netns-exporter
```

## Contribution
Want to contribute! That's awesome! Check out [CONTRIBUTING documentation](https://github.com/jexia/jexia-cli/blob/master/CONTRIBUTING.rst).

