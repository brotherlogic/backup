package main

import (
	"os"
	"time"

	pb "github.com/brotherlogic/backup/proto"
)

func (s *Server) processFile(path string, info os.FileInfo, err error) error {
	s.seen[path] = true

	found := false
	for _, file := range s.config.GetFiles() {
		if file.GetPath() == path {
			found = true
		}
	}

	if !found {
		s.config.Files = append(s.config.Files, &pb.BackupFile{Path: path, DateSeen: time.Now().Unix()})
	}

	return nil
}
