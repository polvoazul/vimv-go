package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/jwalton/go-supportscolor"
)

var color = supportscolor.Stdout().SupportsColor
var cleanup_afterwards = true

func main() {
	defer handleExit() // Needs to be on top

	files := removeEmptyLines(os.Args[1:])
	validateInput(files)
	tmpfolder, filelist := getTmpFile(files)
	defer cleanup(tmpfolder)

	// Spawn Vim to edit the temporary file
	cmd := exec.Command("vim", filelist)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("vim returned non 0. Aborting!", err)
		panic(Exit{1})
	}

	new_filenames_str, err := os.ReadFile(filelist)
	if err != nil {
		fmt.Println("Error reading temporary file:", err)
		panic(Exit{1})
	}

	new_filenames := removeEmptyLines(strings.Split(string(new_filenames_str), "\n"))
	validate(files, new_filenames)

	var to_rename []FilePair
	for i := 0; i < len(files); i++ {
		if files[i] != new_filenames[i] {
			to_rename = append(to_rename, FilePair{from: files[i], to: new_filenames[i]})
		}
	}

	report(to_rename)
	errors := rename(to_rename)

	// Finalizing
	if color {
		fmt.Printf("\033[1;32mRenamed %d files successfully.\033[0m\n", len(to_rename)-len(errors))
	} else {
		fmt.Printf("Renamed %d files successfully.\n", len(to_rename)-len(errors))
	}
	if errors != nil {
		if color {
			fmt.Printf("\033[1;31mError renaming %d files.\033[0m\n", len(errors))
		} else {
			fmt.Printf("Error renaming %d files.\n", len(errors))
		}
	}
}

type FilePair struct {
	from string
	to   string
}

func report(to_rename []FilePair) {
	if color {
		fmt.Printf("\033[1;34mTotal files to be renamed: %d\033[0m\n", len(to_rename))
	} else {
		fmt.Printf("Total files to be renamed: %d\n", len(to_rename))
	}

	for _, fp := range to_rename {
		if color {
			fmt.Printf("\033[1;34m%s\033[0m -> \033[1;33m%s\033[0m\n", fp.from, fp.to)
		} else {
			fmt.Printf("%s -> %s\n", fp.from, fp.to)
		}
	}

	if len(to_rename) == 0 {
		return
	}

	// Confirm
	fmt.Print("Press '(y)' to continue, 'n' to abort: ")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		if err.Error() != "unexpected newline" {
			fmt.Println("Error reading response:", err)
			panic(Exit{1})
		}
	}
	if strings.ToLower(response) != "y" && response != "" {
		fmt.Println("Operation aborted by user.")
		cleanup_afterwards = false
		panic(Exit{1})
	}
}

func rename(to_rename []FilePair) []error {
	var errs []error
	for _, fp := range to_rename {
		err := os.Rename(fp.from, fp.to)
		if err != nil {
			e := fmt.Errorf("error renaming %s to %s: %v", fp.from, fp.to, err)
			fmt.Println(e)
			errs = append(errs, e)
		}
	}
	return errs
}

func getTmpFile(files []string) (string, string) {
	// Create a temporary directory using os.MkdirTemp
	tmpDir, err := os.MkdirTemp("", "vimv-")
	if err != nil {
		fmt.Println("Error creating temporary directory:", err)
		panic(Exit{2})
	}
	tmpFile := filepath.Join(tmpDir, "file_list.txt")
	err = os.WriteFile(tmpFile, []byte(strings.Join(files, "\n")), 0644)
	if err != nil {
		fmt.Println("Error creating temporary file:", err)
		panic(Exit{2})
	}
	return tmpDir, tmpFile
}

func cleanup(tmpDir string) {
	if !cleanup_afterwards {
		fmt.Printf("Aborted: Leaving your edited file at: %s\n", filepath.Join(tmpDir, "file_list.txt"))
		return
	}
	err := os.RemoveAll(tmpDir)
	if err != nil {
		fmt.Println("Error removing temporary file:", err)
		panic(Exit{2})
	}
}

func validate(original []string, new []string) {
	if len(original) != len(new) {
		fmt.Println("Error: Number of original files does not match number of new files")
		panic(Exit{1})
	}

	// Check prohibited chars
	var prohibitedChars []string
	if runtime.GOOS == "windows" {
		prohibitedChars = []string{"<", ">", ":", "\"", "/", "|", "?", "*"}
	} else {
		prohibitedChars = []string{}
	}
	for _, newName := range new {
		for _, char := range prohibitedChars {
			if strings.Contains(newName, char) {
				fmt.Printf("Error: Prohibited character '%s' found in file name: %s\n", char, newName)
				panic(Exit{1})
			}
		}
	}

	// Check empty
	for _, name := range new {
		if name == "" {
			fmt.Println("Error: Empty file name is not allowed")
			panic(Exit{1})
		}
	}

	checkDuplicates(new)
}

func validateInput(files []string) {
	if len(files) == 0 {
		fmt.Println("Error: No files provided")
		panic(Exit{1})
	}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Error: File %s does not exist\n", file)
			panic(Exit{1})
		}
	}
	checkDuplicates(files)
}

func checkDuplicates(new []string) {
	duplicateMap := make(map[string]bool)
	for _, name := range new {
		if duplicateMap[name] {
			fmt.Printf("Error: Duplicate file name found: %s\n", name)
			panic(Exit{1})
		}
		duplicateMap[name] = true
	}
}

func removeEmptyLines(files []string) []string {
	var newFiles []string
	for _, file := range files {
		if file != "" {
			newFiles = append(newFiles, file)
		}
	}
	return newFiles
}

// Needed for defers to run properly. Dont use os.Exit directly, instead panic(Exit{code}). https://stackoverflow.com/a/27630092
type Exit struct{ Code int }

func handleExit() {
	if e := recover(); e != nil {
		if exit, ok := e.(Exit); ok {
			os.Exit(exit.Code)
		}
		panic(e) // not an Exit, bubble up
	}
}
