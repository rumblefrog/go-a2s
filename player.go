package a2s

const (
	A2S_PLAYER_HEADER = 0x44 // Source & up
)

type PlayerInfo struct {
	// Always equal to 'D' (0x44)
	Header uint8

	// Number of players whose information was gathered.
	Players uint8
}

type Player struct {
	// Index of player chunk starting from 0.
	Index uint8

	// Name of the player.
	Name string

	// 	Player's score (usually "frags" or "kills".)
	Score uint32

	// Time (in seconds) player has been connected to the server.
	Duration float32

	// The Ship additional player info
	TheShip *TheShipPlayer
}

type TheShipPlayer struct {
	// Player's deaths
	Deaths uint32

	// Player's money
	Money uint32
}

func (c *Client) QueryPlayer() {

}
