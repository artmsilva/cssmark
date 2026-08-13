package dtcg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type node map[string]any

type Token struct {
	ID    string
	Name  string
	Type  string
	Value any
	Modes map[string]any
}

// Import flattens DTCG token files. Later files overlay an existing token's modes,
// matching the generated dense-mode overlay convention used by Hummingbird.
func Import(paths []string, prefix string) ([]Token, error) {
	byID := map[string]Token{}
	for _, path := range paths {
		body, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		var root node
		if err := json.Unmarshal(body, &root); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		if err := walk(root, nil, nil, byID, prefix); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
	}
	out := make([]Token, 0, len(byID))
	for _, token := range byID {
		out = append(out, token)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func walk(value node, path []string, inheritedType *string, byID map[string]Token, prefix string) error {
	tokenType := ""
	if raw, ok := value["$type"].(string); ok {
		tokenType = raw
	} else if inheritedType != nil {
		tokenType = *inheritedType
	}
	_, hasValue := value["$value"]
	_, hasExtensions := value["$extensions"]
	if hasValue || hasExtensions {
		id := strings.Join(path, ".")
		current, exists := byID[id]
		if !exists {
			current = Token{ID: id, Name: cssName(prefix, id), Type: inferredType(tokenType, id), Modes: map[string]any{}}
		}
		if raw, ok := value["$value"]; ok {
			current.Value = raw
		}
		if extensions, ok := value["$extensions"].(map[string]any); ok {
			if mode, ok := extensions["mode"].(map[string]any); ok {
				for name, modeValue := range mode {
					current.Modes[name] = modeValue
				}
			}
		}
		byID[id] = current
		return nil
	}
	for key, child := range value {
		if strings.HasPrefix(key, "$") {
			continue
		}
		childNode, ok := child.(map[string]any)
		if !ok {
			continue
		}
		if err := walk(childNode, append(path, key), stringPtr(tokenType), byID, prefix); err != nil {
			return err
		}
	}
	return nil
}

func stringPtr(value string) *string { return &value }

func inferredType(tokenType, id string) string {
	if tokenType != "" {
		return tokenType
	}
	// Some legacy DTCG groups omit $type. Their IDs still carry the intended
	// design-token category, so avoid an unhelpful "untyped" source file.
	switch strings.Split(id, ".")[0] {
	case "color":
		return "color"
	case "space", "radius":
		return "dimension"
	case "duration":
		return "duration"
	default:
		return "unknown"
	}
}

func cssName(prefix, id string) string {
	id = strings.ReplaceAll(id, "type-sh", "type")
	return "--" + prefix + "-" + strings.ReplaceAll(id, ".", "-")
}

// Expand expands directory inputs to sorted JSON files; plain files retain CLI order.
func Expand(paths []string) ([]string, error) {
	var out []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, err
		}
		if !info.IsDir() {
			out = append(out, path)
			continue
		}
		var matches []string
		err = filepath.WalkDir(path, func(candidate string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
				matches = append(matches, candidate)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		sort.Strings(matches)
		out = append(out, matches...)
	}
	return out, nil
}
