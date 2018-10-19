package a2s

type TheShipMode int

/*
	0 for Hunt
	1 for Elimination
	2 for Duel
	3 for Deathmatch
	4 for VIP Team
	5 for Team Elimination
*/

const (
	TheShipMode_Unknown TheShipMode = iota
	TheShipMode_Hunt
	TheShipMode_Elimination
	TheShipMode_Duel
	TheShipMode_Deathmatch
	TheShipMode_VIP_Team
	TheShipMode_Team_Elimination
)

func ParseTheShipMode(m uint8) TheShipMode {
	switch m {
	case 0:
		return TheShipMode_Hunt
	case 1:
		return TheShipMode_Elimination
	case 2:
		return TheShipMode_Duel
	case 3:
		return TheShipMode_Deathmatch
	case 4:
		return TheShipMode_VIP_Team
	case 5:
		return TheShipMode_Team_Elimination
	}

	return TheShipMode_Unknown
}

func (m TheShipMode) String() string {
	switch m {
	case TheShipMode_Hunt:
		return "Hunt"
	case TheShipMode_Elimination:
		return "Elimination"
	case TheShipMode_Duel:
		return "Duel"
	case TheShipMode_Deathmatch:
		return "Deathmatch"
	case TheShipMode_VIP_Team:
		return "VIP Team"
	case TheShipMode_Team_Elimination:
		return "Team Elimination"
	default:
		return "Unknown"
	}
}
