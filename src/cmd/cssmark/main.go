package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/artmsilva/cssmark/src/internal/builder"
	"github.com/artmsilva/cssmark/src/internal/differ"
	"github.com/artmsilva/cssmark/src/internal/dtcg"
	"github.com/artmsilva/cssmark/src/internal/parser"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
	case "css":
		runCSS(os.Args[2:])
	case "js":
		runJS(os.Args[2:])
	case "dtcg":
		runDTCG(os.Args[2:])
	case "docs":
		runDocs(os.Args[2:])
	case "validate":
		runValidate(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "version":
		fmt.Printf("cssmark version %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`cssmark - Design token toolchain using CSS @property

Usage:
  cssmark <command> [arguments]

Commands:
  build      Parse tokens and output JSON
  css        Generate clean CSS with :root variables
  js         Generate JavaScript, metadata, and declarations
  dtcg       Flatten DTCG JSON sources into a migration manifest
  docs       Generate static documentation site
  validate   Validate tokens for errors
  diff       Compare two token snapshots
  version    Show version information
  help       Show this help message

Examples:
  cssmark build tokens.css --out tokens.json
  cssmark css tokens.css --out variables.css
  cssmark js tokens.css --out tokens.js --meta-out tokens.meta.js --dts-out tokens.d.ts
  cssmark dtcg context palettes --out token-migration.json
  cssmark docs tokens.css --out ./docs
  cssmark validate tokens.css
  cssmark diff tokens.old.json tokens.new.json`)
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	out := fs.String("out", "tokens.json", "Output JSON file")
	fs.String("o", "tokens.json", "Output JSON file (shorthand)")

	// Manually parse to support both -o and --out with positional args in any order
	var positional []string
	outValue := "tokens.json"
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" || args[i] == "-out" || args[i] == "-o" {
			if i+1 < len(args) {
				outValue = args[i+1]
				i++
			}
		} else if strings.HasPrefix(args[i], "-o=") {
			outValue = strings.TrimPrefix(args[i], "-o=")
		} else if strings.HasPrefix(args[i], "--out=") {
			outValue = strings.TrimPrefix(args[i], "--out=")
		} else if !strings.HasPrefix(args[i], "-") {
			positional = append(positional, args[i])
		}
	}
	*out = outValue

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files, err := expandGlobs(positional)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %d tokens parsed\n", len(tokens))

	if err := builder.WriteJSON(tokens, *out); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("→ %s\n", *out)
}

func runCSS(args []string) {
	var positional []string
	modeSelectors := make(map[string]string)
	outValue := "variables.css"
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" || args[i] == "-out" || args[i] == "-o" {
			if i+1 < len(args) {
				outValue = args[i+1]
				i++
			}
		} else if args[i] == "--mode-selector" {
			if i+1 < len(args) {
				parts := strings.SplitN(args[i+1], "=", 2)
				if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
					modeSelectors[parts[0]] = parts[1]
				}
				i++
			}
		} else if strings.HasPrefix(args[i], "-o=") {
			outValue = strings.TrimPrefix(args[i], "-o=")
		} else if strings.HasPrefix(args[i], "--out=") {
			outValue = strings.TrimPrefix(args[i], "--out=")
		} else if !strings.HasPrefix(args[i], "-") {
			positional = append(positional, args[i])
		}
	}

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files, err := expandGlobs(positional)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %d tokens parsed\n", len(tokens))

	if err := builder.WriteCSSWithModeSelectors(tokens, outValue, modeSelectors); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing CSS: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("→ %s\n", outValue)
}

func runJS(args []string) {
	var positional []string
	jsPath, metaPath, dtsPath := "tokens.js", "tokens.meta.js", "tokens.d.ts"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out", "-o":
			if i+1 < len(args) {
				jsPath = args[i+1]
				i++
			}
		case "--meta-out":
			if i+1 < len(args) {
				metaPath = args[i+1]
				i++
			}
		case "--dts-out":
			if i+1 < len(args) {
				dtsPath = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				positional = append(positional, args[i])
			}
		}
	}
	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}
	files, err := expandGlobs(positional)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := builder.WriteJS(tokens, jsPath, metaPath, dtsPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing JavaScript: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %d tokens written to %s\n", len(tokens), jsPath)
}

func runDTCG(args []string) {
	var positional []string
	prefix, out := "hb", "token-migration.json"
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--out", "-o":
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		case "--prefix":
			if i+1 < len(args) {
				prefix = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") {
				positional = append(positional, args[i])
			}
		}
	}
	if len(positional) == 0 {
		fmt.Fprintln(os.Stderr, "Error: No DTCG files or directories specified")
		os.Exit(1)
	}
	paths, err := dtcg.Expand(positional)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := dtcg.Import(paths, prefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	body, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, body, 0644); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("✓ %d DTCG tokens written to %s\n", len(tokens), out)
}

func runDocs(args []string) {
	// Manually parse to support both -o and --out with positional args in any order
	var positional []string
	out := "./docs"
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" || args[i] == "-out" || args[i] == "-o" {
			if i+1 < len(args) {
				out = args[i+1]
				i++
			}
		} else if strings.HasPrefix(args[i], "-o=") {
			out = strings.TrimPrefix(args[i], "-o=")
		} else if strings.HasPrefix(args[i], "--out=") {
			out = strings.TrimPrefix(args[i], "--out=")
		} else if !strings.HasPrefix(args[i], "-") {
			positional = append(positional, args[i])
		}
	}

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files, err := expandGlobs(positional)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %d tokens parsed\n", len(tokens))

	if err := builder.WriteDocs(tokens, out, files...); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("→ %s\n", out)
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files, err := expandGlobs(fs.Args())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	errors := parser.Validate(tokens)
	deprecated := 0
	for _, t := range tokens {
		if t.Deprecated {
			deprecated++
		}
	}

	fmt.Printf("✓ %d tokens parsed\n", len(tokens))
	if deprecated > 0 {
		fmt.Printf("⚠ %d deprecated\n", deprecated)
	}
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Printf("✗ %s\n", e)
		}
		os.Exit(1)
	}
}

func runDiff(args []string) {
	if len(args) < 2 {
		fmt.Fprintln(os.Stderr, "Error: Need two JSON files to compare")
		fmt.Fprintln(os.Stderr, "Usage: cssmark diff old.json new.json")
		os.Exit(1)
	}

	oldFile, newFile := args[0], args[1]

	oldTokens, err := loadTokensFromJSON(oldFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", oldFile, err)
		os.Exit(1)
	}

	newTokens, err := loadTokensFromJSON(newFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", newFile, err)
		os.Exit(1)
	}

	diff := differ.Compare(oldTokens, newTokens)
	diff.Print()
}

func loadTokensFromJSON(path string) ([]parser.Token, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var tokens []parser.Token
	if err := json.Unmarshal(data, &tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

func expandGlobs(patterns []string) ([]string, error) {
	var files []string
	for _, pattern := range patterns {
		if strings.ContainsAny(pattern, "*?[") {
			matches, err := filepath.Glob(pattern)
			if err != nil {
				return nil, err
			}
			if len(matches) == 0 {
				return nil, fmt.Errorf("no files matched %q", pattern)
			}
			files = append(files, matches...)
		} else {
			files = append(files, pattern)
		}
	}
	if len(files) == 0 {
		return nil, fmt.Errorf("no input files matched")
	}
	return files, nil
}
