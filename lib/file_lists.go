package projects

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
)

type FileLister struct {
	RootDirectory string
	CachePath     string
	MaxDepth      int
}

func NewFileLister(root, cache string, maxDepth int) *FileLister {
	return &FileLister{root, cache, maxDepth}
}

type Match struct {
	Name     string
	FullPath string
}
type MatchList [](*Match)

func MatchFromLine(line string) *Match {
	parts := strings.SplitN(strings.TrimSpace(line), "=", 2)
	if len(parts) != 2 {
		panic("Expected 2 items, got less.")
	}
	return &Match{parts[0], parts[1]}
}

func (m *Match) ToLine() string {
	return fmt.Sprintf("%s=%s\n", m.Name, m.FullPath)
}

func (lister *FileLister) LiveMatchList() (MatchList, error) {
	items := make(MatchList, 0)
	// We run the scanner over children.
	scanner := NewDirectoryScanner(lister.RootDirectory, lister.MaxDepth)

	lock := new(sync.Mutex)

	error := scanner.Scan(func(m *Match) {
		lock.Lock()
		items = append(items, m)
		lock.Unlock()
	})

	if error != nil {
		return MatchList{}, error
	}

	return items, nil
}

func (lister *FileLister) CachedMatchList() (MatchList, error) {

	items := MatchList{}

	file, err := os.Open(lister.CachePath)
	if err != nil {

		if os.IsNotExist(err) {
			// When not exist, we just return an empty list.
			return items, nil
		} else {
			// Otherwise, we return the error.
			return items, err
		}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		// Append to the list of matches.,
		items = append(items, MatchFromLine(scanner.Text()))
	}
	// Now, ensure the scanner was valid, otherwise return nothing.
	if err := scanner.Err(); err != nil {
		return MatchList{}, err
	} else {
		return items, nil
	}
}

// Given a list of matches, writes it to the cached file list.
func (lister *FileLister) WriteCachedList(list MatchList) error {
	file, err := os.Create(lister.CachePath)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, match := range list {
		file.WriteString(match.ToLine())
	}
	return nil
}