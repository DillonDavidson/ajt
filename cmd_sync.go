package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var FFprobeSubsCodecArgs = []string{
	"-v", "error",
	"-select_streams", "s:0",
	"-show_entries", "stream=codec_name",
	"-of", "csv=p=0",
}

func probeSubsCodec(fileName string) (string, error) {
	args := make([]string, 0, len(FFprobeSubsCodecArgs)+1)
	args = append(args, FFprobeSubsCodecArgs...)
	args = append(args, fileName)

	out, err := exec.Command("ffprobe", args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("ffprobe subs codec failed for %q: %w\n%s",
			fileName, err, out)
	}

	return strings.TrimSpace(string(out)), nil
}

func getSubFiles(mkvFile string, extension string) (string, string, error) {
	// Eventually, I'll add some more complex logic to this,
	//   but I'm leaving this 'useless' switch statement for now.
	// ass and srt files are easy to handle, but hdmv are not.
	// I'll probably need more complex logic for that, but for now oh well

	switch extension {
	case "ass":
		extension = "ass"
	case "subrip":
		extension = "srt"
	default:
		return "", "", errors.New("not handling that subtitle type '" + extension + "'")
	}

	r := []rune(mkvFile)
	name := string(r[:len(r)-3]) + extension
	tempName := string(r[:len(r)-4]) + "eng." + extension

	return name, tempName, nil
}

func getMKVfiles() ([]string, error) {
	var mkvFiles []string

	entries, err := os.ReadDir(".")
	if err != nil {
		return mkvFiles, err
	}

	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".mkv" {
			mkvFiles = append(mkvFiles, e.Name())
		}
	}

	return mkvFiles, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}

	if errors.Is(err, os.ErrNotExist) {
		return false
	}

	return false
}

func getBaseName(name string) string {
	baseName := string(name[:len(name)-4])
	return baseName
}

func findExternalSubs(baseName string) (string, error) {
	if fileExists(baseName + ".ass") {
		return ".ass", nil
	}
	if fileExists(baseName + ".srt") {
		return ".srt", nil
	}

	return "", errors.New("external subs not detected")
}

func align(externalSubs string, baseName string) error {
	/*
		alass-cli "$output_sub" "${basename}.ass" "${basename}temp.ass"
		rm "${basename}.ass"
		mv "${basename}temp.ass" "${basename}.ass"
	*/

	extension, err := findExternalSubs(baseName)
	if err != nil {
		return err
	}

	baseNameExt := baseName + extension
	baseNameTemp := "./" + baseName + "temp" + extension

	cmd := exec.Command("alass-cli", externalSubs, baseNameExt, baseNameTemp)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("alass-cli failed: %w", err)
	}

	if err := os.Remove(baseNameExt); err != nil {
		return fmt.Errorf("failed to remove original sub: %w", err)
	}

	if err := os.Rename(baseNameTemp, baseNameExt); err != nil {
		return fmt.Errorf("failed to rename temp sub: %w", err)
	}

	return nil
}

func sea() error {
	_, err := exec.LookPath("ffprobe")
	if err != nil {
		return err
	}

	_, err = exec.LookPath("alass-cli")
	if err != nil {
		return err
	}

	mkvFiles, err := getMKVfiles()
	if err != nil {
		return errors.New("failed to read directory")
	}

	if len(mkvFiles) == 0 {
		return errors.New("no mkv files found")
	}

	for _, mkvFile := range mkvFiles {
		fmt.Printf("Extracting subtitles from %s...\n", mkvFile)

		subExtension, err := probeSubsCodec(mkvFile)
		if err != nil {
			log.Printf("Skipping %s: ffprobe failed: %v", mkvFile, err)
			continue
		}

		baseNameOrig, internalSubs, err := getSubFiles(mkvFile, subExtension)
		if err != nil {
			log.Printf("Skipping %s: %v", mkvFile, err)
			continue
		}

		baseName := getBaseName(baseNameOrig)

		if fileExists(internalSubs) {
			log.Printf("Extracted sub %s already exists, skipping extraction.", internalSubs)
			continue
		}

		cmd := exec.Command("ffmpeg", "-y", "-i", mkvFile, "-map", "0:s:0", internalSubs)
		if err := cmd.Run(); err != nil {
			log.Printf("Failed to extract subtitle for %s: %v", mkvFile, err)
			continue
		}

		if !fileExists(internalSubs) {
			log.Printf("Extraction seemingly succeeded but file '%s' does not exist", internalSubs)
			continue
		}

		fmt.Printf("Successfully extracted: %s\n", internalSubs)

		err = align(internalSubs, baseName)
		if err != nil {
			log.Printf("Alignment failed or skipped for %s: %v", mkvFile, err)
		} else {
			fmt.Printf("Successfully aligned %s\n", mkvFile)
		}

		if err := os.Remove(internalSubs); err != nil {
			log.Printf("Warning: failed to clean up %s: %v", internalSubs, err)
		}
	}

	return nil
}
