package rpc

import (
	"context"
	"log"
	"net"
	"time"

	common "github.com/exc-git/mini-kv-store/internal/common"
	"github.com/exc-git/mini-kv-store/internal/raft"
	"github.com/exc-git/mini-kv-store/internal/store"
	"google.golang.org/grpc"
)

// server implements the KVStoreServer interface
type server struct {
	UnimplementedKVStoreServer
	raftNode *raft.Node
	store    store.Store
}

// NewServer creates a new gRPC server instance
func NewServer(raftNode *raft.Node, store store.Store) *server {
	return &server{
		raftNode: raftNode,
		store:    store,
	}
}

// Get handles key lookup requests
func (s *server) Get(ctx context.Context, req *GetRequest) (*GetResponse, error) {
	value, err := s.store.Get(req.Key)
	if err != nil {
		return nil, err
	}
	return &GetResponse{Value: value}, nil
}

// Set handles key-value storage requests
func (s *server) Set(ctx context.Context, req *SetRequest) (*SetResponse, error) {
	ttl := time.Duration(req.Ttl) * time.Second
	err := s.raftNode.Set(req.Key, req.Value, ttl)
	return &SetResponse{Success: err == nil}, err
}

// Delete handles key removal requests
func (s *server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	err := s.raftNode.Delete(req.Key)
	return &DeleteResponse{Success: err == nil}, err
}

// Join handles node join requests
func (s *server) Join(ctx context.Context, req *common.JoinRequest) (*common.JoinResponse, error) {
	err := s.raftNode.Join(ctx, req.NodeId, req.RaftAddress)
	return &common.JoinResponse{Success: err == nil}, err
}

// Stats returns cluster statistics
func (s *server) Stats(ctx context.Context, req *common.StatsRequest) (*common.StatsResponse, error) {
	statsMap := s.raftNode.Stats()
	statsResponse := &common.StatsResponse{
		Leader:      statsMap["leader"].(string),
		Nodes:       statsMap["nodes"].([]string),
		CommitIndex: statsMap["commit_index"].(uint64),
		LastApplied: statsMap["last_applied"].(uint64),
	}
	return statsResponse, nil
}

// StartServer starts the gRPC server
func StartServer(addr string, raftNode *raft.Node, store store.Store) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	RegisterKVStoreServer(s, NewServer(raftNode, store))

	log.Printf("gRPC server running on %s", addr)
	return s.Serve(lis)
}
