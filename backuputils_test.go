package main

import (
	"testing"

	"golang.org/x/net/context"

	pbcdp "github.com/brotherlogic/cdprocessor/proto"
)

type testCdprocessor struct{}

func (p *testCdprocessor) getRipped(ctx context.Context, req *pbcdp.GetRippedRequest) (*pbcdp.GetRippedResponse, error) {
	return &pbcdp.GetRippedResponse{
		Ripped: []*pbcdp.Rip{
			&pbcdp.Rip{
				Tracks: []*pbcdp.Track{
					&pbcdp.Track{
						FlacPath: "/yes/",
					},
				},
			},
		},
	}, nil
}

func InitTestServer() *Server {
	s := Init()
	s.cdprocessor = &testCdprocessor{}
	return s
}

func TestCountFlacs(t *testing.T) {
	s := InitTestServer()
	s.processFlacs(context.Background())

	if s.flacsToBackup != 1 {
		t.Errorf("Wrong number of backups: %v", s.flacsToBackup)
	}
}
