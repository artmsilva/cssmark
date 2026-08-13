package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTokenSourceLowersImportsStateMapsAndRefs(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "color.css"), []byte(`@token color {
  @token action {
    default: ref(color.brand.blue.450);
    hover: #fff;
    @mode dark {
      hover: #000;
    }
  }
}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.css"), []byte("@tokens { namespace: hb; }\n@import \"./color.css\";\n"), 0644); err != nil {
		t.Fatal(err)
	}
	tokens, err := ParseTokenSource(filepath.Join(dir, "index.css"))
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 2 || tokens[0].ID != "color.action.default" || tokens[0].Name != "--hb-color-action-default" || tokens[0].InitialValue != "var(--hb-color-brand-blue-450)" || tokens[1].ID != "color.action.hover" || tokens[1].Modes["dark"] != "#000" {
		t.Fatalf("got %#v", tokens)
	}
}
