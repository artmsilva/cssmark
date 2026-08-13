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

func TestWriteFlatCSSReplacesExplicitDenseModesWithDerivation(t *testing.T) {
	dir := t.TempDir()
	err := WriteFlatCSS([]Token{{ID: "type.body", Type: "typography", Value: map[string]any{"fontSize": "16px", "fontFamily": "Inter"}, Modes: map[string]any{"dense": map[string]any{"fontSize": "14px", "fontFamily": "Inter"}}}}, dir)
	if err != nil {
		t.Fatal(err)
	}
	body, err := os.ReadFile(filepath.Join(dir, "typography.css"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "@mode dense") || !strings.Contains(string(body), "@derive dense") {
		t.Fatalf("expected derived dense mode:\n%s", body)
	}
}

func TestWriteFlatCSSUsesScaleAndGroupModeDefaults(t *testing.T) {
	dir := t.TempDir()
	err := WriteFlatCSS([]Token{
		{ID: "space.1", Type: "dimension", Value: "0.25rem"},
		{ID: "duration.fast", Type: "duration", Value: "150ms", Modes: map[string]any{"wireframe": "0ms"}},
		{ID: "duration.slow", Type: "duration", Value: "400ms", Modes: map[string]any{"wireframe": "0ms"}},
	}, dir)
	if err != nil {
		t.Fatal(err)
	}
	dimension, err := os.ReadFile(filepath.Join(dir, "dimension.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(dimension), "@scale space {") || strings.Contains(string(dimension), "@token 1 {") {
		t.Fatalf("expected scale source:\n%s", dimension)
	}
	duration, err := os.ReadFile(filepath.Join(dir, "duration.css"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(duration), "@mode wireframe { all: 0ms; }") {
		t.Fatalf("expected group mode:\n%s", duration)
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
