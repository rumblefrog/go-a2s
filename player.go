package a2s

const (
	A2S_PLAYER_HEADER = 0x44 // Source & up
)

type PlayerInfo struct {
	// Always equal to 'D' (0x44)
	Header uint8

	// Number of players whose information was gathered.
	Players uint8 `json:"Players"`
}

type Player struct {
	// Index of player chunk starting from 0.
	Index uint8 `json:"Index"`

	// Name of the player.
	Name string `json:"Name"`

	// 	Player's score (usually "frags" or "kills".)
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

func (c *Client) QueryPlayer() {

}
