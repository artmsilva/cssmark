#!/usr/bin/env node
const fs = require('fs');
const path = require('path');
const os = require('os');

const platform = os.platform();
const arch = os.arch();

const platformKey = `${platform}-${arch}`;
const binDir = path.join(__dirname, '..', 'bin', platformKey);
const binaryName = platform === 'win32' ? 'cssmark.exe' : 'cssmark';
const binaryPath = path.join(binDir, binaryName);

if (!fs.existsSync(binaryPath)) {
  console.error(`cssmark binary not found for ${platformKey}`);
  console.error(`Supported platforms: darwin-arm64, darwin-x64, linux-x64, linux-arm64, win32-x64`);
  process.exit(1);
}

// Make sure binary is executable
try {
  fs.chmodSync(binaryPath, 0o755);
} catch (e) {
  // Ignore on Windows
}

console.log(`cssmark installed for ${platformKey}`);
