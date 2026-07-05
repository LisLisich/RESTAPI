package core_http_types

import (
	"encoding/json"
	"testing"
)

type nullablePayload struct {
	Name Nullable[string] `json:"name"`
}

func TestNullableUnmarshalTracksExplicitValue(t *testing.T) {
	var payload nullablePayload

	err := json.Unmarshal([]byte(`{"name":"Ivan"}`), &payload)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !payload.Name.Set {
		t.Fatal("expected nullable field to be set")
	}
	if payload.Name.Value == nil || *payload.Name.Value != "Ivan" {
		t.Fatalf("expected value %q, got %v", "Ivan", payload.Name.Value)
	}
}

func TestNullableUnmarshalTracksExplicitNull(t *testing.T) {
	var payload nullablePayload

	err := json.Unmarshal([]byte(`{"name":null}`), &payload)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !payload.Name.Set {
		t.Fatal("expected nullable field to be set")
	}
	if payload.Name.Value != nil {
		t.Fatalf("expected nil value, got %v", *payload.Name.Value)
	}
}

func TestNullableUnmarshalLeavesAbsentFieldUnset(t *testing.T) {
	var payload nullablePayload

	err := json.Unmarshal([]byte(`{}`), &payload)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if payload.Name.Set {
		t.Fatal("expected absent nullable field to be unset")
	}
}
