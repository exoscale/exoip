package exoip

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
)

const payloadLength = 24

// NewPayload builds a Payload from a raw buffer (length of 24)
//
// The layout of the payload is as follows:
//
//      2bytes  2bytes  4 bytes         16 bytes
//     +-------+-------+---------------+-------------------------------+
//     | PROTO | PRIO  |    EIP        |   NicID (128bit UUID)         |
//     +-------+-------+---------------+-------------------------------+
//
func NewPayload(buf []byte) (*Payload, error) {
	protobuf := make([]byte, 2)
	protobuf = buf[0:2]
	uuidbuf := make([]byte, 16)
	uuidbuf = buf[8:24]

	version := hex.EncodeToString(protobuf)
	if ProtoVersion != version {
		Logger.Warning(fmt.Sprintf("bad protocol version, got %d", version))
		return nil, errors.New("bad protocol version")
	}

	if buf[2] != buf[3] {
		Logger.Warning("bad payload (priority should repeat)")
		return nil, errors.New("bad payload (priority should repeat)")
	}

	payload := &Payload{
		NicID:    UUIDToStr(uuidbuf),
		Priority: buf[2],
		ExoIP:    net.IPv4(buf[4], buf[5], buf[6], buf[7]),
	}

	return payload, nil
}

// NetworkLoop starts the UDP server
func (engine *Engine) NetworkLoop(listenAddress string) error {
	ServerAddr, err := net.ResolveUDPAddr("udp", listenAddress)
	AssertSuccess(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	AssertSuccess(err)

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
