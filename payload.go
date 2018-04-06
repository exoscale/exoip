package exoip

import (
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"strings"
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

	nicId, err := UUIDToStr(uuidbuf)
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		NicID:    nicId,
		Priority: buf[2],
		IP:       net.IPv4(buf[4], buf[5], buf[6], buf[7]),
	}

	return payload, nil
}

// UUIDToStr converts a UUID into a str
func UUIDToStr(buf []byte) (string, error) {
	if len(buf) != 16 {
		return "", fmt.Errorf("UUID length (%d) mismatch, need 16", len(buf))
	}

	uuid := fmt.Sprintf("%x-%x-%x-%x", buf[0:4], buf[4:6], buf[6:8], buf[8:10], buf[10:16])

	return uuid, nil
}

// StrToUUID convert a str into a UUID
func StrToUUID(uuid string) ([]byte, error) {
	if len(uuid) != 36 {
		return nil, fmt.Errorf("UUID source length (%d) mismatch, need 36", len(uuid))
	}

	uuid = strings.Replace(strings.ToLower(uuid), "-", "", -1)
	if len(uuid) != 32 { // 36 - 4
		return nil, errors.New("UUID has wrong length, need 32")
	}

	ba, err := hex.DecodeString(uuid)
	if err != nil {
		return nil, errors.New("Invalid UUID")
	}

	if len(ba) != 16 { // 32 / 2
		return nil, fmt.Errorf("UUID converted length (%d) mismatch, need 16", len(ba))
	}

	return ba, nil
}
