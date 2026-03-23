package main

import (
	"fmt"
	"os"
)

func rename(ext string) error {
	files, err := findFiles(ext)
	if err != nil {
		return err
	}

	for count, file := range files {
		newName := fmt.Sprintf("%02d.%s", count+1, ext)
		os.Rename(file, newName)
	}

	return nil
}
