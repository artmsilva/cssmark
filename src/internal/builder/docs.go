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
func WriteDocs(tokens []parser.Token, outDir string) error {
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	// Group tokens by category
	groups := groupByCategory(tokens)

	// Parse templates
	tmpl, err := template.New("index.html").Funcs(template.FuncMap{
		"isColor": isColor,
		"join":    strings.Join,
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

	data := struct {
		Groups     []CategoryGroup
		TotalCount int
	}{
		Groups:     groups,
		TotalCount: len(tokens),
	}

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	// Copy styles
	stylesContent, err := templates.ReadFile("templates/styles.css")
	if err != nil {
		return err
	}

	stylesPath := filepath.Join(outDir, "styles.css")
	if err := os.WriteFile(stylesPath, stylesContent, 0644); err != nil {
		return err
	}

	return nil
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
