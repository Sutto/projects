package main

import (
	"fmt"
	"github.com/Sutto/projects/lib"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
)

var projectList *projects.FileLister

func dieWithError(e error) {
	fmt.Fprintf(os.Stderr, "An error occured running %s: %s\n", os.Args[0], e.Error())
	os.Exit(1)
}

func init() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	// Our project list will be based on the current user.

	user, err := user.Current()
	if err != nil {
		dieWithError(err)
	}

	// First, get the project cache path.
	cachedProjectsList := os.Getenv("PROJECTS_CACHE_PATH")
	if cachedProjectsList == "" {
		cachedProjectsList = filepath.Join(user.HomeDir, ".cached-projects-list")
	}

	// Now, the search directory.
	codeDirectory := os.Getenv("PROJECTS_CODE_PATH")
	if codeDirectory == "" {
		codeDirectory = filepath.Join(user.HomeDir, "Code")
	}

	// We've finally initialized the filter lister to the correct path.
	projectList = projects.NewFileLister(codeDirectory, cachedProjectsList)
}

func regenerateList() {
	err := projectList.GenerateCached()
	if err != nil {
		dieWithError(err)
	}
}

func listLiveItems() {
	list, err := projectList.LiveMatchList()
	if err != nil {
		dieWithError(err)
	}
	for _, v := range list {
		fmt.Println(v.Name)
	}
}

func listCachedItems() {
	list, err := projectList.CachedMatchList()
	if err != nil {
		dieWithError(err)
	}
	for _, v := range list {
		fmt.Println(v.Name)
	}
}

func pathForItem() {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "You must provide an extra arg for the name to %s\n", os.Args[0])
		os.Exit(1)
	}

	// Now, we need to find them.
	items, err := projectList.CachedMatchList()

	if err != nil {
		dieWithError(err)
	}

	expected := os.Args[2]

	for _, match := range items {
		if match.Name == expected {
			fmt.Println(match.FullPath)
			os.Exit(0)
		}
	}

	os.Exit(2)

}

func usage() {
	fmt.Printf("Usage: %s {regenerate|list|path} [optional path]\n", os.Args[0])
	os.Exit(1)
}

func main() {

	if len(os.Args) < 2 {
		usage()
		return
	}

	command := os.Args[1]
	switch command {
	case "regenerate":
		regenerateList()
	case "list", "ls":
		listCachedItems()
	case "live":
		listLiveItems()
	case "path":
		pathForItem()
	default:
		usage()
	}
}
