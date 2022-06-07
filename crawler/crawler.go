package crawler

import (
	"crypto/ecdsa"
	"log"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/p2p/enode"
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
	// The log used inside the crawler.
	log *log.Logger
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
		log:        config.Logger,
	}
}
