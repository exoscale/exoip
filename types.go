package exoip

import (
	"log"
	"log/syslog"
	"net"
	"sync"
	"time"

	"github.com/exoscale/egoscale"
)

// Peer represents a peer machine
type Peer struct {
	VirtualMachineID *egoscale.UUID
	UDPAddr          *net.UDPAddr
	Dead             bool
	Priority         byte
	LastSeen         time.Time
	NicID            *egoscale.UUID
	conn             *net.UDPConn
}

// Payload represents a message of our protocol
type Payload struct {
	Priority byte
	IP       net.IP
	NicID    *egoscale.UUID
}

type wrappedLogger struct {
	syslog       bool
	syslogWriter *syslog.Writer
	stdWriter    *log.Logger
}

// Engine represents the ExoIP engine structure
type Engine struct {
	client            *egoscale.Client
	listenPort        int
	ListenAddress     string
	DeadRatio         int
	Interval          time.Duration
	priority          byte
	SendBuf           []byte
	peers             map[string]*Peer
	peersMu           sync.RWMutex
	State             State
	LastSend          time.Time
	InitHoldOff       time.Time
	ElasticIP         net.IP
	VirtualMachineID  *egoscale.UUID
	SecurityGroupName string
	NicID             *egoscale.UUID
	ZoneID            *egoscale.UUID
}
