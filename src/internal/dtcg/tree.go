package dtcg

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
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
		if token.Type == "typography-shorthand" || token.ID == "mode" || strings.HasPrefix(token.ID, "space.") {
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
	if files["dimension"] == nil {
		files["dimension"] = &tokenGroup{children: map[string]*tokenGroup{}}
	}
	for typ, root := range files {
		name := typ + ".css"
		var body strings.Builder
		body.WriteString("/* cssmark token source */\n\n")
		body.WriteString("@token " + typ + " {\n")
		if root.token != nil {
			writeTokenFields(&body, *root.token, 1, nil, "")
		}
		root.write(&body, 1, typ, "", nil)
		writeDenseDerivation(&body, typ, root)
		body.WriteString("}\n")
		if typ == "dimension" {
			body.WriteString("\n@scale space {\n  unit: 0.25rem;\n  steps: 0, .25, .5, 1, 1.5, 2, 2.5, 3, 3.5, 4, 4.5, 5, 5.5, 6, 6.5, 7, 7.5, 8, 8.5, 9, 9.5, 10, 11, 12, 13, 14, 15, 16, 18, 20, 22, 24, 26, 28, 30, 35, 40, 45, 50, 100;\n  sentinel 999: 249.75rem;\n}\n")
		}
		if err := os.WriteFile(filepath.Join(directory, name), []byte(body.String()), 0644); err != nil {
			return err
		}
		names = append(names, name)
	}
	sort.Strings(names)
	var index strings.Builder
	index.WriteString("@tokens {\n  namespace: hb;\n}\n\n")
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
	if len(path) == 0 {
		group.token = &token
		return
	}
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

func (group *tokenGroup) write(out *strings.Builder, indent int, typ, inheritedAll string, path []string) {
	keys := make([]string, 0, len(group.children))
	for key := range group.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if len(keys) > 1 && keys[0] != "regular" {
		for i, key := range keys {
			if key == "regular" {
				keys[0], keys[i] = keys[i], keys[0]
				break
			}
		}
	}
	if all := commonScalarMode(group); all != "" {
		out.WriteString(strings.Repeat("  ", indent) + "@mode wireframe { all: " + all + "; }\n")
		inheritedAll = "wireframe"
	}
	var regular *Token
	if typ == "typography" && group.children["regular"] != nil {
		regular = group.children["regular"].token
	}
	for _, key := range keys {
		child := group.children[key]
		childPath := append(append([]string{}, path...), key)
		pad := strings.Repeat("  ", indent)
		if isAliasGroup(typ, childPath, child) {
			out.WriteString(pad + "@alias " + key + " {\n")
			writeAliases(out, child, indent+1)
			out.WriteString(pad + "}\n")
			continue
		}
		out.WriteString(pad + "@token " + key + " {\n")
		if isStateMap(child) {
			writeStateMap(out, child, indent+1)
			out.WriteString(pad + "}\n")
			continue
		}
		if child.token != nil {
			parent := (*Token)(nil)
			if child.token != regular && inheritsTypography(*child.token, regular) {
				parent = regular
			}
			writeTokenFields(out, *child.token, indent+1, parent, inheritedAll)
		}
		child.write(out, indent+1, typ, inheritedAll, childPath)
		out.WriteString(pad + "}\n")
	}
}

func writeTokenFields(out *strings.Builder, token Token, indent int, parent *Token, inheritedAll string) {
	pad := strings.Repeat("  ", indent)
	if parent != nil {
		out.WriteString(pad + "extends: regular;\n")
	}
	if value, ok := token.Value.(map[string]any); ok {
		keys := []string{"fontFamily", "fontSize", "fontStyle", "fontWeight", "lineHeight", "duration", "delay", "timingFunction"}
		parentValue, _ := func() (map[string]any, bool) {
			if parent == nil {
				return nil, false
			}
			value, ok := parent.Value.(map[string]any)
			return value, ok
		}()
		for _, key := range keys {
			if field, exists := value[key]; exists && !reflect.DeepEqual(field, parentValue[key]) {
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
	baseFields, composite := token.Value.(map[string]any)
	for _, mode := range modes {
		if mode == "dense" || mode == inheritedAll {
			continue
		}
		value := token.Modes[mode]
		if composite {
			fields, ok := value.(map[string]any)
			if !ok {
				continue
			}
			keys := make([]string, 0, len(fields))
			parentMode := map[string]any{}
			if parent != nil {
				parentMode, _ = parent.Modes[mode].(map[string]any)
			}
			for key, field := range fields {
				if !reflect.DeepEqual(field, baseFields[key]) && !reflect.DeepEqual(field, parentMode[key]) {
					keys = append(keys, key)
				}
			}
			if len(keys) == 0 {
				continue
			}
			sort.Strings(keys)
			out.WriteString(fmt.Sprintf("%s@mode %s {\n", pad, mode))
			for _, key := range keys {
				out.WriteString(fmt.Sprintf("%s  %s: %s;\n", pad, cssField(key), cssComposite(fields[key])))
			}
			out.WriteString(pad + "}\n")
			continue
		}
		if reflect.DeepEqual(value, token.Value) {
			continue
		}
		out.WriteString(fmt.Sprintf("%s@mode %s {\n%s  value: %s;\n%s}\n", pad, mode, pad, cssComposite(value), pad))
	}
}

func isAliasGroup(typ string, path []string, group *tokenGroup) bool {
	return typ == "dimension" && strings.Join(path, ".") == "decorative.space" && allScalarLeaves(group)
}

func allScalarLeaves(group *tokenGroup) bool {
	if group.token != nil {
		_, composite := group.token.Value.(map[string]any)
		return !composite
	}
	if len(group.children) == 0 {
		return false
	}
	for _, child := range group.children {
		if !allScalarLeaves(child) {
			return false
		}
	}
	return true
}

func writeAliases(out *strings.Builder, group *tokenGroup, indent int) {
	keys := make([]string, 0, len(group.children))
	for key := range group.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pad := strings.Repeat("  ", indent)
	for _, key := range keys {
		token := group.children[key].token
		if token == nil {
			continue
		}
		out.WriteString(fmt.Sprintf("%s%s: %s;\n", pad, key, reference(token.Value)))
	}
}

func isStateMap(group *tokenGroup) bool {
	if group.token != nil || len(group.children) < 2 {
		return false
	}
	for _, child := range group.children {
		if child.token == nil || len(child.children) != 0 {
			return false
		}
		if _, composite := child.token.Value.(map[string]any); composite {
			return false
		}
	}
	// `100: value` is not valid CSS declaration syntax, so palette/scale stops
	// stay as nested @token blocks. This keeps editor parsing/highlighting intact.
	for name := range group.children {
		if name[0] >= '0' && name[0] <= '9' {
			return false
		}
	}
	return true
}

func writeStateMap(out *strings.Builder, group *tokenGroup, indent int) {
	keys := make([]string, 0, len(group.children))
	for key := range group.children {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	pad := strings.Repeat("  ", indent)
	for _, key := range keys {
		out.WriteString(fmt.Sprintf("%s%s: %s;\n", pad, key, reference(group.children[key].token.Value)))
	}
	modes := map[string]bool{}
	for _, child := range group.children {
		for mode := range child.token.Modes {
			modes[mode] = true
		}
	}
	modeNames := make([]string, 0, len(modes))
	for mode := range modes {
		modeNames = append(modeNames, mode)
	}
	sort.Strings(modeNames)
	for _, mode := range modeNames {
		var overrides []string
		for _, key := range keys {
			child := group.children[key].token
			value, ok := child.Modes[mode]
			if !ok || reflect.DeepEqual(value, child.Value) {
				continue
			}
			overrides = append(overrides, fmt.Sprintf("%s  %s: %s;\n", pad, key, reference(value)))
		}
		if len(overrides) > 0 {
			out.WriteString(fmt.Sprintf("%s@mode %s {\n%s%s}\n", pad, mode, strings.Join(overrides, ""), pad))
		}
	}
}

func reference(value any) string {
	if raw, ok := value.(string); ok {
		if strings.HasPrefix(raw, "{") && strings.HasSuffix(raw, "}") {
			return "ref(" + strings.TrimSuffix(strings.TrimPrefix(raw, "{"), "}") + ")"
		}
	}
	return cssComposite(value)
}

func commonScalarMode(group *tokenGroup) string {
	if len(group.children) < 2 {
		return ""
	}
	var value any
	for _, child := range group.children {
		if child.token == nil || len(child.children) != 0 {
			return ""
		}
		candidate, ok := child.token.Modes["wireframe"]
		if !ok {
			return ""
		}
		if value == nil {
			value = candidate
			continue
		}
		if !reflect.DeepEqual(value, candidate) {
			return ""
		}
	}
	if scalar, ok := value.(string); ok {
		return cssComposite(scalar)
	}
	return ""
}

func inheritsTypography(token Token, regular *Token) bool {
	if regular == nil {
		return false
	}
	value, ok := token.Value.(map[string]any)
	if !ok {
		return false
	}
	base, ok := regular.Value.(map[string]any)
	if !ok {
		return false
	}
	// A sibling is worthwhile to inherit when it differs only by the conventional
	// weight/style axes; retain explicit special typography variants.
	for _, key := range []string{"fontFamily", "fontSize", "lineHeight"} {
		if !reflect.DeepEqual(value[key], base[key]) {
			return false
		}
	}
	return true
}

func writeDenseDerivation(out *strings.Builder, typ string, root *tokenGroup) {
	if typ == "typography" {
		out.WriteString("\n  @derive dense {\n    font-size: step-down(1, floor: var(--hb-space-3));\n    line-height: step-down(1, floor: var(--hb-space-4));\n  }\n")
	}
	if typ == "dimension" && hasSemanticSpace(root, nil) {
		out.WriteString("\n  @derive dense decorative.space {\n    value: step-down(2, floor: var(--hb-space-1));\n    override: --hb-space-2 => var(--hb-space-1-5);\n  }\n")
	}
}

func hasSemanticSpace(group *tokenGroup, path []string) bool {
	if group.token != nil && strings.HasPrefix(strings.Join(path, "."), "decorative.space.") {
		return true
	}
	for name, child := range group.children {
		if hasSemanticSpace(child, append(path, name)) {
			return true
		}
	}
	return false
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
	if len(parts) > 0 && parts[0] == typ {
		parts = parts[1:]
	}
	if typ == "typography" && len(parts) > 1 && (parts[0] == "action" || parts[0] == "decorative") && parts[1] == "type" {
		// `action.type.*` and `decorative.type.*` retain their legacy runtime
		// segment through lowering; source avoids repeating it under typography.
		parts = append(parts[:1], parts[2:]...)
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
