package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/exc-git/mini-kv-store/configs"
	"github.com/exc-git/mini-kv-store/internal/store"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"github.com/hashicorp/raft"
	raft2 "github.com/exc-git/mini-kv-store/internal/raft"
)

func main() {
	cfg := configs.DefaultConfig()

	// Setup data directory
	if err := os.MkdirAll(cfg.DataDir, 0700); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// Create stores
	store := store.NewMemoryStore()
	fsm := raft2.NewFSM(store)

	logStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-log.db"))
	if err != nil {
		log.Fatal("Failed to create log store:", err)
	}

	stableStore, err := raftboltdb.NewBoltStore(filepath.Join(cfg.DataDir, "raft-stable.db"))
	if err != nil {
		log.Fatal("Failed to create stable store:", err)
	}

	snapshotStore, err := raft.NewFileSnapshotStore(cfg.DataDir, 2, os.Stderr)
	if err != nil {
		log.Fatal("Failed to create snapshot store:", err)
	}

	// Create hraft transport
	addr, err := net.ResolveTCPAddr("tcp", cfg.BindAddr)
	if err != nil {
		log.Fatal("Failed to resolve address:", err)
	}
	transport, err := raft.NewTCPTransport(
		addr.String(),
		addr,
		3,
		10*time.Second,
		os.Stdout,
	)

	if err != nil {
		log.Fatal("Failed to create transport:", err)
	}

	// Setup hraft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(cfg.NodeID)
	raftConfig.SnapshotInterval = cfg.Raft.SnapshotInterval
	raftConfig.SnapshotThreshold = cfg.Raft.SnapshotThreshold

	raftNode, err := raft2.NewNode(raftConfig, fsm, logStore, stableStore, snapshotStore, transport)
	if err != nil {
		log.Fatalf("Failed to create raft node: %v", err)
	}

	configuration := raft.Configuration{Servers: []raft.Server{{ID: raft.ServerID(cfg.NodeID), Address: transport.LocalAddr()}},
	}
	if cfg.JoinAddr == "" {
		f := raftNode.BootstrapCluster(configuration)
		if err := f.Error(); err != nil {
			log.Fatal("Failed to bootstrap cluster:", err)
		}
	} else {
		time.Sleep(5 * time.Second)
		if err := joinCluster(cfg.JoinAddr, cfg.NodeID, cfg.BindAddr); err != nil {
			log.Printf("Failed to join cluster: %v", err)
		}
	}

	// Start admin server
	go startAdminServer(raftNode, cfg.HTTPAddr)

	// Wait for shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	if err := raftNode.Shutdown().Error(); err != nil {
		log.Printf("Error shutting down raft node: %v", err)
	}
}

func joinCluster(joinAddr, nodeID, raftAddr string) error {
	client := http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/join?node_id=%s&raft_addr=%s", joinAddr, nodeID, raftAddr))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to join cluster: %s", resp.Status)
	}
	return nil
}

func startAdminServer(raftNode *raft2.Node, addr string) {
	http.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		nodeID := r.URL.Query().Get("node_id")
		raftAddr := r.URL.Query().Get("raft_addr")

		if err := raftNode.Join(r.Context(), nodeID, raftAddr); err != nil {
			log.Printf("Error joining node: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)

			return
		}
		w.WriteHeader(http.StatusOK)
	})

	log.Println("Admin server running on", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
