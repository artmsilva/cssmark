# cssmark

> CSS @property as the single source of truth for design token authoring,
> documentation, and distribution.

A toolchain that treats `@property` blocks as the canonical token definition format. Extended descriptors carry metadata that browsers ignore but the tool consumes.

**Zero dependencies.** Single binary. Written in Go.

## Installation

### npm (Recommended for Node.js projects)

```bash
npm install @artmsilva/cssmark
npx cssmark build tokens.css -o tokens.json
```

Add to your package.json scripts:

```json
{
  "scripts": {
    "tokens:build": "cssmark build src/tokens.css -o dist/tokens.json",
    "tokens:docs": "cssmark docs src/tokens.css -o dist/docs",
    "tokens:validate": "cssmark validate src/tokens.css"
  }
}
```

See [docs/NPM_USAGE.md](./docs/NPM_USAGE.md) for integration examples with Vite, Webpack, Next.js, and more.

### Download Binary

Download from [releases](https://github.com/artmsilva/cssmark/releases).

### Build from Source

```bash
git clone https://github.com/artmsilva/cssmark
cd cssmark
go build -o cssmark ./src/cmd/cssmark
```

### Go Install

```bash
go install github.com/artmsilva/cssmark/src/cmd/cssmark@latest
```

## Token Authoring Format

Tokens are authored as standard `@property` rules with additional descriptors for metadata:

```css
@property --color-brand-primary {
  /* Standard @property descriptors — browser-native */
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;

  /* Extended descriptors — tool-only */
  description: "Primary brand color for interactive elements.";
  category: "color.brand";
  type: "color";
  aliases: "--color-primary, --color-action";
  deprecated: false;
  examples: "background: var(--color-brand-primary); border-color: var(--color-brand-primary);";
}
```

## Commands

### Build (JSON Export)

```bash
cssmark build tokens.css --out tokens.json
```

### Generate Documentation

```bash
cssmark docs tokens.css --out ./docs
```

### Validate Tokens

```bash
cssmark validate tokens.css
```

### Diff Token Snapshots

```bash
cssmark diff tokens.old.json tokens.new.json
```

## Extended Descriptors

| Descriptor   | Type    | Required | Description                              |
|--------------|---------|----------|------------------------------------------|
| description  | string  | no       | Human-readable explanation of the token  |
| category     | string  | no       | Dot-separated group path: `color.brand`  |
| type         | string  | no       | Semantic hint: color, size, duration     |
| aliases      | string  | no       | Comma-separated list of related props    |
| deprecated   | boolean | no       | Marks token deprecated                   |
| examples     | string  | no       | Semicolon-separated CSS usage examples   |

## Output Formats

### JSON

```json
[
  {
    "name": "--color-brand-primary",
    "syntax": "<color>",
    "inherits": false,
    "initialValue": "#0055ff",
    "description": "Primary brand color for interactive elements.",
    "category": "color.brand",
    "type": "color",
    "aliases": ["--color-primary", "--color-action"],
    "deprecated": false,
    "examples": ["background: var(--color-brand-primary);"]
  }
]
```

### Static Documentation Site

A minimal, fast reference site with:
- Sidebar with category tree
- Token cards grouped by category
- Color swatches, type badges, examples
- Deprecated token warnings

## License

MIT
