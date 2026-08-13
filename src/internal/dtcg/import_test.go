package dtcg

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImportInheritsTypesAndMergesModeOnlyOverlay(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "base.json")
	overlay := filepath.Join(dir, "dense.json")
	if err := os.WriteFile(base, []byte(`{"color":{"$type":"color","primary":{"$value":"{palette.blue}","$extensions":{"mode":{"dark":"#000"}}}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(overlay, []byte(`{"color":{"primary":{"$extensions":{"mode":{"dense":"#111"}}}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	tokens, err := Import([]string{base, overlay}, "hb")
	if err != nil {
		t.Fatal(err)
	}
	if len(tokens) != 1 {
		t.Fatalf("got %d tokens", len(tokens))
	}
	token := tokens[0]
	if token.ID != "color.primary" || token.Name != "--hb-color-primary" || token.Type != "color" {
		t.Fatalf("unexpected token: %#v", token)
	}
	if token.Value != "{palette.blue}" || token.Modes["dark"] != "#000" || token.Modes["dense"] != "#111" {
		t.Fatalf("unexpected values: %#v", token)
	}
}

func TestImportRenamesTypographyShorthandForCSS(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "tokens.json")
	if err := os.WriteFile(path, []byte(`{"type-sh":{"$type":"typography","body":{"$value":"x"}}}`), 0644); err != nil {
		t.Fatal(err)
	}
	tokens, err := Import([]string{path}, "hb")
	if err != nil {
		t.Fatal(err)
	}
	if tokens[0].Name != "--hb-type-body" {
		t.Fatal(tokens[0].Name)
	}
}
