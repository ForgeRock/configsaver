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

// Package implements the config saver server
package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	f "github.com/ForgeRock/configsaver/internal/fileutils"
	git "github.com/ForgeRock/configsaver/internal/git"

	pb "github.com/ForgeRock/configsaver/proto"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

// config saver server context + config
type ConfigServer struct {
	// The top of directory where we serve config from.
	RootDirectory string
	// Map of the relative paths to the product configuration.
	// Example:  am: docker/am/configs/cdk
	ProductPath map[string]string
	*f.FileUtil
	*git.GitRepo
	pb.UnimplementedConfigSaverServer // for gRPC
}

var config *ConfigServer

func main() {
	rootDir := "tmp/frconfig"
	if len(os.Args) > 1 {
		rootDir = os.Args[1]
	}

	// This will look for GIT_REPO and GIT_SSH_PATH environment variables.
	gitRepo, err := git.OpenGitRepo(rootDir, "master")
	if err != nil {
		log.Fatalf("failed to open git repo: %v", err)
	}

	config = &ConfigServer{
		RootDirectory: rootDir,
		ProductPath: map[string]string{
			"am":  "docker/am/config-profiles/cdk",
			"idm": "docker/idm/config-profiles/cdk",
		},
		FileUtil: f.NewFileUtil(rootDir),
		GitRepo:  gitRepo,
	}

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterConfigSaverServer(s, config)
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// GetConfig returns the entire config for a given product. Returns to the caller as tar file
func (s *ConfigServer) GetConfig(ctx context.Context, in *pb.GetConfigRequest) (*pb.GetConfigReply, error) {

	log.Printf("GetConfig product: %s commit: %s", in.ProductId, in.CommitId)
	bytes, err := s.FileUtil.GetAllConfiguration(config.ProductPath[in.ProductId])
	if err != nil {
		return &pb.GetConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}
	fmt.Printf("sending tar file with %d bytes", len(bytes))
	return &pb.GetConfigReply{Status: 0, ErrorMessage: "ok", ConfigTar: bytes}, nil
}

// UpdateConfig is called by the client to pass along config updates to be saved.
func (s *ConfigServer) UpdateConfig(ctx context.Context, in *pb.UpdateConfigRequest) (*pb.UpdateConfigReply, error) {
	log.Printf("UpdateConfig product: %s commit: %s", in.ProductId, in.CommitId)

	// Unpack the tar file containing the changes
	err := s.FileUtil.UnpackTarBuffer(in.ConfigTar, config.ProductPath[in.ProductId])
	if err != nil {
		log.Printf("could not unpack tar buffer: %v\n", err)
		return &pb.UpdateConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}
	// new see if the client removed any files.
	if len(in.DeletedFiles) > 0 {
		err = config.FileUtil.DeleteFiles(in.DeletedFiles, config.ProductPath[in.ProductId])
		if err != nil {
			return &pb.UpdateConfigReply{Status: 1, ErrorMessage: err.Error()}, err
		}
	}
	// Update git...
	if err = s.GitRepo.GitStatusAndCommit(); err != nil {
		fmt.Printf("error commiting changes to git %v", err)
		return &pb.UpdateConfigReply{Status: 1, ErrorMessage: err.Error()}, err
	}

	return &pb.UpdateConfigReply{Status: 0, ErrorMessage: "ok"}, nil
}
