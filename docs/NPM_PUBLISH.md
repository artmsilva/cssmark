# Publishing to npm

## GitHub npm Registry

The package is published to GitHub's npm registry automatically via GitHub Actions.

### Setup (One-time)

1. Create a Personal Access Token (PAT) with `packages:write` scope
2. Add it as a repository secret named `NPM_TOKEN`
3. Push a version tag: `git tag v0.1.0 && git push origin v0.1.0`

### Manual Publishing

```bash
# Login to GitHub npm registry
npm login --registry=https://npm.pkg.github.com

# Build the Go binary first
go build -o src/npm/bin/cssmark ./src/cmd/cssmark

# Publish from npm directory
cd src/npm
npm publish
```

### Installing from GitHub Registry

Users need to configure npm to use GitHub's registry for @artmsilva scope:

```bash
# Create .npmrc in project root
echo "@artmsilva:registry=https://npm.pkg.github.com" >> .npmrc

# Then install normally
npm install @artmsilva/cssmark
```

## How the Package Works

1. User runs `npm install @artmsilva/cssmark`
2. The `postinstall` script (`install.js`) runs automatically
3. It downloads the correct binary for the user's platform from GitHub Releases
4. Binary is placed in `node_modules/@artmsilva/cssmark/bin/`
5. User can run `npx cssmark` or use in npm scripts

### Download URL Pattern

```
https://github.com/artmsilva/cssmark/releases/download/v{VERSION}/cssmark_{Platform}_{Arch}.{ext}
```

Examples:
- macOS arm64: `cssmark_Darwin_arm64.tar.gz`
- Linux x64: `cssmark_Linux_x86_64.tar.gz`
- Windows x64: `cssmark_Windows_x86_64.zip`
