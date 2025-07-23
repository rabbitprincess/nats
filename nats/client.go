package nats

import (
	"fmt"

	nats_go "github.com/nats-io/nats.go"
)

type Client struct {
	Conn *nats_go.Conn
	Host string
	Port int
}

func NewClient(host string, port int) *Client {
	return &Client{
		Host: host,
		Port: port,
	}
}

func (c *Client) Address() string {
	return fmt.Sprintf("nats://%s:%d", c.Host, c.Port)
}

func (c *Client) Connect() error {
	var err error
	c.Conn, err = nats_go.Connect(c.Address())
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) Drain() {
	if c.Conn != nil {
		c.Conn.Drain()
	}
}
