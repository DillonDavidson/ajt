package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

func streamTypeFor(kind string) (string, error) {
	switch kind {
	case "subtitle":
		return "s", nil
	case "audio":
		return "a", nil
	case "video":
		return "v", nil
	default:
		return "", fmt.Errorf("kind %s is not allowed, what are you doing?", kind)
	}
}

// selectTrack lists all tracks of the given kind from sample, auto-selects
// if exactly one matches preferredLang, otherwise prompts the user, and
// returns the chosen track's absolute ID and its relative ffmpeg index.
func selectTrack(sample, kind, preferredLang string) (tracks []Track, selectedID string, relIndex int, err error) {
	if _, err = streamTypeFor(kind); err != nil {
		return nil, "", -1, err
	}

	tracks, err = getTracks(sample, kind)
	if err != nil {
		return nil, "", -1, fmt.Errorf("failed to get %s tracks: %v", kind, err)
	}
	if len(tracks) == 0 {
		return nil, "", -1, fmt.Errorf("no %s tracks found", kind)
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

	var targetTracks []Track
	if preferredLang != "" {
		for _, t := range tracks {
			props := parseProperties(t.Properties)
			if props.Language == preferredLang {
				targetTracks = append(targetTracks, t)
			}
		}
	}

	if len(targetTracks) == 1 {
		selectedID = fmt.Sprint(targetTracks[0].ID)
		fmt.Printf("\nauto-selected %s %s track: ID %s\n", preferredLang, kind, selectedID)
	} else {
		selectedID = promptInput("which " + kind + " track ID are we keeping? ")
	}

	relIndex, err = findRelativeIndex(tracks, selectedID)
	if err != nil {
		return nil, "", -1, err
	}

	return tracks, selectedID, relIndex, nil
}

func findRelativeIndex(tracks []Track, selectedID string) (int, error) {
	for i, t := range tracks {
		if fmt.Sprint(t.ID) == selectedID {
			return i, nil
		}
	}
	return -1, fmt.Errorf("track ID %s not found in track list", selectedID)
}

func promptInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
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
