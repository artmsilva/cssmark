package builder

import (
	"encoding/json"
	"os"

	"github.com/artmsilva/cssmark/src/internal/parser"
)

// WriteJSON writes tokens to a JSON file
func WriteJSON(tokens []parser.Token, path string) error {
	data, err := json.MarshalIndent(tokens, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ToJSON converts tokens to JSON bytes
func ToJSON(tokens []parser.Token) ([]byte, error) {
	return json.MarshalIndent(tokens, "", "  ")
}
