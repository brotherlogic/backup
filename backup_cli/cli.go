package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/brotherlogic/goserver/utils"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/backup/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	conn, err := grpc.Dial("discovery:///backup", grpc.WithInsecure(), grpc.WithBalancerName("my_pick_first"))
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewBackupServiceClient(conn)
	ctx, cancel := utils.BuildContext("backup-cli", "backup")
	defer cancel()

	switch os.Args[1] {
	case "token":
		tokenFlags := flag.NewFlagSet("TokenFlags", flag.ExitOnError)
		var token = tokenFlags.String("token", "", "filename to load")
		if err := tokenFlags.Parse(os.Args[2:]); err == nil {
			data, err := ioutil.ReadFile(*token)
			if err != nil {
				log.Fatalf("Token read error: %v", err)
			}
			token := &pb.Token{JsonToken: data}
			_, err = client.SaveToken(ctx, &pb.TokenRequest{Token: token})
			if err != nil {
				log.Fatalf("Save error: %v", err)
			}
		}
	}

}
