package configs

import "time"

type NodeConfig struct {
	ID      string
	Address string
}

type ClusterConfig struct {
	NodeID   string
	BindAddr string
	DataDir  string
	JoinAddr string
	Nodes    []NodeConfig
	Raft     RaftConfig
}

func DefaultClusterConfig() *ClusterConfig {
	return &ClusterConfig{
		NodeID:   "node1",
		BindAddr: "127.0.0.1:9090",
		DataDir:  "./data",
		Nodes: []NodeConfig{
			{ID: "node1", Address: "127.0.0.1:9090"},
			{ID: "node2", Address: "127.0.0.1:9091"},
			{ID: "node3", Address: "127.0.0.1:9092"},
		},
		Raft: RaftConfig{
			HeartbeatTimeout:  1 * time.Second,
			ElectionTimeout:   1 * time.Second,
			SnapshotInterval:  30 * time.Second,
			SnapshotThreshold: 1000,
			TrailingLogs:      10000,
		},
	}
}
