package configs

import "time"

type Config struct {
	NodeID   string
	BindAddr string
	DataDir  string
	Raft     RaftConfig
}

type RaftConfig struct {
	HeartbeatTimeout  time.Duration
	ElectionTimeout   time.Duration
	SnapshotInterval  time.Duration
	SnapshotThreshold uint64
}

func DefaultConfig() *Config {
	return &Config{
		NodeID:   "node1",
		BindAddr: "127.0.0.1:9090",
		DataDir:  "./data",
		Raft: RaftConfig{
			HeartbeatTimeout:  1 * time.Second,
			ElectionTimeout:   1 * time.Second,
			SnapshotInterval:  30 * time.Second,
			SnapshotThreshold: 1000,
		},
	}
}
