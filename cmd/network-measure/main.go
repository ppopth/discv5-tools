package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ppopth/discv5-tools/crawler"
	"github.com/ppopth/discv5-tools/measure"
)

var (
	bootnodesFlag = flag.String("bootnodes", "", "Comma separated nodes used for bootstrapping")
)

func main() {
	flag.Parse()
	log.Print("started discv5-tools/network-measure")

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
		BootNodes: bootNodes,
		Logger:    log.New(os.Stderr, "crawler: ", log.LstdFlags|log.Lmsgprefix),
	}
	cr := crawler.New(cfg)
	cr.Start()
	defer cr.Stop()

	client, err := measure.Listen()
	if err != nil {
		log.Fatalf("the measurement client cannot be created: %v", err)
	}
	for {
		nd, err := cr.GetNode()
		if err != nil {
			log.Fatalf("the crawler stopped unexpectedly: %v", err)
		}
		go func(){
			client.Run(nd)
		}()
	}
}
