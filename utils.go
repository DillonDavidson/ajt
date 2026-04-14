package main

import (
	"errors"
	"os"
	"strings"
)

func shiftTime(timestamp string, deltaSeconds float32, subType SubtitleType) (string, error) {
	switch subType {
	case SRT:
		return shiftSRTTime(timestamp, deltaSeconds)
	case ASS:
		return shiftASSTime(timestamp, deltaSeconds)
	default:
		return "", errors.New("no idea what subtitle type")
	}
}

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
