package main

import (
	"golang.org/x/net/context"

	pbcdp "github.com/brotherlogic/cdprocessor/proto"
)

func (s *Server) processFlacs(ctx context.Context) {
	ripped, err := s.cdprocessor.getRipped(ctx, &pbcdp.GetRippedRequest{})
	flacs := 0
	if err == nil {
		for _, r := range ripped.Ripped {
			for _, t := range r.Tracks {
				if len(t.FlacPath) > 0 {
					flacs++
				}
			}
		}
	}
	s.flacsToBackup = int64(flacs)
}
