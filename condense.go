package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

const (
	PAD float32 = 0.2
)

type Segment struct {
	Start, End float64
}

func condenseCmd() error {
	mkvFiles, err := findFiles("mkv")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("audio", 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	sem := make(chan struct{}, runtime.NumCPU())
	var wg sync.WaitGroup

	for _, file := range mkvFiles {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()
			condense(file)
		})
	}

	wg.Wait()

	return nil
}

func condense(file string) {
	fmt.Println("Processing ", file)

	tempFile := "/tmp/" + file + ".srt"
	clearTempFile(tempFile)

	if err := extractInternalSubs(file, tempFile); err != nil {
		fmt.Printf("warning: skipping %q: %v\n", file, err)
		return
	}

	if err := shiftSubtitleFile(tempFile, SRT, -PAD, PAD); err != nil {
		fmt.Println(err)
		return
	}

	segments, err := srtToSegments(tempFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	baseName := strings.TrimSuffix(filepath.Base(file), ".mkv") + ".ogg"
	outputFile := filepath.Join("audio", baseName)
	if err := runFFmpeg(file, outputFile, buildFilterGraph(segments)); err != nil {
		fmt.Printf("warning: skipping %q: %v\n", file, err)
		return
	}

	clearTempFile(tempFile)
}

func clearTempFile(file string) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		os.Remove(file)
	}
}

func extractInternalSubs(file, subFile string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", file, "-map", "0:s:0", subFile)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func srtToSegments(subFile string) ([]Segment, error) {
	data, err := os.ReadFile(subFile)
	if err != nil {
		return nil, err
	}

	var segments []Segment
	for line := range strings.SplitSeq(string(data), "\n") {
		if !strings.Contains(line, "-->") {
			continue
		}
		parts := strings.Split(line, " --> ")
		if len(parts) != 2 {
			return nil, fmt.Errorf("malformed timestamp line: %q", line)
		}
		start, err := srtTimeToSeconds(strings.TrimSpace(parts[0]))
		if err != nil {
			return nil, err
		}
		end, err := srtTimeToSeconds(strings.TrimSpace(parts[1]))
		if err != nil {
			return nil, err
		}
		segments = append(segments, Segment{Start: start, End: end})
	}
	return segments, nil
}

func buildFilterGraph(segments []Segment) string {
	var sb strings.Builder

	for i, seg := range segments {
		fmt.Fprintf(&sb, "[0:a]atrim=start=%.3f:end=%.3f,asetpts=PTS-STARTPTS[a%d]; ", seg.Start, seg.End, i)
	}

	for i := range segments {
		fmt.Fprintf(&sb, "[a%d]", i)
	}

	fmt.Fprintf(&sb, "concat=n=%d:v=0:a=1[out]", len(segments))

	return sb.String()
}

func runFFmpeg(input, output, filter string) error {
	cmd := exec.Command("ffmpeg",
		"-nostdin",
		"-loglevel", "error",
		"-hide_banner",
		"-y",
		"-vn",
		"-sn",
		"-i", input,
		"-filter_complex", filter,
		"-map", "[out]",
		"-map_metadata", "-1",
		"-ac", "2",
		"-vbr", "on",
		"-compression_level", "10",
		"-application", "voip",
		"-acodec", "libopus",
		output,
	)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
