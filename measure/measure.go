package measure

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover/v5wire"
	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ppopth/discv5-tools/wire"
)

const (
	maxPacketSize = 1280
	maxRequests   = 20
	numAttempts   = 100
	timeout       = 3 * time.Second
)

var (
	errTimeout = errors.New("the request reached the timeout")
)

type Result struct {
	Rtt      time.Duration
	LossRate float64
}

type call struct {
	nd     *enode.Node
	head   *v5wire.Header
	respCh chan<- *v5wire.Header
}

type Client struct {
	ln      *enode.LocalNode
	usocket *net.UDPConn
	// Used to access activeCallByNonce from multiple routines.
	lock sync.Mutex
	// The map used to find the active call by the nonce.
	activeCallByNonce map[v5wire.Nonce]call
	// The semaphore to limit the number of active calls.
	semaphore chan interface{}
	// Shutdown stuff.
	closeOnce sync.Once
	// Used to wait for the goroutines to finish.
	loopWG sync.WaitGroup
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

	client := &Client{
		ln:      ln,
		usocket: usocket,

		activeCallByNonce: make(map[v5wire.Nonce]call),
		semaphore:         make(chan interface{}, maxRequests),
	}
	client.loopWG.Add(1)
	go client.readLoop()

	return client, nil
}

func (c *Client) readLoop() {
	defer c.loopWG.Done()
	buf := make([]byte, maxPacketSize)
	for {
		nbytes, _, err := c.usocket.ReadFromUDP(buf)
		if err != nil {
			return
		}
		content := buf[:nbytes]
		head, msgData, err := wire.DecodeRawPacket(content, c.ln.ID())
		if err != nil {
			// TODO: Log the error
			continue
		}
		_, err = wire.DecodeWhoareyouAuthData(head)
		if err != nil {
			// TODO: Log the error
			continue
		}
		if len(msgData) != 0 {
			// TODO: Log the error
			continue
		}

		c.lock.Lock()
		cl, ok := c.activeCallByNonce[head.Nonce]
		if ok {
			delete(c.activeCallByNonce, head.Nonce)
		}
		c.lock.Unlock()

		if !ok {
			// TODO: Log the error
			continue
		}
		cl.respCh <- head
	}
}

func (c *Client) Close() {
	c.closeOnce.Do(func() {
		c.usocket.Close()
		c.loopWG.Wait()
	})
}

func (c *Client) Send(nd *enode.Node) (*v5wire.Header, error) {
	// Use the semaphore to limit the number of active calls.
	var empty interface{}
	c.semaphore <- empty
	defer func() {
		<-c.semaphore
	}()

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

	c.lock.Lock()
	ch := make(chan *v5wire.Header)
	cl := call{nd, &head, ch}
	c.activeCallByNonce[head.Nonce] = cl
	c.lock.Unlock()

	addr := &net.UDPAddr{IP: nd.IP(), Port: nd.UDP()}
	_, err = c.usocket.WriteToUDP(encoded, addr)
	if err != nil {
		return nil, err
	}

	select {
	case <-time.After(timeout):
		c.lock.Lock()
		delete(c.activeCallByNonce, head.Nonce)
		c.lock.Unlock()
		return nil, errTimeout
	case respHead := <-ch:
		return respHead, nil
	}
}

func (c *Client) Run(nd *enode.Node) (*Result, error) {
	avgRtt := int64(0)
	timeouts := 0
	for i := 0; i < numAttempts; i++ {
		start := time.Now()
		_, err := c.Send(nd)
		if err == errTimeout {
			timeouts++
			continue
		} else if err != nil {
			return nil, err
		}
		elapsed := time.Since(start)
		avgRtt += int64(elapsed)
	}
	avgRtt /= numAttempts
	result := &Result{
		Rtt:      time.Duration(avgRtt),
		LossRate: float64(timeouts)/numAttempts,
	}
	return result, nil
}
