package a2s

import (
	"encoding/json"
	"flag"
	"fmt"
	"testing"
)

var testHost = flag.String("test-host", "s1.zhenyangli.me", "Remote hostname to use for unit tests.")

func TestInfo(t *testing.T) {
	c, err := NewClient(*testHost)
	if err != nil {
		t.Error(err)
		return
	}
	defer c.Close()

	i, err := c.QueryInfo()

	if err != nil {
		t.Error(err)
		return
	}

	JSON, _ := json.Marshal(i)

	fmt.Println(string(JSON))
}
