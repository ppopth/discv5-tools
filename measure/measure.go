package measure

import (
	"net"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ppopth/discv5-tools/wire"
)

type Result struct {
	Rtt      time.Duration
	LossRate float64
}

type Client struct {
	ln      *enode.LocalNode
	usocket *net.UDPConn
}

func Listen() (*Client, error) {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		return nil, err
	}

	// By putting the empty string, it will create a memory database instead
	// of a persistent database.
	db, err := enode.OpenDB("")
	if err != nil {
		return nil, err
	}

	// Create a new local ethereum p2p node.
	ln := enode.NewLocalNode(db, privateKey)
	// Bind to some UDP port.
	addr := "0.0.0.0:0"
	socket, err := net.ListenPacket("udp4", addr)
	if err != nil {
		return nil, err
	}
	usocket := socket.(*net.UDPConn)

	return &Client{ln, usocket}, nil
}

func (c *Client) Run(nd *enode.Node) (*Result, error) {
	// Generate random packet.
	head, msgData, err := wire.GenRandomPacket(c.ln.ID(), nd.ID())
	if err != nil {
		return nil, err
	}

	// Encode the raw packet which is ready to be sent.
	encoded, err := wire.EncodeRawPacket(nd.ID(), head, msgData)
	if err != nil {
		return nil, err
	}

	addr := &net.UDPAddr{IP: nd.IP(), Port: nd.UDP()}
	_, err = c.usocket.WriteToUDP(encoded, addr)
	return &Result{time.Duration(0), 0}, nil
}
