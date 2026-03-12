package redis

import "fmt"

// Client is a placeholder until the concrete Redis integration is added.
type Client struct {
	URL string
}

func Connect(url string) (*Client, error) {
	if url == "" {
		return nil, fmt.Errorf("redis url is required")
	}
	return &Client{URL: url}, nil
}

func (c *Client) Health() error {
	if c == nil || c.URL == "" {
		return fmt.Errorf("redis not configured")
	}
	return nil
}
