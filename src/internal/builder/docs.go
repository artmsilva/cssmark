package builder

import (
	"embed"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/artmsilva/cssmark/src/internal/parser"
)

//go:embed templates/*
var templates embed.FS

// CategoryGroup holds tokens grouped by category
type CategoryGroup struct {
	Name   string
	Tokens []parser.Token
}

// WriteDocs generates a static documentation site
func WriteDocs(tokens []parser.Token, outDir string, sourceFiles ...string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Group tokens by category
	groups := groupByCategory(tokens)

	// Parse templates
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"isColor":    isColor,
		"isSpacing":  isSpacing,
		"isRadius":   isRadius,
		"isFont":     isFont,
		"isFontSize": isFontSize,
		"isShadow":   isShadow,
		"join":       strings.Join,
		"tokenID":    tokenID,
		"safeCSS":    func(s string) template.CSS { return template.CSS(s) },
	}).ParseFS(templates, "templates/index.html")
	if err != nil {
		return err
	}

	// Write index.html
	indexPath := filepath.Join(outDir, "index.html")
	f, err := os.Create(indexPath)
	if err != nil {
		return err
	}
	defer f.Close()

	// Collect unique mode names from tokens
	modeSet := make(map[string]bool)
	deprecatedCount := 0
	for _, t := range tokens {
		if t.Deprecated {
			deprecatedCount++
		}
		for m := range t.Modes {
			modeSet[m] = true
		}
	}
	var modes []string
	for m := range modeSet {
		modes = append(modes, m)
	}
	sort.Strings(modes)

	sourceCSS, readableSourceFiles, err := readSourceCSS(sourceFiles)
	if err != nil {
		return err
	}

	data := struct {
		Groups          []CategoryGroup
		TotalCount      int
		Modes           []string
		SourceFiles     []string
		SourceCSS       string
		DeprecatedCount int
	}{
		Groups:          groups,
		TotalCount:      len(tokens),
		Modes:           modes,
		SourceFiles:     readableSourceFiles,
		SourceCSS:       sourceCSS,
		DeprecatedCount: deprecatedCount,
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	// Generate tokens.css with :root block containing all token values
	tokensPath := filepath.Join(outDir, "tokens.css")
	if err := writeTokensCSS(tokens, tokensPath); err != nil {
		return err
	}

	// Create .nojekyll for GitHub Pages
	nojekyllPath := filepath.Join(outDir, ".nojekyll")
	if err := os.WriteFile(nojekyllPath, []byte{}, 0644); err != nil {
		return err
	}

	return nil
}

// writeTokensCSS writes tokens as CSS using the full ToCSS builder (includes mode overrides)
func writeTokensCSS(tokens []parser.Token, outPath string) error {
	return os.WriteFile(outPath, []byte(ToCSS(tokens)), 0644)
}

func readSourceCSS(files []string) (string, []string, error) {
	var parts []string
	var readableFiles []string
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return "", nil, err
		}

		displayName := filepath.Clean(file)
		if filepath.IsAbs(displayName) {
			displayName = filepath.Base(displayName)
		}
		readableFiles = append(readableFiles, displayName)
		if len(files) > 1 {
			parts = append(parts, "/* "+sanitizeCSSComment(displayName)+" */\n"+string(content))
		} else {
			parts = append(parts, string(content))
		}
	}
	return strings.Join(parts, "\n\n"), readableFiles, nil
}

func sanitizeCSSComment(value string) string {
	return strings.ReplaceAll(value, "*/", "* /")
}

func tokenID(name string) string {
	id := strings.TrimPrefix(name, "--")
	if id == "" {
		return "token"
	}
	var b strings.Builder
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r + ('a' - 'A'))
		default:
			b.WriteRune('-')
		}
	}
	return "token-" + strings.Trim(b.String(), "-")
}

func groupByCategory(tokens []parser.Token) []CategoryGroup {
	groups := make(map[string][]parser.Token)

	for _, token := range tokens {
		cat := token.Category
		if cat == "" {
			cat = "uncategorized"
		}
		groups[cat] = append(groups[cat], token)
	}

	// Convert to slice and sort
	var result []CategoryGroup
	for name, tokens := range groups {
		result = append(result, CategoryGroup{
			Name:   name,
			Tokens: tokens,
		})
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})

	return result
}

func isColor(token parser.Token) bool {
	if token.Type == "color" {
		return true
	}
	if strings.Contains(token.Syntax, "<color>") {
		return true
	}
	return false
}

func isSpacing(token parser.Token) bool {
	if strings.Contains(token.Category, "spacing") {
		return true
	}
	if strings.HasPrefix(token.Name, "--space") {
		return true
	}
	return false
}

func isRadius(token parser.Token) bool {
	if strings.Contains(token.Category, "radius") {
		return true
	}
	if strings.Contains(token.Name, "radius") {
		return true
	}
	return false
}

func isFont(token parser.Token) bool {
	if token.Type == "font" {
		return true
	}
	if strings.Contains(token.Category, "typography.family") {
		return true
	}
	if strings.Contains(token.Name, "--font-sans") || strings.Contains(token.Name, "--font-mono") || strings.Contains(token.Name, "--font-serif") {
		return true
	}
	return false
}

func isFontSize(token parser.Token) bool {
	if strings.Contains(token.Category, "typography.size") {
		return true
	}
	if strings.Contains(token.Name, "font-size") {
		return true
	}
	return false
}

func isShadow(token parser.Token) bool {
	if token.Type == "shadow" {
		return true
	}
	if strings.Contains(token.Category, "effects") && strings.Contains(token.Name, "shadow") {
		return true
	}
	if strings.Contains(token.Name, "shadow") {
		return true
	}
	return false
}
