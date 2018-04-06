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

// SkewMillis how much time to way
const SkewMillis = 100

// Skew how much time to way
const Skew time.Duration = 100 * time.Millisecond

// Verbose makes the client talkative
var Verbose = false

// Logger represents a wrapped version of syslog
var Logger *wrappedLogger

// NewEngineWatchdog creates an new watchdog engine
func NewEngineWatchdog(client *egoscale.Client, ip, instanceID string, interval int,
	prio int, deadRatio int, peers []string) *Engine {

	nicid, err := FetchMyNic(client, instanceID)
	uuidbuf, err := StrToUUID(nicid)
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
		DeadRatio:   deadRatio,
		Interval:    interval,
		Priority:    sendbuf[2],
		SendBuf:     sendbuf,
		Peers:       make([]*Peer, 0),
		State:       StateBackup,
		NicID:       nicid,
		ExoIP:       netip,
		Exo:         client,
		InstanceID:  instanceID,
		InitHoldOff: CurrentTimeMillis() + (1000 * int64(deadRatio) * int64(interval)) + SkewMillis,
	}
	for _, p := range peers {
		engine.Peers = append(engine.Peers, NewPeer(client, p))
	}
	return engine
}

// NewEngine creates a new engine
func NewEngine(client *egoscale.Client, ip, instanceID string) *Engine {

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

	engine := &Engine{
		ExoIP:      netip,
		Exo:        client,
		InstanceID: instanceID,
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

// FetchNicAndVM fetches our NIC and the VirtualMachine
func (engine *Engine) FetchNicAndVM() {

	vmInfo := &egoscale.VirtualMachine{
		ID: engine.InstanceID,
	}

	err := engine.Exo.Get(vmInfo)
	assertSuccess(err)

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

// FindPeer finds a peer by IP
func (engine *Engine) FindPeer(addr net.UDPAddr) *Peer {
	for i, _ := range engine.Peers {
		peer := engine.Peers[i]
		if peer.IP != nil && peer.IP.Equal(addr.IP) {
			return engine.Peers[i]
		}
	}
	return nil
}

// UpdatePeer update the state of the given peer
func (engine *Engine) UpdatePeer(addr net.UDPAddr, payload *Payload) {

	if !engine.ExoIP.Equal(payload.ExoIP) {
		Logger.Warning("peer sent message for wrong EIP")
		return
	}

	if peer := engine.FindPeer(addr); peer != nil {
		peer.Priority = payload.Priority
		peer.NicID = payload.NicID
		peer.LastSeen = CurrentTimeMillis()
		return
	}

	Logger.Warning("peer not found in configuration")
}

// PeerIsNewlyDead contains the logic to say if the peer is considered dead
func (engine *Engine) PeerIsNewlyDead(now int64, peer *Peer) bool {
	peerDiff := now - peer.LastSeen
	dead := peerDiff > int64(engine.Interval*engine.DeadRatio)*1000
	if dead != peer.Dead {
		if dead {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, considering dead.", peer.IP, peerDiff))
		} else {
			Logger.Info(fmt.Sprintf("peer %s last seen %dms ago, is now back alive.", peer.IP, peerDiff))
		}
		peer.Dead = dead
		return dead
	}
	return false
}

// BackupOf tells if we are a backup of the given peer
func (engine *Engine) BackupOf(peer *Peer) bool {
	return (!peer.Dead && peer.Priority < engine.Priority)
}

// HandleDeadPeer releases the NIC
func (engine *Engine) HandleDeadPeer(peer *Peer) {
	engine.ReleaseNic(peer.NicID)
}
