package rpc

import (
	"context"
	"time"

	common "github.com/exc-git/mini-kv-store/internal/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client wraps the gRPC client connection
type Client struct {
	conn   *grpc.ClientConn
	client KVStoreClient
}

// NewClient creates a new gRPC client
func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, err
	}

	return &Client{
		conn:   conn,
		client: NewKVStoreClient(conn),
	}, nil
}

// Close terminates the client connection
func (c *Client) Close() error {
	return c.conn.Close()
}

// Get retrieves a value by key
func (c *Client) Get(ctx context.Context, key string) (string, error) {
	resp, err := c.client.Get(ctx, &GetRequest{Key: key})
	if err != nil {
		return "", err
	}
	return resp.Value, nil
}

// Set stores a key-value pair
func (c *Client) Set(ctx context.Context, key, value string) error {
	_, err := c.client.Set(ctx, &SetRequest{Key: key, Value: value})
	return err
}

// SetWithTTL stores a key-value pair with expiration
func (c *Client) SetWithTTL(ctx context.Context, key, value string, ttl time.Duration) error {
	_, err := c.client.SetWithTTL(ctx, &SetRequest{
		Key:   key,
		Value: value,
		Ttl:   int64(ttl.Seconds()),
	})
	return err
}

// Delete removes a key
func (c *Client) Delete(ctx context.Context, key string) error {
	_, err := c.client.Delete(ctx, &DeleteRequest{Key: key})
	return err
}

// Stats returns cluster statistics
func (c *Client) Stats(ctx context.Context) (*common.StatsResponse, error) {
	return c.client.Stats(ctx, &common.StatsRequest{})
}

// Join adds a node to the cluster
func (c *Client) Join(ctx context.Context, nodeID, raftAddr string) error {
	_, err := c.client.Join(ctx, &common.JoinRequest{
		NodeId:      nodeID,
		RaftAddress: raftAddr,
	})
	return err
}
