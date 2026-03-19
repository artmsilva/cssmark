# cssmark

Design token toolchain using CSS `@property` as the single source of truth for design token authoring, documentation, and distribution.

## Installation

This package is published to the GitHub npm registry. Configure npm to use GitHub Packages for the `@artmsilva` scope:

```bash
echo "@artmsilva:registry=https://npm.pkg.github.com" >> .npmrc
npm install @artmsilva/cssmark
```

## Usage

```bash
# Parse tokens and output JSON
npx cssmark build tokens.css --out tokens.json

# Generate static documentation site
npx cssmark docs tokens.css --out ./docs

# Generate production CSS from tokens
npx cssmark css tokens.css --out tokens.production.css

# Validate tokens
npx cssmark validate tokens.css

# Diff two JSON snapshots
npx cssmark diff tokens.old.json tokens.new.json

# Watch mode
npx cssmark docs tokens.css --out ./docs --watch
```

## Supported Platforms

- macOS (arm64, x64)
- Linux (arm64, x64)
- Windows (x64)

## License

MIT
