package exoip

import (
	"errors"
	"encoding/hex"
	"net"
	"time"
	"os"
)

func BufToPayload(buf []byte) (*Payload, error) {

	protobuf := make([]byte, 2)
	protobuf = buf[0:2]
	uuidbuf := make([]byte, 16)
	uuidbuf = buf[8:24]

	if ProtoVersion != hex.EncodeToString(protobuf) {
		Logger.Warning("bad protocol version")
		return nil, errors.New("bad protocol version")
	}

	if buf[2] != buf[3] {
		Logger.Warning("bad payload (priority should repeat)")
		return nil, errors.New("bad payload (priority should repeat)")
	}

	ip := net.IPv4(buf[4], buf[5], buf[6], buf[7])
	return &Payload{NicId: UUIDToStr(uuidbuf), Priority: buf[2], ExoIP: ip}, nil
}

func (engine *Engine) NetworkLoop(listen_address string) error {
	ServerAddr,err := net.ResolveUDPAddr("udp", listen_address)
	AssertSuccess(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	AssertSuccess(err)
	buf := make([]byte, 24)
	for {
		n, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			Logger.Crit("network server died")
			os.Exit(1)

		}
		if n != 24 {
			Logger.Warning("bad network payload")

		}
		payload, err := BufToPayload(buf)
		if err != nil {
			Logger.Warning("unparseable payload")
		} else {
			engine.UpdatePeer(*addr, payload)
		}
	}
}

func (engine *Engine) NetworkAdvertise() {
	for {
		time.Sleep(time.Duration(engine.Interval) * time.Second)
		go func() {
			for _, peer := range(engine.Peers) {
				/* do not account for errors */
				peer.Conn.Write(engine.SendBuf)
			}
			engine.LastSend = CurrentTimeMillis()
		}()
		go engine.CheckState()
	}
}
