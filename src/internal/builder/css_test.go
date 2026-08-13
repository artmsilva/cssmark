package builder

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/artmsilva/cssmark/src/internal/parser"
)

func TestReadSourceCSSUsesRelativePathsAndHidesAbsoluteDirectories(t *testing.T) {
	dir := t.TempDir()
	absolutePath := filepath.Join(dir, "tokens.css")
	if err := os.WriteFile(absolutePath, []byte("@property --color-a {}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, files, err := readSourceCSS([]string{absolutePath})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "tokens.css" {
		t.Fatalf("expected absolute path to be displayed as basename, got %v", files)
	}

	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	if err := os.MkdirAll("examples", 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("examples/tokens.css", []byte("@property --color-b {}"), 0644); err != nil {
		t.Fatal(err)
	}

	_, files, err = readSourceCSS([]string{"examples/tokens.css"})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || files[0] != "examples/tokens.css" {
		t.Fatalf("expected relative path to be preserved, got %v", files)
	}
}

func TestWriteDocsIncludesReferenceTools(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "tokens.css")
	if err := os.WriteFile(source, []byte(`
@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
  description: "Primary brand color.";
  category: "color.brand";
  type: "color";
}
`), 0644); err != nil {
		t.Fatal(err)
	}
	tokens, err := parser.ParseFiles([]string{source})
	if err != nil {
		t.Fatal(err)
	}

	out := filepath.Join(dir, "site")
	if err := WriteDocs(tokens, out, source); err != nil {
		t.Fatal(err)
	}
	index, err := os.ReadFile(filepath.Join(out, "index.html"))
	if err != nil {
		t.Fatal(err)
	}
	html := string(index)
	for _, expected := range []string{
		`id="token-filter"`,
		`class="density-toggle"`,
		`class="copy-token"`,
		`id="token-color-primary"`,
		`data-token-name="--color-primary"`,
	} {
		if !strings.Contains(html, expected) {
			t.Fatalf("expected generated docs to include %s", expected)
		}
	}
}

func TestToCSSUsesConfiguredModeSelector(t *testing.T) {
	css := ToCSSWithModeSelectors([]parser.Token{{
		Name:         "--space-sm",
		InitialValue: "4px",
		Modes:        map[string]string{"dense": "2px"},
	}}, map[string]string{"dense": ":root[data-density='dense'], [data-density='dense']"})

	if !strings.Contains(css, ":root[data-density='dense'], [data-density='dense'] {") {
		t.Fatalf("expected configured dense selector, got:\n%s", css)
	}
	if strings.Contains(css, ":root[data-color-mode='dense']") {
		t.Fatalf("expected configured selector to replace the default, got:\n%s", css)
	}
}

func TestToCSSSanitizesCategoryComments(t *testing.T) {
	css := ToCSS([]parser.Token{{
		Name:         "--color-primary",
		InitialValue: "#0055ff",
		Category:     "color */ body { color: red } /*",
	}})

	if strings.Contains(css, "*/ body") {
		t.Fatalf("expected category comment terminator to be sanitized, got:\n%s", css)
	}
	if !strings.Contains(css, "* / body") {
		t.Fatalf("expected sanitized comment marker, got:\n%s", css)
	}
}
