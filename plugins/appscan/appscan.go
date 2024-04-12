// Package appscan Scans installed apps (MacOS only for now)
package appscan

import (
	"log"
	"os"
	"strings"
)

// AppScan scans installed app in MacOS (for now)
func AppScan(searchTerm string, appsChannel chan<- string) {
	defer close(appsChannel)

	dir, err := os.Open("/Applications/")
	if err != nil {
		log.Fatalln(err)
	}

	entries, err := dir.Readdirnames(-1)
	if err != nil {
		log.Fatalln(err)
	}

	for _, entry := range entries {
		if strings.Contains(entry, searchTerm) {
			select {
			case appsChannel <- entry:
			default:
				return
			}
		}
	}
}
