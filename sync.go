package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func syncCmd(args []string) error {
	fs := flag.NewFlagSet("sync", flag.ExitOnError)

	dryRun := fs.Bool("dry-run", false, "preview changes without modifying files")
	verbose := fs.Bool("verbose", false, "verbose output")
	recursive := fs.Bool("recursive", false, "recurse into subdirectories")

	fs.Parse(args)
	paths := fs.Args()

	if len(paths) < 1 {
		return fmt.Errorf("expected directory or file(s)")
	}

	mkvs := findAllMkvs(paths, *recursive)

	return synchronize(mkvs, *dryRun, *verbose)
}

func findAllMkvs(paths []string, recurse bool) []string {
	if len(paths) == 0 {
		return nil
	}

	var mkvs []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			continue
		}

		if info.IsDir() {
			entries, err := os.ReadDir(path)
			if err != nil {
				continue
			}

			var next []string
			for _, entry := range entries {
				if entry.IsDir() && !recurse {
					continue
				}
				next = append(next, filepath.Join(path, entry.Name()))

			}

			mkvs = append(mkvs, findAllMkvs(next, recurse)...)
			continue
		}

		if strings.HasSuffix(path, ".mkv") {
			mkvs = append(mkvs, path)
		}
	}

	return mkvs
}

func synchronize(files []string, dryRun, verbose bool) error {
	if !dryRun {
		_, err := exec.LookPath("ffprobe")
		if err != nil {
			return err
		}

		_, err = exec.LookPath("alass-cli")
		if err != nil {
			return err
		}
	}

	if len(files) == 0 {
		return errors.New("no mkv files found")
	}

	for _, file := range files {
		if err := processFile(file, dryRun, verbose); err != nil {
			log.Printf("skipping %s: %v", file, err)
		}
	}

	return nil
}

func processFile(file string, dryRun, verbose bool) error {
	fmt.Printf("extracting subtitles from %s...\n", file)

	subExtension, err := probeSubsCodec(file, dryRun)
	if err != nil {
		return fmt.Errorf("skipping %s: ffprobe failed: %v", file, err)
	}

	baseNameOrig, internalSubs, err := getSubFiles(file, subExtension)
	if err != nil {
		return err
	}

	baseName := string(baseNameOrig[:len(baseNameOrig)-4])

	if err := extractSubs(file, internalSubs, dryRun); err != nil {
		return err
	}

	if err := align(internalSubs, baseName, dryRun, verbose); err != nil {
		return fmt.Errorf("alignment failed or skipped for %s: %v", file, err)
	}

	fmt.Printf("successfully aligned %s\n", file)

	if !dryRun {
		if err := os.Remove(internalSubs); err != nil {
			log.Printf("warning: failed to clean up %s: %v", internalSubs, err)
		}
	}

	return nil
}

func extractSubs(file, internalSubs string, dryRun bool) error {
	if fileExists(internalSubs) && !dryRun {
		log.Printf("extracted sub %s already exists, skipping extraction.", internalSubs)
		return nil
	}

	if dryRun {
		fmt.Printf("ffmpeg -y -i %s -map 0:s:0 %s\n", file, internalSubs)
		return nil
	}

	cmd := exec.Command("ffmpeg", "-y", "-i", file, "-map", "0:s:0", internalSubs)
	if err := cmd.Run(); err != nil {
		log.Printf("failed to extract subtitle for %s: %v", file, err)
		return nil
	}

	if !fileExists(internalSubs) {
		log.Printf("extraction seemingly succeeded but file '%s' does not exist", internalSubs)
		return nil
	}

	fmt.Printf("successfully extracted: %s\n", internalSubs)
	return nil
}

func align(externalSubs string, baseName string, dryRun, verbose bool) error {
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

	if dryRun {
		fmt.Println("alass-cli", externalSubs, baseNameExt, baseNameTemp)
		fmt.Println("rm", baseNameExt)
		fmt.Println("mv", baseNameTemp, baseNameExt)
		return nil
	}

	cmd := exec.Command("alass-cli", externalSubs, baseNameExt, baseNameTemp)
	if verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
	}

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

// mostly helper functions and stuff

var FFprobeSubsCodecArgs = []string{
	"-v", "error",
	"-select_streams", "s:0",
	"-show_entries", "stream=codec_name",
	"-of", "csv=p=0",
}

func probeSubsCodec(fileName string, dryRun bool) (string, error) {
	args := make([]string, 0, len(FFprobeSubsCodecArgs)+1)
	args = append(args, FFprobeSubsCodecArgs...)
	args = append(args, fileName)

	if dryRun {
		fmt.Printf("ffprobe ")
		for _, arg := range FFprobeSubsCodecArgs {
			fmt.Printf(" %s", arg)
		}
		fmt.Printf("\n")
		return "subrip", nil
	}

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
		break
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

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
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
