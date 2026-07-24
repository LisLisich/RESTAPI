package domain

import (
	"errors"
	"strings"
	"testing"

	core_errors "github.com/LisLisich/RESTAPI/internal/core/errors"
)

func TestUserValidateFullNameLength(t *testing.T) {
	tests := []struct {
		name     string
		fullName string
		wantErr  bool
	}{
		{
			name:     "too short",
			fullName: "Iv",
			wantErr:  true,
		},
		{
			name:     "minimum length",
			fullName: "Iva",
			wantErr:  false,
		},
		{
			name:     "maximum unicode length",
			fullName: strings.Repeat("я", 100),
			wantErr:  false,
		},
		{
			name:     "too long unicode length",
			fullName: strings.Repeat("я", 101),
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUserUninitialized(tt.fullName, nil)

			err := user.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestUserValidatePhoneNumber(t *testing.T) {
	tests := []struct {
		name        string
		phoneNumber *string
		wantErr     bool
	}{
		{
			name:        "nil phone number",
			phoneNumber: nil,
			wantErr:     false,
		},
		{
			name:        "minimum length",
			phoneNumber: stringPtr("+123456789"),
			wantErr:     false,
		},
		{
			name:        "maximum length",
			phoneNumber: stringPtr("+12345678901234"),
			wantErr:     false,
		},
		{
			name:        "too short",
			phoneNumber: stringPtr("+12345678"),
			wantErr:     true,
		},
		{
			name:        "too long",
			phoneNumber: stringPtr("+123456789012345"),
			wantErr:     true,
		},
		{
			name:        "missing plus",
			phoneNumber: stringPtr("1234567890"),
			wantErr:     true,
		},
		{
			name:        "contains non digit",
			phoneNumber: stringPtr("+123456789a"),
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := NewUserUninitialized("Ivan Ivanov", tt.phoneNumber)

			err := user.Validate()

			if tt.wantErr && !errors.Is(err, core_errors.ErrInvalidArgument) {
				t.Fatalf("expected ErrInvalidArgument, got %v", err)
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
		})
	}
}

func TestUserPatchValidateNullableFields(t *testing.T) {
	tests := []struct {
		name    string
		patch   UserPatch
		wantErr bool
	}{
		{
			name: "full name null",
			patch: NewUserPatch(
				Nullable[string]{
					Set:   true,
					Value: nil,
				},
				Nullable[string]{},
			),
			wantErr: true,
		},
		{
			name: "phone number null",
			patch: NewUserPatch(
				Nullable[string]{},
				Nullable[string]{
					Set:   true,
					Value: nil,
				},
			),
			wantErr: false,
		},
		{
			name: "all fields absent",
			patch: NewUserPatch(
				Nullable[string]{},
				Nullable[string]{},
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

func TestUserApplyPatchUpdatesOnlySetFields(t *testing.T) {
	user := NewUser(1, 1, "Ivan Ivanov", stringPtr("+79998887766"))
	patch := NewUserPatch(
		Nullable[string]{
			Set:   true,
			Value: stringPtr("Petr Petrov"),
		},
		Nullable[string]{
			Set:   true,
			Value: nil,
		},
	)

	err := user.ApplyPatch(patch)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user.FullName != "Petr Petrov" {
		t.Fatalf("expected full name %q, got %q", "Petr Petrov", user.FullName)
	}
	if user.PhoneNumber != nil {
		t.Fatalf("expected phone number to be nil, got %q", *user.PhoneNumber)
	}
}

func TestUserApplyPatchDoesNotPartiallyMutateOnInvalidResult(t *testing.T) {
	originalPhoneNumber := "+79998887766"
	user := NewUser(1, 1, "Ivan Ivanov", &originalPhoneNumber)
	patch := NewUserPatch(
		Nullable[string]{
			Set:   true,
			Value: stringPtr("Iv"),
		},
		Nullable[string]{
			Set:   true,
			Value: nil,
		},
	)

	err := user.ApplyPatch(patch)

	if !errors.Is(err, core_errors.ErrInvalidArgument) {
		t.Fatalf("expected ErrInvalidArgument, got %v", err)
	}
	if user.FullName != "Ivan Ivanov" {
		t.Fatalf("expected full name to remain %q, got %q", "Ivan Ivanov", user.FullName)
	}
	if user.PhoneNumber == nil || *user.PhoneNumber != originalPhoneNumber {
		t.Fatalf("expected phone number to remain %q, got %v", originalPhoneNumber, user.PhoneNumber)
	}
}
