package exoip

import (
	"net"

	"github.com/exoscale/egoscale"
)

// Peer represents a peer machine
type Peer struct {
	IP       net.IP
	Dead     bool
	Priority byte
	LastSeen int64
	NicID    string
	Conn     *net.UDPConn
}

// Payload represents a message of our protocol
type Payload struct {
	Priority byte
	ExoIP    net.IP
	NicID    string
}

// State represents the state : backup, master
type State int

const (
	// StateBackup represents the backup state
	StateBackup State = iota
	// StateMaster represents the master state
	StateMaster State = iota
)

// Engine represents the ExoIP engine structure
type Engine struct {
	DeadRatio        int
	Interval         int
	Priority         byte
	VHID             byte
	SendBuf          []byte
	Peers            []*Peer
	State            State
	LastSend         int64
	InitHoldOff      int64
	VirtualMachineID string
	NicID            string
	ExoIP            net.IP
	Exo              *egoscale.Client
	Async            egoscale.AsyncInfo
	InstanceID       string
}
