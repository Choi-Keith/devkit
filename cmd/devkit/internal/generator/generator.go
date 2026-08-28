// Package generator materializes a template directory onto disk.
package generator

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Options controls how a template is generated.
type Options struct {
	// Module, when set for a Go template, rewrites the go.mod module path
	// and all internal import paths to this value.
	Module string
	// Force allows writing into a non-empty target directory.
	Force bool
}

// Junk that should never be copied into a generated project.
var excludeDirs = map[string]bool{
	".git": true, ".claude": true, ".idea": true, ".vscode": true,
	"node_modules": true, "dist": true, "build": true,
	".venv": true, "__pycache__": true, ".pytest_cache": true, ".ruff_cache": true,
	"uploads": true,
}

var excludeFiles = map[string]bool{
	".DS_Store": true, "db.sqlite3": true,
}

// Generate copies srcDir into destDir and applies post-processing.
func Generate(srcDir, destDir string, isGo bool, opts Options) error {
	if err := checkDestDir(destDir, opts.Force); err != nil {
		return err
	}

	var goFiles []string
	err := filepath.WalkDir(srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if d.IsDir() {
			if rel != "." && excludeDirs[d.Name()] {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(destDir, rel), 0o755)
		}
		if excludeFiles[d.Name()] {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		perm := info.Mode().Perm()
		if perm == 0 {
			perm = 0o644
		}
		target := filepath.Join(destDir, rel)
		if err := os.WriteFile(target, data, perm); err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".go") {
			goFiles = append(goFiles, target)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("write template: %w", err)
	}

	if opts.Module != "" {
		if !isGo {
			return fmt.Errorf("--module is only supported for Go templates")
		}
		if err := renameModule(destDir, opts.Module, goFiles); err != nil {
			return err
		}
	}
	return nil
}

func checkDestDir(dir string, force bool) error {
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(entries) > 0 && !force {
		return fmt.Errorf("directory %s already exists and is not empty (use --force to overwrite)", dir)
	}
	return nil
}

// renameModule rewrites the module path in go.mod and all Go import paths.
func renameModule(destDir, newModule string, goFiles []string) error {
	gomod := filepath.Join(destDir, "go.mod")
	data, err := os.ReadFile(gomod)
	if err != nil {
		return fmt.Errorf("read go.mod: %w", err)
	}
	line, _, _ := strings.Cut(string(data), "\n")
	oldModule := strings.TrimSpace(strings.TrimPrefix(line, "module"))
	if oldModule == "" || oldModule == newModule {
		return nil
	}

	data = bytes.Replace(data, []byte("module "+oldModule), []byte("module "+newModule), 1)
	if err := os.WriteFile(gomod, data, 0o644); err != nil {
		return err
	}

	for _, f := range goFiles {
		data, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		// Import paths appear quoted, e.g. "old/module/internal/foo".
		if !bytes.Contains(data, []byte(`"`+oldModule+`/`)) && !bytes.Contains(data, []byte(`"`+oldModule+`"`)) {
			continue
		}
		data = bytes.ReplaceAll(data, []byte(`"`+oldModule), []byte(`"`+newModule))
		if err := os.WriteFile(f, data, 0o644); err != nil {
			return err
		}
	}
	return nil
}
