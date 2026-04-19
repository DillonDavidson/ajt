package main

import (
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
		err = syncCmd(os.Args[2:])
	case "shift":
		err = shiftCmd(os.Args[2:])
	case "rename":
		err = renameCmd(os.Args[2:])
	case "sub-drop":
		err = subDropCmd()
	case "audio-drop":
		err = audioDropCmd()
	case "condense":
		err = condenseCmd()
	case "impd":
		err = impd()
	case "help":
		fallthrough
	default:
		printUsage()
	}

	if err != nil {
		log.Fatal(err)
	}
}
