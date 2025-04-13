package raft

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hashicorp/raft"
)

var ErrNotLeader = errors.New("not the leader")

type Node struct {
	*raft.Raft
}

func NewNode(config *raft.Config, fsm raft.FSM, logStore raft.LogStore, stableStore raft.StableStore, snapshots raft.SnapshotStore, transport raft.Transport) (*Node, error) {
	r, err := raft.NewRaft(config, fsm, logStore, stableStore, snapshots, transport)

	if err != nil {
		return nil, err
	}

	return &Node{r}, nil
}

func (n *Node) Set(key, value string, ttl time.Duration) error {
	if n.State() != raft.Leader {
		return ErrNotLeader
	}

	cmd := command{
		Op:    "set",
		Key:   key,
		Value: value,
	}
	if ttl > 0 {
		cmd.TTL = int64(ttl.Seconds())
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	f := n.Apply(cmdBytes, 10*time.Second)
	return f.Error()
}

func (n *Node) Delete(key string) error {
	if n.State() != raft.Leader {
		return ErrNotLeader
	}

	cmd := command{
		Op:  "delete",
		Key: key,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	f := n.Apply(cmdBytes, 10*time.Second)
	return f.Error()
}

// Add these methods to your Node struct
func (n *Node) Join(ctx context.Context, nodeID, raftAddr string) error {
	if n.State() != raft.Leader {
		return ErrNotLeader
	}

	f := n.AddVoter(
		raft.ServerID(nodeID),
		raft.ServerAddress(raftAddr),
		0, 10*time.Second,
	)
	return f.Error()
}

func (n *Node) Stats() map[string]any {
	stats := map[string]any{
		"leader":        string(n.Leader()),
		"nodes":         n.getClusterNodes(),
		"commit_index":  n.CommitIndex(),
		"last_applied":  n.LastIndex(),
	}
	return stats
}

func (n *Node) getClusterNodes() []string {
	future := n.GetConfiguration()
	if err := future.Error(); err != nil {
		return nil
	}

	var nodes []string
	for _, server := range future.Configuration().Servers {
		nodes = append(nodes, string(server.Address))
	}
	return nodes
}
