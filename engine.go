package exoip

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/exoscale/egoscale"
)

// DefaultPort used by exoip
const DefaultPort = 12345

// ProtoVersion version of the protocol
const ProtoVersion = "0201"

// Skew how much time to wait
const Skew = 100 * time.Millisecond

// Verbose makes the client talkative
var Verbose = false

// Logger represents a wrapped version of syslog
var Logger *wrappedLogger

// NewEngineWatchdog creates an new watchdog engine
func NewEngineWatchdog(client *egoscale.Client, addr string, ip net.IP, instanceID egoscale.UUID, interval int,
	prio int, deadRatio int, peers []string, securityGroupName string) *Engine {

	zoneID, nicID, err := fetchMyInfo(client, instanceID)
	assertSuccessOrExit(err)

	sendbuf := make([]byte, payloadLength)
	protobuf, err := hex.DecodeString(ProtoVersion)
	assertSuccessOrExit(err)
	netip := ip.To4()
	if netip == nil {
		Logger.Crit("IPv6 addresses are unsupported")
		_, errP := fmt.Fprintf(os.Stderr, "IPv6 addresses are unsupported %q\n", ip)
		if errP != nil {
			panic(errP)
		}
		os.Exit(1)
	}

	netbytes := []byte(netip)

	sendbuf[0] = protobuf[0]
	sendbuf[1] = protobuf[1]
	sendbuf[2] = byte(prio)
	sendbuf[3] = byte(prio)
	sendbuf[4] = netbytes[0]
	sendbuf[5] = netbytes[1]
	sendbuf[6] = netbytes[2]
	sendbuf[7] = netbytes[3]

	for i, b := range nicID.UUID {
		sendbuf[i+8] = b
	}

	engine := &Engine{
		client:            client,
		ListenAddress:     addr,
		DeadRatio:         deadRatio,
		Interval:          time.Duration(interval) * time.Second,
		priority:          sendbuf[2],
		SendBuf:           sendbuf,
		peers:             make(map[string]*Peer),
		SecurityGroupName: securityGroupName,
		State:             StateBackup,
		NicID:             nicID,
		ElasticIP:         netip,
		VirtualMachineID:  &instanceID,
		ZoneID:            zoneID,
		InitHoldOff:       time.Now().Add(time.Duration(int64(interval)*int64(deadRatio))*time.Second + Skew),
	}

	for _, peerAddress := range peers {
		peer, err := engine.FetchPeer(peerAddress)
		assertSuccessOrExit(err)

		engine.peers[peerAddress] = peer
	}

	return engine
}

// NewEngine creates a new engine
func NewEngine(client *egoscale.Client, ipAddress net.IP, instanceID egoscale.UUID) *Engine {
	ipAddress = ipAddress.To4()
	if ipAddress == nil {
		Logger.Crit("IPv6 addresses are not supported %q", ipAddress)
		_, err := fmt.Fprintf(os.Stderr, "IPv6 addresses are not supported %q\n", ipAddress)
		if err != nil {
			panic(err)
		}
		os.Exit(1)
	}

	engine := &Engine{
		ElasticIP:        ipAddress,
		client:           client,
		VirtualMachineID: &instanceID,
	}
	engine.FetchNicAndVM()
	return engine
}

// NetworkLoop starts the UDP server
func (engine *Engine) NetworkLoop() error {
	ServerAddr, err := net.ResolveUDPAddr("udp", engine.ListenAddress)
	assertSuccessOrExit(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	assertSuccessOrExit(err)

	Logger.Info("listening on %s", ServerAddr)
	buf := make([]byte, payloadLength)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			Logger.Crit("network server died")
			os.Exit(1)
		}

		if bytes.Contains(buf, []byte("info")) {
			engine.Info()
			continue
		}

		if n != payloadLength {
			Logger.Warning("bad network payload")
			continue
		}

		payload, err := NewPayload(buf)
		if err != nil {
			Logger.Warning("unparseable payload")
		} else {
			engine.UpdatePeer(*addr, payload)
		}
	}
}

// Info logs the exoip current state (for debugging)
func (engine *Engine) Info() {
	Logger.Info("VirtualMachine IP: %s", engine.VirtualMachineID)
	Logger.Info("Nic IP: %s", engine.NicID)
	Logger.Info("Elastic IP: %s", engine.ElasticIP.String())
	Logger.Info("Dead ratio: %d", engine.DeadRatio)
	Logger.Info("Priority: %d", engine.priority)
	Logger.Info("State: %s", engine.State)
	Logger.Info("Last Sent: %s", engine.LastSend.Format(time.RFC3339))

	engine.peersMu.RLock()
	defer engine.peersMu.RUnlock()

	for k, peer := range engine.peers {
		Logger.Info("Peer: %s", k)
		peer.Info()
	}
}

// PingPeers sends the SendBuf to each peer
func (engine *Engine) PingPeers() error {
	engine.peersMu.RLock()
	defer engine.peersMu.RUnlock()

	for _, peer := range engine.peers {
		_, err := peer.Send(engine.SendBuf)
		if err != nil {
			Logger.Crit("failure sending to peer %s: %s", peer.VirtualMachineID, err.Error())
		}
	}
	engine.LastSend = time.Now()
	return nil
}

// FetchPeer fetches a Peer from its IP address
func (engine *Engine) FetchPeer(peerAddress string) (*Peer, error) {
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", peerAddress, DefaultPort))
	if err != nil {
		return nil, err
	}

	client := engine.client
	vm := &egoscale.VirtualMachine{
		Nic: []egoscale.Nic{{
			IPAddress: addr.IP,
			IsDefault: true,
		}},
		ZoneID: engine.ZoneID,
	}

	if err := client.Get(vm); err != nil {
		return nil, err
	}

	nic := vm.DefaultNic()
	if nic == nil {
		return nil, fmt.Errorf("peer (%v) has no default nic", peerAddress)
	}

	return NewPeer(engine.ListenAddress, addr, *vm.ID, *nic.ID), nil
}

// FetchNicAndVM fetches our NIC and the VirtualMachine
func (engine *Engine) FetchNicAndVM() {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}
	err := client.Get(vm)
	assertSuccessOrExit(err)

	nic := vm.DefaultNic()
	if nic == nil {
		assertSuccessOrExit(fmt.Errorf("cannot find self default nic"))
	}

	engine.NicID = nic.ID
}

// ObtainNic add the elastic IP to the given NIC
func (engine *Engine) ObtainNic(nicID egoscale.UUID) error {
	client := engine.client

	_, err := client.Request(&egoscale.AddIPToNic{
		NicID:     &nicID,
		IPAddress: engine.ElasticIP,
	})

	if err != nil {
		Logger.Crit("could not add ip %s to nic %s: %s",
			engine.ElasticIP,
			nicID,
			err)
		return err
	}

	Logger.Info("claimed ip %s on nic %s", engine.ElasticIP, nicID)
	return nil
}

// ReleaseMyNic releases the elastic IP from the NIC
func (engine *Engine) ReleaseMyNic() error {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}

	if err := client.Get(vm); err != nil {
		Logger.Crit("could not get virtualmachine: %s. %s", vm.ID, err)
		return err
	}

	var nicAddressID *egoscale.UUID
	nic := vm.DefaultNic()
	if nic != nil {
		for _, secIP := range nic.SecondaryIP {
			if secIP.IPAddress == nil {
				continue
			}

			if secIP.IPAddress.String() == engine.ElasticIP.String() {
				nicAddressID = secIP.ID
				break
			}
		}
	}

	if nicAddressID == nil {
		Logger.Warning("could not remove ip from nic: unknown association")
		return fmt.Errorf("could not remove ip from nic: unknown association")
	}

	req := &egoscale.RemoveIPFromNic{
		ID: nicAddressID,
	}
	if err := client.BooleanRequest(req); err != nil {
		Logger.Crit("could not disassociate ip %s (%s): %s",
			engine.ElasticIP.String(), nicAddressID, err)
		return err
	}

	Logger.Info("released ip %s", engine.ElasticIP.String())
	return nil
}

// ReleaseNic removes the Elastic IP from the given NIC
func (engine *Engine) ReleaseNic(vmID, nicID egoscale.UUID) error {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: &vmID,
	}
	err := client.Get(vm)
	if err != nil {
		Logger.Crit("could not remove IP from NIC VM:%s. %s", vmID, err)
		return err
	}

	var nicAddressID *egoscale.UUID
	nic := vm.DefaultNic()
	if nic != nil && nic.ID.Equal(nicID) {
		for _, secIP := range nic.SecondaryIP {
			if secIP.IPAddress.Equal(engine.ElasticIP) {
				nicAddressID = secIP.ID
				break
			}
		}
	}

	if nicAddressID == nil {
		Logger.Warning("vm %s doesn't hold the ipaddress %s", vmID, engine.ElasticIP)
		return fmt.Errorf("vm %s doesn't hold the ipaddress %s", vmID, engine.ElasticIP)
	}

	req := &egoscale.RemoveIPFromNic{ID: nicAddressID}
	if err := client.BooleanRequest(req); err != nil {
		Logger.Crit("could not remove ip from nic %s (%s): %s", nicID, nicAddressID, err)
		return err
	}

	Logger.Info("released ip %s from nic %s", engine.ElasticIP.String(), nicID)
	return nil
}

// UpdateNic checks if the EIP must be reattached to self
func (engine *Engine) UpdateNic() error {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}
	err := client.Get(vm)
	if err != nil {
		return fmt.Errorf("error fetching VM %s information, %s", engine.VirtualMachineID, err)
	}

	nic := vm.DefaultNic()
	if nic == nil {
		return fmt.Errorf("no default nic found for self")
	}

	if !nic.ID.Equal(*engine.NicID) {
		return fmt.Errorf("default nic ID doesn't match")
	}

	found := false
	for _, secIP := range nic.SecondaryIP {
		if secIP.IPAddress.Equal(engine.ElasticIP) {
			// we still hold the EIP
			found = true
			break
		}
	}

	// disassociate the IP from self if still present and backup
	if engine.State == StateBackup && found {
		Logger.Warning("state is %s but the eip was found, release", engine.State)
		return engine.ReleaseNic(*engine.VirtualMachineID, *engine.NicID)
	}

	// associate the IP to self if missing and Master
	if engine.State == StateMaster && !found {
		Logger.Warning("state is %s but the eip was missing, obtain", engine.State)
		return engine.ObtainNic(*engine.NicID)
	}

	return nil
}

// UpdatePeers refreshes the list of the peers based on the security group
func (engine *Engine) UpdatePeers() error {
	if engine.SecurityGroupName == "" {
		// skip
		return nil
	}

	client := engine.client
	vm := &egoscale.VirtualMachine{
		State:  "Running",
		ZoneID: engine.ZoneID,
	}

	Logger.Info("updating peers %s (zone: %s)", engine.SecurityGroupName, engine.ZoneID)
	vms, err := client.List(vm)
	if err != nil {
		return err
	}

	// grab the right to alter the Peers
	engine.peersMu.Lock()
	defer engine.peersMu.Unlock()

	knownPeers := make(map[string]interface{})
	for key := range engine.peers {
		knownPeers[key] = nil
	}

	for _, v := range vms {
		vm := v.(*egoscale.VirtualMachine)

		// skip self
		if vm.ID == engine.VirtualMachineID {
			continue
		}

		ip := vm.IP()
		if ip == nil {
			continue
		}

		for _, sg := range vm.SecurityGroup {
			if sg.Name != engine.SecurityGroupName {
				continue
			}

			key := ip.String()
			if _, ok := engine.peers[key]; !ok {
				// add peer
				addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", key, DefaultPort))
				if err != nil {
					Logger.Warning(err.Error())
					return err
				}

				Logger.Info("found new peer %s (vm: %s)", key, vm.ID)
				nic := vm.DefaultNic()
				if nic == nil {
					Logger.Warning("no default nic found for %q", vm.ID)
				} else {
					engine.peers[key] = NewPeer(engine.ListenAddress, addr, *vm.ID, *nic.ID)
				}
			} else {
				delete(knownPeers, key)
			}
		}
	}

	// Remove extra peers from list of known peers
	for key := range knownPeers {
		Logger.Info("removing peer %s", key)
		delete(engine.peers, key)
	}

	return nil
}

// UpdatePeer update the state of the given peer
func (engine *Engine) UpdatePeer(addr net.UDPAddr, payload *Payload) {
	if !engine.ElasticIP.Equal(payload.IP) {
		Logger.Warning("peer sent message for wrong EIP, got %s", payload.IP.String())
		return
	}

	engine.peersMu.Lock()
	defer engine.peersMu.Unlock()
	if peer, ok := engine.peers[addr.IP.String()]; ok {
		peer.Priority = payload.Priority
		peer.NicID = payload.NicID
		peer.LastSeen = time.Now()
		return
	}

	Logger.Warning("peer %s not found in configuration", addr.IP.String())
}

// PeerIsNewlyDead contains the logic to say if the peer is considered dead
func (engine *Engine) PeerIsNewlyDead(now time.Time, peer *Peer) bool {
	peerDiff := now.Sub(peer.LastSeen)
	dead := peerDiff > (engine.Interval * time.Duration(engine.DeadRatio))
	if dead != peer.Dead {
		if dead {
			Logger.Info("peer %s last seen %s (%dms ago), considering dead.", peer.UDPAddr.IP, peer.LastSeen.Format(time.RFC3339), peerDiff/time.Millisecond)
		} else {
			Logger.Info("peer %s, is now back alive.", peer.UDPAddr.IP)
		}
		peer.Dead = dead
		return dead
	}
	return false
}

// BackupOf tells if we are a backup of the given peer
func (engine *Engine) BackupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.priority)
}

// PerformStateTransition transition to the given state
func (engine *Engine) PerformStateTransition(state State) {

	if engine.State == state {
		return
	}

	Logger.Info("switching state to %s", state)

	var err error
	if state == StateBackup {
		err = engine.ReleaseNic(*engine.VirtualMachineID, *engine.NicID)
	} else {
		err = engine.ObtainNic(*engine.NicID)
	}

	if err != nil {
		Logger.Crit("could not switch state. %s", err)
		return
	}

	engine.State = state
}

// CheckState updates the states of our peers
func (engine *Engine) CheckState() {

	time.Sleep(Skew)

	now := time.Now()

	if now.Before(engine.InitHoldOff) {
		return
	}

	deadPeers := make([]*Peer, 0)
	bestAdvertisement := true

	engine.peersMu.RLock()
	defer engine.peersMu.RUnlock()

	for _, peer := range engine.peers {
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

	// Disconnect the dead peers from their NIC
	// and reobtain the Nic for ourself (split-brain)
	if len(deadPeers) > 0 {
		for _, peer := range deadPeers {
			err := engine.ReleaseNic(*peer.VirtualMachineID, *peer.NicID)
			if err != nil {
				Logger.Crit(err.Error())
			}
		}

		if err := engine.UpdateNic(); err != nil {
			Logger.Crit(err.Error())
		}
	}
}

// LowerPriority lowers the priority value (making it more important)
func (engine *Engine) LowerPriority() (byte, error) {
	if engine.priority > 1 {
		engine.priority--
		engine.SendBuf[2] = engine.priority
		engine.SendBuf[3] = engine.priority
		return engine.priority, nil
	}
	return engine.priority, fmt.Errorf("priority cannot be lowered any more")
}

// RaisePriority raises the priority value (making it less important)
func (engine *Engine) RaisePriority() (byte, error) {
	if engine.priority < 255 {
		engine.priority++
		engine.SendBuf[2] = engine.priority
		engine.SendBuf[3] = engine.priority
		return engine.priority, nil
	}
	return engine.priority, fmt.Errorf("priority cannot be raised any more")
}
