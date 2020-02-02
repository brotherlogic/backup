package main

import (
	"fmt"
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

func (s *Server) processCloudFile(path string) error {
	for _, file := range s.config.GetFiles() {
		if file.GetPath() == "/media/raid1/"+path {
			s.Log(fmt.Sprintf("Found match %v and %v", file, path))
			return nil
		}
	}

	s.Log(fmt.Sprintf("No match found for %v", path))
	return nil
}
