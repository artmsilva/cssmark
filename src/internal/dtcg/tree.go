package dtcg

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// WriteFlatCSS writes an easy-to-browse cssmark authoring tree. Each leaf is a
// normal @property source file; index.css is only the flat composition order.
func WriteFlatCSS(tokens []Token, directory string) error {
	files := map[string]*strings.Builder{}
	for _, token := range tokens {
		file := sourceFile(token)
		if files[file] == nil {
			files[file] = &strings.Builder{}
			files[file].WriteString("/* cssmark token source */\n\n")
		}
		block, err := PropertySource(token)
		if err != nil {
			return err
		}
		files[file].WriteString(block)
	}
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	var names []string
	for name, body := range files {
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

func declarations(name string, value any) []string {
	switch value := value.(type) {
	case string:
		return []string{fmt.Sprintf("  %s: %s;", name, cssValue(value))}
	case map[string]any:
		if isTypography(value) {
			family := cssComposite(value["fontFamily"])
			size, style := cssComposite(value["fontSize"]), cssComposite(value["fontStyle"])
			weight, height := cssComposite(value["fontWeight"]), cssComposite(value["lineHeight"])
			return []string{
				fmt.Sprintf("  %s: %s %s %s/%s %s;", name, style, weight, size, height, family),
				fmt.Sprintf("  %s-font-family: %s;", name, family),
				fmt.Sprintf("  %s-font-size: %s;", name, size),
				fmt.Sprintf("  %s-font-style: %s;", name, style),
				fmt.Sprintf("  %s-font-weight: %s;", name, weight),
				fmt.Sprintf("  %s-line-height: %s;", name, height),
			}
		}
		if duration, ok := value["duration"]; ok {
			return []string{fmt.Sprintf("  %s: %s %s %s;", name, cssComposite(duration), cssComposite(value["delay"]), cssComposite(value["timingFunction"]))}
		}
	}
	return nil
}

func mergeComposite(base, override any) any {
	baseMap, baseOK := base.(map[string]any)
	overrideMap, overrideOK := override.(map[string]any)
	if !baseOK || !overrideOK {
		return override
	}
	merged := map[string]any{}
	for key, value := range baseMap {
		merged[key] = value
	}
	for key, value := range overrideMap {
		merged[key] = value
	}
	return merged
}
func sourceFile(token Token) string {
	// Group by the CSS concept authors look for, not raw DTCG wire types.
	switch token.Type {
	case "", "string":
		return "mode.css"
	case "cubicBezier":
		return "timing-function.css"
	case "fontFamily":
		return "font-family.css"
	case "fontWeight":
		return "font-weight.css"
	default:
		return strings.ReplaceAll(token.Type, "/", "-") + ".css"
	}
}
func modeSelector(mode string) string {
	switch mode {
	case "wireframe":
		return "[data-color-mode=\"wireframe\"]"
	case "dense":
		return ":root[data-density=\"dense\"], [data-density=\"dense\"]"
	default:
		return fmt.Sprintf(":root[data-color-mode=\"%s\"]", mode)
	}
}
func uniqueSorted(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	sort.Strings(out)
	return out
}
func layerFor(file string) string {
	if strings.HasPrefix(file, "mode-") {
		return "hb-tokens.modes"
	}
	return "hb-tokens.base"
}
func writeLayer(path, layer, selector string, declarations []string) error {
	body := fmt.Sprintf("@layer %s {\n%s {\n%s\n}\n}\n", layer, selector, strings.Join(declarations, "\n"))
	return os.WriteFile(path, []byte(body), 0644)
}
