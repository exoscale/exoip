package exoip

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
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

func removeDash(r rune) rune {
	if r == '-' {
		return -1
	}
	return r
}

// StrToUUID covert a str into a UUID
func StrToUUID(ustr string) ([]byte, error) {

	if len(ustr) != 36 {
		return nil, fmt.Errorf("NicId %s has wrong length", ustr)
	}

	ustr = strings.ToLower(ustr)
	for _, c := range ustr {
		if !(c >= 'a' && c <= 'f') &&
			!(c >= '0' && c <= '9') &&
			!(c == '-') {
			return nil, errors.New("Bad characters in NicId")
		}
	}
	ustr = strings.Map(removeDash, ustr)
	if len(ustr) != 32 {
		return nil, errors.New("NicId has wrong length")
	}

	ba, err := hex.DecodeString(ustr)
	if err != nil {
		return nil, err
	}
	if len(ba) != 16 {
		return nil, errors.New("Bad NicID byte length")
	}
	return ba, nil
}

// UUIDToStr converts a UUID into a str
func UUIDToStr(uuidbuf []byte) string {

	str := hex.EncodeToString(uuidbuf)

	hexuuid := fmt.Sprintf("%s-%s-%s-%s-%s",
		str[0:8], str[8:12], str[12:16], str[16:20], str[20:32])
	return hexuuid
}

// NewWatchdogEngine creates an new watchdog engine
func NewWatchdogEngine(client *egoscale.Client, ip, instanceID string, interval int,
	prio int, deadRatio int, peers []string) *Engine {

	nicid, err := FetchMyNic(client, instanceID)
	uuidbuf, err := StrToUUID(nicid)
	AssertSuccess(err)
	sendbuf := make([]byte, 24)
	protobuf, err := hex.DecodeString(ProtoVersion)
	AssertSuccess(err)
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

	engine := Engine{
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
		Async:       egoscale.AsyncInfo{Retries: 3, Delay: 20},
		InitHoldOff: currentTimeMillis() + (1000 * int64(deadRatio) * int64(interval)) + SkewMillis,
	}
	for _, p := range peers {
		engine.Peers = append(engine.Peers, NewPeer(client, p))
	}
	return &engine
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

	engine := Engine{
		ExoIP:      netip,
		Exo:        client,
		InstanceID: instanceID,
	}
	engine.FetchNicAndVM()
	return &engine
}

func getVirtualMachine(cs *egoscale.Client, instanceID string) (*egoscale.VirtualMachine, error) {
	resp, err := cs.Request(&egoscale.ListVirtualMachines{
		ID: instanceID,
	})
	if err != nil {
		return nil, err
	}
	return resp.(*egoscale.ListVirtualMachinesResponse).VirtualMachine[0], nil
}

func listVirtualMachines(cs *egoscale.Client) ([]*egoscale.VirtualMachine, error) {
	resp, err := cs.Request(&egoscale.ListVirtualMachines{})
	if err != nil {
		return nil, err
	}
	return resp.(*egoscale.ListVirtualMachinesResponse).VirtualMachine, nil
}
