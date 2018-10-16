package a2s

import (
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultTimeout = time.Second * 10
	DefaultPort    = 27015
)

type Client struct {
	addr    string
	conn    net.Conn
	timeout time.Duration
}

func TimeoutOption(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout

		return nil
	}
}

func NewClient(addr string, options ...func(*Client) error) (c *Client, err error) {
	c = &Client{
		timeout: DefaultTimeout,
		addr:    addr,
	}

	for _, f := range options {
		if f == nil {
			return nil, ErrNilOption
		}
		if err = f(c); err != nil {
			return nil, err
		}
	}

	if !strings.Contains(c.addr, ":") {
		c.addr = fmt.Sprintf("%s:%d", c.addr, DefaultPort)
	}

	if c.conn, err = net.DialTimeout("udp", c.addr, c.timeout); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
