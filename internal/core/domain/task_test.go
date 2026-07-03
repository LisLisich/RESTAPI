package domain

import (
	"errors"
	"strings"
	"testing"
	"time"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestTaskValidateTitleLength(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		wantErr bool
	}{
		{
			name:    "too short",
			title:   "ab",
			wantErr: true,
		},
		{
			name:    "minimum length",
			title:   "abc",
			wantErr: false,
		},
		{
			name:    "maximum length",
			title:   strings.Repeat("a", 100),
			wantErr: false,
		},
		{
			name:    "too long",
			title:   strings.Repeat("a", 101),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTaskUninitialized(tt.title, nil, 1)

			err := task.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func stringPtr(value string) *string {
	return &value
}

func TestTaskValidateDescriptionLength(t *testing.T) {
	tests := []struct {
		name        string
		description *string
		wantErr     bool
	}{
		{
			name:        "nil description",
			description: nil,
			wantErr:     false,
		},
		{
			name:        "empty description",
			description: stringPtr(""),
			wantErr:     true,
		},
		{
			name:        "minimum length",
			description: stringPtr("a"),
			wantErr:     false,
		},
		{
			name:        "maximum length",
			description: stringPtr(strings.Repeat("a", 1000)),
			wantErr:     false,
		},
		{
			name:        "too long",
			description: stringPtr(strings.Repeat("a", 1001)),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTaskUninitialized("valid title", tt.description, 1)

			err := task.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestTaskValidateCompletionState(t *testing.T) {
	createdAt := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	beforeCreatedAt := createdAt.Add(-time.Second)
	afterCreatedAt := createdAt.Add(time.Second)

	tests := []struct {
		name        string
		completed   bool
		completedAt *time.Time
		wantErr     bool
	}{
		{
			name:        "completed without completed at",
			completed:   true,
			completedAt: nil,
			wantErr:     true,
		},
		{
			name:        "completed before created at",
			completed:   true,
			completedAt: &beforeCreatedAt,
			wantErr:     true,
		},
		{
			name:        "completed after created at",
			completed:   true,
			completedAt: &afterCreatedAt,
			wantErr:     false,
		},
		{
			name:        "not completed without completed at",
			completed:   false,
			completedAt: nil,
			wantErr:     false,
		},
		{
			name:        "not completed with completed at",
			completed:   false,
			completedAt: &afterCreatedAt,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := NewTask(
				1,
				1,
				"valid title",
				nil,
				tt.completed,
				createdAt,
				tt.completedAt,
				1,
			)

			err := task.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestTaskPatchValidateNullableFields(t *testing.T) {
	tests := []struct {
		name    string
		patch   TaskPatch
		wantErr bool
	}{
		{
			name: "title null",
			patch: NewTaskPatch(
				Nullable[string]{
					Set:   true,
					Value: nil,
				},
				Nullable[string]{},
				Nullable[bool]{},
			),
			wantErr: true,
		},
		{
			name: "completed null",
			patch: NewTaskPatch(
				Nullable[string]{},
				Nullable[string]{},
				Nullable[bool]{
					Set:   true,
					Value: nil,
				},
			),
			wantErr: true,
		},
		{
			name: "description null",
			patch: NewTaskPatch(
				Nullable[string]{},
				Nullable[string]{
					Set:   true,
					Value: nil,
				},
				Nullable[bool]{},
			),
			wantErr: false,
		},
		{
			name: "all fields absent",
			patch: NewTaskPatch(
				Nullable[string]{},
				Nullable[string]{},
				Nullable[bool]{},
			),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.patch.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}
