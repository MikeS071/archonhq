package nats

import (
	"context"
	"fmt"
	"time"

	natsgo "github.com/nats-io/nats.go"
)

type Client struct {
	Conn *natsgo.Conn
}

func Connect(url string) (*Client, error) {
	nc, err := natsgo.Connect(url, natsgo.Timeout(3*time.Second))
	if err != nil {
		return nil, fmt.Errorf("connect nats: %w", err)
	}
	return &Client{Conn: nc}, nil
}

func (c *Client) Health(_ context.Context) error {
	if c.Conn == nil || !c.Conn.IsConnected() {
		return fmt.Errorf("nats not connected")
	}
	return nil
}

func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
}
