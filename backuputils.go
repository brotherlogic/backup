package main

import (
	"fmt"
	"os"
)

func (s *Server) processFile(path string, info os.FileInfo, err error) error {
	s.Log(fmt.Sprintf("Processing %v (%v)", path, info))
	return nil
}
