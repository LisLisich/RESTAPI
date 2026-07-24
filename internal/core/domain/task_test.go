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

func TestTaskApplyPatchUpdatesOnlySetFields(t *testing.T) {
	createdAt := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	task := NewTask(
		1,
		1,
		"old title",
		stringPtr("old description"),
		false,
		createdAt,
		nil,
		1,
	)
	patch := NewTaskPatch(
		Nullable[string]{
			Set:   true,
			Value: stringPtr("new title"),
		},
		Nullable[string]{
			Set:   true,
			Value: nil,
		},
		Nullable[bool]{
			Set:   true,
			Value: boolPtr(true),
		},
	)

	err := task.ApplyPatch(patch)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if task.Title != "new title" {
		t.Fatalf("expected title %q, got %q", "new title", task.Title)
	}
	if task.Description != nil {
		t.Fatalf("expected description to be nil, got %q", *task.Description)
	}
	if !task.Completed {
		t.Fatal("expected task to be completed")
	}
	if task.CompletedAt == nil {
		t.Fatal("expected completed task to have CompletedAt")
	}
	if task.CompletedAt.Before(createdAt) {
		t.Fatalf("expected CompletedAt to be after CreatedAt, got %v before %v", task.CompletedAt, createdAt)
	}
}

func TestTaskApplyPatchDoesNotPartiallyMutateOnInvalidResult(t *testing.T) {
	task := NewTaskUninitialized("old title", stringPtr("old description"), 1)
	patch := NewTaskPatch(
		Nullable[string]{
			Set:   true,
			Value: stringPtr("ab"),
		},
		Nullable[string]{
			Set:   true,
			Value: nil,
		},
		Nullable[bool]{},
	)

	err := task.ApplyPatch(patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if task.Title != "old title" {
		t.Fatalf("expected title to remain %q, got %q", "old title", task.Title)
	}
	if task.Description == nil || *task.Description != "old description" {
		t.Fatalf("expected description to remain %q, got %v", "old description", task.Description)
	}
}

func TestTaskApplyPatchWrapsPatchValidationWithTaskContext(t *testing.T) {
	task := NewTaskUninitialized("old title", nil, 1)
	patch := NewTaskPatch(
		Nullable[string]{
			Set:   true,
			Value: nil,
		},
		Nullable[string]{},
		Nullable[bool]{},
	)

	err := task.ApplyPatch(patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if !strings.Contains(err.Error(), "validate task patch") {
		t.Fatalf("expected error to contain %q, got %q", "validate task patch", err.Error())
	}
}

func TestTaskCompletionDuration(t *testing.T) {
	createdAt := time.Date(2026, 7, 3, 12, 0, 0, 0, time.UTC)
	completedAt := createdAt.Add(2 * time.Hour)

	tests := []struct {
		name         string
		completed    bool
		completedAt  *time.Time
		wantDuration *time.Duration
	}{
		{
			name:         "not completed",
			completed:    false,
			completedAt:  nil,
			wantDuration: nil,
		},
		{
			name:         "completed without completed at",
			completed:    true,
			completedAt:  nil,
			wantDuration: nil,
		},
		{
			name:         "completed with completed at",
			completed:    true,
			completedAt:  &completedAt,
			wantDuration: durationPtr(2 * time.Hour),
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

			duration := task.CompletionDuration()

			if tt.wantDuration == nil {
				if duration != nil {
					t.Fatalf("expected nil duration, got %v", *duration)
				}
				return
			}
			if duration == nil {
				t.Fatalf("expected duration %v, got nil", *tt.wantDuration)
			}
			if *duration != *tt.wantDuration {
				t.Fatalf("expected duration %v, got %v", *tt.wantDuration, *duration)
			}
		})
	}
}

func boolPtr(value bool) *bool {
	return &value
}

func durationPtr(value time.Duration) *time.Duration {
	return &value
}
