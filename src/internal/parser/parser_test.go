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
