package main

import (
	"os"
	"strings"
)

func findFiles(ext string) ([]string, error) {
	entries, err := os.ReadDir(".")
	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ext) {
			files = append(files, name)
		}
	}

	return files, nil
}
