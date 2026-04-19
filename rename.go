package main

import (
	"flag"
	"fmt"
	"os"
)

func renameCmd(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("expected a file extension")
	}

	fs := flag.NewFlagSet("rename", flag.ExitOnError)

	dryRun := fs.Bool("dry-run", false, "preview changes without modifying files")

	fs.Parse(args)
	extensions := fs.Args()

	if len(extensions) < 1 {
		return fmt.Errorf("expected directory or file(s)")
	}

	for _, extension := range extensions {
		files, err := findFiles(extension)
		if err != nil {
			return err
		}

		for count, file := range files {
			newName := fmt.Sprintf("%02d.%s", count+1, extension)
			if *dryRun {
				fmt.Println("mv", file, newName)
			} else {
				os.Rename(file, newName)
			}
		}
	}

	return nil
}
