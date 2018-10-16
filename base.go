package a2s

import (
	"bytes"
	"encoding/binary"
)

const (
	SinglePacket = -1
	MultiPacket  = -2
)

type (
	short    int16
	long     int32
	float    float32
	longlong int64
)

type SourcePacket struct {
	Header           long
	ID               long
	IsCompressed     bool
	Total            byte
	Number           byte
	MaxSize          short
	DecompressedSize long
	CRC32Sum         long
}

func (c *Client) ReadInternal(buffer *bytes.Buffer, length int) error {
	if buffer.Len() == 0 {
		return ErrInvalidPacket
	}

	Packet := &SourcePacket{}

	binary.Read(buffer, binary.LittleEndian, &Packet.Header)

	switch Packet.Header {
	case SinglePacket:
		{
			// Nothing
		}
	case MultiPacket:
		{
			var (
				Packets  [][]byte
				ReadMore bool = false
			)

			binary.Read(buffer, binary.LittleEndian, &Packet.ID)

			Packet.IsCompressed = (Packet.ID & (1<<31 - 1)) != 0
		}
	}

	return nil
}
