package exoip

import (
	"errors"
	"fmt"

	"github.com/exoscale/egoscale"
)

// fetchMyInfo fetches the nic of the current instance
func fetchMyInfo(ego *egoscale.Client, instanceID egoscale.UUID) (*egoscale.UUID, *egoscale.UUID, error) {

	vm := &egoscale.VirtualMachine{
		ID: &instanceID,
	}
	if err := ego.Get(vm); err != nil {
		return nil, nil, err
	}

	nic := vm.DefaultNic()
	if nic == nil {
		return nil, nil, errors.New("cannot find virtual machine default nic")
	}

	return vm.ZoneID, nic.ID, nil
}

// VMHasSecurityGroup tells whether the VM has any security groups
func VMHasSecurityGroup(vm *egoscale.VirtualMachine, sgname string) bool {

	for _, sg := range vm.SecurityGroup {
		if sg.Name == sgname {
			return true
		}
	}
	return false
}

// FindPeerNic return the NIC ID of a given peer
func FindPeerNic(ego *egoscale.Client, ip string) (*egoscale.UUID, error) {

	vms, err := ego.List(new(egoscale.VirtualMachine))
	if err != nil {
		return nil, err
	}

	for _, i := range vms {
		vm := i.(egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic != nil && nic.IPAddress != nil && nic.IPAddress.String() == ip {
			return vm.DefaultNic().ID, nil
		}
	}

	return nil, fmt.Errorf("cannot find nic")
}
