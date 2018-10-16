package a2s

import (
	"bytes"
	"io"
)

func (c *Client) Read(length int) (*bytes.Buffer, error) {
	buf := make([]byte, length)

	_, err := io.ReadFull(c.conn, buf)

	if err != nil {
		return nil, err
	}

	buffer := bytes.NewBuffer(buf)

	c.ReadInternal(buffer, length)

	return buffer, nil
}
