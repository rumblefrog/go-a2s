package a2s

const (
	MULTI_PACKET_RESPONSE_HEADER = -2
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

func (c *Client) ParseMultiplePacketHeader(data []byte) (*MultiPacketHeader, error) {
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

	if !c.pre_orange {
		header.SplitSize = reader.ReadUint16()
	}

	header.Size = reader.Pos()

	header.Payload = data[header.Size:]

	return header, nil
}
