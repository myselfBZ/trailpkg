package update

import (
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CheckForUpdateInManifest returns an array of packages that have a new version available
func CheckForUpdateInManifest(path string) ([]string, error) {
	files, err := getChangedFiles(path)

	if err != nil {
		return nil, err
	}

	var result []string

	for _, f := range files {
		if f != "README.md" {
			result = append(result, f)
		}
	}

	return result, nil
}

func getChangedFiles(path string) ([]string, error) {
	repo, _ := git.PlainOpen(path)

	_ = repo.Fetch(&git.FetchOptions{
		RemoteName: "origin",
	})

	ref, _ := repo.Head()
	localCommit, _ := repo.CommitObject(ref.Hash())

	remoteRef, _ := repo.Reference(plumbing.ReferenceName("refs/remotes/origin/main"), true)
	remoteCommit, _ := repo.CommitObject(remoteRef.Hash())

	localTree, _ := localCommit.Tree()
	remoteTree, _ := remoteCommit.Tree()

	changes, _ := object.DiffTree(localTree, remoteTree)

	result := make([]string, len(changes))

	for i, change := range changes {
		name := ""
		
		if change.From.Name != "" {
			name = change.From.Name
		} else {
			name = change.To.Name
		}

		result[i] = name
	}

	return result, nil
}
