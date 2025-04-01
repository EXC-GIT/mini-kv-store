package rpc

import (
	"context"
	"log"
	"net"
	"time"

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
	var err error
	if req.Ttl > 0 {
		err = s.store.SetWithTTL(req.Key, req.Value, time.Duration(req.Ttl)*time.Second)
	} else {
		err = s.store.Set(req.Key, req.Value)
	}
	return &SetResponse{Success: err == nil}, err
}

// Delete handles key removal requests
func (s *server) Delete(ctx context.Context, req *DeleteRequest) (*DeleteResponse, error) {
	err := s.store.Delete(req.Key)
	return &DeleteResponse{Success: err == nil}, err
}

// Join handles node join requests
func (s *server) Join(ctx context.Context, req *JoinRequest) (*JoinResponse, error) {
	err := s.raftNode.Join(ctx, req.NodeId, req.RaftAddress)
	return &JoinResponse{Success: err == nil}, err
}

// Stats returns cluster statistics
func (s *server) Stats(ctx context.Context, req *StatsRequest) (*StatsResponse, error) {
	return s.raftNode.Stats(), nil
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
