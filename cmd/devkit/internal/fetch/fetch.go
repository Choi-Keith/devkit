// Package fetch downloads a template subdirectory from a git repository.
package fetch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// Template shallow-clones repoURL into a temporary directory and returns the
// path of the requested subdirectory inside the clone. The returned cleanup
// function removes the temporary clone.
//
// ref may be a branch or tag; empty means the repository's default branch.
func Template(repoURL, ref, subdir string) (dir string, cleanup func(), err error) {
	if _, err := exec.LookPath("git"); err != nil {
		return "", nil, fmt.Errorf("git is required to download templates (or use --local)")
	}

	tmp, err := os.MkdirTemp("", "devkit-clone-*")
	if err != nil {
		return "", nil, err
	}
	cleanup = func() { os.RemoveAll(tmp) }

	args := []string{"clone", "--depth", "1", "--quiet"}
	if ref != "" {
		args = append(args, "--branch", ref)
	}
	args = append(args, repoURL, tmp)
	if out, err := exec.Command("git", args...).CombinedOutput(); err != nil {
		cleanup()
		return "", nil, fmt.Errorf("git clone %s: %w\n%s", repoURL, err, out)
	}

	dir = filepath.Join(tmp, filepath.FromSlash(subdir))
	st, err := os.Stat(dir)
	if err != nil || !st.IsDir() {
		cleanup()
		return "", nil, fmt.Errorf("template directory %q not found in %s", subdir, repoURL)
	}
	return dir, cleanup, nil
}
