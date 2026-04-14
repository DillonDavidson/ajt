package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	SUBTITLE_FILE string  = "/tmp/subs.srt"
	PAD           float32 = 0.2
)

type Segment struct {
	Start, End float64
}

// To-Do: Parallelize
func condense() error {
	var failed []string
	mkvFiles, err := findFiles("mkv")
	if err != nil {
		return err
	}

	if err := os.MkdirAll("audio", 0o755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	for _, file := range mkvFiles {
		fmt.Println("Processing ", file)
		clearTempFile(SUBTITLE_FILE)

		if err := extractInternalSubs(file); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", file, err)
			failed = append(failed, file)
			continue
		}

		if err := shiftSubtitleFile(SUBTITLE_FILE, SRT, -PAD, PAD); err != nil {
			return err
		}

		segments, err := srtToSegments()
		if err != nil {
			return err
		}

		baseName := strings.TrimSuffix(filepath.Base(file), ".mkv") + ".ogg"
		outputFile := filepath.Join("audio", baseName)
		if err := runFFmpeg(file, outputFile, buildFilterGraph(segments)); err != nil {
			fmt.Fprintf(os.Stderr, "warning: skipping %q: %v\n", file, err)
			failed = append(failed, file)
			continue
		}
	}

	if len(failed) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d file(s) failed:\n", len(failed))
		for _, f := range failed {
			fmt.Fprintf(os.Stderr, "  %s\n", f)
		}
		return fmt.Errorf("%d file(s) failed condensing", len(failed))
	}

	return nil
}

func clearTempFile(file string) {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		os.Remove(file)
	}
}

func extractInternalSubs(file string) error {
	cmd := exec.Command("ffmpeg", "-y", "-i", file, "-map", "0:s:0", SUBTITLE_FILE)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}

func srtToSegments() ([]Segment, error) {
	data, err := os.ReadFile(SUBTITLE_FILE)
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
	fmt.Println(cmd)
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	return cmd.Run()
}
