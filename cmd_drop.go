package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

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

func getAudioTracks(file string) ([]Track, error) {
	out, err := exec.Command("mkvmerge", "-J", file).Output()
	if err != nil {
		return nil, err
	}

	var info MKVInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}

	audioTracks := []Track{}
	for _, t := range info.Tracks {
		if t.Type == "audio" {
			audioTracks = append(audioTracks, t)
		}
	}
	return audioTracks, nil
}

func parseProperties(raw json.RawMessage) TrackProperties {
	var props TrackProperties
	if err := json.Unmarshal(raw, &props); err != nil {
		// Default to undetermined language if parsing fails
		props.Language = "und"
		props.TrackName = ""
	}
	return props
}

func checkBatchSafety(files []string) {
	var reference []string

	for _, f := range files {
		out, err := exec.Command("mkvmerge", "-J", f).Output()
		if err != nil {
			log.Fatalf("Failed to run mkvmerge on %s: %v", f, err)
		}

		var info MKVInfo
		if err := json.Unmarshal(out, &info); err != nil {
			log.Fatalf("Failed to parse JSON for %s: %v", f, err)
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
				log.Fatalf("Batch safety check failed for %s: track layout does not match reference file", f)
			}
		}
	}
}

func promptInput(prompt string) string {
	fmt.Print(prompt)
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func audioDrop() error {
	files, err := filepath.Glob("*.mkv")
	if err != nil || len(files) == 0 {
		return errors.New("no mkv files found")
	}

	checkBatchSafety(files)

	// Detect audio tracks from first file
	sample := files[0]
	fmt.Printf("Detecting audio tracks from: %s\n\n", sample)
	audioTracks, err := getAudioTracks(sample)
	if err != nil {
		log.Fatalf("Failed to get audio tracks: %v", err)
	}
	if len(audioTracks) == 0 {
		log.Fatal("No audio tracks found.")
	}

	fmt.Println("Available audio tracks:")
	for _, t := range audioTracks {
		props := parseProperties(t.Properties)
		name := ""
		if props.TrackName != "" {
			name = " (" + props.TrackName + ")"
		}
		fmt.Printf("  ID %d: %s | lang=%s%s\n", t.ID, t.Codec, props.Language, name)
	}

	// Auto-select Japanese track
	jpnTracks := []Track{}
	for _, t := range audioTracks {
		props := parseProperties(t.Properties)
		if props.Language == "jpn" {
			jpnTracks = append(jpnTracks, t)
		}
	}

	var selectedID string
	if len(jpnTracks) == 1 {
		selectedID = fmt.Sprint(jpnTracks[0].ID)
		fmt.Printf("\nAuto-selected Japanese track: ID %s\n", selectedID)
	} else {
		if len(jpnTracks) == 0 {
			fmt.Println("\nNo Japanese (jpn) audio track found.")
		} else {
			fmt.Println("\nMultiple Japanese audio tracks found.")
		}
		selectedID = promptInput("Which audio track ID are we keeping? ")
	}

	// Re-mux each MKV file
	for _, f := range files {
		base := strings.TrimSuffix(f, ".mkv")
		tempName := fmt.Sprintf("temp%s.mkv", base)
		fmt.Printf("Processing %s...\n", f)

		cmd := exec.Command("mkvmerge", "-o", tempName, "--audio-tracks", selectedID, f)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to remux %s: %v", f, err)
		}

		if err := os.Remove(f); err != nil {
			log.Fatalf("Failed to remove original file %s: %v", f, err)
		}
		if err := os.Rename(tempName, f); err != nil {
			log.Fatalf("Failed to rename temp file: %v", err)
		}
	}
	fmt.Println("All files processed successfully!")

	return nil
}

func getSubtitleTracks(file string) ([]Track, error) {
	out, err := exec.Command("mkvmerge", "-J", file).Output()
	if err != nil {
		return nil, err
	}

	var info MKVInfo
	if err := json.Unmarshal(out, &info); err != nil {
		return nil, err
	}

	subTracks := []Track{}
	for _, t := range info.Tracks {
		if t.Type == "subtitles" {
			subTracks = append(subTracks, t)
		}
	}
	return subTracks, nil
}

func subDrop() error {
	files, err := filepath.Glob("*.mkv")
	if err != nil || len(files) == 0 {
		return errors.New("no mkv files found")
	}

	checkBatchSafety(files)

	// Detect subtitle tracks from first file
	sample := files[0]
	fmt.Printf("Detecting subtitle tracks from: %s\n\n", sample)

	subTracks, err := getSubtitleTracks(sample)
	if err != nil {
		log.Fatalf("Failed to get subtitle tracks: %v", err)
	}
	if len(subTracks) == 0 {
		log.Fatal("No subtitle tracks found.")
	}

	fmt.Println("Available subtitle tracks:")
	for _, t := range subTracks {
		props := parseProperties(t.Properties)
		name := ""
		if props.TrackName != "" {
			name = " (" + props.TrackName + ")"
		}
		fmt.Printf("  ID %d: %s | lang=%s%s\n", t.ID, t.Codec, props.Language, name)
	}

	// Auto-select English subtitle track
	engTracks := []Track{}
	for _, t := range subTracks {
		props := parseProperties(t.Properties)
		if props.Language == "eng" {
			engTracks = append(engTracks, t)
		}
	}

	var selectedID string
	if len(engTracks) == 1 {
		selectedID = fmt.Sprint(engTracks[0].ID)
		fmt.Printf("\nAuto-selected English subtitle track: ID %s\n", selectedID)
	} else {
		if len(engTracks) == 0 {
			fmt.Println("\nNo English (eng) subtitle track found.")
		} else {
			fmt.Println("\nMultiple English subtitle tracks found.")
		}
		selectedID = promptInput("Which subtitle track ID are we keeping? ")
	}

	// Re-mux each MKV file
	for _, f := range files {
		base := strings.TrimSuffix(f, ".mkv")
		tempName := fmt.Sprintf("temp%s.mkv", base)
		fmt.Printf("Processing %s...\n", f)

		cmd := exec.Command(
			"mkvmerge",
			"-o", tempName,
			"--subtitle-tracks", selectedID,
			f,
		)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			log.Fatalf("Failed to remux %s: %v", f, err)
		}

		if err := os.Remove(f); err != nil {
			log.Fatalf("Failed to remove original file %s: %v", f, err)
		}
		if err := os.Rename(tempName, f); err != nil {
			log.Fatalf("Failed to rename temp file: %v", err)
		}
	}

	fmt.Println("All files processed successfully!")
	return nil
}
