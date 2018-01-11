package exoip

import (
	"errors"
	"fmt"
	"os"

	"github.com/exoscale/egoscale"
)

// FetchMyNic fetches the nic of the current instance
func FetchMyNic(ego *egoscale.Client, instanceID string) (string, error) {

	vmInfo, err := getVirtualMachine(ego, instanceID)
	if err != nil {
		return "", err
	}
	if len(vmInfo.Nic) < 1 {
		return "", errors.New("cannot find virtual machine Nic ID")
	}
	return vmInfo.Nic[0].ID, nil
}

// FetchNicAndVM fetches our NIC and the VirtualMachine
func (engine *Engine) FetchNicAndVM() {

	vmInfo, err := getVirtualMachine(engine.Exo, engine.InstanceID)
	AssertSuccess(err)

	if len(vmInfo.Nic) < 1 {
		Logger.Crit("cannot find virtual machine Nic ID")
		fmt.Fprintln(os.Stderr, "cannot find virtual machine Nic ID")
		os.Exit(1)
	}
	engine.VirtualMachineID = vmInfo.ID
	engine.NicID = vmInfo.Nic[0].ID
}

// ObtainNic add the elastic IP to the given NIC
func (engine *Engine) ObtainNic(nicID string) error {

	_, err := engine.Exo.AddIPToNic(nicID, engine.ExoIP.String())
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not add ip %s to nic %s: %s",
			engine.ExoIP.String(),
			nicID,
			err))
		return err
	}
	Logger.Info(fmt.Sprintf("claimed ip %s on nic %s", engine.ExoIP.String(), nicID))
	return nil
}

// ReleaseMyNic releases the elastic IP from the NIC
func (engine *Engine) ReleaseMyNic() error {
	vm, err := getVirtualMachine(engine.Exo, engine.VirtualMachineID)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could get virtualmachine: %s",
			err))
		return err
	}
	nicAddressID := ""
	for _, secIP := range vm.Nic[0].SecondaryIP {
		if secIP.IPAddress.String() == engine.ExoIP.String() {
			nicAddressID = secIP.ID
			break
		}
	}
	if len(nicAddressID) == 0 {
		Logger.Warning("could not remove ip from nic: unknown association")
		return fmt.Errorf("could not remove ip from nic: unknown association")
	}

	err = engine.Exo.RemoveIPFromNic(nicAddressID)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not dissociate ip %s: %s",
			engine.ExoIP.String(), err))
		return err
	}
	Logger.Info(fmt.Sprintf("released ip %s", engine.ExoIP.String()))
	return nil
}

// ReleaseNic removes the Elastic IP from the given NIC
func (engine *Engine) ReleaseNic(nicID string) {

	vms, err := listVirtualMachines(engine.Exo)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could not list virtualmachines: %s",
			err))
		return
	}

	nicAddressID := ""
	for _, vm := range vms {
		if vm.Nic[0].ID == nicID {
			for _, secIP := range vm.Nic[0].SecondaryIP {
				if secIP.IPAddress.String() == engine.ExoIP.String() {
					nicAddressID = secIP.ID
					break
				}
			}
		}
	}

	if len(nicAddressID) == 0 {
		Logger.Warning("could not remove ip from nic: unknown association")
		return
	}

	err = engine.Exo.RemoveIPFromNic(nicAddressID)
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic %s (%s): %s",
			nicID, nicAddressID, err))
	}
	Logger.Info(fmt.Sprintf("released ip %s from nic %s", engine.ExoIP.String(), nicID))
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

// GetSecurityGroupPeers returns the other machines within the same security group
func GetSecurityGroupPeers(ego *egoscale.Client, sgname string) ([]string, error) {

	peers := make([]string, 0)
	vms, err := listVirtualMachines(ego)
	if err != nil {
		return nil, err
	}

	for _, vm := range vms {
		if VMHasSecurityGroup(&vm, sgname) {
			primaryIP := vm.Nic[0].IPAddress
			peers = append(peers, fmt.Sprintf("%s:%d", primaryIP, DefaultPort))
		}
	}

	return peers, nil
}

// FindPeerNic return the NIC ID of a given peer
func FindPeerNic(ego *egoscale.Client, ip string) (string, error) {

	vms, err := listVirtualMachines(ego)
	if err != nil {
		return "", err
	}

	for _, vm := range vms {

		if vm.Nic[0].IPAddress.String() == ip {
			return vm.Nic[0].ID, nil
		}
	}

	return "", fmt.Errorf("cannot find nic")
}
