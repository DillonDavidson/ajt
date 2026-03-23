package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type SubtitleType int

const (
	SRT SubtitleType = iota
	ASS
)

func enumToString(subType SubtitleType) string {
	switch subType {
	case ASS:
		return ".ass"
	case SRT:
		return ".srt"
	default:
		return ""
	}
}

func processFile(name string, subType SubtitleType, delta float32) error {
	in, err := os.Open(name)
	if err != nil {
		return err
	}
	defer in.Close()

	tempName := name + ".temp"

	out, err := os.Create(tempName)
	if err != nil {
		return err
	}
	defer out.Close()

	if err := shiftLines(in, out, subType, delta); err != nil {
		return err
	}

	if err := os.Remove(name); err != nil {
		return err
	}

	return os.Rename(tempName, name)
}

func shiftLines(in *os.File, out *os.File, subType SubtitleType, delta float32) error {
	scanner := bufio.NewScanner(in)
	writer := bufio.NewWriter(out)

	for scanner.Scan() {
		line := strings.ReplaceAll(scanner.Text(), "\r", "")

		var err error

		switch subType {
		case ASS:
			err = processASSLine(line, writer, delta)
		case SRT:
			err = processSRTLine(line, writer, delta)
		default:
			return errors.New("unsupported extension")
		}

		if err != nil {
			return err
		}
	}

	writer.Flush()
	return scanner.Err()
}

func readShiftAmount() (float32, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter the number of seconds to shift: ")

	text, err := reader.ReadString('\n')
	if err != nil {
		return 0, err
	}

	text = strings.TrimSpace(text)

	val, err := strconv.ParseFloat(text, 32)
	if err != nil {
		return 0, err
	}

	return float32(val), nil
}

func shift(subType SubtitleType) error {
	deltaSeconds, err := readShiftAmount()
	if err != nil {
		return err
	}

	files, err := findFiles(enumToString(subType))
	if err != nil {
		return err
	}

	for _, file := range files {
		if err := processFile(file, subType, deltaSeconds); err != nil {
			return err
		}
	}

	return nil
}
