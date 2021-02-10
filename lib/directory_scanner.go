package projects

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type ScannerCallback func(*Match)

type DirectoryScanner struct {
	Root string
}

type directoryEntry struct {
	path  string
	info  os.FileInfo
	depth uint
}

const MaxProcessingDepth = uint(3)

// Returns a directory scanner. This iterates over the given directory,
// invoking the callback passed to .Scan or returning an error.
func NewDirectoryScanner(root string) *DirectoryScanner {
	return &DirectoryScanner{root}
}

func (e *directoryEntry) ToMatch() *Match {
	return &Match{filepath.Base(e.path), e.path}
}

func directoryScanProcessor(input chan *directoryEntry, output chan *Match, scanner *DirectoryScanner, scanWG *sync.WaitGroup, processWG *sync.WaitGroup) {
	for {
		e := <-input
		entries, err := ioutil.ReadDir(e.path)

		if err != nil {
			scanWG.Done()
			continue
		}

		for _, entry := range entries {
			if entry.Name() == ".DS_Store" {
				continue
			}
			path := filepath.Join(e.path, entry.Name())
			pathToCheck := filepath.Join(path, ".git")
			child := &directoryEntry{path, entry, e.depth + 1}
			if _, err := os.Stat(pathToCheck); err == nil {
				processWG.Add(1)
				output <- child.ToMatch()
			} else if e.depth < MaxProcessingDepth {
				scanWG.Add(1)
				input <- child
			}
		}

		scanWG.Done()
	}
}

func processResults(input chan *Match, callback ScannerCallback, wg *sync.WaitGroup) {
	for {
		result := <-input
		callback(result)
		wg.Done()
	}
}

func (scanner *DirectoryScanner) Scan(callback ScannerCallback) error {
	concurrency := runtime.NumCPU()
	info, err := os.Stat(scanner.Root)
	if err != nil {
		return err
	}

	entries := make(chan *directoryEntry, concurrency*100)
	matches := make(chan *Match, concurrency*100)

	var scanWG sync.WaitGroup
	var processWG sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		go directoryScanProcessor(entries, matches, scanner, &scanWG, &processWG)
	}
	go processResults(matches, callback, &processWG)

	entry := &directoryEntry{scanner.Root, info, 1}

	scanWG.Add(1)
	entries <- entry

	scanWG.Wait()
	processWG.Wait()

	return err
}
