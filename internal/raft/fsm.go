package raft

import (
	"encoding/json"
	"io"
	"sync"
	"time"

	"github.com/exc-git/mini-kv-store/internal/store"
	"github.com/hashicorp/raft"
)

type command struct {
	Op    string `json:"op"`
	Key   string `json:"key"`
	Value string `json:"value,omitempty"`
	TTL   int64  `json:"ttl,omitempty"`
}

type FSM struct {
	store store.Store
	mu    sync.Mutex
}

func NewFSM(store store.Store) *FSM {
	return &FSM{store: store}
}

func (f *FSM) Apply(log *raft.Log) interface{} {
	var cmd command
	if err := json.Unmarshal(log.Data, &cmd); err != nil {
		return err
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	switch cmd.Op {
	case "set":
		if cmd.TTL > 0 {
			return f.store.SetWithTTL(cmd.Key, cmd.Value, time.Duration(cmd.TTL)*time.Second)
		}
		return f.store.Set(cmd.Key, cmd.Value)
	case "delete":
		return f.store.Delete(cmd.Key)
	default:
		return raft.ErrUnsupportedProtocol
	}
}

func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	return nil, nil // Implement if needed
}

func (f *FSM) Restore(rc io.ReadCloser) error {
	return nil // Implement if needed
}
