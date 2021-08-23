package maintest

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"time"
)

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

func main() {

	var updateMap = make(map[string]time.Time)

	for {

		newPaths := make(map[string]bool)

		err := filepath.WalkDir("../forgeops/docker/am", func(path string, d fs.DirEntry, err error) error {
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
		time.Sleep(time.Second * 2)
	}
}
