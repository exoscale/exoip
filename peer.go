package exoip

import (
	"net"

	"github.com/exoscale/egoscale"
)

// NewPeer creates a new peer
func NewPeer(ego *egoscale.Client, peer string) *Peer {
	addr, err := net.ResolveUDPAddr("udp", peer)
	assertSuccess(err)

	ip := addr.IP
	conn, err := net.DialUDP("udp", nil, addr)
	assertSuccess(err)

	peerNic, err := FindPeerNic(ego, ip.String())
	assertSuccess(err)

	return &Peer{
		IP:    ip,
		NicID: peerNic,
		Conn:  conn,
	}
}
