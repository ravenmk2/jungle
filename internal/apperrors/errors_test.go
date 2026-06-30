package apperrors

import "testing"

func TestNewAndError(t *testing.T) {
	e := New(NotFound, "workspace not found")
	if e.Code != NotFound {
		t.Fatalf("code = %s, want %s", e.Code, NotFound)
	}
	if e.Error() != "workspace not found" {
		t.Fatalf("Error() = %q", e.Error())
	}
}

func TestNewWithDetails(t *testing.T) {
	d := []Detail{{Code: "REQUIRED", Message: "workspace is required", Target: "workspace"}}
	e := NewWithDetails(ValidationError, "validation failed", d)
	if len(e.Details) != 1 || e.Details[0].Target != "workspace" {
		t.Fatalf("details = %+v", e.Details)
	}
}
