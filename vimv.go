package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/andreyvit/diff"
	"github.com/fatih/color"
	"github.com/jwalton/go-supportscolor"
)

var cleanup_afterwards = true

func main() {
	defer handleExit() // Needs to be on top

	// Flags
	editor := flag.String("editor", "vim", "Which editor to use?")
	flag.Parse()
	files := flag.Args()

	color.NoColor = !supportscolor.Stdout().SupportsColor

	files = removeEmptyLines(files)
	files = validateInput(files)
	tmpfolder, filelist := writeTmpFile(files)
	defer cleanup(tmpfolder)

	// Spawn Vim to edit the temporary file
	cmd := exec.Command(*editor, filelist)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println("Editor returned non 0. Aborting!", err)
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
	assert_no_conflicts(to_rename)

	report(to_rename)
	errors := rename(to_rename)

	// Finalizing
	color.Green("Renamed %d files successfully.", len(to_rename)-len(errors))
	if len(errors) > 0 {
		color.Red("Error renaming %d files.", len(errors))
	}
}

func assert_no_conflicts(to_rename []FilePair) {
	targetNames := make(map[string]bool)

	for _, pair := range to_rename {
		if pair.to == pair.from {
			continue
		}

		// Check if the target name already exists
		if _, err := os.Stat(pair.to); err == nil {
			die("Error: Destination file '%s' already exists.\n", pair.to)
		} else if !os.IsNotExist(err) {
			// some other error occurred while checking the file
			die("Error: Could not stat destination file '%s': %v\n", pair.to, err)
		}

		// Check for duplicate target names
		if targetNames[pair.to] {
			die("Error: Multiple files are being renamed to '%s', which is unsupported.\n", pair.to)
		}
		targetNames[pair.to] = true
	}
}

type FilePair struct {
	from string
	to   string
}

func report(to_rename []FilePair) {
	color.Cyan("Total files to be renamed: %d", len(to_rename))
	if len(to_rename) == 0 {
		return
	}

	for _, fp := range to_rename {
		fmt.Printf("%s -> %s\n", color.CyanString(fp.from), color.YellowString(fp.to))
	}

	for {
		switch prompt_user() {
		case ResponseYes:
			return
		case ResponseNo:
			die("Operation aborted by user.")
		case ResponseDiff:
			show_diff(to_rename)
		}
	}
}

func die(s string, a ...any) {
	fmt.Printf(s, a...)
	cleanup_afterwards = false
	panic(Exit{1})
}

type UserResponse int

const (
	ResponseYes UserResponse = iota
	ResponseNo
	ResponseDiff
)

func prompt_user() UserResponse {
	// Confirm
	fmt.Print("Press '(y)' to continue, 'd' to show name diff, 'n' to abort:")
	var response string
	_, err := fmt.Scanln(&response)
	if err != nil {
		if err.Error() != "unexpected newline" {
			fmt.Println("Error reading response:", err)
			panic(Exit{1})
		}
	}
	if strings.ToLower(response) == "y" || response == "" {
		return ResponseYes
	}
	if strings.ToLower(response) == "d" {
		return ResponseDiff
	}
	// else
	return ResponseNo
}

func show_diff(pairs []FilePair) {
	type Segment struct {
		text  string
		type_ string
	}
	// Create a list of ordered Segments. Types are +(added) ~(deleted) or =(common)
	for _, fp := range pairs {
		the_diff := diff.CharacterDiff(fp.from, fp.to)
		var segments []Segment
		re := regexp.MustCompile(`\(?(\+\+|~~)\)?`)
		delimiters := re.FindAllStringIndex(the_diff, -1)
		idx := 0
		inside_delimiter := false
		for _, delimiter := range delimiters {
			var type_ string
			if inside_delimiter {
				type_ = string(the_diff[delimiter[0]+1]) // +1 to avoid brackets
			} else {
				type_ = "="
			}
			segments = append(segments, Segment{the_diff[idx:delimiter[0]], type_})
			idx = delimiter[1]
			inside_delimiter = !inside_delimiter
		}
		segments = append(segments, Segment{the_diff[idx:], "="})
		// Print Original Filename
		fmt.Print(color.CyanString("From: "))
		for _, seg := range segments {
			switch seg.type_ {
			case "+":
				// noop
			case "~":
				fmt.Print(color.RedString(seg.text))
			default:
				fmt.Print(color.WhiteString(seg.text))
			}
		}
		fmt.Println()
		// Print New Filename
		fmt.Print(color.YellowString("To:   "))
		for _, seg := range segments {
			switch seg.type_ {
			case "+":
				fmt.Print(color.GreenString(seg.text))
			case "~":
				// noop
			default:
				fmt.Print(color.WhiteString(seg.text))
			}
		}
		fmt.Println()
	}
}

func rename(to_rename []FilePair) []error {
	var errs []error
	for _, fp := range to_rename {
		if _, err := os.Stat(fp.to); err == nil {
			e := fmt.Errorf("destination file %s already exists", fp.to)
			fmt.Println(e)
			errs = append(errs, e)
			continue
		}
		err := os.Rename(fp.from, fp.to)
		if err != nil {
			e := fmt.Errorf("error renaming %s to %s: %v", fp.from, fp.to, err)
			fmt.Println(e)
			errs = append(errs, e)
		}
	}
	return errs
}

func writeTmpFile(files []string) (string, string) {
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

func validateInput(files []string) []string {
	if len(files) == 0 {
		var err error
		files, err = filepath.Glob("./*")
		if err != nil {
			fmt.Println("Error listing files:", err)
			panic(Exit{1})
		}
	}
	for _, file := range files {
		if _, err := os.Stat(file); os.IsNotExist(err) {
			fmt.Printf("Error: File %s does not exist\n", file)
			panic(Exit{1})
		}
	}
	checkDuplicates(files)
	return files
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
