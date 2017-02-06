package exoip

import (
	"errors"
	"fmt"
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

func VMHasSecurityGroup(vm *egoscale.VirtualMachine, sgname string) bool {

	for _, sg := range(vm.SecurityGroups) {
		if sg.Name == sgname {
			return true
		}
	}
	return false
}

func GetSecurityGroupPeers(ego *egoscale.Client, sgname string) ([]string, error) {

	peers := make([]string, 0)
	vms, err := ego.ListVirtualMachines()
	if err != nil {
		return nil, err
	}

	for _, vm := range(vms) {
		if VMHasSecurityGroup(vm, sgname) {
			primary_ip := vm.Nic[0].Ipaddress
			peers = append(peers, fmt.Sprintf("%s:%d", primary_ip, DefaultPort))
		}
	}

	for _, p := range(peers) {
		fmt.Println("found peer:", p)
	}
	return peers, nil
}
