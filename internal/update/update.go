package update

import (
	"errors"
	"fmt"

	"github.com/go-git/go-git/v5"
)

func ManifestUpdate(path string) error {
	fmt.Println("Checking for updates in the manifest...")
	repo, err := git.PlainOpen(path)
	if err != nil {
		return err
	}

	worktree, _ := repo.Worktree()

	err = worktree.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		fmt.Println("No updates available")
		return nil
	} else if err != nil {
		return err
	} 

	return nil
}
