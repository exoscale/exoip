package exoip

func (engine *Engine) ObtainNic(nic_id string) {

	_, err := engine.Exo.AddIpToNic(nic_id, engine.ExoIP.String())
	if err != nil {
		Logger.Crit("could not add ip to nic")
	}
}

func (engine *Engine) ReleaseNic(nic_id string) {
	_, err := engine.Exo.RemoveIpFromNic(nic_id)
	if err != nil {
		Logger.Crit("could not remove ip from nic")
	}
}
