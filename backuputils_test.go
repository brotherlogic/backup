package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/brotherlogic/keystore/client"

	pb "github.com/brotherlogic/backup/proto"
)

func InitTestServer() *Server {
	s := Init()
	s.SkipLog = true
	s.seen = make(map[string]bool)
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".test")
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

func TestMatch(t *testing.T) {
	s := InitTestServer()
	s.config.Files = append(s.config.Files, &pb.BackupFile{Path: "madeup"})

	err := s.processCloudFile("madeup")
	if err != nil {
		t.Errorf("bad proc: %v", err)
	}
}

func TestNoMatch(t *testing.T) {
	s := InitTestServer()
	s.config.Files = append(s.config.Files, &pb.BackupFile{Path: "madeup"})

	err := s.processCloudFile("madeup2")
	if err != nil {
		t.Errorf("bad proc: %v", err)
	}
}
