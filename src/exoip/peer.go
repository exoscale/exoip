package exoip

import (
	"fmt"
	"net"
)

func NewPeer(peer string) *Peer {
	addr, err := net.ResolveUDPAddr("udp", peer)
	if err != nil {
		panic(err)
	}

	ip := addr.IP
	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}
	return &Peer{IP: ip, LastSeen: 0, Conn: conn, Dead: false,}
}

func (engine *Engine) FindPeer(addr net.UDPAddr) *Peer {
	for _, p := range(engine.Peers) {
		if p.IP.Equal(addr.IP) {
			return p
		}
	}
	return nil
}

func (engine *Engine) UpdatePeer(addr net.UDPAddr, payload *Payload) {

	if !engine.ExoIP.Equal(payload.ExoIP) {
		Logger.Warning("peer sent message for wrong EIP")
		return
	}

	peer := engine.FindPeer(addr)
	if peer == nil {
		Logger.Warning("peer not found in configuration")
		return
	}
	peer.Priority = payload.Priority
	peer.NicId = payload.NicId
	peer.LastSeen = CurrentTimeMillis()
}

func (engine *Engine) PeerIsNewlyDead(now int64, peer *Peer) bool {
	peer_diff := now - peer.LastSeen
	dead := peer_diff > int64(engine.Interval * engine.DeadRatio) * 1000
	if dead != peer.Dead {
		if dead {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, considering dead.", peer.IP, peer_diff))
		} else {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, is now back alive.", peer.IP, peer_diff))
		}
		peer.Dead = dead
		return dead
	}
	return false
}

func (engine *Engine) BackupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.Priority)
}

func (engine *Engine) HandleDeadPeer(peer *Peer) {
	engine.ReleaseNic(peer.NicId)
}
