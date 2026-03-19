# cssmark

Design token toolchain using CSS `@property` as the single source of truth for design token authoring, documentation, and distribution.

## Installation

```bash
npm install @artmsilva/cssmark
```

## Usage

```bash
# Parse tokens and output JSON
npx cssmark build tokens.css --out tokens.json

# Generate static documentation site
npx cssmark docs tokens.css --out ./docs

# Validate tokens
npx cssmark validate tokens.css

# Diff two JSON snapshots
npx cssmark diff tokens.old.json tokens.new.json
```

## License

MIT
