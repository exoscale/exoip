package exoip

import (
	"fmt"
	"time"
)

// PerformStateTransition transition to the given state
func (engine *Engine) PerformStateTransition(state State) {

	if engine.State == state {
		return
	}

	Logger.Info(fmt.Sprintf("swiching state to %s", state))

	var err error
	if state == StateBackup {
		err = engine.ReleaseNic(engine.NicID)
	} else {
		err = engine.ObtainNic(engine.NicID)
	}

	if err != nil {
		Logger.Crit(fmt.Sprintf("could not switch state. %s", err))
		return
	}

	engine.State = state
}

// CheckState updates the states of our peers
func (engine *Engine) CheckState() {

	time.Sleep(Skew)

	now := CurrentTimeMillis()

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

	if bestAdvertisement {
		engine.PerformStateTransition(StateMaster)
	} else {
		engine.PerformStateTransition(StateBackup)
	}

	for _, peer := range deadPeers {
		engine.HandleDeadPeer(peer)
	}
}
