package main

import (
	"flag"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/storage"
	"github.com/brotherlogic/goserver"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/brotherlogic/backup/proto"
	pbg "github.com/brotherlogic/goserver/proto"
)

const (
	//TOKEN - the json auth token
	TOKEN = "/github.com/brotherlogic/backup/jsontoken"

	//CONFIG - the overall state
	CONFIG = "/github.com/brotherlogic/backup/config"

	//WAITTIME - time between runs
	WAITTIME = time.Minute * 5
)

var (
	backedup = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backup_backedup",
		Help: "Total number of files backed up",
	})
	notbackedup = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "backup_notbackedup",
		Help: "Total number of files backed up",
	})
)

//Server main server type
type Server struct {
	*goserver.GoServer
	config    *pb.Config
	token     *pb.Token
	seen      map[string]bool
	hashMutex *sync.Mutex
	hashMap   map[int32]string
	resolver  map[int32]map[int32]string
}

func (s *Server) intHashPath(ctx context.Context, path string) (int32, int32) {
	slashIndex := strings.LastIndex(path, "/")
	if slashIndex == -1 {
		slashIndex = 0
	}

	h := fnv.New32a()
	h.Write([]byte(path[:slashIndex]))
	pSum := int32(h.Sum32())

	h2 := fnv.New32a()
	h2.Write([]byte(path[slashIndex+1:]))
	fSum := int32(h2.Sum32())

	s.hashMutex.Lock()
	defer s.hashMutex.Unlock()

	if val, ok := s.hashMap[pSum]; !ok {
		s.hashMap[pSum] = path[:slashIndex]
	} else {
		if path[:slashIndex] != val {
			s.RaiseIssue(ctx, "Hash Clash", fmt.Sprintf("%v and %v have clashed", val, path[:slashIndex]), false)
		}
	}

	if val, ok := s.hashMap[fSum]; !ok {
		s.hashMap[fSum] = path[slashIndex+1:]
	} else {
		if path[slashIndex+1:] != val {
			s.RaiseIssue(ctx, "Hash Clash", fmt.Sprintf("%v and %v have clashed", val, path[slashIndex+1:]), false)
		}
	}

	return pSum, fSum
}

func (s *Server) hashPath(ctx context.Context, path string) string {
	h := fnv.New64a()
	h.Write([]byte(path))
	bs := h.Sum(nil)
	hash := string(bs)

	return hash
}

// Init builds the server
func Init() *Server {
	s := &Server{
		GoServer:  &goserver.GoServer{},
		config:    &pb.Config{},
		hashMap:   make(map[int32]string),
		hashMutex: &sync.Mutex{},
		resolver:  make(map[int32]map[int32]string),
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

func (s *Server) loadConfig(ctx context.Context) error {
	config := &pb.Config{}
	data, _, err := s.KSclient.Read(ctx, CONFIG, config)

	if err != nil {
		return err
	}

	config, ok := data.(*pb.Config)
	if !ok {
		return fmt.Errorf("Unable to unwrap config: %v", err)
	}
	s.config = config

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
	err := s.loadConfig(ctx)
	if status.Convert(err).Code() == codes.NotFound {
		s.config.LastBackup = 1
		err = s.KSclient.Save(ctx, CONFIG, s.config)
	}
	return err
}

// GetState gets the state of the server
func (s *Server) GetState() []*pbg.State {
	rep := ""
	siz := int64(0)
	if len(s.config.GetFiles()) > 0 {
		rep = fmt.Sprintf("%v", s.config.GetFiles()[0])
		siz = int64(proto.Size(s.config.GetFiles()[0]))
	}
	largest := ""
	biggest := 0
	for _, f := range s.config.GetFiles() {
		if proto.Size(f) > biggest {
			biggest = proto.Size(f)
			largest = fmt.Sprintf("%v", f)
		}
	}

	return []*pbg.State{
		&pbg.State{Key: "config_size", Value: int64(proto.Size(s.config))},
		&pbg.State{Key: "files", Value: int64(len(s.config.GetFiles()))},
		&pbg.State{Key: "sample", Text: rep},
		&pbg.State{Key: "sample_size", Value: siz},
		&pbg.State{Key: "largest", Text: largest},
		&pbg.State{Key: "last_sync", TimeValue: s.config.GetLastBackup()},
	}
}

func (s *Server) gcWalk(ctx context.Context) (time.Time, error) {
	err := s.loadToken(ctx)
	if err != nil {
		return time.Now().Add(time.Minute * 5), err
	}
	err = s.loadConfig(ctx)
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
	count := 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return time.Now().Add(time.Minute * 5), err
		}

		s.processCloudFile(ctx, attrs.Name)
		count++
	}

	s.Log(fmt.Sprintf("Processed %v cloud files (%v)", count, err))

	return time.Now().Add(WAITTIME), s.KSclient.Save(ctx, CONFIG, s.config)
}

func (s *Server) fsWalk(ctx context.Context) (time.Time, error) {
	err := s.loadConfig(ctx)
	s.hashMutex.Lock()
	s.hashMap = make(map[int32]string)
	s.hashMutex.Unlock()
	if err != nil {
		return time.Now().Add(time.Minute * 5), err
	}
	t, err := time.Now().Add(WAITTIME), filepath.Walk("/media/raid1/", s.processFile)

	if err == nil {
		err = s.KSclient.Save(ctx, CONFIG, s.config)
	}

	s.Log(fmt.Sprintf("Now there's %v files (but %v) -> %v => %v", len(s.config.GetFiles()), err, proto.Size(s.config), s.config.GetFiles()[0]))

	return t, err
}

func (s *Server) monitor(ctx context.Context) error {
	bucount := 0
	nbucount := 0
	stats, _ := s.GetStats(ctx, &pb.StatsRequest{})
	s.Log(fmt.Sprintf("Processed %v files", len(stats.GetStats())))
	for _, stat := range stats.GetStats() {
		if stat.GetState() == pb.BackupFile_NOT_BACKED_UP {
			nbucount += int(stat.GetCount())
		}
		if stat.GetState() == pb.BackupFile_BACKED_UP {
			bucount += int(stat.GetCount())
		}
	}

	backedup.Set(float64(bucount))
	notbackedup.Set(float64(nbucount))
	return nil
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
	server.RegisterLockingTask(server.alertOnMismatch, "alert_on_mismatch")
	server.RegisterRepeatingTaskNonMaster(server.monitor, "monitor", time.Minute)

	fmt.Printf("%v", server.Serve())
}
