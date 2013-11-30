package projects

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type ScannerCallback func(*Match)

type DirectoryScanner struct {
	Root      string
	Processor *ConcurrentProcessor
}

type directoryEntry struct {
	path  string
	info  os.FileInfo
}

// Returns a directory scanner. This iterates over the given directory,
// invoking the callback passed to .Scan or returning an error.
func NewDirectoryScanner(root string) *DirectoryScanner {
	processor := NewConcurrentProcessor()
	return &DirectoryScanner{root, processor}
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
				scanner.Processor.AddJob(child.ScanChildrenJob(scanner, callback))
			}
		}
	}
	return nil
}

// Wraps returning a func, just so we make sure we have a correct reference to the scanner and callback.
func (e *directoryEntry) ScanChildrenJob(scanner *DirectoryScanner, callback ScannerCallback) Job {
	return func() {
		e.ScanChildren(scanner, callback)
	}
}

func (scanner *DirectoryScanner) Scan(callback ScannerCallback) error {
	info, err := os.Stat(scanner.Root)
	if err != nil {
		return err
	}

	entry := &directoryEntry{scanner.Root, info}

	processor := scanner.Processor

	processor.StartManager()
	processor.StartWorkers(8)
	processor.AddJob(entry.ScanChildrenJob(scanner, callback))
	// Now, start the workers.
	processor.WaitForCompletion()
	return nil
}
