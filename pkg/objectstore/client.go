package objectstore

import "fmt"

// Config controls the object-store client wiring.
type Config struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	Region    string
}

type Client struct {
	Config Config
}

func New(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("object store endpoint and bucket are required")
	}
	return &Client{Config: cfg}, nil
}

func (c *Client) Health() error {
	if c == nil || c.Config.Endpoint == "" {
		return fmt.Errorf("object store not configured")
	}
	return nil
}
