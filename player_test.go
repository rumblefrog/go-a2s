package a2s

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayer(t *testing.T) {
	c, err := NewClient(*testHost)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	p, err := c.QueryPlayer()

	if err != nil {
		t.Error(err)
		return
	}

	JSON, _ := json.Marshal(p)

	fmt.Println(string(JSON))
}

// Example response from https://developer.valvesoftware.com/wiki/Server_queries#Response_Format_2
func validPlayerInfoPacket() []byte {
	return []byte{
		0xFF, 0xFF, 0xFF, 0xFF, 0x44, 0x02, 0x01, 0x5B, 0x44, 0x5D, 0x2D, 0x2D, 0x2D, 0x2D, 0x3E, 0x54,
		0x2E, 0x4E, 0x2E, 0x57, 0x3C, 0x2D, 0x2D, 0x2D, 0x2D, 0x00, 0x0E, 0x00, 0x00, 0x00, 0xB4, 0x97,
		0x00, 0x44, 0x02, 0x4B, 0x69, 0x6C, 0x6C, 0x65, 0x72, 0x20, 0x21, 0x21, 0x21, 0x00, 0x05, 0x00,
		0x00, 0x00, 0x69, 0x24, 0xD9, 0x43,
	}
}

func TestParsePlayerInfo(t *testing.T) {
	c, err := NewClient(*testHost, disableDial())
	assert.Nil(t, err, "NewClient should not return an error")
	defer c.Close()

	info, err := c.parsePlayerInfo(validPlayerInfoPacket())
	assert.Nil(t, err, "player info should not fail for a valid packet")
	assert.Equal(t, uint8(0x2), info.Count, "Player count should match")
	assert.Equal(t, 2, len(info.Players), "Player count should match actual number of players parsed")
	assert.Equal(t, uint8(1), info.Players[0].Index, "Player index should match")
	assert.Equal(t, "[D]---->T.N.W<----", info.Players[0].Name, "Player name should match")
	assert.Equal(t, float32(514.37036), info.Players[0].Duration, "Player duration should match")
	assert.Equal(t, uint32(14), info.Players[0].Score, "Player score should match")
	assert.Equal(t, uint8(2), info.Players[1].Index, "Player index should match")
	assert.Equal(t, "Killer !!!", info.Players[1].Name, "Player name should match")
	assert.Equal(t, float32(434.28445), info.Players[1].Duration, "Player duration should match")
	assert.Equal(t, uint32(5), info.Players[1].Score, "Player score should match")
}

func FuzzParsePlayerInfo(f *testing.F) {

	validPacket := validPlayerInfoPacket()

	// seed corpus from a valid packet
	for i := 0; i < len(validPacket); i++ {
		f.Add(validPacket[i:], int32(0))
		f.Add(validPacket[:i], int32(0))
		f.Add(validPacket[i:], int32(App_TheShip))
		f.Add(validPacket[:i], int32(App_TheShip))
	}

	f.Fuzz(func(t *testing.T, a []byte, appId int32) {
		c, err := NewClient(*testHost, SetAppID(appId), disableDial())
		assert.Nil(f, err, "NewClient should not return an error")
		defer c.Close()
		_, err = c.parsePlayerInfo(a)
		// sometimes the fuzzer can actually generate valid player info, if so, skip since we are only checking for panics when an invalid packet is returned from a server
		if err == nil {
			t.Skip()
		}
		assert.NotNil(t, err, "invalid parsePlayerInfo fuzzing should return an error")
	})
}
