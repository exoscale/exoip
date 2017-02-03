package exoip

import (
	"net"
)

type Peer struct {
	IP		net.IP
	Dead            bool
	Priority	byte
	LastSeen	int64
	Conn		*net.UDPConn
}

type State int

const (
	StateBackup State = iota
	StateMaster State = iota
)

type Engine struct {
	DeadRatio	int
	Interval	int
	Priority	byte
	VHID		byte
	SendBuf		[]byte
	Peers		[]*Peer
	State		State
	LastSend	int64
	InitHoldOff	int64
}
