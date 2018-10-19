package a2s

type ServerOS int

const (
	ServerOS_Unknown ServerOS = iota
	ServerOS_Linux
	ServerOS_Windows
	ServerOS_Mac
)

func ParseServerOS(env uint8) ServerOS {
	switch env {
	case uint8('l'):
		return ServerOS_Linux
	case uint8('w'):
		return ServerOS_Windows
	case uint8('m'), uint8('o'):
		return ServerOS_Mac
	}

	return ServerOS_Unknown
}

func (os ServerOS) String() string {
	switch os {
	case ServerOS_Linux:
		return "Linux"
	case ServerOS_Windows:
		return "Windows"
	case ServerOS_Mac:
		return "Mac"
	default:
		return "Unknown"
	}
}
