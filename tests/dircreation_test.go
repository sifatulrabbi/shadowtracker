package tests

import (
	"fmt"
	"os"
	"testing"
)

func TestRootDirAccess(t *testing.T) {
	entries, err := os.ReadDir("/var/log")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println(entries)
}
