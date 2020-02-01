package main

import (
	"fmt"
	"os"
	"testing"
)

func InitTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.seen = make(map[string]bool)
	return s
}

func TestSpecRead(t *testing.T) {
	s := InitTestServer()
	fi, err := os.Stat("backuputils_test.go")
	err = s.processFile("path", fi, fmt.Errorf("What"))
	if err != nil {
		t.Errorf("Bad pull: %v", err)
	}

	err = s.processFile("path", fi, fmt.Errorf("What"))
	if err != nil {
		t.Errorf("Bad pull: %v", err)
	}

	if len(s.config.GetFiles()) != 1 {
		t.Errorf("Too many files added")
	}

}
