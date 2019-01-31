package main

import (
	"os"
	"path/filepath"
	"regexp"

	pb "github.com/brotherlogic/backup/proto"
)

func match(reg, path string) bool {
	match, _ := regexp.MatchString(reg, path)
	return match
}

func (s *Server) mapConfig(mapping *pb.BackupSpec) ([]string, error) {
	files := []string{}

	err := filepath.Walk(mapping.BaseDirectory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			if match(mapping.MatchRegex, path) {
				files = append(files, path)
			}
		}

		return nil
	})

	return files, err
}
