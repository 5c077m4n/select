// Package filescan scans files
package filescan

import (
	"context"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

// FileScan scans files
func FileScan(ctx context.Context, root, searchTerm string, pathsChannel chan<- string) {
	defer close(pathsChannel)

	visit := func(path string, _ fs.DirEntry, err error) error {
		if err != nil {
			log.Fatalln(err)
		}

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
