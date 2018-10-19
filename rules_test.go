package a2s

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRules(t *testing.T) {
	c, err := NewClient(TestHost)

	defer c.Close()

	if err != nil {
		t.Error(err)
		return
	}

	r, err := c.QueryRules()

	if err != nil {
		t.Error(err)
		return
	}

	JSON, _ := json.Marshal(r)

	fmt.Println(string(JSON))
}
