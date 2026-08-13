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

### var() References

Tokens can reference other tokens using `var()`. References are resolved to literal values in JSON output, while CSS output preserves the original `var()` references.

```css
@property --color-blue-500 {
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
}

@property --color-primary {
  syntax: "<color>";
  inherits: false;
  initial-value: var(--color-blue-500);
}
```

JSON output resolves the chain:

```json
{ "name": "--color-primary", "initialValue": "#0055ff" }
```

CSS output preserves the reference:

```css
:root {
  --color-blue-500: #0055ff;
  --color-primary: var(--color-blue-500);
}
```

Chained references (`var(--a)` → `var(--b)` → `#fff`) are fully resolved. Circular and unknown references are kept as-is.

### Mode Descriptors

Define alternative values for different modes using `mode-*` descriptors. These generate `:root[data-color-mode='...']` override blocks in CSS output.

```css
@property --color-bg {
  syntax: "<color>";
  inherits: false;
  initial-value: #ffffff;
  mode-dark: #1a1a2e;
  mode-high-contrast: #000000;
}
```

CSS output:

```css
:root {
  --color-bg: #ffffff;
}

:root[data-color-mode='dark'] {
  --color-bg: #1a1a2e;
}

:root[data-color-mode='high-contrast'] {
  --color-bg: #000000;
}
```

JSON output includes modes as a map:

```json
{ "name": "--color-bg", "initialValue": "#ffffff", "modes": { "dark": "#1a1a2e", "high-contrast": "#000000" } }
```

Mode values also support `var()` references, which are resolved the same way.

Use `--mode-selector mode=selector` when a mode is not a color mode. For example, dense spacing can target both the document root and nested containers:

```sh
cssmark css tokens.css --out tokens.css \
  --mode-selector "dense=:root[data-density='dense'], [data-density='dense']"
```

## Commands

### Build (JSON Export)

```bash
cssmark build tokens.css --out tokens.json
```

### Import DTCG JSON for migration

`dtcg` is a migration aid for existing DTCG token trees. It recursively flattens JSON files, inherits group `$type`, merges later mode-only overlays by token path, and preserves stable IDs, CSS names, values, and modes in a JSON manifest.

```bash
cssmark dtcg context palettes --prefix hb --out token-migration.json
```

Add `--tree-out tokens` to produce a flat, composable **cssmark authoring** tree: every token file lives directly under `tokens/` (`color.css`, `action.css`, `decorative.css`, `type.css`, etc.) and `index.css` contains only the ordered imports. It does not emit runtime `:root` or mode override files. `--css-out tokens.source.css` remains available for a single-file authoring artifact.

### Generate JavaScript artifacts

Use `id` to preserve a stable programmatic key when the CSS variable name is not enough to reconstruct it.

```css
@property --hb-color-action-primary {
  id: "color.action.primary";
  syntax: "<color>";
  inherits: false;
  initial-value: #0055ff;
  mode-dark: #66aaff;
}
```

```bash
cssmark js tokens.css \
  --out tokens.js \
  --meta-out tokens.meta.js \
  --dts-out tokens.d.ts
```

This emits an ESM `tokens` map, per-token `modes`, a `token(id, mode)` helper, metadata, and a declaration file.

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
| mode-*       | string  | no       | Override value for a named mode           |

## Output Formats

### JSON

```json
[
  {
    "name": "--color-brand-primary",
    "syntax": "<color>",
    "inherits": false,
    "initialValue": "#0055ff",
    "modes": {
      "dark": "#66aaff"
    },
    "description": "Primary brand color for interactive elements.",
    "category": "color.brand",
    "type": "color",
    "aliases": ["--color-primary", "--color-action"],
    "deprecated": false,
    "examples": ["background: var(--color-brand-primary);"]
  }
]
```

When a token uses `var()` references, `initialValue` contains the resolved literal value. The `modes` field is omitted when no `mode-*` descriptors are defined.

### Static Documentation Site

A minimal, fast reference site with:
- Sidebar with category tree
- Token cards grouped by category
- Color swatches, type badges, examples
- Deprecated token warnings
- Mobile drawer navigation
- Optional source stylesheet viewer with tiny built-in syntax highlighting

The project site and generated example docs are separate:
- **Project docs:** [artmsilva.github.io/cssmark](https://artmsilva.github.io/cssmark)
- **Generated example docs:** [artmsilva.github.io/cssmark/example](https://artmsilva.github.io/cssmark/example)

## Trust Model and Current Parser Scope

cssmark is a build-time tool for trusted source files in your repository. The generated docs inline token values into preview styles and display the source stylesheet, so do not generate public docs from secrets or untrusted CSS.

The parser intentionally targets cssmark's authoring subset: standard `@property` blocks plus cssmark extended descriptors. It is not a full CSS parser yet. Complex CSS outside token definitions may be ignored, but cross-file `var()` references between parsed tokens are resolved after all input files are loaded.

## License

MIT
