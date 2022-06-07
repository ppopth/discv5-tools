package crawler

import (
	"crypto/ecdsa"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/enode"
)

var (
	errCrawlerRunning = errors.New("crawler already running")
	errCrawlerStopped = errors.New("crawler stopped")
)

// A shadow interface of discover.UDPv5, so we can do dependency injection
// with a fake one.
type discv5 interface {
	RandomNodes() enode.Iterator
	RequestENR(*enode.Node) (*enode.Node, error)
	Close()
}

// Config is a configuration used to create Crawler.
type Config struct {
	// The list of ethereum nodes used to bootstrap the network.
	BootNodes []*enode.Node
	// Logger.
	Logger *log.Logger
}

// Crawler is a container for states of a cralwer node.
type Crawler struct {
	config *Config
	// The interface used to communicate with the ethereum DHT.
	disc discv5
	// The private key used to run the ethereum node.
	privateKey *ecdsa.PrivateKey
	// The store containing the IDs of alive nodes that have been seen.
	store map[enode.ID]bool
	// The log used inside the crawler.
	log *log.Logger

	// Used to enter critical sections.
	lock sync.Mutex
	// Used to wait for the goroutines to finish.
	loopWG sync.WaitGroup
	// Used to send a signal when the crawler stops.
	quit chan struct{}
	// Used to indicate if the crawling is running.
	running bool
	// Used to send a new node out when the user wants it.
	ndCh chan *enode.Node
}

// New creates a new crawler.
func New(config *Config) *Crawler {
	// If no private key is provided, generate a new key.
	// We can ignore an error over here, because it's just a key generation
	// and there will be no error.
	privateKey, _ := crypto.GenerateKey()

	if config.Logger == nil {
		config.Logger = log.Default()
	}

	return &Crawler{
		config:     config,
		privateKey: privateKey,
		store:      make(map[enode.ID]bool),
		log:        config.Logger,
	}
}

func (c *Crawler) GetNode() (*enode.Node, error) {
	c.lock.Lock()
	if !c.running {
		return nil, errCrawlerStopped
	}
	c.lock.Unlock()

	select {
	case nd := <-c.ndCh:
		return nd, nil
	case <-c.quit:
		return nil, errCrawlerStopped
	}
}

func (c *Crawler) Start() error {
	c.lock.Lock()
	defer c.lock.Unlock()
	if c.running {
		return errCrawlerRunning
	}
	c.running = true
	c.quit = make(chan struct{})
	c.ndCh = make(chan *enode.Node)

	if err := c.setupDiscovery(); err != nil {
		return err
	}

	c.loopWG.Add(1)
	go c.run()
	return nil
}

func (c *Crawler) Stop() {
	c.lock.Lock()
	if !c.running {
		c.lock.Unlock()
		return
	}
	c.running = false
	// Send a signal to the running routines that it is stopping.
	close(c.quit)
	c.disc.Close()
	c.lock.Unlock()
	c.loopWG.Wait()
}

func (c *Crawler) run() {
	iter := c.disc.RandomNodes()
	defer iter.Close()
	defer c.loopWG.Done()
	for iter.Next() {
		n := iter.Node()
		if _, ok := c.store[n.ID()]; ok {
			c.log.Printf("found duplicated node (id=%s)", n.ID().TerminalString())
			continue
		}
		// We have to directly request the ENR from the node to make sure that
		// the node is alive.
		nn, err := c.disc.RequestENR(n)
		if err != nil {
			// If it's not alive, log and skip to the next node.
			c.log.Printf("found unalive node (id=%s)", n.ID().TerminalString())
			continue
		}
		// Save the alive node to check for the duplication later.
		c.store[n.ID()] = true
		c.log.Printf("found alive node (id=%s)", nn.ID().TerminalString())
		c.ndCh <- nn
	}
}

// Run all the necessary steps to produce `c.disc`.
func (c *Crawler) setupDiscovery() error {
	cfg := discover.Config{
		PrivateKey: c.privateKey,
		Bootnodes:  c.config.BootNodes,
	}
	// By putting the empty string, it will create a memory database instead
	// of a persistent database.
	db, err := enode.OpenDB("")
	if err != nil {
		return err
	}

	// Create a new local ethereum p2p node.
	ln := enode.NewLocalNode(db, cfg.PrivateKey)
	// Bind to some UDP port.
	addr := "0.0.0.0:0"
	socket, err := net.ListenPacket("udp4", addr)
	if err != nil {
		return err
	}
	usocket := socket.(*net.UDPConn)

	// SetFallbackIP and SetFallbackUDP set the last-resort IP address.
	// This address is used if no endpoint prediction can be made.
	uaddr := socket.LocalAddr().(*net.UDPAddr)
	if uaddr.IP.IsUnspecified() {
		ln.SetFallbackIP(net.IP{127, 0, 0, 1})
	} else {
		ln.SetFallbackIP(uaddr.IP)
	}
	ln.SetFallbackUDP(uaddr.Port)

	// ListenV5 listens on the given connection. It creates many goroutines to
	// handle events and incoming packets.
	c.disc, err = discover.ListenV5(usocket, ln, cfg)
	if err != nil {
		return err
	}
	return nil
}
