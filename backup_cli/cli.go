package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/brotherlogic/goserver/utils"

	pb "github.com/brotherlogic/backup/proto"

	//Needed to pull in gzip encoding init
	_ "google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/resolver"
)

func init() {
	resolver.Register(&utils.DiscoveryClientResolverBuilder{})
}

func main() {
	ctx, cancel := utils.ManualContext("backupcli-"+os.Args[1], time.Minute)
	defer cancel()
	conn, err := utils.LFDialServer(ctx, "backup")
	if err != nil {
		log.Fatalf("Unable to dial: %v", err)
	}
	defer conn.Close()

	client := pb.NewBackupServiceClient(conn)

	switch os.Args[1] {
	case "run":
		_, err := client.RunBackup(ctx, &pb.RunBackupRequest{})
		fmt.Printf("RAN: %v\n", err)
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
	case "stats":
		res, err := client.GetStats(ctx, &pb.StatsRequest{})
		if err != nil {
			log.Fatalf("Save error: %v", err)
		}
		for _, stat := range res.GetStats() {
			fmt.Printf("%v - %v\n", stat.GetCount(), stat.GetState())
		}
	}

}
