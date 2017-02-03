package exoip

import (
	"fmt"
	"time"
)



func (engine *Engine) CheckState() {

	time.Sleep(Skew)

	now := CurrentTimeMillis()

	if now <= engine.InitHoldOff {
		return
	}

	dead_peers := make([]*Peer, 0)
	best_advertisement := true

	for _, peer := range(engine.Peers) {
		if engine.PeerIsNewlyDead(now, peer) {
			dead_peers = append(dead_peers, peer)
		} else {
			if engine.BackupOf(peer) {
				best_advertisement = false
			}
		}
	}

	if best_advertisement == false {
		Logger.Info("host is backup")
	} else {
		Logger.Info("host is master")
	}

	for _, peer := range(dead_peers) {
		Logger.Info(fmt.Sprintf("found dead peer:", peer.IP))
	}
}
