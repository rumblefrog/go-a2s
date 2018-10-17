package a2s

type Visiblity int

const (
	Visiblity_Unknown Visiblity = iota
	Visiblity_Public
	Visiblity_Private
)

func ParseVisbility(v uint8) Visiblity {
	switch v {
	case 0:
		return Visiblity_Public
	case 1:
		return Visiblity_Private
	}

	return Visiblity_Unknown
}
