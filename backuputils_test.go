package main

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/brotherlogic/keystore/client"
	"github.com/golang/protobuf/proto"
	"golang.org/x/net/context"

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
	pNum, fNum := s.intHashPath(context.Background(), "/media/raid1/madeup")
	s.config.Files = append(s.config.Files, &pb.BackupFile{FilenameHash: fNum, DirectoryHash: pNum})

	err := s.processCloudFile(context.Background(), "madeup")
	if err != nil {
		t.Errorf("bad proc: %v", err)
	}

	log.Printf("CONFIG %v", s.config)
	log.Printf("Size %v", proto.Size(s.config.Files[0]))
}

func TestNoMatch(t *testing.T) {
	s := InitTestServer()
	//s.config.Files = append(s.config.Files, &pb.BackupFile{Path: "madeup"})

	err := s.processCloudFile(context.Background(), "madeup2")
	if err != nil {
		t.Errorf("bad proc: %v", err)
	}
}

func TestAlertOnMismatch(t *testing.T) {
	s := InitTestServer()
	val, err := s.alertOnMismatch(context.Background())
	if err != nil {
		t.Errorf("Alerting caused an error: %v -> %v", err, val)
	}
}

func TestAlertOnSstats(t *testing.T) {
	s := InitTestServer()
	s.alertOnBadStats(context.Background(), []*pb.Stat{&pb.Stat{State: pb.BackupFile_NOT_BACKED_UP}})
}

func TestSize(t *testing.T) {
	b := &pb.BackupFile{DateSeen: 1588041103, State: pb.BackupFile_NOT_BACKED_UP, DirectoryHash: 438442931, FilenameHash: -896055949}
	log.Printf("SIZE = %v", proto.Size(b))
}
