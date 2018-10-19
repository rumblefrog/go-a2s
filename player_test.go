package a2s

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestPlayer(t *testing.T) {
	c, err := NewClient(TestHost)

	defer c.Close()

	if err != nil {
		t.Error(err)
		return
	}

	p, err := c.QueryPlayer()

	if err != nil {
		t.Error(err)
		return
	}

	JSON, _ := json.Marshal(p)

	fmt.Println(string(JSON))
}
