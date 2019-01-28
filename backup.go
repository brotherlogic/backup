package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"time"

	"github.com/brotherlogic/goserver"
	"github.com/brotherlogic/goserver/utils"
	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/backup/proto"
	pbcdp "github.com/brotherlogic/cdprocessor/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

type cdprocessor interface {
	getRipped(ctx context.Context, req *pbcdp.GetRippedRequest) (*pbcdp.GetRippedResponse, error)
}

type prodCdprocessor struct{}

func (p *prodCdprocessor) getRipped(ctx context.Context, req *pbcdp.GetRippedRequest) (*pbcdp.GetRippedResponse, error) {
	ip, port, err := utils.Resolve("cdprocessor")
	if err != nil {
		return &pbcdp.GetRippedResponse{}, err
	}

	conn, err := grpc.Dial(ip+":"+strconv.Itoa(int(port)), grpc.WithInsecure())
	if err != nil {
		return &pbcdp.GetRippedResponse{}, err
	}
	defer conn.Close()

	client := pbcdp.NewCDProcessorClient(conn)
	return client.GetRipped(ctx, req)
}

//Server main server type
type Server struct {
	*goserver.GoServer
	flacsToBackup int64
	cdprocessor   cdprocessor
	config        *pb.Config
}

// Init builds the server
func Init() *Server {
	s := &Server{
		&goserver.GoServer{},
		int64(0),
		&prodCdprocessor{},
		&pb.Config{},
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

// Mote promotes/demotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	return nil
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	return []*pbg.State{
		&pbg.State{Key: "flacs_to_backup", Value: s.flacsToBackup},
		&pbg.State{Key: "specs", Value: int64(len(s.config.Specs))},
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

	server.GoServer.KSclient = *keystoreclient.GetClient(server.GetIP)

	err := server.RegisterServer("backup", false)
	if err != nil {
		log.Fatalf("Unable to register: %v", err)
	}

	server.RegisterRepeatingTask(server.processFlacs, "process_flacs", time.Minute*5)

	fmt.Printf("%v", server.Serve())
}
