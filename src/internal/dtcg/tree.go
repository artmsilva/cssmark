package dtcg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WriteFlatCSS writes flat, type-grouped cssmark source. The custom @token
// grouping is authoring syntax; nesting mirrors semantic paths without forcing
// every path segment into a separate file.
func WriteFlatCSS(tokens []Token, directory string) error {
	files := map[string]*tokenGroup{}
	for _, token := range tokens {
		// Cobalt's type-sh tokens duplicate regular typography only to produce a
		// font shorthand. cssmark derives that shorthand from typography axes.
		if token.Type == "typography-shorthand" {
			continue
		}
		typ := sourceType(token)
		if files[typ] == nil {
			files[typ] = &tokenGroup{children: map[string]*tokenGroup{}}
		}
		files[typ].add(sourcePath(token, typ), token)
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	var names []string
	for typ, root := range files {
		name := typ + ".css"
		var body strings.Builder
		body.WriteString("/* cssmark token source */\n\n")
		body.WriteString("@token " + typ + " {\n")
		root.write(&body, 1)
		body.WriteString("}\n")
		if err := os.WriteFile(filepath.Join(directory, name), []byte(body.String()), 0644); err != nil {
			return err
		}
		names = append(names, name)
	}
	sort.Strings(names)
	var index strings.Builder
	for _, name := range names {
		index.WriteString(fmt.Sprintf("@import \"./%s\";\n", name))
	}
	return os.WriteFile(filepath.Join(directory, "index.css"), []byte(index.String()), 0644)
}

type tokenGroup struct {
	children map[string]*tokenGroup
	token    *Token
}

func (group *tokenGroup) add(path []string, token Token) {
	part := path[0]
	child := group.children[part]
	if child == nil {
		child = &tokenGroup{children: map[string]*tokenGroup{}}
		group.children[part] = child
	}
	if len(path) == 1 {
		child.token = &token
		return
	}
	child.add(path[1:], token)
}

func (group *tokenGroup) write(out *strings.Builder, indent int) {
	keys := make([]string, 0, len(group.children))
	for key := range group.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		child := group.children[key]
		pad := strings.Repeat("  ", indent)
		out.WriteString(pad + "@token " + key + " {\n")
		if child.token != nil {
			writeTokenFields(out, *child.token, indent+1)
		}
		child.write(out, indent+1)
		out.WriteString(pad + "}\n")
	}
}

func writeTokenFields(out *strings.Builder, token Token, indent int) {
	pad := strings.Repeat("  ", indent)
	if value, ok := token.Value.(map[string]any); ok {
		keys := []string{"fontFamily", "fontSize", "fontStyle", "fontWeight", "lineHeight", "duration", "delay", "timingFunction"}
		for _, key := range keys {
			if field, exists := value[key]; exists {
				out.WriteString(fmt.Sprintf("%s%s: %s;\n", pad, cssField(key), cssComposite(field)))
			}
		}
	} else {
		out.WriteString(fmt.Sprintf("%svalue: %s;\n", pad, cssComposite(token.Value)))
	}
	modes := make([]string, 0, len(token.Modes))
	for mode := range token.Modes {
		modes = append(modes, mode)
	}
	sort.Strings(modes)
	for _, mode := range modes {
		out.WriteString(fmt.Sprintf("%s@mode %s {\n", pad, mode))
		value := token.Modes[mode]
		if fields, ok := value.(map[string]any); ok {
			keys := make([]string, 0, len(fields))
			for key := range fields {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				out.WriteString(fmt.Sprintf("%s  %s: %s;\n", pad, cssField(key), cssComposite(fields[key])))
			}
		} else {
			out.WriteString(fmt.Sprintf("%s  value: %s;\n", pad, cssComposite(value)))
		}
		out.WriteString(pad + "}\n")
	}
}

func sourceType(token Token) string {
	switch token.Type {
	case "", "string":
		return "mode"
	case "cubicBezier":
		return "timing-function"
	case "fontFamily":
		return "font-family"
	case "fontWeight":
		return "font-weight"
	default:
		return strings.ReplaceAll(token.Type, "/", "-")
	}
}

func sourcePath(token Token, typ string) []string {
	parts := strings.Split(token.ID, ".")
	if len(parts) > 1 && parts[0] == typ {
		parts = parts[1:]
	}
	if typ == "typography" {
		for i, part := range parts {
			if part == "type" {
				parts = append(parts[:i], parts[i+1:]...)
				break
			}
		}
	}
	return parts
}

func cssField(key string) string {
	switch key {
	case "fontFamily":
		return "font-family"
	case "fontSize":
		return "font-size"
	case "fontStyle":
		return "font-style"
	case "fontWeight":
		return "font-weight"
	case "lineHeight":
		return "line-height"
	case "timingFunction":
		return "timing-function"
	default:
		return key
	}
}
