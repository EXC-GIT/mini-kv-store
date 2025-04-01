package configs

import "time"

type RaftConfig struct {
	HeartbeatTimeout  time.Duration
	ElectionTimeout   time.Duration
	SnapshotInterval  time.Duration
	SnapshotThreshold uint64
	TrailingLogs      uint64
}

type Config struct {
	NodeID      string
	BindAddr    string
	RPCAddr     string
	HTTPAddr    string
	MetricsAddr string
	DataDir     string
	JoinAddr    string
	Raft        RaftConfig
}

func DefaultConfig() *Config {
	return &Config{
		NodeID:      "node1",
		BindAddr:    "127.0.0.1:9090",
		RPCAddr:     "127.0.0.1:9091",
		HTTPAddr:    "127.0.0.1:8080",
		MetricsAddr: "127.0.0.1:9100",
		DataDir:     "./data",
		Raft: RaftConfig{
			HeartbeatTimeout:  1 * time.Second,
			ElectionTimeout:   1 * time.Second,
			SnapshotInterval:  30 * time.Second,
			SnapshotThreshold: 1000,
			TrailingLogs:      10000,
		},
	}
}
