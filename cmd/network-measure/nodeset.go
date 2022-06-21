package main

import (
	clist "container/list"
	"fmt"
	"log"
	"time"

	"github.com/ethereum/go-ethereum/p2p/enode"
	"github.com/ppopth/discv5-tools/measure"
)

const (
	timeout = 10 * time.Minute
)

type node struct {
	nd     *enode.Node
	value  measure.Result
	expiry time.Time
}

type nodeSet struct {
	l   *clist.List
	ht  map[enode.ID]*clist.Element
	log *log.Logger
}

func newNodeset(logger *log.Logger) *nodeSet {
	return &nodeSet{
		l:   clist.New(),
		ht:  make(map[enode.ID]*clist.Element),
		log: logger,
	}
}

func (s *nodeSet) last() *node {
	if s.l.Len() == 0 {
		return nil
	} else {
		// We need to copy the node. Otherwise, it can be changed on the fly.
		return &(*s.l.Back().Value.(*node))
	}
}

func (s *nodeSet) remove(id enode.ID) {
	e := s.ht[id]
	if e != nil {
		s.l.Remove(e)
		s.log.Printf("removed id=%s nodeset={%v}", id.TerminalString(), s)
		delete(s.ht, id)
	}
}

func (s *nodeSet) refresh(id enode.ID) {
	e := s.ht[id]
	if e != nil {
		s.l.MoveToFront(e)
		e.Value.(*node).expiry = time.Now().Add(timeout)
		s.log.Printf("refreshed id=%s nodeset={%v}", id.TerminalString(), s)
	}
}

func (s *nodeSet) String() string {
	return fmt.Sprintf("len=%v", s.len())
}

func (s *nodeSet) len() int {
	return s.l.Len()
}

// Try to add a node and if it can be added, but won't actually add it.
func (s *nodeSet) dryAdd(n *enode.Node) bool {
	e, ok := s.ht[n.ID()]
	if !ok {
		// The node is not in the set.
		return true
	}
	if n.Seq() > e.Value.(*node).nd.Seq() {
		// The new node has a higher seq number.
		return true
	}
	return false
}

func (s *nodeSet) add(n *enode.Node, res measure.Result) {
	e, ok := s.ht[n.ID()]
	if !ok {
		// The node is not in the set.
		el := s.l.PushFront(&node{n, res, time.Now().Add(timeout)})
		s.ht[n.ID()] = el
		s.log.Printf("added id=%s result=%v nodeset={%v}", n.ID().TerminalString(), res, s)
		return
	}
	if n.Seq() > e.Value.(*node).nd.Seq() {
		// The new node has a higher seq number.
		e.Value = &node{n, res, time.Now().Add(timeout)}
		s.l.MoveToFront(e)
		s.log.Printf("updated id=%s result=%v nodeset={%v}", n.ID().TerminalString(), res, s)
	}
}
