package projects

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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
			// fmt.Printf("Scanned: %v", e)
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
				fmt.Printf("Queued: %v\n", child.ToMatch())
				output <- child.ToMatch()
			} else if e.depth < MaxProcessingDepth {
				scanWG.Add(1)
				fmt.Printf("Scanning: %v\n", child)
				input <- child
			}
		}

		fmt.Printf("Scanned: %v\n", e)
		scanWG.Done()
	}
}

func processResults(input chan *Match, callback ScannerCallback, wg *sync.WaitGroup) {
	for {
		result := <-input
		// fmt.Printf("Processed: %v - wg: %v\n", result, wg)
		callback(result)
		// fmt.Printf("Called the callback: %v\n", result)
		wg.Done()
	}
}

func (scanner *DirectoryScanner) Scan(callback ScannerCallback) error {
	info, err := os.Stat(scanner.Root)
	channel := make(chan *directoryEntry)
	result := make(chan *Match)

	var scanWG sync.WaitGroup
	var processWG sync.WaitGroup

	for i := 0; i < 8; i++ {
		go directoryScanProcessor(channel, result, scanner, &scanWG, &processWG)
	}

	// go processResults(result, callback, &processWG)

	if err != nil {
		return err
	}

	entry := &directoryEntry{scanner.Root, info, 1}
	scanWG.Add(1)
	channel <- entry

	for {
		item := <-result
		callback(item)
	}

	fmt.Printf("Now done scanning")
	scanWG.Wait()

	return err
}
