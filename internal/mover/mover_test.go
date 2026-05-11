package mover

import (
	"os"
	"path/filepath"
	"testing"
)

func TestMoveDirectory_Success(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "source")
	dst := filepath.Join(base, "dest", "source")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	// Put a file inside so we can verify it moved
	if err := os.WriteFile(filepath.Join(src, "file.txt"), []byte("data"), 0644); err != nil {
		t.Fatal(err)
	}

	m := New(false)
	if err := m.MoveDirectory(src, dst); err != nil {
		t.Fatalf("MoveDirectory error: %v", err)
	}

	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Errorf("source should no longer exist after move")
	}
	if _, err := os.Stat(filepath.Join(dst, "file.txt")); err != nil {
		t.Errorf("file should exist at destination: %v", err)
	}
}

func TestMoveDirectory_DryRun(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "source")
	dst := filepath.Join(base, "dest", "source")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}

	m := New(true)
	if err := m.MoveDirectory(src, dst); err != nil {
		t.Fatalf("MoveDirectory dry-run error: %v", err)
	}

	if _, err := os.Stat(src); err != nil {
		t.Errorf("dry-run should not move source: %v", err)
	}
	if _, err := os.Stat(dst); !os.IsNotExist(err) {
		t.Errorf("dry-run should not create destination")
	}
}

func TestMoveDirectory_SourceNotExist(t *testing.T) {
	m := New(false)
	err := m.MoveDirectory("/nonexistent/path", "/also/nonexistent")
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

func TestMoveDirectory_SourceIsFile(t *testing.T) {
	base := t.TempDir()
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	m := New(false)
	err := m.MoveDirectory(file, filepath.Join(base, "dest"))
	if err == nil {
		t.Error("expected error when source is a file, got nil")
	}
}

func TestValidateMoveOperation_Success(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "source")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(base, "zombies", "source")

	m := New(false)
	if err := m.ValidateMoveOperation(src, dst); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}
}

func TestValidateMoveOperation_DestinationAlreadyExists(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "source")
	dst := filepath.Join(base, "dest")
	for _, d := range []string{src, dst} {
		if err := os.MkdirAll(d, 0755); err != nil {
			t.Fatal(err)
		}
	}

	m := New(false)
	err := m.ValidateMoveOperation(src, dst)
	if err == nil {
		t.Error("expected error when destination already exists, got nil")
	}
}

func TestValidateMoveOperation_SourceNotExist(t *testing.T) {
	m := New(false)
	err := m.ValidateMoveOperation("/nonexistent/src", "/some/dst")
	if err == nil {
		t.Error("expected error for missing source, got nil")
	}
}

func TestCleanupEmptyDirs_RemovesEmptySubdirs(t *testing.T) {
	root := t.TempDir()
	empty := filepath.Join(root, "empty")
	if err := os.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}

	m := New(false)
	if err := m.CleanupEmptyDirs(root); err != nil {
		t.Fatalf("CleanupEmptyDirs error: %v", err)
	}

	if _, err := os.Stat(empty); !os.IsNotExist(err) {
		t.Errorf("expected empty subdir to be removed, but it still exists")
	}
}

func TestCleanupEmptyDirs_LeavesNonEmptyDirs(t *testing.T) {
	root := t.TempDir()
	sub := filepath.Join(root, "nonempty")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "file.txt"), []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}

	m := New(false)
	if err := m.CleanupEmptyDirs(root); err != nil {
		t.Fatalf("CleanupEmptyDirs error: %v", err)
	}

	if _, err := os.Stat(sub); err != nil {
		t.Errorf("non-empty subdir should remain: %v", err)
	}
}

func TestCleanupEmptyDirs_DryRunLeavesEmptyDir(t *testing.T) {
	root := t.TempDir()
	empty := filepath.Join(root, "empty")
	if err := os.MkdirAll(empty, 0755); err != nil {
		t.Fatal(err)
	}

	m := New(true)
	if err := m.CleanupEmptyDirs(root); err != nil {
		t.Fatalf("CleanupEmptyDirs dry-run error: %v", err)
	}

	if _, err := os.Stat(empty); err != nil {
		t.Errorf("dry-run should not remove empty dir: %v", err)
	}
}
