package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Printf(
		`ajt - Another Japanese Tool

Usage: %s [OPTIONS]

Options:
	audio-drop
		for all mkv files, view the audio tracks and choose one to keep
	condense
		for all mkv files, condense into ogg files
	impd
		use impd to turn all mkv files into ogg files for passive immersion
	rename [-dry-run] <file_extension...>
		rename all <file_extension...> files numerically by extension type
	shift <subtitle_file>
		shift all of <subtitle_file>'s times
	sub-drop
		for all mkv files, view the subtitle tracks and choose one to keep
	sync [-dry-run] [-verbose] [-recursive]
		for all mkv files with matching srt/ass file names, sync the external subtitles with the internal
	help
		print this message
`, os.Args[0])
}
