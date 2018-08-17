package exoip

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	"github.com/exoscale/egoscale"
)

const payloadLength = 24

// NewPayload builds a Payload from a raw buffer (length of 24)
//
// The layout of the payload is as follows:
//
//      2bytes  2bytes      4 bytes
//     ┏━━━━━━━┳━━━━━━━┳━━━━━━━━━━━━━━━┓
//     ┃ PROTO ┃ PRIO  ┃    EIP        ┃            16 bytes
//     ┣━━━━━━━┻━━━━━━━┻━━━━━━━━━━━━━━━┻━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┓
//     ┃ NicID (128bit UUID)                                          ┃
//     ┗━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━┛
//
func NewPayload(buf []byte) (*Payload, error) {
	version := hex.EncodeToString(buf[0:2])
	if ProtoVersion != version {
		Logger.Warning(fmt.Sprintf("bad protocol version, got %v", version))
		return nil, errors.New("bad protocol version")
	}

	if buf[2] != buf[3] {
		Logger.Warning("bad payload (priority should repeat)")
		return nil, errors.New("bad payload (priority should repeat)")
	}

	nicID, err := egoscale.ParseUUID(fmt.Sprintf(
		"%s-%s-%s-%s-%s",
		hex.EncodeToString(buf[8:12]),
		hex.EncodeToString(buf[12:14]),
		hex.EncodeToString(buf[14:16]),
		hex.EncodeToString(buf[16:18]),
		hex.EncodeToString(buf[18:24]),
	))
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		NicID:    nicID,
		Priority: buf[2],
		IP:       net.IPv4(buf[4], buf[5], buf[6], buf[7]),
	}

	return payload, nil
}
