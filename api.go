package exoip

import (
	"errors"
	"fmt"

	"github.com/exoscale/egoscale"
)

// FetchMyNic fetches the nic of the current instance
func FetchMyNic(ego *egoscale.Client, instanceID string) (string, error) {

	vmInfo := &egoscale.VirtualMachine{
		ID: instanceID,
	}
	if err := ego.Get(vmInfo); err != nil {
		return "", err
	}
	nic := vmInfo.DefaultNic()
	if nic == nil {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return nic.ID, nil
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

// getSecurityGroupPeers returns the other machines within the same security group
func getSecurityGroupPeers(client *egoscale.Client, zoneId string, securityGroupName string) ([]string, error) {

	peers := make([]string, 0)
	vms, err := client.List(&egoscale.VirtualMachine{
		ZoneID: zoneId,
	})

	if err != nil {
		return nil, err
	}

	for _, i := range vms {
		vm := i.(egoscale.VirtualMachine)

		if VMHasSecurityGroup(&vm, securityGroupName) {
			nic := vm.DefaultNic()
			if nic != nil && nic.IPAddress != nil {
				primaryIP := nic.IPAddress.String()
				peers = append(peers, fmt.Sprintf("%s:%d", primaryIP, DefaultPort))
			}
		}
	}

	return peers, nil
}

// FindPeerNic return the NIC ID of a given peer
func FindPeerNic(ego *egoscale.Client, ip string) (string, error) {

	vms, err := ego.List(new(egoscale.VirtualMachine))
	if err != nil {
		return "", err
	}

	for _, i := range vms {
		vm := i.(egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic != nil && nic.IPAddress != nil && nic.IPAddress.String() == ip {
			return vm.DefaultNic().ID, nil
		}
	}

	return "", fmt.Errorf("cannot find nic")
}
