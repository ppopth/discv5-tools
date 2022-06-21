package main

import (
	"flag"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ethereum/go-ethereum/params"
	"github.com/ppopth/discv5-tools/crawler"
	"github.com/ppopth/discv5-tools/measure"
)

const (
	maxMeasurements = 20
	maxRefreshs     = 40
)

var (
	bootnodesFlag = flag.String("bootnodes", "", "Comma separated nodes used for bootstrapping")
	crawlFlag     = flag.Bool("crawl", false, "Crawl the DHT and measure every node found")
	enrFlag       = flag.String("enr", "", "The ENR of the node you want to measure")
)

var (
	// The mutex used to access the nodeset in multiple routines.
	lock    sync.Mutex
	nodeset *nodeSet
	timer   chan interface{}
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

	if *crawlFlag {
		crawl(bootNodes)
	} else if *enrFlag == "" {
		log.Fatal("please provide the ENR of the node you want to measure")
	} else {
		nd := enode.MustParse(*enrFlag)
		client, err := measure.Listen()
		if err != nil {
			log.Fatalf("the measurement client cannot be created: %v", err)
		}
		result, err := client.Run(nd)
		// TODO: display result
	}
}

func crawl(bootNodes []*enode.Node) {
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

	nodeset = newNodeset(log.New(os.Stderr, "nodeset: ", log.LstdFlags|log.Lmsgprefix))
	// Run a routine to check the nodes in the nodeset regularly if they are
	// still alive.
	timer = make(chan interface{})
	go gc(client)

	// This semaphore is used to limit the number of concurrent measurements.
	semaphore := make(chan interface{}, maxMeasurements)
	for {
		nd, err := cr.GetNode()
		if err != nil {
			log.Fatalf("the crawler stopped unexpectedly: %v", err)
		}
		// Check if we are interested in the ENR we just found.
		// If it's the ENR we already have or it's older than the one we
		// have, we aren't. Otherwise, we are.
		lock.Lock()
		if !nodeset.dryAdd(nd) {
			lock.Unlock()
			continue
		}
		lock.Unlock()

		var empty interface{}
		semaphore <- empty
		go func() {
			defer func() { <-semaphore }()
			result, err := client.Run(nd)
			if err != nil {
				log.Printf("error: %v\n", err)
				return
			}
			lock.Lock()
			defer lock.Unlock()
			emptied := nodeset.len() == 0
			nodeset.add(nd, *result)
			if emptied && nodeset.len() == 1 {
				go func() {
					<-time.After(nodeset.last().expiry.Sub(time.Now()))
					var empty interface{}
					timer <- empty
				}()
			}
		}()
	}
}

func gc(client *measure.Client) {
	// This semaphore is used to limit the number of concurrent refreshes.
	semaphore := make(chan interface{}, maxRefreshs)
	for range timer {
		var wg sync.WaitGroup
		lock.Lock()
		for e := nodeset.l.Back(); e != nil && e.Value.(*node).expiry.Before(time.Now()); e = e.Prev() {
			n := e.Value.(*node)
			var empty interface{}
			semaphore <- empty
			wg.Add(1)
			go func(n node) {
				defer wg.Done()
				defer func() { <-semaphore }()
				success := false
				for i := 0; i < 3; i++ {
					_, err := client.Send(n.nd)
					if err != nil {
						continue
					}
					success = true
					break
				}
				lock.Lock()
				defer lock.Unlock()
				// Check if the ENR of the node has changed or not.
				e := nodeset.ht[n.nd.ID()]
				if e == nil || e.Value.(*node).nd.Seq() != n.nd.Seq() {
					return
				}

				if success {
					nodeset.refresh(n.nd.ID())
				} else {
					nodeset.remove(n.nd.ID())
				}
			}(*n)
		}
		lock.Unlock()
		wg.Wait()

		// Check if we need to set another timer.
		lock.Lock()
		if nodeset.len() != 0 {
			duration := nodeset.last().expiry.Sub(time.Now())
			go func() {
				<-time.After(duration)
				var empty interface{}
				timer <- empty
			}()
		}
		lock.Unlock()
	}
}
