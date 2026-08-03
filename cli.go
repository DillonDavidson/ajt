package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Printf(
		`ajt - Utilities for managing video and subtitles files for AJATTing

Usage:
	%s <command> [options]

Options:
	audio-drop
		Interactively keep audio one track in each MKV
	condense
		Condense all mkv files into ogg files
	rename 
		Rename all files numerically by extension type
	shift 
		Shift subtitle timestamps
	sub-drop
		Interactively keep one subtitle track in each MKV
	sync 
		Synchronize external subtitles with internal subtitles in each MKV
	help
		Show this message
	version
		Display version information
`, os.Args[0],
	)
}
