package exoip

import (
	"errors"
	"github.com/pyr/egoscale/src/egoscale"

)

func FetchMyNic(ego *egoscale.Client, mserver string) (string, error) {

	instance_id, err := FetchMetadata(mserver, "/latest/instance-id")
	if err != nil {
		return "", err
	}
	vm_info, err := ego.GetVirtualMachine(instance_id)
	if err != nil {
		return "", err
	}
	if len(vm_info.Nic) < 1 {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return vm_info.Nic[0].Id, nil
}

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
