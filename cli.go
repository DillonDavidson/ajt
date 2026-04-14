package main

import (
	"fmt"
	"os"
)

func printUsage() {
	fmt.Println("ajt - Another Japanese Tool")
	fmt.Println()
	fmt.Printf("Usage: %s [OPTIONS]\n\n\n", os.Args[0])

	fmt.Println("  audio-drop")
	fmt.Println("        for all mkv files, view the audio tracks and choose one to keep")
	fmt.Println("  condense")
	fmt.Println("        for all mkv files, condense into ogg files")
	fmt.Println("  impd")
	fmt.Println("        use impd to turn all mkv files into ogg files for passive immersion")
	fmt.Println("  rename <file_extension>")
	fmt.Println("        rename all <file_extension> files numerically")
	fmt.Println("  shift <subtitle_file>")
	fmt.Println("        shift all of <subtitle_file>'s times")
	fmt.Println("  sub-drop")
	fmt.Println("        for all mkv files, view the subtitle tracks and choose one to keep")
	fmt.Println("  sync")
	fmt.Println("        for all mkv files with matching srt/ass file names, sync the external subtitles with the internal")
}
