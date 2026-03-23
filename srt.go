package main

import (
	"bufio"
	"fmt"
	"strings"
	"time"
)

func processSRTLine(line string, writer *bufio.Writer, delta float32) error {
	if !strings.Contains(line, "-->") {
		fmt.Fprintln(writer, line)
		return nil
	}

	parts := strings.Split(line, " --> ")

	newStart, err := shiftTime(parts[0], delta, SRT)
	if err != nil {
		return err
	}

	newEnd, err := shiftTime(parts[1], delta, SRT)
	if err != nil {
		return err
	}

	fmt.Fprintf(writer, "%s --> %s\n", newStart, newEnd)

	return nil
}

func shiftSRTTime(timestamp string, deltaSeconds float32) (string, error) {
	layout := "15:04:05,000" // idk, vibe coded this part
	t, err := time.Parse(layout, timestamp)
	if err != nil {
		return "", err
	}

	delta := time.Duration(deltaSeconds * float32(time.Second))
	newTime := t.Add(delta)

	return newTime.Format(layout), nil
}
