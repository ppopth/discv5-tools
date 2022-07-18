package main

import (
	sha "crypto/sha256"
	"encoding/hex"
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ppopth/discv5-tools/crawler"
)

var (
	bootnodesFlag = flag.String("bootnodes", "", "Comma separated nodes used for bootstrapping")
)

func main() {
	var bootUrls []string
	if *bootnodesFlag != "" {
		bootUrls = strings.Split(*bootnodesFlag, ",")
	} else {
		bootUrls = params.V5Bootnodes
	}

	var bootNodes []*enode.Node
	for _, url := range bootUrls {
		bootNodes = append(bootNodes, enode.MustParse(url))
	}
	cfg := &crawler.Config{
		BootNodes:     bootNodes,
		Logger:        log.New(os.Stderr, "crawler: ", log.LstdFlags|log.Lmsgprefix),
		CheckLiveness: false,
	}
	cr := crawler.New(cfg)
	cr.Start()
	defer cr.Stop()

	seen := make(map[string]bool)
	for {
		nd, err := cr.GetNode()
		if err != nil {
			log.Fatalf("the crawler stopped unexpectedly: %v", err)
		}
		hash := sha.Sum256([]byte(nd.String()))
		encoded := hex.EncodeToString(hash[:])
		if _, ok := seen[nd.String()]; ok {
			continue
		}
		seen[nd.String()] = true
		log.Printf("found new ENR (hash=%v, len=%v)\n", encoded, len(seen))
		if nd.IP().IsPrivate() {
			log.Printf("found private IP in the ENR (hash=%v, len=%v)\n", encoded, len(seen))
		}
	}
}
