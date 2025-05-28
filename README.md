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

### Manually:
You can download it from here: https://github.com/polvoazul/vimv-go/releases, download it, make it executable and place it somewhere like /usr/local/bin. Alternatively run this command that does exactly this:

```bash
# Detect OS and architecture
OS=$(uname -s)
ARCH=$(uname -m)

# Map OS and architecture to download URL format
case $OS in "Darwin") OS_NAME="Darwin" ;; "Linux") OS_NAME="Linux" ;; "MINGW"*|"MSYS"*|"CYGWIN"*) OS_NAME="Windows" ;; *) echo "Unsupported OS: $OS" ; exit 1 ;; esac
case $ARCH in "x86_64") ARCH_NAME="x86_64" ;; "arm64"|"aarch64") ARCH_NAME="arm64" ;; "i386"|"i686") ARCH_NAME="i386" ;; *) echo "Unsupported architecture: $ARCH" ; exit 1 ;; esac

# Get latest version
VERSION=$(curl -s https://api.github.com/repos/polvoazul/vimv-go/releases/latest | grep -o '"tag_name": "v[^"]*"' | cut -d'"' -f4)

# Download and install
if [ "$OS_NAME" = "Windows" ]; then
    echo Please do it manually by going to: "https://github.com/polvoazul/vimv-go/releases/download/${VERSION}/vimv-go_${OS_NAME}_${ARCH_NAME}.zip"
    exit 1
fi
(
set -e
cd /tmp/
curl -L "https://github.com/polvoazul/vimv-go/releases/download/${VERSION}/vimv-go_${OS_NAME}_${ARCH_NAME}.tar.gz" -o /tmp/vimv.tar.gz
tar xzf vimv.tar.gz
chmod +x vimv-go
sudo mv vimv-go /usr/local/bin/vimv
rm vimv.tar.gz
echo Installed successfully!
)
```

## Contributing

Contributions to Vimv are welcome! If you have any suggestions or bug fixes, please open an issue or submit a pull request.


## License

Vimv is licensed under the WTFPL (do whatever you want).