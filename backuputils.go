package main

import (
	"fmt"
	"os"
	"time"

	pb "github.com/brotherlogic/backup/proto"
	"github.com/brotherlogic/goserver/utils"
	"golang.org/x/net/context"
)

func (s *Server) processFile(cpath string, info os.FileInfo, err error) error {
	ctx, cancel := utils.ManualContext("process-file", "process-file", time.Minute)
	defer cancel()
	path := s.hashPath(ctx, cpath)
	s.seen[path] = true

	found := false
	for _, file := range s.config.GetFiles() {
		if file.GetPath() == path {
			found = true
		}
	}

	if !found {
		s.config.Files = append(s.config.Files, &pb.BackupFile{Path: path, DateSeen: time.Now().Unix(), State: pb.BackupFile_NOT_BACKED_UP})
	}

	return nil
}

func (s *Server) processCloudFile(ctx context.Context, cpath string) error {
	path := s.hashPath(ctx, "/media/raid1/"+cpath)
	for _, file := range s.config.GetFiles() {
		if file.GetPath() == path {
			file.State = pb.BackupFile_BACKED_UP
			return nil
		}
	}

	return nil
}

func (s *Server) alertOnMismatch(ctx context.Context) (time.Time, error) {
	stats, _ := s.GetStats(ctx, &pb.StatsRequest{})
	s.alertOnBadStats(ctx, stats.GetStats())
	return time.Now().Add(time.Hour * 24), nil
}

func (s *Server) alertOnBadStats(ctx context.Context, stats []*pb.Stat) {
	for _, stat := range stats {
		if stat.GetState() == pb.BackupFile_NOT_BACKED_UP {
			s.RaiseIssue(ctx, "Backup Issue", fmt.Sprintf("Some files are noot backed up"), false)
		}
	}
}
