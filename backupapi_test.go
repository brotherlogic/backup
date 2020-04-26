package main

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/backup/proto"
)

func TestSaveToken(t *testing.T) {
	s := InitTestServer()
	_, err := s.SaveToken(context.Background(), &pb.TokenRequest{Token: &pb.Token{JsonToken: []byte("hello")}})
	if err != nil {
		t.Errorf("Bad token save: %v", err)
	}
}

func TestBuildState(t *testing.T) {
	s := InitTestServer()
	s.config.Files = append(s.config.Files, &pb.BackupFile{})
	s.config.Files = append(s.config.Files, &pb.BackupFile{})

	stats, err := s.GetStats(context.Background(), &pb.StatsRequest{})

	if err != nil {
		t.Errorf("Bad stats")
	}

	if len(stats.GetStats()) != 1 || stats.GetStats()[0].Count != 2 {
		t.Errorf("Bad stats: %v", stats)
	}
}
