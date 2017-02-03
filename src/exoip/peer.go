package exoip

import (
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

func (engine *Engine) UpdatePeer(addr net.UDPAddr, vhid byte, prio byte) {

	if vhid != engine.VHID {
		Logger.Warning("peer sent message for unknown VHID")
		return
	}

	peer := engine.FindPeer(addr)
	if peer == nil {

		Logger.Warning("peer not found in configuration")
		return
	}
	peer.Priority = prio
	peer.LastSeen = CurrentTimeMillis()
}

func (engine *Engine) PeerIsNewlyDead(now int64, peer *Peer) bool {
	dead := (peer.LastSeen < (now - (int64(engine.Interval * engine.DeadRatio) * 1000)))
	if dead != peer.Dead {
		peer.Dead = dead
		return dead
	}
	return false
}

func (engine *Engine) BackupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.Priority)
}
