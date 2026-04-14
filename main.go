package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	var err error

	switch os.Args[1] {
	case "sync":
		err = sea()
	case "shift":
		if len(os.Args) < 3 {
			fmt.Println("Error: expected a subtitle file extension")
			os.Exit(1)
		}

		switch os.Args[2] {
		case "srt":
			err = shift(SRT)
		case "ass":
			err = shift(ASS)
		default:
			fmt.Printf("Error: '%s' is not a valid extension", os.Args[2])
			os.Exit(1)
		}
	case "rename":
		if len(os.Args) < 3 {
			fmt.Println("Error: expected a file extension")
			os.Exit(1)
		}

		err = rename(os.Args[2])
	case "sub-drop":
		err = subDrop()
	case "audio-drop":
		err = audioDrop()
	case "condense":
		err = condense()
	case "impd":
		err = impd()
	default:
		printUsage()
	}

	if err != nil {
		log.Fatal(err)
	}
}
