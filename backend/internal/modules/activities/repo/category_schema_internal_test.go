package repo

import "testing"

func TestPlaceholderForFieldReturnsStoredPlaceholder(t *testing.T) {
	field := fieldSchemaRow{Placeholder: "Enter location"}

	if got := placeholderForField(field); got != "Enter location" {
		t.Fatalf("expected stored placeholder, got %q", got)
	}
}
