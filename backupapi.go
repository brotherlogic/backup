package main

import (
	"fmt"

	pb "github.com/brotherlogic/backup/proto"
	epb "github.com/brotherlogic/executor/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
)

var (
	serverCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backup_servercount",
		Help: "Push Size",
	})
)

func (s *Server) RunBackup(ctx context.Context, _ *pb.RunBackupRequest) (*pb.RunBackupResponse, error) {
	servers, err := s.getAllServers(ctx)
	if err != nil {
		return nil, err
	}

	serverCount.Set(float64(len(servers)))

	for _, server := range servers {
		conn, err := s.FDialSpecificServer(ctx, "executor", s.Registry.Identifier)
		if err != nil {
			return nil, err
		}
		defer conn.Close()

		client := epb.NewExecutorServiceClient(conn)
		_, err = client.QueueExecute(ctx, &epb.ExecuteRequest{Command: &epb.Command{
			Binary: "rsync",
			Parameters: []string{
				"-avz",
				"--progress",
				fmt.Sprintf("%v:/media/scratch/dlogs/", server),
				"/media/raid/dlog-backup/"}}})
		if err != nil {
			return nil, err
		}

	}

	return &pb.RunBackupResponse{}, nil
}

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
			stats.Stats = append(stats.Stats, &pb.Stat{State: f.GetState(), Count: 1, Example: f})
		}
	}

	return stats, nil
}
