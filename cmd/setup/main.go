package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/go-git/go-git/v5"
)

func setup() {

	root := os.Getenv("TRAILPKG_ROOT")
	if root == "" {
		fmt.Println("error: couldn't find TRAILPKG_ROOT env var. Please set it in your shell config")
		return
	}

	if _, err := os.Stat(path.Join(root, "state.json")); err == nil {
		fmt.Println("trailpkg has already been set up")
		return
	}

	localDirs  := []string{"bin", "etc", "build", "store"}

	for _, d := range localDirs {
		if err := os.MkdirAll(path.Join(root, d), 0755); err != nil {
			fmt.Printf("error: codln't create %s: %v\n", d, err)
		}
	}

	stateFile := "state.json"


	file, err := os.Create(path.Join(root, stateFile)) 
	if err != nil {
		fmt.Println("error: couldn't create the state.json file", err)
		return
	}

	json.NewEncoder(file).Encode(map[string]any{
		"installed":[]any{},
		"last_manifest_update": time.Now(),
	})

	file.Close()

	repoURL := "https://github.com/myselfBZ/trailpkg-manifest.git"
	localPath := path.Join(root, "manifest")


	_, err = git.PlainClone(localPath, false, &git.CloneOptions{
		URL:      repoURL,
	})

	if err != nil {
		fmt.Printf("Clone failed: %v\n", err)
		return
	}

	fmt.Println("Set up complete!")
}

func main() {
	setup()
}
