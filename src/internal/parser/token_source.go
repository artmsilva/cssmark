package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var (
	tokenOpen    = regexp.MustCompile(`^\s*@(token|alias)\s+([^\s{]+)\s*\{\s*$`)
	modeOpen     = regexp.MustCompile(`^\s*@mode\s+([^\s{]+)\s*\{\s*$`)
	deriveOpen   = regexp.MustCompile(`^\s*@derive\s+`)
	fieldLine    = regexp.MustCompile(`^\s*([\w-]+):\s*(.+);\s*$`)
	importRef    = regexp.MustCompile(`@import\s+"([^"]+)"`)
	namespaceRef = regexp.MustCompile(`namespace:\s*([\w-]+)`)
	refValue     = regexp.MustCompile(`ref\(([^)]+)\)`)
	scaleBlock   = regexp.MustCompile(`(?s)@scale\s+space\s*\{\s*unit:\s*([\d.]+)rem;\s*steps:\s*([^;]+);(?:\s*sentinel\s+999:\s*([\d.]+)rem;)?\s*\}`)
)

// ParseTokenSource lowers cssmark's compact authoring form. Entry points declare
// a namespace then import the flat token files.
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
	var source strings.Builder
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
		source.Write(content)
	}
	return applyDerivations(tokens, source.String()), nil
}

type sourceNode struct {
	path   []string
	fields map[string]string
	modes  map[string]map[string]string
}

func parseTokenText(body, filename, namespace string) ([]Token, error) {
	var stack []*sourceNode
	var nodes []*sourceNode
	skipDepth := 0
	for _, line := range strings.Split(body, "\n") {
		if skipDepth > 0 {
			skipDepth += strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}
		if deriveOpen.MatchString(line) {
			skipDepth = strings.Count(line, "{") - strings.Count(line, "}")
			continue
		}
		if match := tokenOpen.FindStringSubmatch(line); len(match) == 3 {
			path := []string{match[2]}
			if len(stack) > 0 {
				path = append(append([]string{}, stack[len(stack)-1].path...), match[2])
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
			} else {
				nodes = append(nodes, node)
			}
			continue
		}
		if match := fieldLine.FindStringSubmatch(line); len(match) == 3 && len(stack) > 0 {
			stack[len(stack)-1].fields[match[1]] = match[2]
		}
	}
	byPath := map[string]*sourceNode{}
	for _, node := range nodes {
		byPath[strings.Join(node.path, ".")] = node
	}
	var tokens []Token
	for _, node := range nodes {
		tokens = append(tokens, lowerNode(node, byPath, namespace, filename)...)
	}
	tokens = append(tokens, lowerSpaceScale(body, namespace, filename)...)
	return dedupeTokens(tokens), nil
}

func lowerNode(node *sourceNode, byPath map[string]*sourceNode, namespace, filename string) []Token {
	fields, modes := resolvedFields(node, byPath, nil)
	base := sourceName(namespace, node.path)
	if hasTypography(fields) {
		return lowerTypography(base, fields, modes, filename)
	}
	if hasTransition(fields) {
		return lowerTransition(base, fields, modes, filename)
	}
	if value, ok := fields["value"]; ok {
		t := Token{Name: base, InitialValue: lowerValue(value, namespace), Modes: lowerModes(modes, "value", namespace), Source: Source{File: filename}}
		return []Token{t}
	}
	if len(node.path) > 0 && node.path[0] == "timing-function" {
		var out []Token
		keys := make([]string, 0, len(fields))
		for key := range fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			out = append(out, Token{Name: "--" + namespace + "-timingFunction-" + key, InitialValue: "cubic-bezier(" + fields[key] + ")", Source: Source{File: filename}})
		}
		return out
	}
	// State maps and aliases are arbitrary scalar leaf fields.
	var keys []string
	for key := range fields {
		if key != "extends" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	var out []Token
	for _, key := range keys {
		out = append(out, Token{Name: base + "-" + key, InitialValue: lowerValue(fields[key], namespace), Modes: lowerModes(modes, key, namespace), Source: Source{File: filename}})
	}
	return out
}

func resolvedFields(node *sourceNode, byPath map[string]*sourceNode, seen map[string]bool) (map[string]string, map[string]map[string]string) {
	if seen == nil {
		seen = map[string]bool{}
	}
	key := strings.Join(node.path, ".")
	if seen[key] {
		return node.fields, node.modes
	}
	seen[key] = true
	fields, modes := map[string]string{}, map[string]map[string]string{}
	if parent, ok := node.fields["extends"]; ok {
		path := append(append([]string{}, node.path[:len(node.path)-1]...), parent)
		if base := byPath[strings.Join(path, ".")]; base != nil {
			fields, modes = resolvedFields(base, byPath, seen)
		}
	}
	copyFields := func(from map[string]string) map[string]string {
		to := map[string]string{}
		for k, v := range from {
			to[k] = v
		}
		return to
	}
	fields = copyFields(fields)
	for k, v := range node.fields {
		fields[k] = v
	}
	modeCopy := map[string]map[string]string{}
	for mode, values := range modes {
		modeCopy[mode] = copyFields(values)
	}
	for mode, values := range node.modes {
		if modeCopy[mode] == nil {
			modeCopy[mode] = map[string]string{}
		}
		for k, v := range values {
			modeCopy[mode][k] = v
		}
	}
	return fields, modeCopy
}

func lowerTypography(base string, fields map[string]string, modes map[string]map[string]string, file string) []Token {
	axes := []string{"font-family", "font-size", "font-style", "font-weight", "line-height"}
	var out []Token
	for _, axis := range axes {
		if value, ok := fields[axis]; ok {
			out = append(out, Token{Name: base + "-" + axis, InitialValue: lowerValue(value, "hb"), Modes: lowerModes(modes, axis, "hb"), Source: Source{File: file}})
		}
	}
	if len(out) == 5 {
		out = append(out, Token{Name: base, InitialValue: fmt.Sprintf("var(%s-font-style) var(%s-font-weight) var(%s-font-size)/var(%s-line-height) var(%s-font-family)", base, base, base, base, base), Source: Source{File: file}})
	}
	return out
}
func lowerTransition(base string, fields map[string]string, modes map[string]map[string]string, file string) []Token {
	axes := []string{"duration", "delay", "timing-function"}
	var out []Token
	for _, axis := range axes {
		if value, ok := fields[axis]; ok {
			out = append(out, Token{Name: base + "-" + axis, InitialValue: lowerValue(value, "hb"), Modes: lowerModes(modes, axis, "hb"), Source: Source{File: file}})
		}
	}
	if len(out) == 3 {
		out = append(out, Token{Name: base, InitialValue: fmt.Sprintf("var(%s-duration) var(%s-delay) var(%s-timing-function)", base, base, base), Source: Source{File: file}})
	}
	return out
}
func lowerModes(modes map[string]map[string]string, field, namespace string) map[string]string {
	out := map[string]string{}
	for mode, fields := range modes {
		if value, ok := fields[field]; ok {
			out[mode] = lowerValue(value, namespace)
		}
	}
	return out
}
func hasTypography(fields map[string]string) bool { _, ok := fields["font-size"]; return ok }
func hasTransition(fields map[string]string) bool { _, ok := fields["duration"]; return ok }

func sourceName(namespace string, path []string) string {
	parts := append([]string{}, path...)
	if len(parts) > 0 && parts[0] == "typography" {
		parts = parts[1:]
		// Cobalt puts `type` after semantic roots but before top-level type paths.
		if len(parts) == 0 || parts[0] != "type" {
			if len(parts) > 0 && (parts[0] == "decorative" || parts[0] == "action") {
				parts = append(parts[:1], append([]string{"type"}, parts[1:]...)...)
			} else {
				parts = append([]string{"type"}, parts...)
			}
		}
	}
	if len(parts) > 0 && parts[0] == "dimension" {
		parts = parts[1:]
	}
	if len(parts) > 0 && parts[0] == "timing-function" {
		parts[0] = "timingFunction"
	}
	return "--" + namespace + "-" + strings.Join(parts, "-")
}
func lowerValue(value, namespace string) string {
	return refValue.ReplaceAllStringFunc(value, func(match string) string {
		id := strings.TrimSuffix(strings.TrimPrefix(match, "ref("), ")")
		return "var(--" + namespace + "-" + strings.ReplaceAll(id, ".", "-") + ")"
	})
}

func lowerSpaceScale(body, namespace, file string) []Token {
	match := scaleBlock.FindStringSubmatch(body)
	if len(match) == 0 {
		return nil
	}
	unit, _ := strconv.ParseFloat(match[1], 64)
	var out []Token
	for _, raw := range strings.Split(match[2], ",") {
		step := strings.TrimSpace(raw)
		n, _ := strconv.ParseFloat(step, 64)
		id := strings.ReplaceAll(strings.TrimPrefix(step, "."), ".", "-")
		if strings.HasPrefix(step, ".") {
			id = "0-" + id
		}
		out = append(out, Token{Name: "--" + namespace + "-space-" + id, InitialValue: strconv.FormatFloat(unit*n, 'f', -1, 64) + "rem", Source: Source{File: file}})
	}
	if len(match) > 3 && match[3] != "" {
		out = append(out, Token{Name: "--" + namespace + "-space-999", InitialValue: match[3] + "rem", Source: Source{File: file}})
	}
	return out
}
func applyDerivations(tokens []Token, source string) []Token {
	if !strings.Contains(source, "@derive dense") {
		return tokens
	}
	byName := map[string]*Token{}
	for i := range tokens {
		byName[tokens[i].Name] = &tokens[i]
	}
	for i := range tokens {
		token := &tokens[i]
		if token.Modes == nil {
			token.Modes = map[string]string{}
		}
		if strings.Contains(token.Name, "-type-") {
			if strings.HasSuffix(token.Name, "-font-size") {
				token.Modes["dense"] = stepDown(token.InitialValue, byName, "--hb-space-3")
			}
			if strings.HasSuffix(token.Name, "-line-height") {
				token.Modes["dense"] = stepDown(token.InitialValue, byName, "--hb-space-4")
			}
			// Cobalt emits every typography axis in wireframe, even unchanged axes.
			if strings.HasSuffix(token.Name, "-font-size") || strings.HasSuffix(token.Name, "-font-style") || strings.HasSuffix(token.Name, "-font-weight") || strings.HasSuffix(token.Name, "-line-height") {
				token.Modes["wireframe"] = token.InitialValue
			}
		}
		if strings.Contains(token.Name, "-decorative-space-") {
			token.Modes["dense"] = denseSpace(token.InitialValue, byName)
		}
	}
	// Rebuild a shorthand after modes are known. A direct `font` declaration is
	// required in every mode because custom-property references resolve where
	// assigned, not where a consuming font declaration is evaluated.
	for i := range tokens {
		if strings.Contains(tokens[i].Name, "-type-") && !strings.Contains(tokens[i].Name, "-font-") && !strings.HasSuffix(tokens[i].Name, "-line-height") {
			if tokens[i].Modes == nil {
				tokens[i].Modes = map[string]string{}
			}
			base := tokens[i].Name
			for _, mode := range []string{"dense", "wireframe"} {
				family := modeValue(byName[base+"-font-family"], mode)
				size := modeValue(byName[base+"-font-size"], mode)
				style := modeValue(byName[base+"-font-style"], mode)
				weight := modeValue(byName[base+"-font-weight"], mode)
				height := modeValue(byName[base+"-line-height"], mode)
				if family != "" && size != "" && style != "" && weight != "" && height != "" {
					tokens[i].Modes[mode] = style + " " + weight + " " + size + "/" + height + " " + family
				}
			}
		}
	}
	return tokens
}
func modeValue(token *Token, mode string) string {
	if token == nil {
		return ""
	}
	if value, ok := token.Modes[mode]; ok {
		return value
	}
	return token.InitialValue
}

func stepDown(value string, byName map[string]*Token, floor string) string {
	match := regexp.MustCompile(`--hb-space-([\d-]+)`).FindStringSubmatch(value)
	if len(match) != 2 {
		return value
	}
	ladder := []string{"0", "0-25", "0-5", "1", "1-5", "2", "2-5", "3", "3-5", "4", "4-5", "5", "5-5", "6", "6-5", "7", "7-5", "8", "8-5", "9", "9-5", "10", "11", "12", "13", "14", "15", "16", "18", "20", "22", "24", "26", "28", "30", "35", "40", "45", "50", "100"}
	index := map[string]int{}
	for i, x := range ladder {
		index[x] = i
	}
	current, ok := index[match[1]]
	if !ok {
		return value
	}
	floorID := strings.TrimPrefix(floor, "--hb-space-")
	target := index[floorID]
	if current <= target {
		return value
	}
	return "var(--hb-space-" + ladder[current-1] + ")"
}
func denseSpace(value string, byName map[string]*Token) string {
	return stepDown(stepDown(value, byName, "--hb-space-1"), byName, "--hb-space-1")
}

func dedupeTokens(tokens []Token) []Token {
	seen := map[string]bool{}
	out := make([]Token, 0, len(tokens))
	for _, t := range tokens {
		if !seen[t.Name] {
			seen[t.Name] = true
			out = append(out, t)
		}
	}
	return out
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
