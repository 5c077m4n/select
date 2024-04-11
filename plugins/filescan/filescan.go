package filescan

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

func FileScan(root, searchTerm string, pathChannel chan<- string, ctx context.Context) {
	defer close(pathChannel)

	visit := func(path string, dirEntry fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			log.Println("[file scan] Timeout")
			return filepath.SkipAll
		default:
			if strings.Contains(path, searchTerm) {
				pathChannel <- path
			}
			return nil
		}
	}

	root, err := filepath.Abs(root)
	if err != nil {
		log.Fatalln(err)
	}

	if err := filepath.WalkDir(root, visit); err != nil {
		log.Fatalln(err)
	}
}
