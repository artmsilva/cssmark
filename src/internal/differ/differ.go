package differ

import (
	"fmt"

	"github.com/artmsilva/cssmark/src/internal/parser"
)

// Diff represents the difference between two token sets
type Diff struct {
	Added      []parser.Token
	Removed    []parser.Token
	Changed    []TokenChange
	Deprecated []parser.Token
}

// TokenChange represents a changed token
type TokenChange struct {
	Name     string
	Old      parser.Token
	New      parser.Token
	Fields   []string // which fields changed
}

// Compare compares two token sets and returns the diff
func Compare(oldTokens, newTokens []parser.Token) Diff {
	oldMap := make(map[string]parser.Token)
	newMap := make(map[string]parser.Token)

	for _, t := range oldTokens {
		oldMap[t.Name] = t
	}
	for _, t := range newTokens {
		newMap[t.Name] = t
	}

	var diff Diff

	// Find added and changed
	for name, newToken := range newMap {
		if oldToken, exists := oldMap[name]; exists {
			// Check if changed
			if changes := compareTokens(oldToken, newToken); len(changes) > 0 {
				diff.Changed = append(diff.Changed, TokenChange{
					Name:   name,
					Old:    oldToken,
					New:    newToken,
					Fields: changes,
				})
			}
			// Check if newly deprecated
			if !oldToken.Deprecated && newToken.Deprecated {
				diff.Deprecated = append(diff.Deprecated, newToken)
			}
		} else {
			diff.Added = append(diff.Added, newToken)
		}
	}

	// Find removed
	for name, oldToken := range oldMap {
		if _, exists := newMap[name]; !exists {
			diff.Removed = append(diff.Removed, oldToken)
		}
	}

	return diff
}

func compareTokens(old, new parser.Token) []string {
	var changes []string

	if old.InitialValue != new.InitialValue {
		changes = append(changes, "initialValue")
	}
	if old.Syntax != new.Syntax {
		changes = append(changes, "syntax")
	}
	if old.Description != new.Description {
		changes = append(changes, "description")
	}
	if old.Category != new.Category {
		changes = append(changes, "category")
	}
	if old.Type != new.Type {
		changes = append(changes, "type")
	}
	if old.Deprecated != new.Deprecated {
		changes = append(changes, "deprecated")
	}

	return changes
}

// Print outputs the diff to stdout
func (d Diff) Print() {
	if len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0 && len(d.Deprecated) == 0 {
		fmt.Println("No changes detected.")
		return
	}

	if len(d.Added) > 0 {
		fmt.Printf("\n✚ Added (%d)\n", len(d.Added))
		for _, t := range d.Added {
			fmt.Printf("  + %s = %s\n", t.Name, t.InitialValue)
		}
	}

	if len(d.Removed) > 0 {
		fmt.Printf("\n✖ Removed (%d)\n", len(d.Removed))
		for _, t := range d.Removed {
			fmt.Printf("  - %s\n", t.Name)
		}
	}

	if len(d.Changed) > 0 {
		fmt.Printf("\n◎ Changed (%d)\n", len(d.Changed))
		for _, c := range d.Changed {
			fmt.Printf("  ~ %s\n", c.Name)
			for _, field := range c.Fields {
				fmt.Printf("    %s: %v → %v\n", field, getField(c.Old, field), getField(c.New, field))
			}
		}
	}

	if len(d.Deprecated) > 0 {
		fmt.Printf("\n⚠ Newly Deprecated (%d)\n", len(d.Deprecated))
		for _, t := range d.Deprecated {
			fmt.Printf("  ⚠ %s\n", t.Name)
		}
	}

	fmt.Println()
}

func getField(t parser.Token, field string) interface{} {
	switch field {
	case "initialValue":
		return t.InitialValue
	case "syntax":
		return t.Syntax
	case "description":
		return t.Description
	case "category":
		return t.Category
	case "type":
		return t.Type
	case "deprecated":
		return t.Deprecated
	default:
		return ""
	}
}

// ToMarkdown outputs the diff as markdown (for PR comments)
func (d Diff) ToMarkdown() string {
	if len(d.Added) == 0 && len(d.Removed) == 0 && len(d.Changed) == 0 && len(d.Deprecated) == 0 {
		return "No token changes detected."
	}

	var md string
	md += "## Token Changes\n\n"

	if len(d.Added) > 0 {
		md += fmt.Sprintf("### ✚ Added (%d)\n\n", len(d.Added))
		for _, t := range d.Added {
			md += fmt.Sprintf("- `%s` = `%s`\n", t.Name, t.InitialValue)
		}
		md += "\n"
	}

	if len(d.Removed) > 0 {
		md += fmt.Sprintf("### ✖ Removed (%d)\n\n", len(d.Removed))
		for _, t := range d.Removed {
			md += fmt.Sprintf("- `%s`\n", t.Name)
		}
		md += "\n"
	}

	if len(d.Changed) > 0 {
		md += fmt.Sprintf("### ◎ Changed (%d)\n\n", len(d.Changed))
		for _, c := range d.Changed {
			md += fmt.Sprintf("- `%s`\n", c.Name)
		}
		md += "\n"
	}

	if len(d.Deprecated) > 0 {
		md += fmt.Sprintf("### ⚠ Newly Deprecated (%d)\n\n", len(d.Deprecated))
		for _, t := range d.Deprecated {
			md += fmt.Sprintf("- `%s`\n", t.Name)
		}
		md += "\n"
	}

	return md
}
