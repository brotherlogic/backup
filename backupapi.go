package main

import (
	pb "github.com/brotherlogic/backup/proto"
	"golang.org/x/net/context"
)

//SaveToken saves out the token
func (s *Server) SaveToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	return &pb.TokenResponse{}, s.KSclient.Save(ctx, TOKEN, req.GetToken())
}

// GetStats gets the relevant stats for the system
func (s *Server) GetStats(ctx context.Context, req *pb.StatsRequest) (*pb.StatsResponse, error) {
	stats := &pb.StatsResponse{}
	for _, f := range s.config.Files {
		found := false
		for _, stat := range stats.GetStats() {
			if stat.GetState() == f.GetState() {
				found = true
				stat.Count++
			}
		}

		if !found {
			stats.Stats = append(stats.Stats, &pb.Stat{State: f.GetState(), Count: 1})
		}
	}

	return stats, nil
}
