package exoip

import (
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
)

// NewPeer creates a new peer
func NewPeer(ego *egoscale.Client, peer string) *Peer {
	addr, err := net.ResolveUDPAddr("udp", peer)
	AssertSuccess(err)
	ip := addr.IP
	conn, err := net.DialUDP("udp", nil, addr)
	AssertSuccess(err)
	peerNic, err := FindPeerNic(ego, ip.String())
	AssertSuccess(err)
	return &Peer{IP: ip, NicID: peerNic, LastSeen: 0, Conn: conn, Dead: false}
}

// FindPeer finds a peer by IP
func (engine *Engine) FindPeer(addr net.UDPAddr) *Peer {
	for i, _ := range engine.Peers {
		peer := engine.Peers[i]
		if peer.IP != nil && peer.IP.Equal(addr.IP) {
			return engine.Peers[i]
		}
	}
	return nil
}

// UpdatePeer update the state of the given peer
func (engine *Engine) UpdatePeer(addr net.UDPAddr, payload *Payload) {

	if !engine.ExoIP.Equal(payload.ExoIP) {
		Logger.Warning("peer sent message for wrong EIP")
		return
	}

	if peer := engine.FindPeer(addr); peer != nil {
		peer.Priority = payload.Priority
		peer.NicID = payload.NicID
		peer.LastSeen = CurrentTimeMillis()
		return
	}

	Logger.Warning("peer not found in configuration")
}

// PeerIsNewlyDead contains the logic to say if the peer is considered dead
func (engine *Engine) PeerIsNewlyDead(now int64, peer *Peer) bool {
	peerDiff := now - peer.LastSeen
	dead := peerDiff > int64(engine.Interval*engine.DeadRatio)*1000
	if dead != peer.Dead {
		if dead {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, considering dead.", peer.IP, peerDiff))
		} else {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, is now back alive.", peer.IP, peerDiff))
		}
		peer.Dead = dead
		return dead
	}
	return false
}

// BackupOf tells if we are a backup of the given peer
func (engine *Engine) BackupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.Priority)
}

// HandleDeadPeer releases the NIC
func (engine *Engine) HandleDeadPeer(peer *Peer) {
	engine.ReleaseNic(peer.NicID)
}
