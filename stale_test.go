package a2s

import (
	"encoding/binary"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// fakeServer answers like srcds: a new challenge per request, and only the latest is accepted
type fakeServer struct {
	conn      *net.UDPConn
	mu        sync.Mutex
	challenge uint32
	delay     time.Duration
	before    []byte
}

func newFakeServer(t *testing.T) *fakeServer {
	conn, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1)})
	assert.Nil(t, err, "ListenUDP should not return an error")
	t.Cleanup(func() { conn.Close() })

	s := &fakeServer{conn: conn}

	go s.serve()

	return s
}

func (s *fakeServer) serve() {
	buf := make([]byte, DefaultMaxPacketSize)

	for {
		n, from, err := s.conn.ReadFromUDP(buf)

		if err != nil {
			return
		}

		s.mu.Lock()
		reply := s.reply(buf[:n])
		delay, before := s.delay, s.before
		s.delay, s.before = 0, nil
		s.mu.Unlock()

		go func() {
			time.Sleep(delay)

			if before != nil {
				s.conn.WriteToUDP(before, from)
			}

			s.conn.WriteToUDP(reply, from)
		}()
	}
}

func (s *fakeServer) reply(req []byte) []byte {
	full := fakeInfoPacket("fake")
	challenge := req[len(req)-4:]

	if req[4] == A2S_PLAYER_REQUEST {
		full = validPlayerInfoPacket()
	}

	if binary.LittleEndian.Uint32(challenge) == s.challenge {
		return full
	}

	s.challenge++

	reply := []byte{0xFF, 0xFF, 0xFF, 0xFF, A2S_PLAYER_CHALLENGE_REPLY_HEADER, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(reply[5:], s.challenge)

	return reply
}

func fakeInfoPacket(name string) []byte {
	var b PacketBuilder

	b.WriteBytes([]byte{0xFF, 0xFF, 0xFF, 0xFF, A2S_INFO_RESPONSE, 17})
	b.WriteCString(name)
	b.WriteCString("map")
	b.WriteCString("folder")
	b.WriteCString("game")
	b.WriteBytes([]byte{0, 0, 1, 10, 0, 'd', 'l', 0, 1})
	b.WriteCString("1.0")

	return b.Bytes()
}

func TestQueryInfoAfterLateReply(t *testing.T) {
	s := newFakeServer(t)

	c, err := NewClient(s.conn.LocalAddr().String(), TimeoutOption(100*time.Millisecond))
	assert.Nil(t, err, "NewClient should not return an error")
	defer c.Close()

	s.mu.Lock()
	s.delay = 200 * time.Millisecond
	s.mu.Unlock()

	_, err = c.QueryInfo()
	assert.NotNil(t, err, "QueryInfo should time out while the server is stalled")

	// wait for the late reply to arrive
	time.Sleep(300 * time.Millisecond)

	for i := 0; i < 3; i++ {
		info, err := c.QueryInfo()

		if !assert.Nil(t, err, "QueryInfo should succeed after a late reply") {
			return
		}

		assert.Equal(t, "fake", info.Name, "Server name should match")
	}
}

func TestQueryPlayerSkipsWrongReplyType(t *testing.T) {
	s := newFakeServer(t)

	c, err := NewClient(s.conn.LocalAddr().String(), TimeoutOption(500*time.Millisecond))
	assert.Nil(t, err, "NewClient should not return an error")
	defer c.Close()

	s.mu.Lock()
	s.before = fakeInfoPacket("stale")
	s.mu.Unlock()

	p, err := c.QueryPlayer()

	if !assert.Nil(t, err, "QueryPlayer should skip an unexpected A2S_INFO reply") {
		return
	}

	assert.Equal(t, uint8(2), p.Count, "Player count should match")
}

func splitPacket(id uint32, total, number uint8, payload []byte) []byte {
	b := []byte{0xFE, 0xFF, 0xFF, 0xFF, 0, 0, 0, 0, total, number, 0xE0, 0x04}
	binary.LittleEndian.PutUint32(b[4:], id)

	return append(b, payload...)
}

func TestCollectMultiplePacketResponseSkipsStaleID(t *testing.T) {
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	c := &Client{conn: client, maxPacketSize: DefaultMaxPacketSize, buffer: make([]byte, 0, DefaultMaxPacketSize)}

	full := validPlayerInfoPacket()

	go func() {
		server.Write(splitPacket(8, 2, 0, full[:20]))
		server.Write(splitPacket(7, 2, 1, []byte("stale")))
		server.Write(splitPacket(8, 2, 1, full[20:]))
	}()

	payload, err := c.collectMultiplePacketResponse(splitPacket(7, 2, 0, []byte("stale")))
	assert.Nil(t, err, "collectMultiplePacketResponse should skip fragments of a stale response")
	assert.Equal(t, full, payload, "Payload should match the current response")
}
