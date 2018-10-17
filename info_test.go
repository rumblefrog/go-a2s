package a2s

import (
	"fmt"
	"testing"
)

func TestInfo(t *testing.T) {
	c, err := NewClient("74.91.112.77")

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
}
