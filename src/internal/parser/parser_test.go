package parser

import (
	"testing"
)

func TestParseBasicToken(t *testing.T) {
	css := `
@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	token := tokens[0]
	if token.Name != "--color-primary" {
		t.Errorf("Expected name --color-primary, got %s", token.Name)
	}
	if token.Syntax != "<color>" {
		t.Errorf("Expected syntax <color>, got %s", token.Syntax)
	}
	if token.Inherits != false {
		t.Errorf("Expected inherits false, got true")
	}
	if token.InitialValue != "#0055ff" {
		t.Errorf("Expected initial-value #0055ff, got %s", token.InitialValue)
	}
}

func TestParseExtendedDescriptors(t *testing.T) {
	css := `
@property --color-brand {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
  description: "Primary brand color";
  category: "color.brand";
  type: "color";
  aliases: "--color-primary, --color-action";
  deprecated: false;
  examples: "background: var(--color-brand); border-color: var(--color-brand);";
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	token := tokens[0]
	if token.Description != "Primary brand color" {
		t.Errorf("Expected description, got %s", token.Description)
	}
	if token.Category != "color.brand" {
		t.Errorf("Expected category color.brand, got %s", token.Category)
	}
	if token.Type != "color" {
		t.Errorf("Expected type color, got %s", token.Type)
	}
	if len(token.Aliases) != 2 {
		t.Errorf("Expected 2 aliases, got %d", len(token.Aliases))
	}
	if len(token.Examples) != 2 {
		t.Errorf("Expected 2 examples, got %d", len(token.Examples))
	}
}

func TestParseMultipleTokens(t *testing.T) {
	css := `
@property --color-a {
  syntax: "<color>";
  inherits: false;
  initial-value: red;
}

@property --color-b {
  syntax: "<color>";
  inherits: false;
  initial-value: blue;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(tokens) != 2 {
		t.Fatalf("Expected 2 tokens, got %d", len(tokens))
	}
}

func TestRootOverridesInitialValue(t *testing.T) {
	css := `
@property --space-4 {
  syntax: "<length>";
  inherits: false;
  initial-value: 1rem;
}

:root {
  --space-4: 16px;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	if tokens[0].InitialValue != "16px" {
		t.Errorf("Expected :root value 16px, got %s", tokens[0].InitialValue)
	}
}

func TestVarResolutionDirect(t *testing.T) {
	css := `
@property --color-blue {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
}

@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--color-blue);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var primary *Token
	for i := range tokens {
		if tokens[i].Name == "--color-primary" {
			primary = &tokens[i]
		}
	}

	if primary == nil {
		t.Fatal("--color-primary not found")
	}
	if primary.InitialValue != "#0055ff" {
		t.Errorf("Expected resolved value #0055ff, got %s", primary.InitialValue)
	}
	if primary.RawInitialValue != "var(--color-blue)" {
		t.Errorf("Expected raw value var(--color-blue), got %s", primary.RawInitialValue)
	}
}

func TestVarResolutionChain(t *testing.T) {
	css := `
@property --color-hex {
  syntax: "<color>";
  inherits: false;
  initial-value: #ff0000;
}

@property --color-brand {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--color-hex);
}

@property --color-action {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--color-brand);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var action *Token
	for i := range tokens {
		if tokens[i].Name == "--color-action" {
			action = &tokens[i]
		}
	}

	if action == nil {
		t.Fatal("--color-action not found")
	}
	if action.InitialValue != "#ff0000" {
		t.Errorf("Expected resolved chain value #ff0000, got %s", action.InitialValue)
	}
	if action.RawInitialValue != "var(--color-brand)" {
		t.Errorf("Expected raw value var(--color-brand), got %s", action.RawInitialValue)
	}
}

func TestVarResolutionCircular(t *testing.T) {
	css := `
@property --a {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--b);
}

@property --b {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--a);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	// Should not panic — circular refs are kept as-is
	for _, tok := range tokens {
		if tok.Name == "--a" && tok.RawInitialValue == "" {
			t.Error("Expected --a to have RawInitialValue set")
		}
	}
}

func TestVarResolutionUnknownRef(t *testing.T) {
	css := `
@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--unknown-token);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if tokens[0].InitialValue != "var(--unknown-token)" {
		t.Errorf("Expected unknown ref kept as-is, got %s", tokens[0].InitialValue)
	}
}

func TestVarResolutionWithRootOverride(t *testing.T) {
	css := `
@property --color-blue {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
}

@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: #000000;
}

:root {
  --color-primary: var(--color-blue);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var primary *Token
	for i := range tokens {
		if tokens[i].Name == "--color-primary" {
			primary = &tokens[i]
		}
	}

	if primary == nil {
		t.Fatal("--color-primary not found")
	}
	// :root override with var() should be resolved
	if primary.InitialValue != "#0055ff" {
		t.Errorf("Expected resolved value #0055ff, got %s", primary.InitialValue)
	}
	if primary.RawInitialValue != "var(--color-blue)" {
		t.Errorf("Expected raw value var(--color-blue), got %s", primary.RawInitialValue)
	}
}

func TestVarResolutionLiteralUnchanged(t *testing.T) {
	css := `
@property --color-literal {
  syntax: "<color>";
  inherits: false;
  initial-value: #ff0000;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	if tokens[0].RawInitialValue != "" {
		t.Errorf("Expected empty RawInitialValue for literal, got %s", tokens[0].RawInitialValue)
	}
	if tokens[0].InitialValue != "#ff0000" {
		t.Errorf("Expected #ff0000, got %s", tokens[0].InitialValue)
	}
}

func TestModeDescriptor(t *testing.T) {
	css := `
@property --color-bg {
  syntax: "<color>";
  inherits: false;
  initial-value: #ffffff;
  mode-dark: #1a1a2e;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tok := tokens[0]
	if tok.InitialValue != "#ffffff" {
		t.Errorf("Expected initial-value #ffffff, got %s", tok.InitialValue)
	}
	if tok.Modes == nil {
		t.Fatal("Expected Modes to be set")
	}
	if tok.Modes["dark"] != "#1a1a2e" {
		t.Errorf("Expected mode-dark #1a1a2e, got %s", tok.Modes["dark"])
	}
}

func TestModeWithVarResolution(t *testing.T) {
	css := `
@property --color-blue-300 {
  syntax: "<color>";
  inherits: false;
  initial-value: #94B4FF;
}

@property --color-blue-450 {
  syntax: "<color>";
  inherits: false;
  initial-value: #2C49EF;
}

@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--color-blue-450);
  mode-dark: var(--color-blue-300);
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	var primary *Token
	for i := range tokens {
		if tokens[i].Name == "--color-primary" {
			primary = &tokens[i]
		}
	}

	if primary == nil {
		t.Fatal("--color-primary not found")
	}

	// initial-value resolved
	if primary.InitialValue != "#2C49EF" {
		t.Errorf("Expected resolved initial-value #2C49EF, got %s", primary.InitialValue)
	}
	if primary.RawInitialValue != "var(--color-blue-450)" {
		t.Errorf("Expected raw initial-value var(--color-blue-450), got %s", primary.RawInitialValue)
	}

	// mode-dark resolved
	if primary.Modes["dark"] != "#94B4FF" {
		t.Errorf("Expected resolved mode-dark #94B4FF, got %s", primary.Modes["dark"])
	}
	if primary.RawModes["dark"] != "var(--color-blue-300)" {
		t.Errorf("Expected raw mode-dark var(--color-blue-300), got %s", primary.RawModes["dark"])
	}
}

func TestMultipleModes(t *testing.T) {
	css := `
@property --color-bg {
  syntax: "<color>";
  inherits: false;
  initial-value: #ffffff;
  mode-dark: #1a1a2e;
  mode-high-contrast: #000000;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tok := tokens[0]
	if len(tok.Modes) != 2 {
		t.Fatalf("Expected 2 modes, got %d", len(tok.Modes))
	}
	if tok.Modes["dark"] != "#1a1a2e" {
		t.Errorf("Expected dark #1a1a2e, got %s", tok.Modes["dark"])
	}
	if tok.Modes["high-contrast"] != "#000000" {
		t.Errorf("Expected high-contrast #000000, got %s", tok.Modes["high-contrast"])
	}
}

func TestModeLiteralNoRawModes(t *testing.T) {
	css := `
@property --color-bg {
  syntax: "<color>";
  inherits: false;
  initial-value: #ffffff;
  mode-dark: #1a1a2e;
}
`
	tokens, err := ParseString(css)
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}

	tok := tokens[0]
	if tok.RawModes != nil {
		t.Errorf("Expected nil RawModes for literal mode values, got %v", tok.RawModes)
	}
}

func TestValidateDuplicates(t *testing.T) {
	tokens := []Token{
		{Name: "--color-a", Source: Source{File: "a.css", Line: 1}},
		{Name: "--color-a", Source: Source{File: "b.css", Line: 5}},
	}

	errors := Validate(tokens)
	if len(errors) == 0 {
		t.Error("Expected duplicate error, got none")
	}
}
