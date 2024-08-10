package main

import (
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type file struct {
	filePath string
	fetching bool
	size     int64 // -1 represents that the size has not yet been fetched.
}

func FetchSize(f *file) {
	fi, err := os.Stat(f.filePath)
	if err != nil {
		log.Fatal(err)
	}

	// Artificially slow down the program to simulate a slow operation and get a visual effect in the TUI.
	time.Sleep(750 * time.Millisecond)

	f.size = fi.Size()
	f.fetching = false
	log.Println("Fetched size for", f.filePath)
}

func prettyPrintBytes(bytes int64) string {
	const (
		KiB = 1024
		MiB = 1024 * KiB
		GiB = 1024 * MiB
	)

	switch {
	case bytes < KiB:
		return fmt.Sprintf("%d B", bytes)
	case bytes < MiB:
		return fmt.Sprintf("%.1f KiB", float64(bytes)/float64(KiB))
	case bytes < GiB:
		return fmt.Sprintf("%.1f MiB", float64(bytes)/float64(MiB))
	default:
		return fmt.Sprintf("%.1f GiB", float64(bytes)/float64(GiB))
	}
}

func listGitProjectFiles() ([]file, error) {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	repo, err := git.PlainOpen(currentWorkingDir)

	if err != nil {
		log.Fatal(err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		log.Fatal(err)
	}

	// I think this finds the exclude patterns in the .gitignore file in the Git repository in this directory.
	patterns, err := gitignore.ReadPatterns(worktree.Filesystem, nil)
	if err != nil {
		log.Fatal(err)
	}

	// I think 'worktree.Excludes' are the ignore patterns in maybe the home directory's .gitignore file. Not really
	//sure.
	patterns = append(patterns, worktree.Excludes...)

	m := gitignore.NewMatcher(patterns)

	var files []file

	err = filepath.WalkDir(".", func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		isDir := info.IsDir()

		// Split the path into its components. For example, if the path is "hello/README.md", the components will be
		// ["hello", "README.md"].
		pathComponents := strings.Split(filepath.Clean(path), string(filepath.Separator))

		ignored := m.Match(pathComponents, isDir)
		if err != nil {
			return err
		}

		if isDir {
			// If the directory is ignored then, we can speed up the file walking process by skipping the directory.
			if ignored {
				return filepath.SkipDir
			}

			if path == ".git" {
				return filepath.SkipDir
			}

			// We don't want to list directories, only files.
			return nil
		}

		if ignored {
			return nil
		}

		files = append(files, file{filePath: path, size: -1})
		return nil
	})

	return files, err
}
