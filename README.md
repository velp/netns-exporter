# Netns prometheus exporter

Prometheus exporter to monitoring interface statistics and statistics in /proc filesystem for each network namespace.

The exporter should be run with root privileges on the host system (don't work from a docker container).

For example, for two network namespaces like:

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