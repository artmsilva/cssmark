package dtcg

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteFlatCSSWritesComposableFilesAndExpandsTypography(t *testing.T) {
	dir := t.TempDir()
	err := WriteFlatCSS([]Token{{
		ID: "action.type.body", Name: "--hb-action-type-body", Value: map[string]any{
			"fontFamily": "{font-family.body}", "fontSize": "{space.4}", "fontStyle": "normal", "fontWeight": "700", "lineHeight": "{space.6}",
		}, Modes: map[string]any{"dense": map[string]any{"fontSize": "{space.3-5}", "lineHeight": "{space.5-5}"}},
	}}, dir)
	if err != nil {
		t.Fatal(err)
	}
	for path, expected := range map[string]string{
		"index.css":      `@import "./action.css" layer(hb-tokens.base);`,
		"action.css":     `--hb-action-type-body-font-size: var(--hb-space-4);`,
		"mode-dense.css": `--hb-action-type-body-font-size: var(--hb-space-3-5);`,
	} {
		body, err := os.ReadFile(filepath.Join(dir, path))
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), expected) {
			t.Fatalf("%s missing %q:\n%s", path, expected, body)
		}
	}
}
