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

// Implements the config-saver client
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	f "github.com/ForgeRock/configsaver/internal/fileutils"
	pb "github.com/ForgeRock/configsaver/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type clientCtx struct {
	server          string
	conn            *grpc.ClientConn
	configDirectory string
	fileUtil        *f.FileUtil
	grpc            pb.ConfigSaverClient
}

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	PermitWithoutStream: true,             // send pings even without active streams
}

// With no args we get the config from the server and exit.
// with one arg (the time in seconds) we scan for changes and upload to the server
func main() {

	if len(os.Args) > 2 {
		log.Fatalf("Usage: %s [scanSeconds]", os.Args[0])
	}

	// where to save the config
	configDir := f.GetEnvOrDefault("CONFIG_DIR", "/tmp")
	// which product we want to config for
	configProduct := f.GetEnvOrDefault("CONFIG_PRODUCT", "am")
	// The config server address:port
	server := f.GetEnvOrDefault("CONFIG_SERVER", "localhost:50051")

	log.Printf("config_client starting. product: %s,  configDir: %s\n", configProduct, configDir)

	// todo: Move to NewFileUtil
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating config directory %s: %v", configDir, err)
		}
	}

	log.Printf("Waiting for server connection %s\n", server)
	// Set up a connection to the server.
	conn, err := grpc.Dial(server, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithKeepaliveParams(kacp))
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewConfigSaverClient(conn)

	client := clientCtx{
		server:          server,
		configDirectory: configDir,
		fileUtil:        f.NewFileUtil(configDir),
		conn:            conn,
		grpc:            c,
	}

	// If there is only one arg, read the config from the server and exit
	if len(os.Args) == 1 {
		client.getConfigFromServer(configProduct)
		os.Exit(0)
	}

	// There is more than org, so we want to iterate looking for changes to send to the server.
	scanSeconds, err := strconv.Atoi(os.Args[1])
	if err != nil || scanSeconds < 1 || scanSeconds > 120 {
		log.Fatalf("Invalid scanSeconds: %s. Must be between 1 and 120", os.Args[2])
	}

	scanDuration := time.Duration(scanSeconds) * time.Second
	client.scanAndSaveToServer(scanDuration, configProduct)

}

func (client *clientCtx) getConfigFromServer(productId string) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*120)
	defer cancel()
	r, err := client.grpc.GetConfig(ctx, &pb.GetConfigRequest{ProductId: productId, CommitId: "master"})
	if err != nil {
		log.Fatalf("could not get configuration for %s from the server: %v", "am", err)
	}
	log.Printf("Status = %d Error message: %s", r.Status, r.ErrorMessage)
	if err := client.fileUtil.UnpackTarBuffer(r.GetConfigTar(), ""); err != nil {
		log.Fatalf("could not unpack configuration: %v", err)
	}
}

// Loops looking for changes to the config directory and uploads to the server
func (client *clientCtx) scanAndSaveToServer(scanDuration time.Duration, productId string) {

	// The first pass through scans the initial files. We do this so all the files don't
	// get flagged as new
	err := client.fileUtil.ScanFiles()
	if err != nil {
		log.Printf("Error scanning files: %v", err)
	}

	// now loop looking for changes
	for {
		tarBytes := make([]byte, 0)

		err := client.fileUtil.ScanFiles()
		if err != nil {
			log.Printf("Error scanning files: %v", err)
		}
		newOrModifiedFiles := len(client.fileUtil.ModifiedFiles) > 0 || len(client.fileUtil.NewFiles) > 0
		if newOrModifiedFiles {
			tarBytes, err = client.fileUtil.TarUpModifiedFiles()
			if err != nil {
				log.Printf("Error creating tar: %v", err)
			}
			log.Printf("Number files modified = %d  new = %d tar file size=%d\n", len(client.fileUtil.ModifiedFiles), len(client.fileUtil.NewFiles), len(tarBytes))
		}

		// if there are new files, modified files, or deleted files, then let the server know
		if newOrModifiedFiles || len(client.fileUtil.DeletedFiles) > 0 {
			for {

				log.Printf("updating server, modified=%d  new=%d deleted=%d  tar_bytes=%d\n",
					len(client.fileUtil.ModifiedFiles), len(client.fileUtil.NewFiles), len(client.fileUtil.DeletedFiles), len(tarBytes))
				// todo: what to do about defer in infinite loop?
				ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
				r, err := client.grpc.UpdateConfig(ctx, &pb.UpdateConfigRequest{
					CommitId:     "master",
					ProductId:    productId,
					ConfigTar:    tarBytes,
					DeletedFiles: client.fileUtil.DeletedFiles,
				})
				cancel()

				if err != nil {
					log.Printf("error updating server %v. Ill try again", err)
				} else {
					log.Printf("response status %d  %s", r.Status, r.ErrorMessage)
					break
				}
				time.Sleep(time.Second * 10)

			}
		}

		time.Sleep(scanDuration)
	}

}
