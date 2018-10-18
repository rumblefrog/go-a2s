package a2s

import (
	"encoding/binary"
	"errors"
)

const (
	A2S_PLAYER_REQUEST  = 0x55
	A2S_PLAYER_RESPONSE = 0x44 // Source & up
)

var (
	ErrBadRulesReply = errors.New("Bad rules reply")
)

type PlayerInfo struct {
	// Always equal to 'D' (0x44)
	Header uint8

	// Number of players whose information was gathered.
	Count uint8 `json:"Count"`

	// Slice of pointers to each Player
	Players []*Player `json:"Players"`
}

type Player struct {
	// Index of player chunk starting from 0.
	Index uint8 `json:"Index"`

	// Name of the player.
	Name string `json:"Name"`

	// Player's score (usually "frags" or "kills".)
	Score uint32 `json:"Score"`

	// Time (in seconds) player has been connected to the server.
	Duration float32 `json:"Duration"`

	// The Ship additional player info
	TheShip *TheShipPlayer `json:"TheShip,omitempty"`
}

type TheShipPlayer struct {
	// Player's deaths
	Deaths uint32 `json:"Deaths"`

	// Player's money
	Money uint32 `json:"Money"`
}

func (c *Client) QueryPlayer() (*PlayerInfo, error) {
	/*
		A2S_PLAYER

		Request Format

		Header	byte	'U' (0x55)
		Challenge	int	Challenge number, or -1 (0xFFFFFFFF) to receive a challenge number.

		FF FF FF FF 55 FF FF FF FF                         ÿÿÿÿUÿÿÿÿ"

		Example A2S_PLAYER request with the received challenge number:

		FF FF FF FF 55 4B A1 D5 22                         ÿÿÿÿUÿÿÿÿ"
	*/

	data, immediate, err := c.GetChallenge(A2S_PLAYER_REQUEST, A2S_PLAYER_RESPONSE)

	if err != nil {
		return nil, err
	}

	if !immediate {
		if err := c.Send([]byte{
			0xff, 0xff, 0xff, 0xff,
			A2S_PLAYER_REQUEST,
			data[0], data[1], data[2], data[3],
		}); err != nil {
			return nil, err
		}

		data, err = c.Receive()

		if err != nil {
			return nil, err
		}
	}

	// Read header (long 4 bytes)
	switch int32(binary.LittleEndian.Uint32(data)) {
	case -1:
		return c.ParsePlayerInfo(data)
	case -2:
		data, err = c.CollectMultiplePacketResponse(data)

		if err != nil {
			return nil, err
		}

		return c.ParsePlayerInfo(data)
	}

	return nil, ErrBadPacketHeader
}

func (c *Client) ParsePlayerInfo(data []byte) (*PlayerInfo, error) {
	reader := NewPacketReader(data)

	// Simple response now

	if reader.ReadInt32() != -1 {
		return nil, ErrBadPacketHeader
	}

	if reader.ReadUint8() != A2S_PLAYER_RESPONSE {
		return nil, ErrBadRulesReply
	}

	info := &PlayerInfo{}

	info.Count = reader.ReadUint8()

	var player *Player

	for i := 0; i < int(info.Count); i++ {
		player = &Player{}

		player.Index = reader.ReadUint8()
		player.Name = reader.ReadString()
		player.Score = reader.ReadUint32()
		player.Duration = reader.ReadFloat32()

		/*
			The Ship additional player info

			Only if client AppID is set to 2400
		*/
		if c.appid == App_TheShip {
			player.TheShip = &TheShipPlayer{}

			player.TheShip.Deaths = reader.ReadUint32()
			player.TheShip.Money = reader.ReadUint32()
		}

		info.Players = append(info.Players, player)
	}

	return info, nil
}
