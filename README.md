exoip: heartbeat monitor for Exoscale Elastic IP Addresses
==========================================================

[![Build Status](https://travis-ci.org/exoscale/exoip.svg?branch=master)](https://travis-ci.org/exoscale/exoip)

**exoip** is a small tool meant to make the process of watching
Exoscale Elastic IP Addresses and performing state transitions much
easier.

```
$ go install github.com/exoscale/exoip/cmd/exoip
```

**exoip** can run in one of three modes:

- *Association Mode* (`-A`): associates an EIP with an instance and exit.

- *Dissociation Mode* (`-D`): dissociates an EIP from an instance and exit.

- *Watchdog Mode* (`-W`): watches for peer liveness and handle necessary state transitions.


## Watchdog protocol

The goal of **exoip** is to assert liveness of peers participating in
the ownership of an *Exoscale Elastic IP*. The assumption is that at
least two peers will participate in the election process.


**exoip** uses a protocol very similar to
[CARP](http://en.wikipedia.org/wiki/Common_Address_Redundancy_Protocol)
and to some extent
[VRRP](http://en.wikipedia.org/wiki/Virtual_Router_Redundancy_Protocol).

The idea is quite simple: for each of its configured peers, **exoip**
sends a 24-byte payload through **UDP**. The payload consists of a
protocol version, a (repeated, for error checking) priority to help
elect masters, the *Elastic IP* that must be shared accross alll
peers, and the peer's Nic ID.

The layout of the payload is as follows:

      2bytes  2bytes  4 bytes         16 bytes
    +-------+-------+---------------+-------------------------------+
    | PROTO | PRIO  |    EIP        |   NicID (128bit UUID)         |
    +-------+-------+---------------+-------------------------------+


When a peer fails to advertise for a configurable period of time, it
is considered dead and action is taken to reclaim its ownership of
the configured *Elastic IP Address*.

## Configuration

**exoip** is configured through command line arguments or an equivalent
environment variable:

    -A
        Association mode (exclusive with -D and -W)
    -D
        Dissociation mode (exclusive with -A and -W)
    -W
        Watchdog mode (exclusive with -A and -D)
    -P int (or IF_HOST_PRIORITY)
        Host priority (lowest wins) (default 10, maximum 255)
    -l string (or IF_BIND_ADDRESS)
        Address to bind to (default ":12345")
    -i string (or IF_EXOSCALE_INSTANCE_ID)
        Instance ID of one self (useful when running from a container)
    -p string (or IF_EXOSCALE_PEERS)
        peers to communicate with (may be repeated and/or comma-separated)
    -G string (or IF_EXOSCALE_PEER_GROUP)
        Security-Group to use to create/maintain the list of peers
    -r int (or IF_DEAD_RATIO)
        Dead ratio (default 3)
    -t int (or IF_ADVERTISEMENT_INTERVAL)
        Advertisement interval in seconds (default 1)
    -xi string (or IF_ADDRESS)
        Exoscale Elastic IP to watch over
    -xk string (or IF_EXOSCALE_API_KEY)
        Exoscale API Key
    -xs string (or IF_EXOSCALE_API_SECRET)
        Exoscale API Secret

## Signals

When running as a Docker container, signals are the best way to interact with the running container.

**exoip** listens to `SIGUSR1` and `SIGUSR2` which will influence the current priority value by respectively doing a -1 or a +1 on it. `SIGUSR1` will promote it to a higher rank while `SIGUSR2` will lower its rank. A simple way to put on backup mode a node without restarting **exoip**.

`SIGTERM` or `SIGINT` will attempt to disassociate the Elastic IP before quitting.

## Building

If you wish to inspect **exoip** and build it by yourself, you can install it by using `go get`.

    go get -u github.com/exoscale/exoip/cmd/exoip

### Updating

It uses [godep](https://github.com/golang/dep), so it should be easy.

    dep status
    dep ensure -update

## Setup using Cloud Init

As shown in the [HAProxy Elastic IP Automatic
failover](https://www.exoscale.ch/syslog/2017/02/07/haproxy-elastic-ip-automatic-failover/)
article, `exoip` can be setup as a _dummy_ net interface. Below is the article
configuration described using [Cloud Init](http://cloudinit.readthedocs.io/)
(supported by Ubuntu, Debian, RHEL, CentOS, etc.)

```yaml
#cloud-config

package_update: true
package_upgrade: true

packages:
- ifupdown

write_files:
- path: /etc/network/interfaces
  content: |
    source /etc/network/interfaces.d/*.cfg
- path: /etc/network/interfaces.d/51-exoip.cfg
  content: |
    auto lo:1
    iface lo:1 inet static
      address 198.51.100.50              # change me
      netmask 255.255.255.255
      exoscale-peer-group load-balancer  # change me
      exoscale-api-key EXO....           # change me
      exoscale-api-secret LZ...          # change me
      up /usr/local/bin/exoip -W &
      down killall exoip

runcmd:
- wget https://github.com/exoscale/exoip/releases/download/0.3.5/exoip
- wget https://github.com/exoscale/exoip/releases/download/0.3.5/exoip.asc
- gpg --recv-keys E458F9F85608DF5A22ECCD158B58C61D4FFE0C86
- gpg --verify --trust-model always exoip.asc
- sudo chmod +x exoip
- sudo mv exoip /usr/local/bin/
- sudo ifup lo:1
```
