package rpc

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "github.com/exc-git/mini-kv-store/api/v1"
	"github.com/exc-git/mini-kv-store/internal/raft"

	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedKVServiceServer
	node *raft.Node
}

func NewServer(node *raft.Node) *Server {
	return &Server{node: node}
}

func (s *Server) Get(ctx context.Context, req *pb.GetRequest) (*pb.GetResponse, error) {
	value, found := s.node.Get(req.Key)
	return &pb.GetResponse{
		Value: value,
		Found: found,
	}, nil
}

func (s *Server) Put(ctx context.Context, req *pb.PutRequest) (*pb.PutResponse, error) {
	err := s.node.ApplyCommand("SET", req.Key, req.Value)
	return &pb.PutResponse{
		Success: err == nil,
	}, err
}

func (s *Server) Delete(ctx context.Context, req *pb.DeleteRequest) (*pb.DeleteResponse, error) {
	err := s.node.ApplyCommand("DELETE", req.Key, "")
	return &pb.DeleteResponse{
		Success: err == nil,
	}, err
}

func StartGRPCServer(node *raft.Node, addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterKVServiceServer(grpcServer, NewServer(node))

	log.Printf("Starting gRPC server on %s", addr)
	return grpcServer.Serve(lis)
}
