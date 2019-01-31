package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	pb "github.com/brotherlogic/backup/proto"
)

func (s *Server) mapConfig(mapping *pb.BackupSpec) ([]string, error) {
	files := []string{}

	err := filepath.Walk(mapping.BaseDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			match, _ := regexp.MatchString(mapping.MatchRegex, path)
			s.Log(fmt.Sprintf("Match %v -> %v", path, match))
			if match {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
