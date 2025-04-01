package rpc

import (
	"context"
	"fmt"

	pb "github.com/exc-git/mini-kv-store/api/v1"

	"google.golang.org/grpc"
)

type Client struct {
	conn   *grpc.ClientConn
	client pb.KVServiceClient
}

func NewClient(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("failed to dial server: %v", err)
	}

	return &Client{
		conn:   conn,
		client: pb.NewKVServiceClient(conn),
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Get(key string) (string, bool, error) {
	resp, err := c.client.Get(context.Background(), &pb.GetRequest{Key: key})
	if err != nil {
		return "", false, err
	}
	return resp.Value, resp.Found, nil
}

func (c *Client) Put(key, value string) error {
	_, err := c.client.Put(context.Background(), &pb.PutRequest{
		Key:   key,
		Value: value,
	})
	return err
}

func (c *Client) Delete(key string) error {
	_, err := c.client.Delete(context.Background(), &pb.DeleteRequest{Key: key})
	return err
}
