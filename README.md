exoip: heartbeat monitor for Exoscale Elastic IP Addresses
==========================================================

[![Build Status](https://travis-ci.org/exoscale/exoip.svg?branch=master)](https://travis-ci.org/exoscale/exoip)

**exoip** is a small tool meant to make the process of watching
Exoscale Elastic IP Addresses and performing state transitions much
easier.

**exoip** can run in one of three modes:

- *Association Mode*: associates an EIP with an instance and exit.
- *Dissociation Mode*: dissociates an EIP from an instance and exit.
- *Watchdog Mode*: watches for peer liveness and handle necessary state transitions.

## Watchdog protocol

The goal of **exoip** is to assert liveness of peers participating in
the ownership of an *Exoscale Elastic IP*. The assumption is that at
least two peers will participate in the election process.


**exoip** uses a protocol very similar to
[CARP](http://en.wikipedia.org/wiki/Common_Addresss_Redundancy_Protocol)
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
    -p string (or IF_EXOSCALE_PEERS)
        peers to communicate with (may be repeated and/or comma-separated)
    -G string (or IF_EXOSCALE_PEER_GROUP)
        Security-Group to build peer list from
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

## Building

If you wish to inspect **exoip** and build it by yourself, you may do so
by cloning [this repository](https://github.com/exoscale/exoip) and 
peforming the following steps:

    make deps
    make
