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

func (n *Node) Set(key, value string, timeout time.Duration) error {
	if n.State() != raft.Leader {
		return ErrNotLeader
	}

	cmd := map[string]string{
		"op":    "set",
		"key":   key,
		"value": value,
	}

	cmdBytes, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	f := n.Apply(cmdBytes, 0)
	return f.Error()
}

func (n *Node) GetLeaderAddress() string {
	return string(n.Leader())
}
