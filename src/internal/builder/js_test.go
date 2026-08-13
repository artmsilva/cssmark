package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artmsilva/cssmark/src/internal/parser"
)

func TestWriteJSUsesStableIDsAndModes(t *testing.T) {
	dir := t.TempDir()
	js, meta, dts := filepath.Join(dir, "tokens.js"), filepath.Join(dir, "tokens.meta.js"), filepath.Join(dir, "tokens.d.ts")
	err := WriteJS([]parser.Token{{
		ID:           "color.action.primary",
		Name:         "--hb-color-action-primary",
		Type:         "color",
		InitialValue: "#123456",
		Modes:        map[string]string{"dark": "#abcdef"},
	}}, js, meta, dts)
	if err != nil {
		t.Fatal(err)
	}
	for path, expected := range map[string]string{
		js:   `"color.action.primary": "#123456"`,
		meta: `"$type": "color"`,
		dts:  `export declare const tokens: Record<string, string>;`,
	} {
		body, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(body), expected) {
			t.Fatalf("%s missing %q:\n%s", path, expected, body)
		}
	}
}
