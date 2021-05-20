package a2s

import (
	"errors"
)

const (
	A2S_INFO_REQUEST  = 0x54
	A2S_INFO_RESPONSE = 0x49 // Source & up
)

var (
	ErrBadPacketHeader   = errors.New("Packet header mismatch")
	ErrUnsupportedHeader = errors.New("Unsupported protocol header")
)

type ServerInfo struct {
	// Protocol version used by the server.
	Protocol uint8 `json:"Protocol"`

	// Name of the server.
	Name string `json:"Name"`

	// Map the server has currently loaded.
	Map string `json:"Map"`

	// Name of the folder containing the game files.
	Folder string `json:"Folder"`

	// Full name of the game.
	Game string `json:"Game"`

	// Steam Application ID of game.
	ID uint16 `json:"AppID"`

	// Number of players on the server.
	Players uint8 `json:"Players"`

	// Maximum number of players the server reports it can hold.
	MaxPlayers uint8 `json:"MaxPlayers"`

	// Number of bots on the server.
	Bots uint8 `json:"Bots"`

	// Indicates the type of server
	// Rag Doll Kung Fu servers always return 0 for "Server type."
	ServerType ServerType `json:"ServerType"`

	// Indicates the operating system of the server
	ServerOS ServerOS `json:"ServerOS"`

	// Indicates whether the server requires a password
	Visibility bool `json:"Visibility"`

	// Specifies whether the server uses VAC
	VAC bool `json:"VAC"`

	// These fields only exist in a response if the server is running The Ship
	TheShip *TheShipInfo `json:"TheShip,omitempty"`

	// Version of the game installed on the server.
	Version string `json:"Version"`

	// If present, this specifies which additional data fields will be included.
	EDF uint8 `json:"EDF,omitempty"`

	ExtendedServerInfo *ExtendedServerInfo `json:"ExtendedServerInfo,omitempty"`

	SourceTV *SourceTVInfo `json:"SourceTV,omitempty"`
}

type TheShipInfo struct {
	Mode      TheShipMode `json:"Mode"`
	Witnesses uint8       `json:"Witnesses"`
	Duration  uint8       `json:"Duration"`
}

type ExtendedServerInfo struct {
	// The server's game port number.
	Port uint16 `json:"Port"`

	// Server's SteamID.
	SteamID uint64 `json:"SteamID"`

	// Tags that describe the game according to the server (for future use.)
	Keywords string `json:"Keywords"`

	// The server's 64-bit GameID. If this is present, a more accurate AppID is present in the low 24 bits. The earlier AppID could have been truncated as it was forced into 16-bit storage.
	GameID uint64 `json:"GameID"`
}

type SourceTVInfo struct {
	// Spectator port number for SourceTV.
	Port uint16 `json:"Port"`

	// Name of the spectator server for SourceTV.
	Name string `json:"Name"`
}

func (c *Client) QueryInfo() (*ServerInfo, error) {
	var builder PacketBuilder

	/*
		(FF FF FF FF) 54 53 6F 75 72 63 65 20 45 6E 67 69   ÿÿÿÿTSource Engi
		6E 65 20 51 75 65 72 79 00                        ne Query.

	*/
	builder.WriteBytes([]byte{
		0xFF, 0xFF, 0xFF, 0xFF, A2S_INFO_REQUEST,
	})

	builder.WriteCString("Source Engine Query")

	data, immediate, err := c.getChallenge(builder.Bytes(), A2S_INFO_RESPONSE)

	if err != nil {
		return nil, err
	}

	if !immediate {
		builder.WriteBytes(data)
		if err := c.send(builder.Bytes()); err != nil {
			return nil, err
		}

		data, err = c.receive()

		if err != nil {
			return nil, err
		}
	}

	/*
		Header	long	Always equal to -1 (0xFFFFFFFF). Means it isn't split.
		Payload
	*/

	reader := NewPacketReader(data)

	if reader.ReadInt32() != -1 {
		return nil, ErrBadPacketHeader
	}

	info := &ServerInfo{}

	header := reader.ReadUint8()
	if header != A2S_INFO_RESPONSE {
		return nil, ErrUnsupportedHeader
	}

	info.Protocol = reader.ReadUint8()

	info.Name = reader.ReadString()
	info.Map = reader.ReadString()
	info.Folder = reader.ReadString()
	info.Game = reader.ReadString()

	info.ID = reader.ReadUint16()

	info.Players = reader.ReadUint8()
	info.MaxPlayers = reader.ReadUint8()
	info.Bots = reader.ReadUint8()

	// Rag Doll Kung Fu servers always return 0 for "Server type."
	info.ServerType = ParseServerType(reader.ReadUint8())

	info.ServerOS = ParseServerOS(reader.ReadUint8())

	info.Visibility = reader.ReadUint8() == 1

	info.VAC = reader.ReadUint8() == 1

	if AppID(info.ID) == App_TheShip {
		info.TheShip = &TheShipInfo{}
		info.TheShip.Mode = ParseTheShipMode(reader.ReadUint8())
		info.TheShip.Witnesses = reader.ReadUint8()
		info.TheShip.Duration = reader.ReadUint8()
	}

	info.Version = reader.ReadString()

	// Start of EDF

	if !reader.More() {
		return info, nil
	}

	info.ExtendedServerInfo = &ExtendedServerInfo{}

	info.EDF = reader.ReadUint8()

	if (info.EDF & 0x80) != 0 {
		info.ExtendedServerInfo.Port = reader.ReadUint16()
	}

	if (info.EDF & 0x10) != 0 {
		info.ExtendedServerInfo.SteamID = reader.ReadUint64()
	}

	if (info.EDF & 0x40) != 0 {
		info.SourceTV = &SourceTVInfo{}
		info.SourceTV.Port = reader.ReadUint16()
		info.SourceTV.Name = reader.ReadString()
	}

	if (info.EDF & 0x20) != 0 {
		info.ExtendedServerInfo.Keywords = reader.ReadString()
	}

	if (info.EDF & 0x01) != 0 {
		info.ExtendedServerInfo.GameID = reader.ReadUint64()
	}

	return info, nil
}
