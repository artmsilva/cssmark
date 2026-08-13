package dtcg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFlatCSSNestsSemanticStateModifiers(t *testing.T) {
	dir := t.TempDir()
	err := WriteFlatCSS([]Token{{
		ID: "color.action.border.focus-visible", Name: "--hb-color-action-border-focus-visible", Type: "color", Value: "{color.brand.blue.300}",
	}}, dir)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "color.css"))
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		"@token color {", "@token action {", "@token border {", "@token focus-visible {", "value: var(--hb-color-brand-blue-300);",
	} {
		if !strings.Contains(string(body), expected) {
			t.Fatalf("missing %q:\n%s", expected, body)
		}
	}
}

func TestWriteFlatCSSUsesTypographyWithoutLegacyTypeSegment(t *testing.T) {
	dir := t.TempDir()
	err := WriteFlatCSS([]Token{{ID: "decorative.type.body.regular", Type: "typography", Value: map[string]any{"fontSize": "{space.4}"}}}, dir)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "typography.css"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "@token type {") || !strings.Contains(string(body), "@token decorative {") {
		t.Fatalf("unexpected tree:\n%s", body)
	}
}
