package exoip

import (
	"errors"
	"time"
	"strings"
	"fmt"
	"net"
	"encoding/hex"
	"os"
	"github.com/pyr/egoscale/src/egoscale"
)

const ProtoVersion = "0201"
const SkewMillis = 100
const Skew time.Duration = 100 * time.Millisecond

func remove_dash(r rune) rune {
	if (r == '-') {
		return -1
	}
	return r
}

func StrToUUID(ustr string) ([]byte, error) {

	if (len(ustr) != 36) {
		return nil, errors.New("NicId has wrong length")
	}

	ustr = strings.ToLower(ustr)
	for _, c := range(ustr) {
		if (!(c >= 'a' && c <= 'f') &&
			!(c >= '0' && c <= '9') &&
			!(c == '-')) {
			return nil, errors.New("Bad characters in NicId")
		}
	}
	ustr = strings.Map(remove_dash, ustr)
	if (len(ustr) != 32) {
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

func UUIDToStr(uuidbuf []byte) string {

	str := hex.EncodeToString(uuidbuf)

	hexuuid := fmt.Sprintf("%s-%s-%s-%s-%s",
		str[0:8], str[8:12], str[12:16], str[16:20], str[20:32])
	return hexuuid
}

func NewEngine(nicid string, ip string, client *egoscale.Client, interval int,
	vhid int, prio int, dead_ratio int, peers []string) *Engine {

	uuidbuf, err := StrToUUID(nicid)
	if err != nil {
		panic(err)
	}

	sendbuf := make([]byte, 24)
	protobuf, err := hex.DecodeString(ProtoVersion)
	if err != nil {
		panic(err)
	}
	netip := net.ParseIP(ip)
	if netip == nil {
		Logger.Crit("could not parse IP")
		os.Exit(1)
	}
	netip = netip.To4()
	if netip == nil {
		Logger.Crit("Unsupported IPv6 Address")
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

	for i, b := range(uuidbuf) {
		sendbuf[i+8] = b
	}
	engine := Engine{
		DeadRatio: dead_ratio,
		Interval: interval,
		Priority: sendbuf[2],
		SendBuf: sendbuf,
		Peers: make([]*Peer, 0),
		State: StateBackup,
		ExoNic: uuidbuf,
		ExoIP: netip,
		Exo: client,
		InitHoldOff: CurrentTimeMillis() + (1000 * int64(dead_ratio) * int64(interval)) + SkewMillis,
	}
	for _, p := range(peers) {
		engine.Peers = append(engine.Peers, NewPeer(p))
	}
	return &engine
}
