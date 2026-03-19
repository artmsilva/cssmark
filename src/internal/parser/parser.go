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
	varRefRe        = regexp.MustCompile(`var\((--[\w-]+)\)`)
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

	// Resolve var() references in initial values
	resolveVarReferences(result)

	return result, nil
}

// resolveVarReferences resolves var(--X) references in InitialValue fields.
// When a token's initial-value contains var(--other-token), the reference is
// chased through the token graph until a literal value is found.
// The original unresolved value is preserved in RawInitialValue for CSS output.
func resolveVarReferences(tokens []Token) {
	byName := make(map[string]*Token)
	for i := range tokens {
		byName[tokens[i].Name] = &tokens[i]
	}

	for i := range tokens {
		if varRefRe.MatchString(tokens[i].InitialValue) {
			tokens[i].RawInitialValue = tokens[i].InitialValue
			tokens[i].InitialValue = resolveVarValue(tokens[i].InitialValue, byName, nil)
		}

		for modeName, modeValue := range tokens[i].Modes {
			if varRefRe.MatchString(modeValue) {
				if tokens[i].RawModes == nil {
					tokens[i].RawModes = make(map[string]string)
				}
				tokens[i].RawModes[modeName] = modeValue
				tokens[i].Modes[modeName] = resolveVarValue(modeValue, byName, nil)
			}
		}
	}
}

func resolveVarValue(value string, byName map[string]*Token, seen map[string]bool) string {
	if seen == nil {
		seen = make(map[string]bool)
	}

	return varRefRe.ReplaceAllStringFunc(value, func(match string) string {
		sub := varRefRe.FindStringSubmatch(match)
		name := sub[1]

		if seen[name] {
			return match // circular reference — keep as-is
		}
		seen[name] = true

		target, exists := byName[name]
		if !exists {
			return match // unknown var — keep as-is
		}

		resolved := target.InitialValue
		if varRefRe.MatchString(resolved) {
			resolved = resolveVarValue(resolved, byName, seen)
		}

		return resolved
	})
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
		default:
			// mode-<name> descriptors (e.g., mode-dark, mode-high-contrast)
			if strings.HasPrefix(key, "mode-") {
				modeName := strings.TrimPrefix(key, "mode-")
				if modeName != "" {
					if token.Modes == nil {
						token.Modes = make(map[string]string)
					}
					token.Modes[modeName] = value
				}
			}
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
