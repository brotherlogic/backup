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
	ctx, cancel := utils.ManualContext("process-file", "process-file", time.Minute, false)
	defer cancel()
	pNum, fNum := s.intHashPath(ctx, cpath)
	//s.seen[path] = true

	found := false
	for _, file := range s.config.GetFiles() {
		b1, ok := s.resolver[file.GetDirectoryHash()]
		if !ok {
			b1 = make(map[int32]string)
		}
		b1[file.GetFilenameHash()] = cpath
		if file.GetDirectoryHash() == pNum && file.GetFilenameHash() == fNum {
			found = true
		}
	}

	if !found {
		s.config.Files = append(s.config.Files, &pb.BackupFile{DirectoryHash: pNum, FilenameHash: fNum, DateSeen: int32(time.Now().Unix()), State: pb.BackupFile_NOT_BACKED_UP})
	}

	return nil
}

func (s *Server) processCloudFile(ctx context.Context, cpath string) error {
	pNum, fNum := s.intHashPath(ctx, "/media/raid1/"+cpath)
	for _, file := range s.config.GetFiles() {
		if file.GetDirectoryHash() == pNum && file.GetFilenameHash() == fNum {
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
			file := "UNKNOWN"
			f, ok := s.resolver[stat.GetExample().GetDirectoryHash()]
			if ok {
				file = f[stat.GetExample().GetFilenameHash()]
			}
			s.RaiseIssue(ctx, "Backup Issue", fmt.Sprintf("Some files are noot backed up; for example %v", file), false)
		}
	}
}
