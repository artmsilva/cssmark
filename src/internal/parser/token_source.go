package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var (
	tokenOpen    = regexp.MustCompile(`^\s*@token\s+([^\s{]+)\s*\{\s*$`)
	modeOpen     = regexp.MustCompile(`^\s*@mode\s+([^\s{]+)\s*\{\s*$`)
	fieldLine    = regexp.MustCompile(`^\s*([\w-]+):\s*(.+);\s*$`)
	importRef    = regexp.MustCompile(`@import\s+"([^"]+)"`)
	namespaceRef = regexp.MustCompile(`namespace:\s*([\w-]+)`)
	refValue     = regexp.MustCompile(`ref\(([^)]+)\)`)
)

// ParseTokenSource lowers the compact @token authoring form into ordinary
// tokens. Source entrypoints declare a namespace then import flat token files.
func ParseTokenSource(path string) ([]Token, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	namespace := ""
	if match := namespaceRef.FindStringSubmatch(string(body)); len(match) == 2 {
		namespace = match[1]
	}
	imports := importRef.FindAllStringSubmatch(string(body), -1)
	if len(imports) == 0 {
		return parseTokenText(string(body), filepath.Base(path), namespace)
	}
	var tokens []Token
	for _, match := range imports {
		child := filepath.Join(filepath.Dir(path), match[1])
		content, err := os.ReadFile(child)
		if err != nil {
			return nil, err
		}
		parsed, err := parseTokenText(string(content), match[1], namespace)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, parsed...)
	}
	return tokens, nil
}

type sourceNode struct {
	path   []string
	fields map[string]string
	modes  map[string]map[string]string
}

func parseTokenText(body, filename, namespace string) ([]Token, error) {
	var stack []*sourceNode
	var tokens []Token
	for lineNo, line := range strings.Split(body, "\n") {
		if match := tokenOpen.FindStringSubmatch(line); len(match) == 2 {
			path := []string{match[1]}
			if len(stack) > 0 {
				path = append(append([]string{}, stack[len(stack)-1].path...), match[1])
			}
			stack = append(stack, &sourceNode{path: path, fields: map[string]string{}, modes: map[string]map[string]string{}})
			continue
		}
		if match := modeOpen.FindStringSubmatch(line); len(match) == 2 && len(stack) > 0 {
			stack = append(stack, &sourceNode{path: append(stack[len(stack)-1].path, "@mode:"+match[1]), fields: map[string]string{}, modes: map[string]map[string]string{}})
			continue
		}
		if strings.TrimSpace(line) == "}" && len(stack) > 0 {
			node := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			if strings.HasPrefix(node.path[len(node.path)-1], "@mode:") && len(stack) > 0 {
				stack[len(stack)-1].modes[strings.TrimPrefix(node.path[len(node.path)-1], "@mode:")] = node.fields
			} else if len(node.fields) > 0 {
				tokens = append(tokens, lowerNode(node, namespace, filename, lineNo+1)...)
			}
			continue
		}
		if match := fieldLine.FindStringSubmatch(line); len(match) == 3 && len(stack) > 0 {
			stack[len(stack)-1].fields[match[1]] = match[2]
		}
	}
	return tokens, nil
}

func lowerNode(node *sourceNode, namespace, filename string, line int) []Token {
	name := "--" + namespace + "-" + strings.Join(node.path, "-")
	if value, ok := node.fields["value"]; ok {
		token := Token{Name: name, InitialValue: lowerValue(value, namespace), Modes: map[string]string{}, Source: Source{File: filename, Line: line}}
		for mode, fields := range node.modes {
			if value, ok := fields["value"]; ok {
				token.Modes[mode] = lowerValue(value, namespace)
			}
		}
		return []Token{token}
	}
	// A leaf with arbitrary scalar fields is a state/variant map. This supports
	// default/hover as well as feedback ramps (lighter/middle/darker) without
	// baking a vocabulary of state names into the compiler.
	composite := map[string]bool{"font-family": true, "font-size": true, "font-style": true, "font-weight": true, "line-height": true, "duration": true, "delay": true, "timing-function": true, "extends": true}
	var out []Token
	keys := make([]string, 0, len(node.fields))
	for key := range node.fields {
		if !composite[key] {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	for _, state := range keys {
		value := node.fields[state]
		token := Token{Name: name + "-" + state, InitialValue: lowerValue(value, namespace), Modes: map[string]string{}, Source: Source{File: filename, Line: line}}
		for mode, fields := range node.modes {
			if value, ok := fields[state]; ok {
				token.Modes[mode] = lowerValue(value, namespace)
			}
		}
		out = append(out, token)
	}
	if len(out) > 0 {
		return out
	}
	value, ok := node.fields["value"]
	if !ok {
		return nil
	}
	token := Token{Name: name, InitialValue: lowerValue(value, namespace), Modes: map[string]string{}, Source: Source{File: filename, Line: line}}
	for mode, fields := range node.modes {
		if value, ok := fields["value"]; ok {
			token.Modes[mode] = lowerValue(value, namespace)
		}
	}
	return []Token{token}
}

func lowerValue(value, namespace string) string {
	return refValue.ReplaceAllStringFunc(value, func(match string) string {
		id := strings.TrimSuffix(strings.TrimPrefix(match, "ref("), ")")
		return "var(--" + namespace + "-" + strings.ReplaceAll(id, ".", "-") + ")"
	})
}

func ParseTokenSources(paths []string) ([]Token, error) {
	var all []Token
	for _, path := range paths {
		tokens, err := ParseTokenSource(path)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		all = append(all, tokens...)
	}
	return all, nil
}
