package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

func impd() error {
	if err := os.MkdirAll("audio", os.ModePerm); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	files, err := findFiles("mkv")
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return errors.New("no .mkv files found")
	}

	var wg sync.WaitGroup

	for _, mkvfile := range files {
		wg.Add(1)

		go func(mkv string) {
			defer wg.Done()

			basename := strings.TrimSuffix(mkv, ".mkv")
			fmt.Println("Processing:", basename)

			var subtitleFile string
			if _, err := os.Stat(basename + ".ass"); err == nil {
				subtitleFile = basename + ".ass"
			} else if _, err := os.Stat(basename + ".srt"); err == nil {
				subtitleFile = basename + ".srt"
			}

			outputFile := filepath.Join("audio", basename+".ogg")

			if subtitleFile != "" {
				cmd := exec.Command("impd", "condense", "-f", "-i", mkv, "-s", subtitleFile, "-o", outputFile)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					fmt.Printf("Error processing %s: %v\n", basename, err)
				}
			} else {
				fmt.Println("No subtitle found for", basename)
			}
		}(mkvfile)
	}

	wg.Wait()
	return nil
}
