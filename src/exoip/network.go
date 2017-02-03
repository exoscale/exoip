package exoip

import (
	"net"
	"time"
)

func (engine *Engine) NetworkLoop(listen_address string) error {
	ServerAddr,err := net.ResolveUDPAddr("udp", listen_address)
	AssertSuccess(err)
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	AssertSuccess(err)
	buf := make([]byte, 2)
	for {
		_, addr, err := ServerConn.ReadFromUDP(buf)
		if err != nil {
			return nil
		}
		engine.UpdatePeer(*addr, buf[0], buf[1])
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
