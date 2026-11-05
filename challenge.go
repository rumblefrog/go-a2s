package a2s

import (
	"errors"
)

const (
	A2S_PLAYER_CHALLENGE_REPLY_HEADER = 0x41

	// MaxChallengeRetries is how many new challenges a query answers after the first one
	// before it gives up with ErrTooManyChallenges. A server can answer a request that
	// carries a challenge with a new challenge, for example after its challenge expired.
	MaxChallengeRetries = 5
)

var (
	ErrBadChallengeResponse = errors.New("Bad challenge response")
	ErrTooManyChallenges    = errors.New("Server kept answering with a new challenge")
)

func (c *Client) getChallenge(header []byte, fullResult byte) ([]byte, bool, error) {
	if err := c.send(header); err != nil {
		return nil, false, err
	}

	data, err := c.receive()

	if err != nil {
		return nil, false, err
	}

	reader := NewPacketReader(data)

	switch int32(reader.ReadUint32()) {
	case -2:
		// We received an unexpected full reply
		return data, true, nil
	case -1:
		// Continue
	default:
		return nil, false, ErrBadPacketHeader
	}

	switch reader.ReadUint8() {
	case A2S_PLAYER_CHALLENGE_REPLY_HEADER:
		// Received a challenge number

		return data[reader.Pos() : reader.Pos()+4], false, nil
	case fullResult:
		// Received full result

		return data, true, nil
	}

	return nil, false, ErrBadChallengeResponse
}

// challengeOf returns the challenge number in data when data is a challenge reply
// (FF FF FF FF 41 followed by 4 bytes), and false otherwise.
func challengeOf(data []byte) ([]byte, bool) {
	if len(data) < 9 {
		return nil, false
	}

	reader := NewPacketReader(data)

	if reader.ReadInt32() != -1 || reader.ReadUint8() != A2S_PLAYER_CHALLENGE_REPLY_HEADER {
		return nil, false
	}

	return data[5:9], true
}

// requestWithChallenge sends the request that request builds for challenge and returns the
// reply. When the server answers with a new challenge instead of the result, the request is
// sent again with that challenge, up to MaxChallengeRetries times.
func (c *Client) requestWithChallenge(request func(challenge []byte) []byte, challenge []byte) ([]byte, error) {
	for retries := 0; ; retries++ {
		if err := c.send(request(challenge)); err != nil {
			return nil, err
		}

		data, err := c.receive()

		if err != nil {
			return nil, err
		}

		next, ok := challengeOf(data)
		if !ok {
			return data, nil
		}

		if retries == MaxChallengeRetries {
			return nil, ErrTooManyChallenges
		}

		challenge = next
	}
}
