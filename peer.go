package exoip

import (
	"net"
)

// NewPeer creates a new peer
func NewPeer(address *net.UDPAddr, id, nicID string) *Peer {
	conn, err := net.DialUDP("udp", nil, address)
	assertSuccess(err)

	return &Peer{
		VirtualMachineID: id,
		UDPAddr:          address,
		NicID:            nicID,
		conn:             conn,
	}
}

// Send writes the given buf to the connection
func (peer *Peer) Send(buf []byte) (int, error) {
	return peer.conn.Write(buf)
}
