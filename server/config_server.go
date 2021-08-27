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

	f "github.com/ForgeRock/configsaver/internal/fileutils"
	git "github.com/ForgeRock/configsaver/internal/git"

	pb "github.com/ForgeRock/configsaver/proto"
	"google.golang.org/grpc"
	// pb "proto/proto"
)

const (
	port = ":50051"
)

// The configuration of the config saver server
// Eventually this will be read from a config file or command line args.
type ConfigServerConfig struct {
	// The top of directory where we serve config from.
	RootDirectory string
	// Map of the relative paths to the product configuration.
	// Example:  am: docker/am/configs/cdk
	ProductPath map[string]string
	// TODO: Various git parameters here when we add git support. Default branch, upstream remotes, etc.
	*f.FileUtil
}

type server struct {
	pb.UnimplementedConfigSaverServer
}

// GetConfig returns the entire config for a given product. Returns to the caller as tar file
func (s *server) GetConfig(ctx context.Context, in *pb.GetConfigRequest) (*pb.GetConfigReply, error) {
	log.Printf("GetConfig product: %s commit: %s", in.ProductId, in.CommitId)
	bytes, err := f.GetAllConfiguration(config.RootDirectory, config.ProductPath[in.ProductId])
	if err != nil {
		return &pb.GetConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}
	fmt.Printf("sending tar file with %d bytes", len(bytes))
	return &pb.GetConfigReply{Status: 0, ErrorMessage: "ok", ConfigTar: bytes}, nil
}

// UpdateConfig is called by the client to pass along config updates to be saved.
func (s *server) UpdateConfig(ctx context.Context, in *pb.UpdateConfigRequest) (*pb.UpdateConfigReply, error) {
	log.Printf("UpdateConfig product: %s commit: %s", in.ProductId, in.CommitId)

	err := config.FileUtil.UnpackTarBuffer(in.ConfigTar)

	if err != nil {
		log.Printf("could not unpack tar buffer: %v\n", err)
		return &pb.UpdateConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}
	return &pb.UpdateConfigReply{Status: 0, ErrorMessage: "ok"}, nil
}

var config *ConfigServerConfig

func main() {

	rootDir := "tmp"

	futil := f.NewFileUtil(rootDir)

	config = &ConfigServerConfig{
		RootDirectory: "tmp/forgeops",
		ProductPath: map[string]string{
			"am":  "docker/am/config-profiles/cdk",
			"idm": "docker/idm/config-profiles/cdk",
		},
		FileUtil: futil,
	}

	log.Println("Getting the git repo")
	_, err := git.OpenGitRepo("https://stash.forgerock.org/scm/cloud/forgeops.git", config.RootDirectory, "master")

	if err != nil {
		log.Fatalf("failed to open git repo: %v", err)
	}

	// _ = git.GitStatusAndCommit()

	// panic("quit")

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
