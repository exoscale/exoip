package exoip

import (
	"time"
)

// SwitchToBackup switches the state to backup mode
func (engine *Engine) SwitchToBackup() {
	Logger.Warning("switching to back-up state")
	engine.ReleaseNic(engine.NicID)
}

// SwitchToMaster switches the state to master mode
func (engine *Engine) SwitchToMaster() {
	Logger.Warning("switching to master state")
	engine.ObtainNic(engine.NicID)
}

// PerformStateTransition transition to the given state
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

// CheckState updates the states of our peers
func (engine *Engine) CheckState() {

	time.Sleep(Skew)

	now := currentTimeMillis()

	if now <= engine.InitHoldOff {
		return
	}

	deadPeers := make([]*Peer, 0)
	bestAdvertisement := true

	for _, peer := range engine.Peers {
		if engine.PeerIsNewlyDead(now, peer) {
			deadPeers = append(deadPeers, peer)
		} else {
			if engine.BackupOf(peer) {
				bestAdvertisement = false
			}
		}
	}

	if bestAdvertisement == false {
		engine.PerformStateTransition(StateBackup)
	} else {
		engine.PerformStateTransition(StateMaster)
	}

	for _, peer := range deadPeers {
		engine.HandleDeadPeer(peer)
	}
}
