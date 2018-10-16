package a2s

import "errors"

var (
	ErrNilOption     = errors.New("Invalid option")
	ErrInvalidPacket = errors.New("Invalid packet")
	ErrOutOfBounds   = errors.New("Read out of bounds")
)
