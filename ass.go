package main

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"
	"time"
)

func processASSLine(line string, writer *bufio.Writer, startDelta, endDelta float32) error {
	if !strings.HasPrefix(line, "Dialogue:") {
		fmt.Fprintln(writer, line)
		return nil
	}

	parts := strings.SplitN(line, ",", 10)
	if len(parts) < 3 {
		fmt.Fprintln(writer, line)
		return nil
	}

	newStart, err := shiftTime(parts[1], startDelta, ASS)
	if err != nil {
		return err
	}

	newEnd, err := shiftTime(parts[2], endDelta, ASS)
	if err != nil {
		return err
	}

	parts[1] = newStart
	parts[2] = newEnd

	fmt.Fprintln(writer, strings.Join(parts, ","))

	return nil
}

func shiftASSTime(timestamp string, deltaSeconds float32) (string, error) {
	// Split H:MM:SS.cc
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		return "", fmt.Errorf("invalid timestamp")
	}

	hours, _ := strconv.Atoi(parts[0])
	minutes, _ := strconv.Atoi(parts[1])

	secParts := strings.Split(parts[2], ".")
	seconds, _ := strconv.Atoi(secParts[0])
	centis, _ := strconv.Atoi(secParts[1])

	// Convert to duration
	total := time.Duration(hours)*time.Hour +
		time.Duration(minutes)*time.Minute +
		time.Duration(seconds)*time.Second +
		time.Duration(centis)*10*time.Millisecond

	// Apply shift
	total += time.Duration(deltaSeconds * float32(time.Second))

	// Clamp negative
	if total < 0 {
		total = 0
	}

	// Convert back
	h := total / time.Hour
	total %= time.Hour

	m := total / time.Minute
	total %= time.Minute

	s := total / time.Second
	total %= time.Second

	cs := total / (10 * time.Millisecond)

	return fmt.Sprintf("%d:%02d:%02d.%02d", h, m, s, cs), nil
}
