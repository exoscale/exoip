package exoip

import (
	"errors"
	"fmt"
	"os"

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

// FetchNicAndVM fetches our NIC and the VirtualMachine
func (engine *Engine) FetchNicAndVM() {

	vmInfo := &egoscale.VirtualMachine{
		ID: engine.InstanceID,
	}

	err := engine.Exo.Get(vmInfo)
	AssertSuccess(err)

	nic := vmInfo.DefaultNic()
	if nic == nil {
		Logger.Crit("cannot find virtual machine Nic ID")
		fmt.Fprintln(os.Stderr, "cannot find virtual machine Nic ID")
		os.Exit(1)
	}

	engine.VirtualMachineID = vmInfo.ID
	engine.NicID = nic.ID
}

// ObtainNic add the elastic IP to the given NIC
func (engine *Engine) ObtainNic(nicID string) error {

	_, err := engine.Exo.Request(&egoscale.AddIPToNic{
		NicID:     nicID,
		IPAddress: engine.ExoIP,
	})

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
	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}

	if err := engine.Exo.Get(vm); err != nil {
		Logger.Crit(fmt.Sprintf("could not get virtualmachine: %s. %s", vm.ID, err))
		return err
	}

	nicAddressID := ""
	nic := vm.DefaultNic()
	if nic != nil {
		for _, secIP := range nic.SecondaryIP {
			if secIP.IPAddress == nil {
				continue
			}

			if secIP.IPAddress.String() == engine.ExoIP.String() {
				nicAddressID = secIP.ID
				break
			}
		}
	}

	if nicAddressID == "" {
		Logger.Warning("could not remove ip from nic: unknown association")
		return fmt.Errorf("could not remove ip from nic: unknown association")
	}

	req := &egoscale.RemoveIPFromNic{
		ID: nicAddressID,
	}
	if err := engine.Exo.BooleanRequest(req); err != nil {
		Logger.Crit(fmt.Sprintf("could not dissociate ip %s (%s): %s",
			engine.ExoIP.String(), nicAddressID, err))
		return err
	}

	Logger.Info(fmt.Sprintf("released ip %s", engine.ExoIP.String()))
	return nil
}

// ReleaseNic removes the Elastic IP from the given NIC
func (engine *Engine) ReleaseNic(nicID string) error {

	vms, err := engine.Exo.List(new(egoscale.VirtualMachine))
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could not list virtualmachines: %s",
			err))
		return err
	}

	nicAddressID := ""
	for _, i := range vms {
		vm := i.(egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic != nil && nic.ID == nicID {
			for _, secIP := range nic.SecondaryIP {
				if secIP.IPAddress.String() == engine.ExoIP.String() {
					nicAddressID = secIP.ID
					break
				}
			}
		}
	}

	if len(nicAddressID) == 0 {
		Logger.Warning("could not remove ip from nic: unknown association")
		return fmt.Errorf("")
	}

	req := &egoscale.RemoveIPFromNic{ID: nicAddressID}
	if err := engine.Exo.BooleanRequest(req); err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic %s (%s): %s",
			nicID, nicAddressID, err))
		return err
	}

	Logger.Info(fmt.Sprintf("released ip %s from nic %s", engine.ExoIP.String(), nicID))
	return nil
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
	vms, err := ego.List(new(egoscale.VirtualMachine))
	if err != nil {
		return nil, err
	}

	for _, i := range vms {
		vm := i.(egoscale.VirtualMachine)
		if VMHasSecurityGroup(&vm, sgname) {
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
