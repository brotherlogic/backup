package main

import (
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
