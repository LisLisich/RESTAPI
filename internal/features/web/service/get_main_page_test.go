package web_service

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

type fakeWebRepository struct {
	getFileCalled bool
	gotFilePath   string
	getFileErr    error
}

var _ WebRepository = (*fakeWebRepository)(nil)

func (r *fakeWebRepository) GetFile(filepath string) ([]byte, error) {
	r.getFileCalled = true
	r.gotFilePath = filepath

	if r.getFileErr != nil {
		return nil, r.getFileErr
	}
	return []byte("<html></html>"), nil
}

func TestGetMainPageReadsPublicIndexFromProjectRoot(t *testing.T) {
	repository := &fakeWebRepository{}
	service := NewWebService(repository)
	t.Setenv("PROJECT_ROOT", "/project")

	html, err := service.GetMainPage()

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if string(html) != "<html></html>" {
		t.Fatalf("expected html content, got %q", html)
	}
	if !repository.getFileCalled {
		t.Fatal("expected GetFile to be called")
	}
	if !strings.Contains(repository.gotFilePath, "public/index.html") {
		t.Fatalf("expected public index path, got %q", repository.gotFilePath)
	}
}

func TestGetMainPageWrapsRepositoryError(t *testing.T) {
	repository := &fakeWebRepository{
		getFileErr: core_errors.ErrNotFound,
	}
	service := NewWebService(repository)
	t.Setenv("PROJECT_ROOT", "/project")

	_, err := service.GetMainPage()

	if !errors.Is(err, core_errors.ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
