package organizer

import (
	"os"
	"path/filepath"
	"testing"
)

func statDir(t *testing.T, path string) os.FileInfo {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return info
}

func TestMoveDirectoryKeepsDirectoryInode(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	before := statDir(t, filepath.Join(source, "book"))

	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(output, "Author", "Book")
	if !os.SameFile(before, statDir(t, target)) {
		t.Fatal("book directory was recreated instead of renamed")
	}
	if got := safetyRead(t, filepath.Join(target, "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(source, "book")); !os.IsNotExist(err) {
		t.Fatalf("source directory still exists: %v", err)
	}
	if len(org.GetSummary().Moves) != 1 {
		t.Fatalf("moves = %d, want 1", len(org.GetSummary().Moves))
	}
}

func TestMoveDirectoryInPlaceKeepsDirectoryInode(t *testing.T) {
	library := t.TempDir()
	safetyWrite(
		t,
		filepath.Join(library, "incoming", "book", "metadata.json"),
		`{"title":"Book","authors":["Author"],"series":["Saga #2"]}`,
	)
	safetyWrite(t, filepath.Join(library, "incoming", "book", "audio.mp3"), "original audio")
	before := statDir(t, filepath.Join(library, "incoming", "book"))

	org, err := NewOrganizer(&OrganizerConfig{
		BaseDir:        library,
		LayoutTemplate: "{author}/{series}/{series_number. }{title}",
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	target := filepath.Join(library, "Author", "Saga", "2. Book")
	if !os.SameFile(before, statDir(t, target)) {
		t.Fatal("book directory was recreated instead of renamed")
	}
}

func TestMoveDirectoryFallsBackWhenTargetExists(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	target := filepath.Join(output, "Author", "Book")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	before := statDir(t, target)

	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	if !os.SameFile(before, statDir(t, target)) {
		t.Fatal("existing target directory was replaced")
	}
	if got := safetyRead(t, filepath.Join(target, "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(source, "book", "audio.mp3")); !os.IsNotExist(err) {
		t.Fatalf("audio still in source: %v", err)
	}
}

func TestMoveDirectoryFallsBackWithSubdirectories(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	extras := filepath.Join(source, "book", "extras", "notes.txt")
	safetyWrite(t, extras, "notes")

	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	if got := safetyRead(t, filepath.Join(output, "Author", "Book", "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if got := safetyRead(t, extras); got != "notes" {
		t.Fatalf("subdirectory content = %q", got)
	}
}

func TestMoveDirectoryDryRunLeavesSource(t *testing.T) {
	org, source, output := safetyOrganizer(t, true)

	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	if got := safetyRead(t, filepath.Join(source, "book", "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("dry run created output: %v", err)
	}
}

func TestMoveDirectoryUndoRestoresFiles(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	if err := org.Execute(); err != nil {
		t.Fatal(err)
	}

	undo, err := NewOrganizer(&OrganizerConfig{BaseDir: source, OutputDir: output, Undo: true})
	if err != nil {
		t.Fatal(err)
	}
	if err := undo.Execute(); err != nil {
		t.Fatal(err)
	}

	if got := safetyRead(t, filepath.Join(source, "book", "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(output, "Author", "Book", "audio.mp3")); !os.IsNotExist(err) {
		t.Fatalf("audio still in target: %v", err)
	}
}

func TestMoveDirectoryRollsBackWhenLogCannotBeSaved(t *testing.T) {
	org, source, output := safetyOrganizer(t, false)
	before := statDir(t, filepath.Join(source, "book"))
	if err := os.MkdirAll(org.GetLogPath(), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := org.Execute(); err == nil {
		t.Fatal("expected log save failure")
	}

	if !os.SameFile(before, statDir(t, filepath.Join(source, "book"))) {
		t.Fatal("book directory was not renamed back")
	}
	if got := safetyRead(t, filepath.Join(source, "book", "audio.mp3")); got != "original audio" {
		t.Fatalf("audio content = %q", got)
	}
	if _, err := os.Stat(filepath.Join(output, "Author", "Book")); !os.IsNotExist(err) {
		t.Fatalf("target directory still exists: %v", err)
	}
}
