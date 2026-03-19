package parser

// Token represents a parsed CSS @property design token
type Token struct {
	Name            string            `json:"name"`
	Syntax          string            `json:"syntax,omitempty"`
	Inherits        bool              `json:"inherits"`
	InitialValue    string            `json:"initialValue,omitempty"`
	RawInitialValue string            `json:"-"` // original var() reference before resolution
	Modes           map[string]string `json:"modes,omitempty"`
	RawModes        map[string]string `json:"-"` // original var() references before resolution
	Description     string            `json:"description,omitempty"`
	Category        string            `json:"category,omitempty"`
	Type            string            `json:"type,omitempty"`
	Aliases         []string          `json:"aliases,omitempty"`
	Deprecated      bool              `json:"deprecated,omitempty"`
	Examples        []string          `json:"examples,omitempty"`
	Source          Source            `json:"source"`
}

// Source tracks where a token was defined
type Source struct {
	File string `json:"file"`
	Line int    `json:"line"`
}
