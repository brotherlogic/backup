package main

import (
	"context"
	"testing"

	pb "github.com/brotherlogic/backup/proto"
)

func TestSaveToken(t *testing.T) {
	s := InitTestServer()
	_, err := s.SaveToken(context.Background(), &pb.TokenRequest{})
	if err != nil {
		t.Errorf("Bad token save: %v", err)
	}
}
