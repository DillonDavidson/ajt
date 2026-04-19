package main

import (
	"bufio"
	"encoding/json"
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

type Track struct {
	ID         int             `json:"id"`
	Type       string          `json:"type"`
	Codec      string          `json:"codec"`
	Properties json.RawMessage `json:"properties"`
}

type TrackProperties struct {
	Language  string `json:"language"`
	TrackName string `json:"track_name"`
}

type MKVInfo struct {
	Tracks []Track `json:"tracks"`
}

func fullDrop(kind string) error {
	if kind != "subtitle" && kind != "audio" {
		return fmt.Errorf("kind %s is not allowed, what are you doing?\n", kind)
	}

	files, err := filepath.Glob("*.mkv")
	if err != nil || len(files) == 0 {
		return fmt.Errorf("no mkv files found")
	}

	if err := checkBatchSafety(files); err != nil {
		return err
	}

	sample := files[0]
	fmt.Printf("detecting %s tracks from: %s\n\n", kind, sample)

	tracks, err := getTracks(sample, kind)
	if err != nil {
		return fmt.Errorf("failed to get %s tracks: %v", kind, err)
	}
	if len(tracks) == 0 {
		return fmt.Errorf("no %s tracks found.", kind)
	}

	fmt.Printf("available %s tracks:\n", kind)
	for _, t := range tracks {
		props := parseProperties(t.Properties)
		name := ""
		if props.TrackName != "" {
			name = " (" + props.TrackName + ")"
		}
		fmt.Printf("  ID %d: %s | lang=%s%s\n", t.ID, t.Codec, props.Language, name)
	}

	var trackLanguage string
	if kind == "subtitle" {
		trackLanguage = "eng"
	} else {
		trackLanguage = "jpn"
	}

	targetTracks := []Track{}
	for _, t := range tracks {
		props := parseProperties(t.Properties)
		if props.Language == trackLanguage {
			targetTracks = append(targetTracks, t)
		}
	}

	var selectedID string
	if len(targetTracks) == 1 {
		selectedID = fmt.Sprint(targetTracks[0].ID)
		fmt.Printf("\nauto-selected %s %s track: ID %s\n", trackLanguage, kind, selectedID)
	} else {
		selectedID = promptInput("which " + kind + " track ID are we keeping? ")
	}

	// Re-mux each MKV file
	for _, file := range files {
		if err := dropTrack(file, selectedID, kind); err != nil {
			return err
		}
	}

	fmt.Println("all files processed successfully!")
	return nil
}

func getTracks(file, kind string) ([]Track, error) {
	var trackType string
	if kind == "subtitle" {
		trackType = "subtitles"
	} else {
		trackType = "audio"
	}

	out, err := exec.Command("mkvmerge", "-J", file).Output()
	if err != nil {
		return nil, err
	}

	var info MKVInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}

	tracks := []Track{}
	for _, t := range info.Tracks {
		if t.Type == trackType {
			tracks = append(tracks, t)
		}
	}

	return tracks, nil
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

func parseProperties(raw json.RawMessage) TrackProperties {
	var props TrackProperties
	if err := json.Unmarshal(raw, &props); err != nil {
		props.Language = "und"
		props.TrackName = ""
	}
	return props
}

func checkBatchSafety(files []string) error {
	var reference []string

	for _, f := range files {
		out, err := exec.Command("mkvmerge", "-J", f).Output()
		if err != nil {
			return fmt.Errorf("failed to run mkvmerge on %s: %v", f, err)
		}

		var info MKVInfo
		if err := json.Unmarshal(out, &info); err != nil {
			return fmt.Errorf("failed to parse json for %s: %v", f, err)
		}

		signature := []string{}
		for _, t := range info.Tracks {
			props := parseProperties(t.Properties)
			signature = append(signature, t.Type+"|"+props.TrackName)
		}

		if reference == nil {
			reference = signature
		} else {
			matched := len(signature) == len(reference)
			if matched {
				for i := range signature {
					if signature[i] != reference[i] {
						matched = false
						break
					}
				}
			}

			if !matched {
				return fmt.Errorf("batch safety check failed for %s: track layout does not match reference file", f)
			}
		}
	}

	return nil
}

func promptInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}
