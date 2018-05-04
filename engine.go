package exoip

import (
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
func NewEngineWatchdog(client *egoscale.Client, ip, instanceID string, interval int,
	prio int, deadRatio int, peers []string, securityGroupName string) *Engine {

	zoneID, nicID, err := fetchMyInfo(client, instanceID)
	assertSuccess(err)

	uuidbuf, err := StrToUUID(nicID)
	assertSuccess(err)

	sendbuf := make([]byte, 24)
	protobuf, err := hex.DecodeString(ProtoVersion)
	assertSuccess(err)
	netip := net.ParseIP(ip)
	if netip == nil {
		Logger.Crit("Could not parse IP")
		fmt.Fprintln(os.Stderr, "Could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		Logger.Crit("Unsupported IPv6 Address")
		fmt.Fprintln(os.Stderr, "Unsupported IPv6 Address")
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

	for i, b := range uuidbuf {
		sendbuf[i+8] = b
	}

	engine := &Engine{
		client:            client,
		DeadRatio:         deadRatio,
		Interval:          time.Duration(interval) * time.Second,
		priority:          sendbuf[2],
		SendBuf:           sendbuf,
		peers:             make(map[string]*Peer),
		SecurityGroupName: securityGroupName,
		State:             StateBackup,
		NicID:             nicID,
		ElasticIP:         netip,
		VirtualMachineID:  instanceID,
		ZoneID:            zoneID,
		InitHoldOff:       time.Now().Add(time.Duration(int64(interval)*int64(deadRatio))*time.Second + Skew),
	}

	for _, peerAddress := range peers {
		peer, err := engine.FetchPeer(peerAddress)
		assertSuccess(err)

		engine.peers[peerAddress] = peer
	}

	return engine
}

// NewEngine creates a new engine
func NewEngine(client *egoscale.Client, ipAddress, instanceID string) *Engine {

	netip := net.ParseIP(ipAddress)
	if netip == nil {
		Logger.Crit("Could not parse IP")
		fmt.Fprintln(os.Stderr, "Could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		Logger.Crit("Unsupported IPv6 Address")
		fmt.Fprintln(os.Stderr, "Unsupported IPv6 Address")
		os.Exit(1)
	}

	engine := &Engine{
		ElasticIP:        netip,
		client:           client,
		VirtualMachineID: instanceID,
	}
	engine.FetchNicAndVM()
	return engine
}

// NetworkLoop starts the UDP server
func (engine *Engine) NetworkLoop(listenAddress string) error {
	ServerAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	assertSuccess(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	assertSuccess(err)

	buf := make([]byte, payloadLength)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			Logger.Crit("network server died")
			os.Exit(1)
		}

		if n != payloadLength {
			Logger.Warning("bad network payload")
		}

		payload, err := NewPayload(buf)
		if err != nil {
			Logger.Warning("unparseable payload")
		} else {
			engine.UpdatePeer(*addr, payload)
		}
	}
}

// PingPeers sends the SendBuf to each peer
func (engine *Engine) PingPeers() error {
	engine.peersMu.RLock()
	defer engine.peersMu.RUnlock()

	for _, peer := range engine.peers {
		// do not account for errors
		peer.Send(engine.SendBuf)
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
		return nil, fmt.Errorf("Peer (%v) has no default nic", peerAddress)
	}

	return NewPeer(addr, vm.ID, nic.ID), nil
}

// FetchNicAndVM fetches our NIC and the VirtualMachine
func (engine *Engine) FetchNicAndVM() {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}
	err := client.Get(vm)
	assertSuccess(err)

	nic := vm.DefaultNic()
	if nic == nil {
		assertSuccess(fmt.Errorf("cannot find self default nic"))
	}

	engine.NicID = nic.ID
}

// ObtainNic add the elastic IP to the given NIC
func (engine *Engine) ObtainNic(nicID string) error {
	client := engine.client

	_, err := client.Request(&egoscale.AddIPToNic{
		NicID:     nicID,
		IPAddress: engine.ElasticIP,
	})

	if err != nil {
		Logger.Crit(fmt.Sprintf("could not add ip %s to nic %s: %s",
			engine.ElasticIP,
			nicID,
			err))
		return err
	}

	Logger.Info(fmt.Sprintf("claimed ip %s on nic %s", engine.ElasticIP.String(), nicID))
	return nil
}

// ReleaseMyNic releases the elastic IP from the NIC
func (engine *Engine) ReleaseMyNic() error {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}

	if err := client.Get(vm); err != nil {
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

			if secIP.IPAddress.String() == engine.ElasticIP.String() {
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
	if err := client.BooleanRequest(req); err != nil {
		Logger.Crit(fmt.Sprintf("could not dissociate ip %s (%s): %s",
			engine.ElasticIP.String(), nicAddressID, err))
		return err
	}

	Logger.Info(fmt.Sprintf("released ip %s", engine.ElasticIP.String()))
	return nil
}

// ReleaseNic removes the Elastic IP from the given NIC
func (engine *Engine) ReleaseNic(nicID string) error {
	client := engine.client

	vms, err := client.List(new(egoscale.VirtualMachine))
	if err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic: could not list virtualmachines: %s",
			err))
		return err
	}

	nicAddressID := ""
	for _, i := range vms {
		vm := i.(*egoscale.VirtualMachine)
		nic := vm.DefaultNic()
		if nic != nil && nic.ID == nicID {
			for _, secIP := range nic.SecondaryIP {
				if secIP.IPAddress.Equal(engine.ElasticIP) {
					nicAddressID = secIP.ID
					break
				}
			}
		}
	}

	if len(nicAddressID) == 0 {
		Logger.Warning("no vm holds the ipaddress")
		return fmt.Errorf("")
	}

	req := &egoscale.RemoveIPFromNic{ID: nicAddressID}
	if err := client.BooleanRequest(req); err != nil {
		Logger.Crit(fmt.Sprintf("could not remove ip from nic %s (%s): %s",
			nicID, nicAddressID, err))
		return err
	}

	Logger.Info(fmt.Sprintf("released ip %s from nic %s", engine.ElasticIP.String(), nicID))
	return nil
}

// UpdateNic checks if the EIP must be reattached to self
func (engine *Engine) UpdateNic() error {
	client := engine.client

	vm := &egoscale.VirtualMachine{
		ID: engine.VirtualMachineID,
	}
	err := client.Get(vm)
	assertSuccess(err)

	nic := vm.DefaultNic()
	if nic == nil {
		return fmt.Errorf("no default nic found for self")
	}

	if nic.ID != engine.NicID {
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

	// disassociate the IP from self if still present and slave
	if engine.State == StateBackup && found {
		return engine.ReleaseNic(engine.NicID)
	}

	// associate the IP to self if missing and Master
	if engine.State == StateMaster && !found {
		return engine.ObtainNic(engine.NicID)
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

	Logger.Info(fmt.Sprintf("Updating peers %s (zone: %s)", engine.SecurityGroupName, engine.ZoneID))
	// grab the right to alter the Peers
	engine.peersMu.Lock()
	defer engine.peersMu.Unlock()

	knownPeers := make(map[string]interface{})
	for key := range engine.peers {
		knownPeers[key] = nil
	}

	vms, err := client.List(vm)
	if err != nil {
		return err
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

				Logger.Info(fmt.Sprintf("Found new peer %s (vm: %s)", key, vm.ID))
				engine.peers[key] = NewPeer(addr, vm.ID, vm.DefaultNic().ID)
			} else {
				delete(knownPeers, key)
			}
		}
	}

	// Remove extra peers from list of known peers
	for key := range knownPeers {
		Logger.Info(fmt.Sprintf("removing peer %s", key))
		delete(engine.peers, key)
	}

	return nil
}

// UpdatePeer update the state of the given peer
func (engine *Engine) UpdatePeer(addr net.UDPAddr, payload *Payload) {
	if !engine.ElasticIP.Equal(payload.IP) {
		Logger.Warning("peer sent message for wrong EIP")
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

	Logger.Warning("peer not found in configuration")
}

// PeerIsNewlyDead contains the logic to say if the peer is considered dead
func (engine *Engine) PeerIsNewlyDead(now time.Time, peer *Peer) bool {
	peerDiff := now.Sub(peer.LastSeen)
	dead := peerDiff > (engine.Interval * time.Duration(engine.DeadRatio))
	if dead != peer.Dead {
		if dead {
			Logger.Info(fmt.Sprintf("peer %s last seen %s (%dms ago), considering dead.", peer.UDPAddr.IP, peer.LastSeen.Format(time.RFC3339), peerDiff/time.Millisecond))
		} else {
			Logger.Info(fmt.Sprintf("peer %s last seen %s (%dms ago), is now back alive.", peer.UDPAddr.IP, peer.LastSeen.Format(time.RFC3339), peerDiff/time.Millisecond))
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

// HandleDeadPeer releases the NIC
func (engine *Engine) HandleDeadPeer(peer *Peer) {
	engine.ReleaseNic(peer.NicID)
}

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

	for _, peer := range deadPeers {
		engine.HandleDeadPeer(peer)
	}
}

// LowerPriority lowers the priority value (making it more important)
func (engine *Engine) LowerPriority() (byte, error) {
	if engine.priority > 1 {
		engine.priority--
		return engine.priority, nil
	}
	return engine.priority, fmt.Errorf("Priority cannot be lowered any more")
}

// RaisePriority raises the priority value (making it less important)
func (engine *Engine) RaisePriority() (byte, error) {
	if engine.priority < 255 {
		engine.priority++
		return engine.priority, nil
	}
	return engine.priority, fmt.Errorf("Priority cannot be raised any more")
}
