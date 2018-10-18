package a2s

import (
	"bytes"
	"compress/bzip2"
	"errors"
	"hash/crc32"
)

const (
	MULTI_PACKET_RESPONSE_HEADER = -2
)

var (
	ErrPacketOutOfBound    = errors.New("Packet out of bound")
	ErrDuplicatePacket     = errors.New("Received same packet of same index")
	ErrWrongBz2Size        = errors.New("Bad bz2 decompression size")
	ErrMismatchBz2Checksum = errors.New("Bz2 decompressed checksum mismatches")
)

type MultiPacketHeader struct {
	// Size of the packet header
	Size int

	// Same as the Goldsource server meaning.
	// However, if the most significant bit is 1, then the response was compressed with bzip2 before being cut and sent.
	ID uint32

	// The total number of packets in the response.
	Total uint8

	// The number of the packet. Starts at 0.
	Number uint8

	/*
		(Orange Box Engine and above only.)
		Maximum size of packet before packet switching occurs.
		The default value is 1248 bytes (0x04E0), but the server administrator can decrease this.
		For older engine versions: the maximum and minimum size of the packet was unchangeable.
		AppIDs which are known not to contain this field: 215, 17550, 17700, and 240 when protocol = 7.
	*/
	SplitSize uint16

	// Indicates if payload is compressed w/bzip2
	Compressed bool

	// Payload
	Payload []byte
}

func (c *Client) parseMultiplePacketHeader(data []byte) (*MultiPacketHeader, error) {
	reader := NewPacketReader(data)

	if reader.ReadInt32() != -2 {
		return nil, ErrBadPacketHeader
	}

	header := &MultiPacketHeader{}

	header.ID = reader.ReadUint32()

	// https://github.com/xPaw/PHP-Source-Query/blob/f713415696d61cdd36639124fa573406360d8219/SourceQuery/BaseSocket.php#L78
	header.Compressed = (header.ID & uint32(0x80000000)) != 0

	header.Total = reader.ReadUint8()

	header.Number = reader.ReadUint8()

	if !c.preOrange {
		header.SplitSize = reader.ReadUint16()
	}

	header.Size = reader.Pos()

	// Include decompressed size & crc32sum as not all packets have this, so we'll read it later from the start if it's compressed
	header.Payload = data[header.Size:]

	return header, nil
}

func (c *Client) collectMultiplePacketResponse(data []byte) ([]byte, error) {
	header, err := c.parseMultiplePacketHeader(data)

	if err != nil {
		return []byte{}, ErrBadPacketHeader
	}

	packets := make([]*MultiPacketHeader, header.Total)

	received := 0
	fullSize := 0

	for {
		if int(header.Number) >= len(packets) {
			return nil, ErrPacketOutOfBound
		}

		if packets[header.Number] != nil {
			return nil, ErrDuplicatePacket
		}

		packets[header.Number] = header

		fullSize += len(header.Payload)

		received++

		if received == len(packets) {
			break
		}

		data, err := c.receive()

		if err != nil {
			return nil, err
		}

		header, err = c.parseMultiplePacketHeader(data)

		if err != nil {
			return nil, err
		}
	}

	payload := make([]byte, fullSize)

	cursor := 0

	for _, header := range packets {
		copy(payload[cursor:cursor+len(header.Payload)], header.Payload)
		cursor += len(header.Payload)
	}

	// Includes decompressed size & crc32 sum as that is unread yet, so it's included as part of payload
	reader := NewPacketReader(payload)

	if packets[0].Compressed {
		decompressedSize := reader.ReadUint32()
		checkSum := reader.ReadUint32()

		if decompressedSize > uint32(1024*1024) {
			return nil, ErrWrongBz2Size
		}

		decompressed := make([]byte, decompressedSize)

		bz2Reader := bzip2.NewReader(bytes.NewReader(data[reader.Pos():]))

		n, err := bz2Reader.Read(decompressed)

		if err != nil {
			return nil, err
		}

		if n != int(decompressedSize) {
			return nil, ErrWrongBz2Size
		}

		if crc32.ChecksumIEEE(decompressed) != checkSum {
			return nil, ErrMismatchBz2Checksum
		}

		payload = decompressed
	}

	return payload, nil
}
