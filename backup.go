package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"github.com/brotherlogic/goserver"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/backup/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	config *pb.Config
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		&pb.Config{
			Specs: []*pb.BackupSpec{
				&pb.BackupSpec{
					Server:        "stack1",
					BaseDirectory: "/media/music/",
					MatchRegex:    ".*flac$",
				},
			},
		},
	}
	return s
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {

}

// ReportHealth alerts if we're not healthy
func (s *Server) ReportHealth() bool {
	return true
}

//Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	files, err := s.mapConfig(s.config.Specs[0])
	example := fmt.Sprintf("%v", err)
	if len(files) > 0 {
		example = files[0]
	}
	return []*pbg.State{
		&pbg.State{Key: "mapping", Value: int64(len(files))},
		&pbg.State{Key: "specs", Value: int64(len(s.config.Specs))},
		&pbg.State{Key: "example_file", Text: example},
	}
}

func main() {
	var quiet = flag.Bool("quiet", false, "Show all output")
	flag.Parse()

	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	server := Init()
	server.PrepServer()
	server.Register = server

	err := server.RegisterServer("backup", false)
	if err != nil {
		log.Fatalf("Unable to register: %v", err)
	}

	fmt.Printf("%v", server.Serve())
}
