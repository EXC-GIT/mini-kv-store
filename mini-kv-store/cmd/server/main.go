package main

import (
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/exc-git/mini-kv-store/internal/raft"
	"github.com/exc-git/mini-kv-store/internal/store"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
)

func main() {
	// Configuration
	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID("node1")

	// Create store and FSM
	store := store.NewMemoryStore()
	fsm := raft.NewFSM(store)

	// Setup data directory
	dataDir := "./data"
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		log.Fatal(err)
	}

	// Create persistent stores
	logStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-log.db"))
	if err != nil {
		log.Fatal(err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(dataDir, "raft-stable.db"))
	if err != nil {
		log.Fatal(err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(dataDir, 2, os.Stderr)
	if err != nil {
		log.Fatal(err)
	}

	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:9090")
	if err != nil {
		log.Fatal(err)
	}
	transport, err := raft.NewTCPTransportWithLogger(
		addr.String(),
		addr,
		3,
		10*time.Second,
		log.New(os.Stdout, "[raft-transport] ", log.LstdFlags),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create Raft node
	raftNode, err := raft.NewNode(config, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatal("Failed to create raft node:", err)
	}

	// Bootstrap cluster
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      config.LocalID,
				Address: transport.LocalAddr(),
			},
		},
	}
	f := raftNode.BootstrapCluster(configuration)
	if err := f.Error(); err != nil && err != raft.ErrCantBootstrap {
		log.Fatal("Failed to bootstrap cluster:", err)
	}

	// Monitor leadership
	go func() {
		for {
			select {
			case <-time.After(5 * time.Second):
				if raftNode.State() == raft.Leader {
					log.Println("I am the leader!")
				} else {
					log.Println("Current leader:", raftNode.Leader())
				}
			}
		}
	}()

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	// Shutdown
	if err := raftNode.Shutdown().Error(); err != nil {
		log.Printf("Error shutting down raft: %v", err)
	}
}
