// Package main implements a client for Greeter service.
package main

import (
	"context"
	"log"
	"time"

	f "github.com/ForgeRock/configsaver/fileutils"

	pb "github.com/ForgeRock/configsaver/proto"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:50051"
	defaultName = "worl FOOOO"
)

func main() {
	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewConfigSaverClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetConfig(ctx, &pb.GetConfigRequest{ProductId: "am", CommitId: "master"})
	if err != nil {
		log.Fatalf("could not get configuration: %v", err)
	}
	log.Printf("Status = %d Error message: %s", r.Status, r.ErrorMessage)
	f.UnpackTarGzBuffer(r.GetConfigTarGz(), "/var/tmp/configsaver")

}
