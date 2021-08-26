package fileutils

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	// Whether to gzip the tar file or not.
	// The gzip compressor triggers tar errors if the number of bytes written is too small
	// TODO: find out why this is the case (extra padding needed?)
	// For now, we can use gRPC compression if performance is a concern
	UseCompression = false
)

// Get the entire configuration for the product as a tarball of bytes
// rootDir is the top of the directory (example, tmp/forgeops).
// productPath is the relative path under that root where the configuration files are (example, docker/am/product-configs/cdk)
func GetAllConfiguration(rootDir, productPath string) ([]byte, error) {
	var paths []string
	dir := filepath.Join(rootDir, productPath)
	// recursively walk the directory
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			// add to the list of paths we want to include in the tarball
			paths = append(paths, path)
			return nil
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	// create the tarball from the list of paths
	buf, err := CreateTarBuffer(dir, paths)
	if err != nil {
		return nil, err
	}

	return buf, nil

}

// TODO: testing the cost of looking for changes.
func SendFiles() {

	var updateMap = make(map[string]time.Time)

	for {

		newPaths := make(map[string]bool)

		err := filepath.WalkDir(".", func(path string, d fs.DirEntry, err error) error {
			_ = walkDirFunction(path, d, updateMap, newPaths)
			return nil
		})

		// look for keys in fmap that are NOT in updateMap - they must be deleted files.
		for k, _ := range updateMap {
			if _, ok := newPaths[k]; !ok {
				fmt.Printf("%s deleted\n", k)
				delete(updateMap, k)
			}
		}

		if err != nil {
			panic(err)
		}
		time.Sleep(time.Second * 5)
	}
}

// Create an in-memory tarball of the listed files.
// Rootdir is the top of the config directory  (example, tmp/forgeops/docker/am/product-configs/cdk).
// filepaths are the list of files to include in the tarball
func CreateTarBuffer(rootDir string, filePaths []string) ([]byte, error) {

	// var buf bytes.Buffer
	var buf bytes.Buffer

	var tarWriter *tar.Writer

	if UseCompression {
		gzipWriter := gzip.NewWriter(&buf)
		defer gzipWriter.Close()

		tarWriter = tar.NewWriter(gzipWriter)
	} else {
		tarWriter = tar.NewWriter(&buf)
	}

	defer tarWriter.Close()

	for _, filePath := range filePaths {
		err := addFileToTarWriter(rootDir, filePath, tarWriter)
		if err != nil {
			return buf.Bytes(), fmt.Errorf("could not add file '%s', to tarball, got error '%v'", filePath, err)
		}
	}

	return buf.Bytes(), nil
}

// Given a tar file in a memory buf, unpack it to the specified rootDir directory.
func UnpackTarBuffer(buf []byte, rootDir string) error {
	reader := bytes.NewReader(buf)
	var tarReader *tar.Reader

	if UseCompression {
		gzipReader, err := gzip.NewReader(reader)
		if err != nil {
			return fmt.Errorf("could not create gzip reader, got error '%v'", err.Error())
		}
		defer gzipReader.Close()
		tarReader = tar.NewReader(gzipReader)

	} else {
		tarReader = tar.NewReader(reader)
	}

	if err := os.MkdirAll(rootDir, 0755); err != nil {
		return fmt.Errorf("could not create directory '%s', got error '%v'", rootDir, err.Error())
	}

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read next tar header, got error '%v'", err.Error())
		}

		path := filepath.Join(rootDir, header.Name)
		// fmt.Printf("got header %s restorePath %s\n", header.Name, path)

		err = os.MkdirAll(filepath.Dir(path), 0755)
		if err != nil {
			return fmt.Errorf("could not create directory for file '%s', got error '%v'", path, err.Error())
		}

		file, err := os.Create(path)
		if err != nil {
			return fmt.Errorf("could not create file '%s', got error '%v'", path, err.Error())
		}
		defer file.Close()

		_, err = io.Copy(file, tarReader)
		if err != nil {
			return fmt.Errorf("could not copy data from tarball to file '%s', got error '%v'", header.Name, err.Error())
		}
	}

	return nil
}

// Add a file to the tarball. The rootDir prefix is stripped from the archive so that
// the receiver can restore the archive to a preferred relative location
func addFileToTarWriter(rootDir, filePath string, tarWriter *tar.Writer) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("could not open file '%s', got error '%s'", filePath, err.Error())
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("could not get stat for file '%s', got error '%v'", filePath, err.Error())
	}

	tarPath := filePath[len(rootDir):]

	log.Printf("Adding %s to tar as %s\n", filePath, tarPath)

	header := &tar.Header{
		Name:    tarPath,
		Size:    stat.Size(),
		Mode:    int64(stat.Mode()),
		ModTime: stat.ModTime(),
	}

	err = tarWriter.WriteHeader(header)
	if err != nil {
		return fmt.Errorf("could not write header for file '%s', got error '%v'", filePath, err.Error())
	}

	_, err = io.Copy(tarWriter, file)
	if err != nil {
		return fmt.Errorf("could not copy the file '%s' data to the tarball, got error '%v'", filePath, err.Error())
	}
	//fmt.Printf("Adding tar file %s relative path = %s size %d\n ", filePath, relativePath, stat.Size())

	return nil
}

// Function called for every file and directory we visit
// Map m is the existing files that we have observed in the previous iteration.
// Map newPaths are the new files in this iteration - we use this to determine if there are files in
// map m that are not in the current iteration. These are files that have been deleted from the filesystem

func walkDirFunction(path string, d fs.DirEntry, m map[string]time.Time, newPaths map[string]bool) error {
	if !d.IsDir() {
		info, _ := d.Info()
		t := info.ModTime()
		// Look up value in the main current map
		if val, ok := m[path]; ok {
			if t != val {
				fmt.Printf("%s changed time %v\n", path, t)
				m[path] = t
			}
		} else {
			fmt.Printf("adding %s to watch\n", path)
		}
		m[path] = t
		newPaths[path] = true
	}

	return nil
}
