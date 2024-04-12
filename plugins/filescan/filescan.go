package filescan

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

func FileScan(root, searchTerm string, pathsChannel chan<- string, ctx context.Context) {
	defer close(pathsChannel)

	visit := func(path string, _ fs.DirEntry, err error) error {
		select {
		case <-ctx.Done():
			log.Println("[file scan] Timeout")
			return filepath.SkipAll
		default:
			if !strings.Contains(path, ".git/") && strings.Contains(path, searchTerm) {
				select {
				case pathsChannel <- path:
				default:
					return filepath.SkipAll
				}
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
