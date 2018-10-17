package a2s

type VAC int

const (
	VAC_Unknown VAC = iota
	VAC_Unsecured
	VAC_Secured
)

func ParseVAC(vac uint8) VAC {
	switch vac {
	case 0:
		return VAC_Unsecured
	case 1:
		return VAC_Secured
	}

	return VAC_Unknown
}
