package exoip

import (
	"fmt"
	"net"
	"time"
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

// Info logs the current state (for debugging)
func (peer *Peer) Info() {
	Logger.Info(fmt.Sprintf("\tVirtualMachine ID: %s", peer.VirtualMachineID))
	Logger.Info(fmt.Sprintf("\tNic ID: %s", peer.NicID))
	Logger.Info(fmt.Sprintf("\tAddress: %s", peer.UDPAddr))
	Logger.Info(fmt.Sprintf("\tDead: %v", peer.Dead))
	Logger.Info(fmt.Sprintf("\tPriority: %d", peer.Priority))
	Logger.Info(fmt.Sprintf("\tLast Seen: %s", peer.LastSeen.Format(time.RFC3339)))
}
