package projects

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type ScannerCallback func(*Match)

type DirectoryScanner struct {
	Root      string
	MaxDepth  int
	Processor *ConcurrentProcessor
}

type directoryEntry struct {
	path  string
	depth int
	info  os.FileInfo
}

// Returns a directory scanner. This iterates over the given directory, up to a specified depth,
// invoking the callback passed to .Scan or returning an error.
func NewDirectoryScanner(root string, maxDepth int) *DirectoryScanner {
	processor := NewConcurrentProcessor()
	return &DirectoryScanner{root, maxDepth, processor}
}

func (e *directoryEntry) shouldRecurseInto(scanner *DirectoryScanner) bool {
	if e.info.IsDir() && e.depth >= scanner.MaxDepth {
		return false
	} else {
		return true
	}
}

func (e *directoryEntry) isGitDirectory() bool {
	return e.info.IsDir() && filepath.Base(e.path) == ".git"
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

	childDepth := e.depth + 1

	// Otherwise, iterate over the entries.
	for _, entry := range entries {
		// We bail on non directories, so we can shortcut the logic here of listing / scanning directories.
		if entry.IsDir() {
			path := filepath.Join(e.path, entry.Name())
			pathToCheck := filepath.Join(path, ".git")
			child := &directoryEntry{path, childDepth, entry}
			if _, err := os.Stat(pathToCheck); err == nil {
				callback(child.ToMatch())
			} else {
				// Here we recurse. This should likely be pushed onto a channel as something we can process using a queue.
				scanner.Processor.AddJob(child.ScanChildrenJob(scanner, callback))
			}
		}
	}
	return nil
}

func (e *directoryEntry) ScanChildrenJob(scanner *DirectoryScanner, callback ScannerCallback) Job {
	return func() {
		e.ScanChildren(scanner, callback)
	}
}

func (scanner *DirectoryScanner) Scan(callback ScannerCallback) error {
	// We recursively invoke the scan function under a given depth.
	info, err := os.Stat(scanner.Root)

	if err != nil {
		return err
	}

	entry := &directoryEntry{scanner.Root, 0, info}

	processor := scanner.Processor

	processor.StartManager()
	processor.StartWorkers(8)
	processor.AddJob(entry.ScanChildrenJob(scanner, callback))
	// Now, start the workers.
	processor.WaitForCompletion()
	return nil
}
