package exoip

import (
	"time"
)

func (engine *Engine) SwitchToBackup() {
	Logger.Warning("switching to back-up state")
}

func (engine *Engine) SwitchToMaster() {
	Logger.Warning("switching to master state")
}

func (engine *Engine) PerformStateTransition(state State) {

	if engine.State == state {
		return
	}

	engine.State = state

	if state == StateBackup {
		engine.SwitchToBackup()
	} else {
		engine.SwitchToMaster()
	}
}

func (engine *Engine) CheckState() {

	time.Sleep(Skew)
	Logger.Info("checking for state changes")

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
		engine.PerformStateTransition(StateBackup)
	} else {
		engine.PerformStateTransition(StateMaster)
	}

	for _, peer := range(dead_peers) {
		engine.HandleDeadPeer(peer)
	}
}
