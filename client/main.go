// Package main implements a client for Greeter service.
package main

import (
	"log"
	"time"

	f "github.com/ForgeRock/configsaver/internal/fileutils"
)

const (
	address = "localhost:50051"
)

func main() {

	// Test walking files
	futil := f.NewFileUtil("/tmp/cs")

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
			// TODO
			log.Printf("updating server, modified=%d  new=%d deleted=%d  tar_bytes=%d\n",
				len(futil.ModifiedFiles), len(futil.NewFiles), len(futil.DeletedFiles), len(tarBytes))

		}

		time.Sleep(time.Second * 5)
	}

	// // Set up a connection to the server.
	// conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	// if err != nil {
	// 	log.Fatalf("did not connect: %v", err)
	// }
	// defer conn.Close()
	// c := pb.NewConfigSaverClient(conn)

	// ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	// defer cancel()
	// r, err := c.GetConfig(ctx, &pb.GetConfigRequest{ProductId: "am", CommitId: "master"})
	// if err != nil {
	// 	log.Fatalf("could not get configuration: %v", err)
	// }
	// log.Printf("Status = %d Error message: %s", r.Status, r.ErrorMessage)
	// if err := f.UnpackTarBuffer(r.GetConfigTar(), "/tmp/cs"); err != nil {
	// 	log.Fatalf("could not unpack configuration: %v", err)
	// }

}
