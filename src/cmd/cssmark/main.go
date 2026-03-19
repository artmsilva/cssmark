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
	"github.com/artmsilva/cssmark/src/internal/parser"
)

const version = "0.1.0"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		runBuild(os.Args[2:])
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
  docs       Generate static documentation site
  validate   Validate tokens for errors
  diff       Compare two token snapshots
  version    Show version information
  help       Show this help message

Examples:
  cssmark build tokens.css --out tokens.json
  cssmark docs tokens.css --out ./docs
  cssmark validate tokens.css
  cssmark diff tokens.old.json tokens.new.json`)
}

func runBuild(args []string) {
	fs := flag.NewFlagSet("build", flag.ExitOnError)
	out := fs.String("out", "tokens.json", "Output JSON file")
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files := expandGlobs(fs.Args())
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

func runDocs(args []string) {
	fs := flag.NewFlagSet("docs", flag.ExitOnError)
	out := fs.String("out", "./docs", "Output directory")

	// Separate flags from positional args
	var positional []string
	for i := 0; i < len(args); i++ {
		if args[i] == "--out" || args[i] == "-out" {
			if i+1 < len(args) {
				*out = args[i+1]
				i++
			}
		} else if !strings.HasPrefix(args[i], "-") {
			positional = append(positional, args[i])
		}
	}

	if len(positional) < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files := expandGlobs(positional)
	tokens, err := parser.ParseFiles(files)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("✓ %d tokens parsed\n", len(tokens))

	if err := builder.WriteDocs(tokens, *out); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing docs: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("→ %s\n", *out)
}

func runValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	fs.Parse(args)

	if fs.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "Error: No input files specified")
		os.Exit(1)
	}

	files := expandGlobs(fs.Args())
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

func expandGlobs(patterns []string) []string {
	var files []string
	for _, pattern := range patterns {
		if strings.Contains(pattern, "*") {
			matches, err := filepath.Glob(pattern)
			if err == nil {
				files = append(files, matches...)
			}
		} else {
			files = append(files, pattern)
		}
	}
	return files
}
