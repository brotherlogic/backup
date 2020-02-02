package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"

	"cloud.google.com/go/storage"
	"github.com/brotherlogic/goserver"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/backup/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

const (
	//TOKEN - the json auth token
	TOKEN = "/github.com/brotherlogic/backup/jsontoken"
)

//Server main server type
type Server struct {
	*goserver.GoServer
	config *pb.Config
	token  *pb.Token
	seen   map[string]bool
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer: &goserver.GoServer{},
		config:   &pb.Config{},
	}
	return s
}

func (s *Server) loadToken(ctx context.Context) error {
	token := &pb.Token{}
	data, _, err := s.KSclient.Read(ctx, TOKEN, token)

	if err != nil {
		return err
	}

	token, ok := data.(*pb.Token)
	if !ok {
		return fmt.Errorf("Unable to unwrap token: %v", err)
	}
	s.token = token

	return nil
}

// DoRegister does RPC registration
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterBackupServiceServer(server, s)
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
	return []*pbg.State{
		&pbg.State{Key: "last_sync", TimeValue: s.config.GetLastBackup()},
	}
}

func (s *Server) gcWalk(ctx context.Context) (time.Time, error) {
	err := s.loadToken(ctx)
	if err != nil {
		return time.Now().Add(time.Minute * 5), err
	}

	creds, err := google.CredentialsFromJSON(ctx, s.token.JsonToken, storage.ScopeReadOnly)
	if err != nil {
		return time.Now().Add(time.Minute * 5), err
	}

	client, err := storage.NewClient(ctx, option.WithCredentials(creds))
	if err != nil {
		return time.Now().Add(time.Minute * 5), err
	}

	bkt := client.Bucket("brotherlogic-archive")

	query := &storage.Query{Prefix: ""}
	it := bkt.Objects(ctx, query)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return time.Now().Add(time.Minute * 5), err
		}

		s.processCloudFile(attrs.Name)
	}

	return time.Now().Add(time.Minute * 5), nil
}

func (s *Server) fsWalk(ctx context.Context) (time.Time, error) {
	s.seen = make(map[string]bool)
	t, err := time.Now().Add(time.Minute*5), filepath.Walk("/media/raid1/", s.processFile)

	// Set other files missing
	for _, f := range s.config.Files {
		if !s.seen[f.GetPath()] {
			f.State = pb.BackupFile_MISSING
		}
	}

	s.Log(fmt.Sprintf("Now there's %v files (but %v)", len(s.config.GetFiles()), err))

	return t, err
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

	err := server.RegisterServerV2("backup", false, false)
	if err != nil {
		return
	}

	server.RegisterLockingTask(server.fsWalk, "fs_walk")
	server.RegisterLockingTask(server.gcWalk, "gc_walk")

	fmt.Printf("%v", server.Serve())
}
