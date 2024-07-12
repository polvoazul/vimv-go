# Vimv - Renaming Files in Bulk using vim

Vimv is a command-line tool that allows you to rename multiple files at once using your favorite text editor, Vim.

Inspired on [vimv](https://github.com/thameera/vimv), but I plan to add more features.

## Features

* Supports renaming multiple files at once
* Utilizes Vim for editing file names, allowing for powerful text manipulation capabilities
* Shows colored diffs before applying changes
* Provides a user-friendly interface for confirming changes before applying them

## Usage

To use Vimv, simply run the command `vimv` followed by the list of files you want to rename. For example:
```bash
vimv file1.txt file2.txt file3.txt
```
This will open Vim with a list of the files, allowing you to edit their names as needed. Once you save and exit Vim, Vimv will apply the changes to the original files.

## Installation

### Homebrew
    Work in progres...

### Manually:
You can download it from here: https://github.com/polvoazul/vimv-go/releases, then `chmod +x` it and add to somewhere in your path (like `/usr/local/bin/`).

## Contributing

Contributions to Vimv are welcome! If you have any suggestions or bug fixes, please open an issue or submit a pull request.


## License

Vimv is licensed under the WTFPL (do whatever you want).