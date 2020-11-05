package a2s

import (
	"errors"
	"fmt"
	"net"
	"strings"
	"time"
)

const (
	DefaultTimeout       = time.Second * 3
	DefaultPort          = 27015
	DefaultMaxPacketSize = 1400
)

var (
	ErrNilOption = errors.New("Invalid client option")
)

type Client struct {
	addr          string
	conn          net.Conn
	timeout       time.Duration
	maxPacketSize uint32
	buffer        []byte
	preOrange     bool
	appid         AppID
	wait          time.Duration
	next          time.Time
}

func TimeoutOption(timeout time.Duration) func(*Client) error {
	return func(c *Client) error {
		c.timeout = timeout

		return nil
	}
}

func PreOrangeBox(pre bool) func(*Client) error {
	return func(c *Client) error {
		c.preOrange = pre

		return nil
	}
}

func SetAppID(appid int32) func(*Client) error {
	return func(c *Client) error {
		c.appid = AppID(appid)

		return nil
	}
}

// SetMaxPacketSize changes the maximum buffer size of a UDP packet
// Note that some games such as squad may use a non-standard packet size
// Refer to the game documentation to see if this needs to be changed
func SetMaxPacketSize(size uint32) func(*Client) error {
	return func(c *Client) error {
		c.maxPacketSize = size

		return nil
	}
}

func NewClient(addr string, options ...func(*Client) error) (c *Client, err error) {
	c = &Client{
		timeout:       DefaultTimeout,
		addr:          addr,
		maxPacketSize: DefaultMaxPacketSize,
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

	c.buffer = make([]byte, 0, c.maxPacketSize)

	return c, nil
}

func (c *Client) send(data []byte) error {
	c.enforceRateLimit()

	defer c.setNextQueryTime()

	if c.timeout > 0 {
		c.conn.SetWriteDeadline(c.extendedDeadline())
	}

	_, err := c.conn.Write(data)

	return err
}

func (c *Client) receive() ([]byte, error) {
	defer c.setNextQueryTime()

	if c.timeout > 0 {
		c.conn.SetReadDeadline(c.extendedDeadline())
	}

	size, err := c.conn.Read(c.buffer[0:c.maxPacketSize])

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

func (c *Client) extendedDeadline() time.Time {
	return time.Now().Add(c.timeout)
}

func (c *Client) setNextQueryTime() {
	if c.wait != 0 {
		c.next = time.Now().Add(c.wait)
	}
}

func (c *Client) enforceRateLimit() {
	if c.wait == 0 {
		return
	}

	wait := c.next.Sub(time.Now())
	if wait > 0 {
		time.Sleep(wait)
	}
}
