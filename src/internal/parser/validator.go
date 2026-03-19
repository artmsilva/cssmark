package parser

import (
	"fmt"
)

// Validate checks tokens for common issues
func Validate(tokens []Token) []string {
	var errors []string
	seen := make(map[string]Source)

	for _, token := range tokens {
		// Check for duplicate names
		if prev, exists := seen[token.Name]; exists {
			errors = append(errors, fmt.Sprintf(
				"duplicate token %s: first in %s:%d, again in %s:%d",
				token.Name, prev.File, prev.Line, token.Source.File, token.Source.Line,
			))
		}
		seen[token.Name] = token.Source

		// Check for missing syntax
		if token.Syntax == "" {
			errors = append(errors, fmt.Sprintf(
				"token %s missing syntax descriptor (%s:%d)",
				token.Name, token.Source.File, token.Source.Line,
			))
		}

		// Check for missing initial-value
		if token.InitialValue == "" {
			errors = append(errors, fmt.Sprintf(
				"token %s missing initial-value (%s:%d)",
				token.Name, token.Source.File, token.Source.Line,
			))
		}
	}

	return errors
}

// ValidateStrict performs additional checks
func ValidateStrict(tokens []Token) []string {
	errors := Validate(tokens)

	for _, token := range tokens {
		// Warn about missing descriptions
		if token.Description == "" {
			errors = append(errors, fmt.Sprintf(
				"token %s missing description (%s:%d)",
				token.Name, token.Source.File, token.Source.Line,
			))
		}

		// Warn about missing categories
		if token.Category == "" {
			errors = append(errors, fmt.Sprintf(
				"token %s missing category (%s:%d)",
				token.Name, token.Source.File, token.Source.Line,
			))
		}
	}

	return errors
}
