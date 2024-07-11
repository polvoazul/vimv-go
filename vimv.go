package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func main() {
	// Create a temporary file with the list of file names
	tmpFile := "file_list.txt"
	files := os.Args[1:]
	err := ioutil.WriteFile(tmpFile, []byte(strings.Join(files, "\n")), 0644)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		return
	}

	// Spawn Vim to edit the temporary file
	cmd := exec.Command("vim", tmpFile)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Println("Error running Vim:", err)
		return
	}

	// Rename the files to their new names
	data, err := ioutil.ReadFile(tmpFile)
	if err != nil {
		fmt.Println("Error reading temporary file:", err)
		return
	}
	newFileNames := strings.Split(string(data), "\n")
	for i, newName := range newFileNames {
		if i < len(files) {
			err = os.Rename(files[i], newName)
			if err != nil {
				fmt.Printf("Error renaming %s to %s: %v\n", files[i], newName, err)
			}
		}
	}

	// Clean up the temporary file
	err = os.Remove(tmpFile)
	if err != nil {
		fmt.Println("Error removing temporary file:", err)
	}
}