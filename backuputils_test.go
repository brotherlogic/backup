package main

import (
	"os"
	"testing"

	pb "github.com/brotherlogic/backup/proto"
)

func InitTestServer() *Server {
	s := Init()
	s.SkipLog = true
	return s
}

func TestSpecRead(t *testing.T) {
	s := InitTestServer()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working dir")
	}

	files, err := s.mapConfig(&pb.BackupSpec{BaseDirectory: pwd + "/testDir/", MatchRegex: ".*blah$"})
	if err != nil {
		t.Fatalf("Error mapping config: %v", err)
	}

	if len(files) != 1 || files[0] != "testfile.blah" {
		t.Errorf("Error running mapper: %v", files)
	}
}

func TestSpecReadWithMadeupDirectory(t *testing.T) {
	s := InitTestServer()

	pwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get working dir")
	}

	files, err := s.mapConfig(&pb.BackupSpec{BaseDirectory: pwd + "/testDirMadeUp/", MatchRegex: ".*blah$"})
	if err == nil {
		t.Errorf("Bad spec did not fail: %v", files)
	}
}
