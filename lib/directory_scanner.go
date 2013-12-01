package projects

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type ScannerCallback func(*Match)

type DirectoryScanner struct {
	Root      string
}

type directoryEntry struct {
	path string
	info os.FileInfo
}

// Returns a directory scanner. This iterates over the given directory,
// invoking the callback passed to .Scan or returning an error.
func NewDirectoryScanner(root string) *DirectoryScanner {
	return &DirectoryScanner{root}
}

func (e *directoryEntry) ToMatch() *Match {
	return &Match{filepath.Base(e.path), e.path}
}

func (e *directoryEntry) ScanChildren(scanner *DirectoryScanner, callback ScannerCallback) error {
	entries, err := ioutil.ReadDir(e.path)
	if err != nil {
		// When the directory we wish to scan has issues, bail early.
		return err
	}

	for _, entry := range entries {
		// We bail on non directories, so we can shortcut the logic here of listing / scanning directories.
		if entry.IsDir() {
			path := filepath.Join(e.path, entry.Name())
			pathToCheck := filepath.Join(path, ".git")
			child := &directoryEntry{path, entry}
			if _, err := os.Stat(pathToCheck); err == nil {
				callback(child.ToMatch())
			} else {
				if err := child.ScanChildren(scanner, callback); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (scanner *DirectoryScanner) Scan(callback ScannerCallback) error {
	info, err := os.Stat(scanner.Root)
	if err != nil {
		return err
	}

	entry := &directoryEntry{scanner.Root, info}
	err 	= entry.ScanChildren(scanner, callback)
	return err
}
