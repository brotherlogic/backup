package main

import (
	pb "github.com/brotherlogic/backup/proto"
	"golang.org/x/net/context"
)

//SaveToken saves out the token
func (s *Server) SaveToken(ctx context.Context, req *pb.TokenRequest) (*pb.TokenResponse, error) {
	return &pb.TokenResponse{}, s.KSclient.Save(ctx, TOKEN, req.GetToken())
}
