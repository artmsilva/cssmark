package parser

import (
	"os"
	"regexp"
	"strconv"
	"strings"
)

var (
	propertyBlockRe = regexp.MustCompile(`@property\s+(--[\w-]+)\s*\{([^}]+)\}`)
	descriptorRe    = regexp.MustCompile(`([\w-]+)\s*:\s*("[^"]*"|[^;]+);?`)
	rootBlockRe     = regexp.MustCompile(`:root\s*\{([^}]+)\}`)
	customPropRe    = regexp.MustCompile(`(--[\w-]+)\s*:\s*([^;]+);`)
)

// ParseFiles parses multiple CSS files and returns all tokens
func ParseFiles(paths []string) ([]Token, error) {
	var allTokens []Token

	for _, path := range paths {
		tokens, err := ParseFile(path)
		if err != nil {
			return nil, err
		}
		allTokens = append(allTokens, tokens...)
	}

	return allTokens, nil
}

// ParseFile parses a single CSS file and returns tokens
func ParseFile(path string) ([]Token, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return Parse(string(content), path)
}

// Parse parses CSS content and returns tokens
func Parse(css string, filename string) ([]Token, error) {
	tokens := make(map[string]*Token)

	// Parse @property blocks
	matches := propertyBlockRe.FindAllStringSubmatchIndex(css, -1)
	for _, match := range matches {
		name := css[match[2]:match[3]]
		body := css[match[4]:match[5]]
		line := countLines(css[:match[0]]) + 1

		token := &Token{
			Name: name,
			Source: Source{
				File: filename,
				Line: line,
			},
		}

		parseDescriptors(body, token)
		tokens[name] = token
	}

	// Parse :root blocks for runtime values
	rootMatches := rootBlockRe.FindAllStringSubmatch(css, -1)
	for _, match := range rootMatches {
		body := match[1]
		propMatches := customPropRe.FindAllStringSubmatch(body, -1)
		for _, pm := range propMatches {
			name := pm[1]
			value := strings.TrimSpace(pm[2])

			if token, exists := tokens[name]; exists {
				// :root value overrides initial-value
				token.InitialValue = value
			}
		}
	}

	// Convert map to slice
	result := make([]Token, 0, len(tokens))
	for _, token := range tokens {
		result = append(result, *token)
	}

	return result, nil
}

func parseDescriptors(body string, token *Token) {
	matches := descriptorRe.FindAllStringSubmatch(body, -1)

	for _, match := range matches {
		key := strings.TrimSpace(match[1])
		value := strings.TrimSpace(match[2])
		value = strings.Trim(value, `"`)

		switch key {
		case "syntax":
			token.Syntax = value
		case "inherits":
			token.Inherits = value == "true"
		case "initial-value":
			token.InitialValue = value
		case "description":
			token.Description = value
		case "category":
			token.Category = value
		case "type":
			token.Type = value
		case "aliases":
			aliases := strings.Split(value, ",")
			for i, a := range aliases {
				aliases[i] = strings.TrimSpace(a)
			}
			token.Aliases = aliases
		case "deprecated":
			token.Deprecated = value == "true"
		case "examples":
			examples := strings.Split(value, ";")
			var cleaned []string
			for _, e := range examples {
				e = strings.TrimSpace(e)
				if e != "" {
					cleaned = append(cleaned, e)
				}
			}
			token.Examples = cleaned
		}
	}
}

func countLines(s string) int {
	return strings.Count(s, "\n")
}

// ParseString is a convenience function for parsing CSS from a string
func ParseString(css string) ([]Token, error) {
	return Parse(css, "<string>")
}

// MustParse parses CSS and panics on error (useful for tests)
func MustParse(css string) []Token {
	tokens, err := ParseString(css)
	if err != nil {
		panic(err)
	}
	return tokens
}

// ParseInt safely parses an integer with a default value
func ParseInt(s string, def int) int {
	if v, err := strconv.Atoi(s); err == nil {
		return v
	}
	return def
}
