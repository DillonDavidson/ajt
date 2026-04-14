package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func processSRTLine(line string, writer *bufio.Writer, startDelta, endDelta float32) error {
	if !strings.Contains(line, "-->") {
		fmt.Fprintln(writer, line)
		return nil
	}

	parts := strings.Split(line, " --> ")

	newStart, err := shiftTime(parts[0], startDelta, SRT)
	if err != nil {
		return err
	}

	newEnd, err := shiftTime(parts[1], endDelta, SRT)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "%s --> %s\n", newStart, newEnd)

	return nil
}

// To-do: Manually parse this and remove time.Time
func shiftSRTTime(timestamp string, deltaSeconds float32) (string, error) {
	layout := "15:04:05,000" // idk, vibe coded this part
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		return "", err
	}

	delta := time.Duration(deltaSeconds * float32(time.Second))
	newTime := t.Add(delta)

	// HACK: time.Parse uses 0000-01-01 as the base date, so we can't clamp
	// against time.Time{} — use explicit midnight instead. Remove when we
	// ditch time.Time for raw milliseconds.
	midnight := time.Date(0, time.January, 1, 0, 0, 0, 0, time.UTC) // hack for now
	if newTime.Before(midnight) {
		newTime = midnight
	}

	return newTime.Format(layout), nil
}

func srtTimeToSeconds(timestamp string) (float64, error) {
	t, err := time.Parse("15:04:05,000", timestamp)
	if err != nil {
		return 0, fmt.Errorf("parsing timestamp %q: %w", timestamp, err)
	}

	return float64(t.Hour())*3600 +
		float64(t.Minute())*60 +
		float64(t.Second()) +
		float64(t.Nanosecond())/1e9, nil
}
