package exoip

import (
	"net"

	"github.com/exoscale/egoscale"
)

type Peer struct {
	IP       net.IP
	Dead     bool
	Priority byte
	LastSeen int64
	NicId    string
	Conn     *net.UDPConn
}

type Payload struct {
	Priority byte
	ExoIP    net.IP
	NicId    string
}

type State int

const (
	StateBackup State = iota
	StateMaster State = iota
)

type Engine struct {
	DeadRatio   int
	Interval    int
	Priority    byte
	VHID        byte
	SendBuf     []byte
	Peers       []*Peer
	State       State
	LastSend    int64
	InitHoldOff int64
	ExoVM       string
	NicId       string
	ExoIP       net.IP
	Exo         *egoscale.Client
	AsyncInfo   egoscale.AsyncInfo
}
