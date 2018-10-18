package a2s

import (
	"fmt"
	"testing"
)

func TestInfo(t *testing.T) {
	c, err := NewClient("74.91.116.5:27015")

	if err != nil {
		fmt.Println(err)
		return
	}

	i, err := c.QueryInfo()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v", i)

	s, err := c.QueryPlayer()

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%v", s.Players)
}
