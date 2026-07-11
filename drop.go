package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func audioDropCmd() error {
	return fullDrop("audio")
}

func subDropCmd() error {
	return fullDrop("subtitle")
}

func fullDrop(kind string) error {
	files, err := filepath.Glob("*.mkv")
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no mkv files found")
	}

	if err := checkBatchSafety(files); err != nil {
		return err
	}

	sample := files[0]
	fmt.Printf("detecting %s tracks from: %s\n\n", kind, sample)

	var preferredLang string
	if kind == "subtitle" {
		preferredLang = "eng"
	} else if kind == "audio" {
		preferredLang = "jpn"
	}

	_, selectedID, _, err := selectTrack(sample, kind, preferredLang)
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := dropTrack(file, selectedID, kind); err != nil {
			return err
		}
	}

	fmt.Println("all files processed successfully!")
	return nil
}

func dropTrack(file, ID, kind string) error {
	var trackFlag string
	if kind == "subtitle" {
		trackFlag = "--subtitle-tracks"
	} else {
		trackFlag = "--audio-tracks"
	}

	base := strings.TrimSuffix(file, ".mkv")
	tempName := fmt.Sprintf("temp%s.mkv", base)
	fmt.Printf("processing %s...\n", file)

	cmd := exec.Command("mkvmerge", "-o", tempName, trackFlag, ID, file)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remux %s: %v", file, err)
	}

	if err := os.Remove(file); err != nil {
		return fmt.Errorf("failed to remove original file %s: %v", file, err)
	}

	if err := os.Rename(tempName, file); err != nil {
		return fmt.Errorf("failed to rename temp file: %v", err)
	}

	return nil
}
