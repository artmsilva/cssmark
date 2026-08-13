package dtcg

import (
	"strings"
	"testing"
)

func TestCSSSourcePreservesScalarAliasesAndCompositeModes(t *testing.T) {
	css, err := CSSSource([]Token{
		{ID: "color.primary", Name: "--hb-color-primary", Type: "color", Value: "{color.blue.500}", Modes: map[string]any{"dark": "#000"}},
		{ID: "type.body", Name: "--hb-type-body", Type: "typography", Value: map[string]any{"fontFamily": "Inter", "fontSize": "{space.4}", "fontStyle": "normal", "fontWeight": "400", "lineHeight": "{space.6}"}, Modes: map[string]any{"dense": map[string]any{"fontSize": "{space.3}"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, expected := range []string{
		`id: "color.primary";`,
		`initial-value: var(--hb-color-blue-500);`,
		`mode-dark: #000;`,
		`@property --hb-type-body-font-size {`,
		`mode-dense: var(--hb-space-3);`,
	} {
		if !strings.Contains(css, expected) {
			t.Fatalf("missing %q:\n%s", expected, css)
		}
	}
}
