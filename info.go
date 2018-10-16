package a2s

const (
	A2SHeader = 0x49
)

type A2SInfo struct {
	Protocol      byte
	Name          string
	Map           string
	Folder        string
	Game          string
	ID            int
	Players       byte
	MaxPlayers    byte
	Bots          byte
	ServerType    byte
	Environment   byte
	Visibility    byte
	VAC           byte
	Version       string
	EDF           byte
	Port          int
	SteamID       int64
	SpectatorPort int
	SpectatorName string
	KeyWords      string
	GameID        int64
}

func (c *Client) GetInfo() {

}
