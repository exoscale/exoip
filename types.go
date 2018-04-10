package exoip

import (
	"log"
	"log/syslog"
	"net"
	"sync"

	"github.com/exoscale/egoscale"
)

// Peer represents a peer machine
type Peer struct {
	VirtualMachineID string
	UDPAddr          *net.UDPAddr
	Dead             bool
	Priority         byte
	LastSeen         int64
	NicID            string
	conn             *net.UDPConn
}

// Payload represents a message of our protocol
type Payload struct {
	Priority byte
	IP       net.IP
	NicID    string
}

type wrappedLogger struct {
	syslog       bool
	syslogWriter *syslog.Writer
	stdWriter    *log.Logger
}

//go:generate stringer -type=State

// State represents the state : backup, master
type State int

const (
	// StateBackup represents the backup state
	StateBackup State = iota
	// StateMaster represents the master state
	StateMaster
)

// Engine represents the ExoIP engine structure
type Engine struct {
	client            *egoscale.Client
	DeadRatio         int
	Interval          int
	Priority          byte
	SendBuf           []byte
	peers             map[string]*Peer
	peersMu           sync.RWMutex
	State             State
	LastSend          int64
	InitHoldOff       int64
	ElasticIP         net.IP
	VirtualMachineID  string
	SecurityGroupName string
	NicID             string
	ZoneID            string
}
