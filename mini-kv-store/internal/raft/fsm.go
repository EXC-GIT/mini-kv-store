package raft

import (
	"encoding/json"
	"io"

	"github.com/exc-git/mini-kv-store/internal/store"
	"github.com/hashicorp/raft"
)

type command struct {
	Op    string `json:"op"` // "set" or "delete"
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
}

type FSM struct {
	store store.Store
}

func NewFSM(store store.Store) *FSM {
	return &FSM{store: store}
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	switch cmd.Op {
	case "set":
		return f.store.Set(cmd.Key, cmd.Value)
	case "delete":
		return f.store.Delete(cmd.Key)
	default:
		return raft.ErrUnsupportedProtocol
	}
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	// Implement if needed
	return nil, nil
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	// Implement if needed
	return nil
}
