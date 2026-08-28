package generator

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, content string, perm os.FileMode) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), perm); err != nil {
		t.Fatal(err)
	}
}

func TestGenerate(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "go.mod"), "module example.com/old\n\ngo 1.24\n", 0o644)
	writeFile(t, filepath.Join(src, "main.go"), "package main\n\nimport _ \"example.com/old/internal/x\"\n", 0o644)
	writeFile(t, filepath.Join(src, "run.sh"), "#!/bin/sh\n", 0o755)
	writeFile(t, filepath.Join(src, ".venv/lib/x.py"), "junk", 0o644)
	writeFile(t, filepath.Join(src, ".DS_Store"), "junk", 0o644)

	dest := filepath.Join(t.TempDir(), "out")
	err := Generate(src, dest, true, Options{Module: "example.com/new"})
	if err != nil {
		t.Fatal(err)
	}

	// go.mod module path rewritten.
	data, err := os.ReadFile(filepath.Join(dest, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if got := string(data); got != "module example.com/new\n\ngo 1.24\n" {
		t.Errorf("go.mod = %q", got)
	}

	// Import paths rewritten.
	data, err = os.ReadFile(filepath.Join(dest, "main.go"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(data, []byte(`"example.com/new/internal/x"`)) {
		t.Errorf("main.go imports not rewritten: %q", data)
	}

	// Executable bit preserved.
	info, err := os.Stat(filepath.Join(dest, "run.sh"))
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Errorf("run.sh perm = %o, want 755", info.Mode().Perm())
	}

	// Junk excluded.
	for _, junk := range []string{".venv", ".DS_Store"} {
		if _, err := os.Stat(filepath.Join(dest, junk)); !os.IsNotExist(err) {
			t.Errorf("%s should not be copied", junk)
		}
	}
}

func TestGenerateRefusesNonEmptyDir(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a", 0o644)
	dest := t.TempDir()
	writeFile(t, filepath.Join(dest, "existing.txt"), "e", 0o644)

	if err := Generate(src, dest, false, Options{}); err == nil {
		t.Error("expected error for non-empty dest")
	}
	if err := Generate(src, dest, false, Options{Force: true}); err != nil {
		t.Errorf("--force should allow non-empty dest: %v", err)
	}
}

func TestGenerateModuleOnlyForGo(t *testing.T) {
	src := t.TempDir()
	writeFile(t, filepath.Join(src, "a.txt"), "a", 0o644)
	dest := filepath.Join(t.TempDir(), "out")
	if err := Generate(src, dest, false, Options{Module: "x"}); err == nil {
		t.Error("expected error when --module used with non-Go template")
	}
}
