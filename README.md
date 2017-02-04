exoip: heartbeat monitor for Exoscale Elastic IP Addresses
==========================================================

**exoip** is a small tool meant to make the process of watching
Exoscale Elastic IP Addresses and performing state transitions much
easier.

## Watchdog protocol

The goal of **exoip** is to assert liveness of peers participating in
the ownership of an *Exoscale Elastic IP*. The assumption is that at
least two peers will participate in the election process.


**exoip** uses a protocol very similar to
[CARP](http://en.wikipedia.org/wiki/Common_Addresss_Redundancy_Protocol)
and to some extent
[VRRP](http://en.wikipedia.org/wiki/Virtual_Router_Redundancy_Protocol).

The idea is quite simple, for each of it's configured peers, **exoip**
sends a 24-byte payload through **UDP**. The payload consists of a
protocol version, a (repeated, for error checking) priority to help
elect masters, the *Elastic IP* that must be shared accross alll
peers, and the peer's Nic ID.

The layout of the payload is as follows:

      2bytes 2bytes  4bytes           16bytes
    +-------+-------+---------------+-------------------------------+
	| PROTO | PRIO  |    EIP        |   NicID (128bit UUID)         |
	+-------+-------+---------------+-------------------------------+

	
When a peer fails to advertise for a configurable period of time, it
is considered dead and action is taken to reclaim its ownership of
the configured *Elastic IP Address*.

## Configuration

**exoip** is configured through command line arguments.

    -P int
    	Host priority (lowest wins) (default 10)
    -i int
    	Cluster ID advertised (default 10)
    -l string
    	Address to bind to (default ":12345")
    -p value
    	peers to communicate with
    -r int
    	Dead ratio (default 3)
    -t int
    	Advertisement interval in seconds (default 1)
    -xi string
	    Exoscale Elastic IP to watch over
    -xk string
    	Exoscale API Key
    -xn string
    	Exoscale NIC ID
    -xs string
    	Exoscale API Secret

## Building

If you wish to inspect **exoip** and build it by yourself, you may do so
by cloning [this repository](https://github.com/exoscale/exoip) and 
peforming the following steps:

    make deps
	make
	
	
