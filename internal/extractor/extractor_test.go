package extractor

import (
	"os"
	"path/filepath"
	"testing"
)

// --- getRARBaseName ---

func TestGetRARBaseName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"archive.rar", "archive"},
		{"archive.part01.rar", "archive"},
		{"archive.part001.rar", "archive"},
		{"archive.part1.rar", "archive"},
		{"Show.S01E01.2160p.WEB.H265-GROUP.rar", "show.s01e01.2160p.web.h265-group"},
		{"Show.S01E01.2160p.WEB.H265-GROUP.part02.rar", "show.s01e01.2160p.web.h265-group"},
	}

	for _, tt := range tests {
		got := getRARBaseName(tt.input)
		if got != tt.want {
			t.Errorf("getRARBaseName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- isRARFile ---

func TestIsRARFile(t *testing.T) {
	yes := []string{
		"archive.rar",
		"archive.part01.rar",
		"archive.part001.rar",
		"/some/path/show.s01e01.rar",
	}
	no := []string{
		"archive.mkv",
		"archive.mp4",
		"archive.zip",
		"archive.r00",
		"",
	}

	for _, p := range yes {
		if !isRARFile(p) {
			t.Errorf("isRARFile(%q) = false, want true", p)
		}
	}
	for _, p := range no {
		if isRARFile(p) {
			t.Errorf("isRARFile(%q) = true, want false", p)
		}
	}
}

// --- IsAlreadyExtracted ---

func TestIsAlreadyExtracted_NoMediaFiles(t *testing.T) {
	dir := t.TempDir()
	// Only RAR files present
	touch(t, filepath.Join(dir, "archive.rar"))
	touch(t, filepath.Join(dir, "archive.part02.rar"))

	rar := RARFile{Directory: dir}
	got, name := IsAlreadyExtracted(rar)
	if got {
		t.Errorf("IsAlreadyExtracted = true (file: %q), want false", name)
	}
}

func TestIsAlreadyExtracted_WithMKV(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "archive.rar"))
	touch(t, filepath.Join(dir, "show.s01e01.mkv"))

	rar := RARFile{Directory: dir}
	got, name := IsAlreadyExtracted(rar)
	if !got {
		t.Errorf("IsAlreadyExtracted = false, want true")
	}
	if name != "show.s01e01.mkv" {
		t.Errorf("IsAlreadyExtracted name = %q, want %q", name, "show.s01e01.mkv")
	}
}

func TestIsAlreadyExtracted_WithMP4(t *testing.T) {
	dir := t.TempDir()
	touch(t, filepath.Join(dir, "movie.mp4"))

	rar := RARFile{Directory: dir}
	got, _ := IsAlreadyExtracted(rar)
	if !got {
		t.Errorf("IsAlreadyExtracted = false for .mp4, want true")
	}
}

// --- FindRARFiles ---

func TestFindRARFiles_Empty(t *testing.T) {
	dir := t.TempDir()
	ext := New(dir)
	files, err := ext.FindRARFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 0 {
		t.Errorf("got %d files, want 0", len(files))
	}
}

func TestFindRARFiles_SingleArchive(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "Show.S01E01")
	must(t, os.MkdirAll(subDir, 0755))
	touch(t, filepath.Join(subDir, "show.s01e01.rar"))

	ext := New(dir)
	files, err := ext.FindRARFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("got %d files, want 1", len(files))
	}
	if files[0].Directory != subDir {
		t.Errorf("Directory = %q, want %q", files[0].Directory, subDir)
	}
}

func TestFindRARFiles_MultiPartDeduped(t *testing.T) {
	dir := t.TempDir()
	subDir := filepath.Join(dir, "Movie")
	must(t, os.MkdirAll(subDir, 0755))
	touch(t, filepath.Join(subDir, "movie.part01.rar"))
	touch(t, filepath.Join(subDir, "movie.part02.rar"))
	touch(t, filepath.Join(subDir, "movie.part03.rar"))

	ext := New(dir)
	files, err := ext.FindRARFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("got %d files (want 1) — multi-part not deduplicated", len(files))
	}
}

func TestFindRARFiles_MultipleArchives(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"ShowA", "ShowB", "ShowC"} {
		sub := filepath.Join(dir, name)
		must(t, os.MkdirAll(sub, 0755))
		touch(t, filepath.Join(sub, name+".rar"))
	}

	ext := New(dir)
	files, err := ext.FindRARFiles()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(files) != 3 {
		t.Errorf("got %d files, want 3", len(files))
	}
}

// --- DeleteRARFiles ---

func TestDeleteRARFiles_RemovesOnlyRARs(t *testing.T) {
	dir := t.TempDir()
	rarPath := filepath.Join(dir, "archive.rar")
	mkvPath := filepath.Join(dir, "show.s01e01.mkv")
	touch(t, rarPath)
	touch(t, mkvPath)

	rar := RARFile{Directory: dir}
	if err := DeleteRARFiles(rar, false); err != nil {
		t.Fatalf("DeleteRARFiles error: %v", err)
	}

	if _, err := os.Stat(rarPath); !os.IsNotExist(err) {
		t.Errorf("expected RAR file to be deleted, but it still exists")
	}
	if _, err := os.Stat(mkvPath); err != nil {
		t.Errorf("expected MKV file to remain, but got: %v", err)
	}
}

func TestDeleteRARFiles_DryRunLeavesFiles(t *testing.T) {
	dir := t.TempDir()
	rarPath := filepath.Join(dir, "archive.rar")
	touch(t, rarPath)

	rar := RARFile{Directory: dir}
	if err := DeleteRARFiles(rar, true); err != nil {
		t.Fatalf("DeleteRARFiles dry-run error: %v", err)
	}

	if _, err := os.Stat(rarPath); err != nil {
		t.Errorf("dry-run should not delete files, but RAR is gone: %v", err)
	}
}

func TestDeleteRARFiles_MultiPart(t *testing.T) {
	dir := t.TempDir()
	parts := []string{"archive.part01.rar", "archive.part02.rar", "archive.part03.rar"}
	for _, p := range parts {
		touch(t, filepath.Join(dir, p))
	}
	touch(t, filepath.Join(dir, "archive.mkv"))

	rar := RARFile{Directory: dir}
	if err := DeleteRARFiles(rar, false); err != nil {
		t.Fatalf("DeleteRARFiles error: %v", err)
	}

	for _, p := range parts {
		if _, err := os.Stat(filepath.Join(dir, p)); !os.IsNotExist(err) {
			t.Errorf("expected %s to be deleted", p)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, "archive.mkv")); err != nil {
		t.Errorf("expected MKV to remain: %v", err)
	}
}

// --- helpers ---

func touch(t *testing.T, path string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("touch %q: %v", path, err)
	}
	f.Close()
}

func must(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}
