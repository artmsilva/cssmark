#!/usr/bin/env node
const { spawn } = require('child_process');
const path = require('path');
const os = require('os');

const platform = os.platform();
const arch = os.arch();

const platformKey = `${platform}-${arch}`;
const binaryName = platform === 'win32' ? 'cssmark.exe' : 'cssmark';
const binaryPath = path.join(__dirname, '..', 'bin', platformKey, binaryName);

const child = spawn(binaryPath, process.argv.slice(2), {
  stdio: 'inherit',
  windowsHide: true,
});

child.on('error', (err) => {
  console.error(`Failed to run cssmark: ${err.message}`);
  console.error(`Binary path: ${binaryPath}`);
  process.exit(1);
});

child.on('close', (code) => {
  process.exit(code || 0);
});
