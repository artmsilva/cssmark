# npm Usage Guide

## Installation

```bash
npm install @artmsilva/cssmark
```

## Usage in package.json Scripts

```json
{
  "scripts": {
    "tokens:build": "cssmark build src/tokens.css -o dist/tokens.json",
    "tokens:docs": "cssmark docs src/tokens.css -o dist/docs",
    "tokens:validate": "cssmark validate src/tokens.css",
    "tokens:diff": "cssmark diff tokens.old.json tokens.new.json"
  },
  "devDependencies": {
    "@artmsilva/cssmark": "^0.1.0"
  }
}
```

## Integration Examples

### Vite

```js
// vite.config.js
import { execSync } from 'child_process';

export default {
  plugins: [
    {
      name: 'cssmark',
      buildStart() {
        execSync('npx cssmark build src/tokens.css -o src/tokens.json');
      }
    }
  ]
}
```

### Webpack

```js
// webpack.config.js
const { execSync } = require('child_process');

module.exports = {
  plugins: [
    {
      apply: (compiler) => {
        compiler.hooks.beforeCompile.tap('CSSMark', () => {
          execSync('npx cssmark build src/tokens.css -o src/tokens.json');
        });
      }
    }
  ]
}
```

### Next.js

```js
// next.config.js
const { execSync } = require('child_process');

// Build tokens before Next.js starts
execSync('npx cssmark build tokens/design.css -o public/tokens.json');

module.exports = {
  // your config
}
```

## How It Works

1. When you run `npm install @artmsilva/cssmark`, the `postinstall` script runs
2. It detects your platform (macOS, Linux, Windows) and architecture (x64, arm64)
3. Downloads the pre-built binary from GitHub releases
4. Extracts to `node_modules/@artmsilva/cssmark/bin/`
5. The binary is available via `npx cssmark` or in npm scripts

## Manual Installation

If automatic download fails:

1. Download the binary manually from [releases](https://github.com/artmsilva/cssmark/releases)
2. Extract and place in `node_modules/@artmsilva/cssmark/bin/cssmark`
3. Make executable: `chmod +x node_modules/@artmsilva/cssmark/bin/cssmark`

## Using Without npm

You can also install the Go binary directly:

```bash
go install github.com/artmsilva/cssmark/src/cmd/cssmark@latest
```
