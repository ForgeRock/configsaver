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
)

const (
	address = "localhost:50051"
)

func main() {

	if len(os.Args) <= 1 || len(os.Args) > 3 {
		log.Fatalf("Usage: %s configDirectory [scanSeconds]", os.Args[0])
	}

	configDir := os.Args[1]

	// todo: Move to NewFileUtil
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		err := os.Mkdir(configDir, os.ModePerm)
		if err != nil {
			log.Fatalf("Error creating config directory %s: %v", configDir, err)

		}
	}

	futil := f.NewFileUtil(configDir)

	log.Printf("Waiting for server connection..\n")

	// Set up a connection to the server.
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("could not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewConfigSaverClient(conn)

	// If there is only one arg, read the config from the server and exit
	if len(os.Args) == 2 {
		getConfigFromServer(c, futil)
		os.Exit(0)
	}

	// There is more than org, so we want to iterate looking for changes to send to the server.
	scanSeconds, err := strconv.Atoi(os.Args[2])
	if err != nil || scanSeconds < 1 || scanSeconds > 120 {
		log.Fatalf("Invalid scanSeconds: %s. Must be between 1 and 120", os.Args[2])
	}

	scanDuration := time.Duration(scanSeconds) * time.Second
	scanAndSaveToServer(c, futil, scanDuration)

}

func getConfigFromServer(c pb.ConfigSaverClient, futil *f.FileUtil) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.GetConfig(ctx, &pb.GetConfigRequest{ProductId: "am", CommitId: "master"})
	if err != nil {
		log.Fatalf("could not get configuration for %s from the server: %v", "am", err)
	}
	log.Printf("Status = %d Error message: %s", r.Status, r.ErrorMessage)
	if err := futil.UnpackTarBuffer(r.GetConfigTar(), ""); err != nil {
		log.Fatalf("could not unpack configuration: %v", err)
	}
}

func scanAndSaveToServer(c pb.ConfigSaverClient, futil *f.FileUtil, scanDuration time.Duration) {
	for {
		tarBytes := make([]byte, 0)

		err := futil.ScanFiles()
		if err != nil {
			log.Printf("Error walking files: %v", err)
		}
		newOrModifiedFiles := len(futil.ModifiedFiles) > 0 || len(futil.NewFiles) > 0
		if newOrModifiedFiles {
			tarBytes, err = futil.TarUpModifiedFiles()
			if err != nil {
				log.Printf("Error creating tar: %v", err)
			}
			log.Printf("Number files modified = %d  new = %d tar file size=%d\n", len(futil.ModifiedFiles), len(futil.NewFiles), len(tarBytes))
		}

		// if there are new files, modified files, or deleted files, then let the server know
		if newOrModifiedFiles || len(futil.DeletedFiles) > 0 {
			log.Printf("updating server, modified=%d  new=%d deleted=%d  tar_bytes=%d\n",
				len(futil.ModifiedFiles), len(futil.NewFiles), len(futil.DeletedFiles), len(tarBytes))
			// todo: what to do about defer in infinite loop?
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			//defer cancel()
			r, err := c.UpdateConfig(ctx, &pb.UpdateConfigRequest{
				CommitId:     "master",
				ProductId:    "am",
				ConfigTar:    tarBytes,
				DeletedFiles: futil.DeletedFiles,
			})
			if err != nil {
				log.Printf("error updating server %v", err)
			} else {
				log.Printf("response status %d  %s", r.Status, r.ErrorMessage)
			}
			cancel()
		}

		time.Sleep(scanDuration)
	}

}
