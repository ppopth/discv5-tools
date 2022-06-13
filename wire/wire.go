package wire

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"

	"github.com/ethereum/go-ethereum/p2p/discover/v5wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

// Packet header flag values.
const (
	flagMessage = iota
	flagWhoareyou
	flagHandshake
)

// Protocol constants.
const (
	version         = 1
	minVersion      = 1
	sizeofMaskingIV = 16

	minMessageSize      = 48 // this refers to data after static headers
	randomPacketMsgSize = 20
)

var protocolID = [6]byte{'d', 'i', 's', 'c', 'v', '5'}

type (
	whoareyouAuthData struct {
		IDNonce   [16]byte // ID proof data
		RecordSeq uint64   // highest known ENR sequence of requester
	}

	messageAuthData struct {
		SrcID enode.ID
	}
)

// Packet sizes.
var (
	sizeofMessageAuthData = binary.Size(messageAuthData{})
)

func EncodeRawPacket(id enode.ID, head v5wire.Header, msgdata []byte) ([]byte, error) {
	// Write the unmasked packet first.
	var buf bytes.Buffer
	buf.Write(head.IV[:])
	binary.Write(&buf, binary.BigEndian, &head.StaticHeader)
	buf.Write(head.AuthData)

	masked := buf.Bytes()[sizeofMaskingIV:]
	// Apply masking.
	block, err := aes.NewCipher(id[:16])
	if err != nil {
		return nil, fmt.Errorf("can't create cipher: %v", err)
	}
	stream := cipher.NewCTR(block, head.IV[:])
	stream.XORKeyStream(masked[:], masked[:])

	// Write the packet message.
	buf.Write(msgdata)
	return buf.Bytes(), nil
}

func GenRandomPacket(fromID enode.ID, toID enode.ID) (v5wire.Header, []byte, error) {
	head := v5wire.Header{
		StaticHeader: v5wire.StaticHeader{
			ProtocolID: protocolID,
			Version:    version,
			Flag:       flagMessage,
			AuthSize:   uint16(sizeofMessageAuthData),
		},
	}

	// Encode auth data.
	auth := messageAuthData{SrcID: fromID}
	if _, err := crand.Read(head.Nonce[:]); err != nil {
		return head, nil, fmt.Errorf("can't get random data: %v", err)
	}

	var headbuf bytes.Buffer // packet header
	binary.Write(&headbuf, binary.BigEndian, auth)
	head.AuthData = headbuf.Bytes()

	var msgctbuf []byte // message data ciphertext
	// Fill message ciphertext buffer with random bytes.
	msgctbuf = append(msgctbuf[:0], make([]byte, randomPacketMsgSize)...)
	crand.Read(msgctbuf)
	// Generate masking IV.
	crand.Read(head.IV[:])

	return head, msgctbuf, nil
}
