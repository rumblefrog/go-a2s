package a2s

import (
	"bytes"
	"encoding/binary"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// rechallengeServer is a local UDP server that answers the first request of a query with a
// challenge, and then answers `extra` more requests with a new challenge each before it sends
// response. A request that does not carry the newest challenge gets yet another challenge,
// like a real server whose challenge expired.
type rechallengeServer struct {
	conn     net.PacketConn
	response []byte
	extra    int

	mu sync.Mutex
	// requests holds the challenge each request carried, in order.
	requests [][]byte
}

func newRechallengeServer(t *testing.T, response []byte, extra int) *rechallengeServer {
	t.Helper()

	conn, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	s := &rechallengeServer{conn: conn, response: response, extra: extra}

	t.Cleanup(func() { conn.Close() })

	go s.serve()

	return s
}

func (s *rechallengeServer) serve() {
	buf := make([]byte, 1400)
	var current uint32 = 1000
	sent := 0

	for {
		n, addr, err := s.conn.ReadFrom(buf)
		if err != nil {
			return
		}

		req := append([]byte{}, buf[:n]...)
		got := req[len(req)-4:]
		s.mu.Lock()
		s.requests = append(s.requests, got)
		s.mu.Unlock()

		challenge := make([]byte, 4)
		binary.LittleEndian.PutUint32(challenge, current)

		var reply []byte
		if sent > s.extra && bytes.Equal(got, challenge) {
			reply = s.response
		} else {
			current++
			binary.LittleEndian.PutUint32(challenge, current)
			reply = append([]byte{0xFF, 0xFF, 0xFF, 0xFF, A2S_PLAYER_CHALLENGE_REPLY_HEADER}, challenge...)
			sent++
		}

		s.conn.WriteTo(reply, addr)
	}
}

func (s *rechallengeServer) seen() [][]byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([][]byte{}, s.requests...)
}

func (s *rechallengeServer) client(t *testing.T) *Client {
	t.Helper()

	c, err := NewClient(s.conn.LocalAddr().String(), TimeoutOption(time.Second))
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { c.Close() })

	return c
}

func challengeBytes(n uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, n)
	return b
}

func infoResponse() []byte {
	var b PacketBuilder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, A2S_INFO_RESPONSE, 17})
	b.WriteCString("Test server")
	b.WriteCString("de_dust2")
	b.WriteCString("cstrike")
	b.WriteCString("Counter-Strike: Source")
	b.WriteBytes([]byte{240, 0, 5, 32, 0, 'd', 'l', 0, 1})
	b.WriteCString("1.0.0.0")
	return b.Bytes()
}

func rulesResponse() []byte {
	var b PacketBuilder
	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, A2S_RULES_RESPONSE, 1, 0})
	b.WriteCString("mp_timelimit")
	b.WriteCString("30")
	return b.Bytes()
}

func TestQueryAnswersNewChallenges(t *testing.T) {
	for _, extra := range []int{0, 1, MaxChallengeRetries} {
		s := newRechallengeServer(t, infoResponse(), extra)
		info, err := s.client(t).QueryInfo()
		assert.Nil(t, err, "QueryInfo with %d new challenges", extra)
		if assert.NotNil(t, info) {
			assert.Equal(t, "de_dust2", info.Map)
			assert.Equal(t, uint8(5), info.Players)
		}
		// The first request carries no challenge; every later one the newest challenge.
		if seen := s.seen(); assert.Len(t, seen, extra+2) {
			for i, got := range seen[1:] {
				assert.Equal(t, challengeBytes(uint32(1001+i)), got, "request %d", i+1)
			}
		}
	}
}

func TestPlayerAndRulesAnswerNewChallenges(t *testing.T) {
	s := newRechallengeServer(t, validPlayerInfoPacket(), 2)
	players, err := s.client(t).QueryPlayer()
	assert.Nil(t, err)
	if assert.NotNil(t, players) {
		assert.Equal(t, uint8(2), players.Count)
	}

	s = newRechallengeServer(t, rulesResponse(), 2)
	rules, err := s.client(t).QueryRules()
	assert.Nil(t, err)
	if assert.NotNil(t, rules) {
		assert.Equal(t, "30", rules.Rules["mp_timelimit"])
	}
}

func TestQueryGivesUpAfterTooManyChallenges(t *testing.T) {
	s := newRechallengeServer(t, infoResponse(), MaxChallengeRetries+1)
	_, err := s.client(t).QueryInfo()
	assert.Equal(t, ErrTooManyChallenges, err)

	s = newRechallengeServer(t, validPlayerInfoPacket(), MaxChallengeRetries+1)
	_, err = s.client(t).QueryPlayer()
	assert.Equal(t, ErrTooManyChallenges, err)
}

func TestChallengeOf(t *testing.T) {
	got, ok := challengeOf([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x41, 0x0b, 0xa8, 0x6e, 0x38})
	assert.True(t, ok)
	assert.Equal(t, []byte{0x0b, 0xa8, 0x6e, 0x38}, got)

	_, ok = challengeOf(infoResponse())
	assert.False(t, ok)

	_, ok = challengeOf([]byte{0xFF, 0xFF, 0xFF, 0xFF, 0x41, 0x0b})
	assert.False(t, ok, "too short to hold a challenge")
}
