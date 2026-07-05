package web_fs_repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestGetFileReturnsFileContent(t *testing.T) {
	dir := t.TempDir()
	filePath := filepath.Join(dir, "index.html")
	content := []byte("<html></html>")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("write test file: %v", err)
	}
	repository := NewWebRepository()

	got, err := repository.GetFile(filePath)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(got) != string(content) {
		t.Fatalf("expected content %q, got %q", content, got)
	}
}

func TestGetFileMapsMissingFileToNotFound(t *testing.T) {
	repository := NewWebRepository()

	_, err := repository.GetFile(filepath.Join(t.TempDir(), "missing.html"))

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
