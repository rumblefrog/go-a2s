package a2s

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultTimeout = time.Second * 10
	DefaultPort    = 27015
)

var (
	ErrNilOption = errors.New("Invalid client option")
)

type Client struct {
	addr       string
	conn       net.Conn
	timeout    time.Duration
	buffer     [MaxPacketSize]byte
	pre_orange bool
	appid      AppID
}

func TimeoutOption(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout

		return nil
	}
}

func PreOrangeBox(pre bool) func(*Client) error {
	return func(c *Client) error {
		c.pre_orange = pre

		return nil
	}
}

func SetAppID(appid int32) func(*Client) error {
	return func(c *Client) error {
		c.appid = AppID(appid)

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

func (c *Client) Send(data []byte) error {
	_, err := c.conn.Write(data)

	return err
}

func (c *Client) Receive() ([]byte, error) {
	size, err := c.conn.Read(c.buffer[0:MaxPacketSize])

	if err != nil {
		return nil, err
	}

	buffer := make([]byte, size)

	copy(buffer, c.buffer[:size])

	return buffer, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
