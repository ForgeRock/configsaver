/*
 *
 * Copyright  2021 ForgeRock AS
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"fmt"
	"log"
	"net"

	f "github.com/ForgeRock/configsaver/fileutils"

	pb "github.com/ForgeRock/configsaver/proto"
	"google.golang.org/grpc"
	// pb "proto/proto"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedConfigSaverServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) GetConfig(ctx context.Context, in *pb.GetConfigRequest) (*pb.GetConfigReply, error) {
	log.Printf("product: %s commit: %s", in.ProductId, in.CommitId)
	bytes, err := f.SendItAll("./proto")
	if err != nil {
		return &pb.GetConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}
	fmt.Printf("sending %d bytes", len(bytes))
	return &pb.GetConfigReply{Status: 0, ErrorMessage: "ok", ConfigTarGz: bytes}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterConfigSaverServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
