package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"google.golang.org/protobuf/proto"

	pb "github.com/brotherlogic/backup/proto"
	epb "github.com/brotherlogic/executor/proto"
	qpb "github.com/brotherlogic/queue/proto"
	google_protobuf "github.com/golang/protobuf/ptypes/any"
)

var (
	serverCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backup_servercount",
		Help: "Push Size",
	})
	rsyncs = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "backup_rsyns",
		Help: "Push Size",
	}, []string{"server"})
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
		_, err = client.QueueExecute(ctx, &epb.ExecuteRequest{
			Command: &epb.Command{
				Binary:           "rsync",
				DeleteOnComplete: true,
				Parameters: []string{
					"-avz",
					"--progress",
					fmt.Sprintf("%v:/media/scratch/dlogs/", server),
					"/media/raid/dlog-backup/"}}})
		if err != nil {
			return nil, err
		}
		rsyncs.With(prometheus.Labels{"server": server}).Inc()

	}

	conn2, err2 := s.FDialServer(ctx, "queue")
	if err2 != nil {
		return nil, err2
	}
	defer conn2.Close()
	qclient := qpb.NewQueueServiceClient(conn2)
	upup := &pb.RunBackupRequest{}
	data, _ := proto.Marshal(upup)
	_, err3 := qclient.AddQueueItem(ctx, &qpb.AddQueueItemRequest{
		QueueName: "run_backup",
		RunTime:   time.Now().Add(time.Hour).Unix(),
		Payload:   &google_protobuf.Any{Value: data},
		Key:       fmt.Sprintf("backup-%v", time.Now().Add(time.Hour).Unix()),
	})

	return &pb.RunBackupResponse{}, err3
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
