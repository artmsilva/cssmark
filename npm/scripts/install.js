#!/usr/bin/env node
const { execSync, spawn } = require('child_process');
const fs = require('fs');
const path = require('path');
const https = require('https');
const os = require('os');

const VERSION = require('../package.json').version;
const REPO = 'artmsilva/cssmark';

function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();

  const platformMap = {
    darwin: 'darwin',
    linux: 'linux',
    win32: 'windows',
  };

  const archMap = {
    x64: 'amd64',
    arm64: 'arm64',
  };

  const p = platformMap[platform];
  const a = archMap[arch];

  if (!p || !a) {
    throw new Error(`Unsupported platform: ${platform}-${arch}`);
  }

  return { platform: p, arch: a };
}

function getBinaryName(platform) {
  return platform === 'windows' ? 'cssmark.exe' : 'cssmark';
}

function downloadFile(url, dest) {
  return new Promise((resolve, reject) => {
    const follow = (url, redirects = 0) => {
      if (redirects > 10) {
        reject(new Error('Too many redirects'));
        return;
      }

      https.get(url, (response) => {
        if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
          follow(response.headers.location, redirects + 1);
          return;
        }

        if (response.statusCode !== 200) {
          reject(new Error(`Download failed: ${response.statusCode}`));
          return;
        }

        const file = fs.createWriteStream(dest);
        response.pipe(file);
        file.on('finish', () => {
          file.close(resolve);
        });
      }).on('error', reject);
    };

    follow(url);
  });
}

async function extractTarGz(archive, dest) {
  execSync(`tar -xzf "${archive}" -C "${dest}"`, { stdio: 'inherit' });
}

async function extractZip(archive, dest) {
  execSync(`unzip -o "${archive}" -d "${dest}"`, { stdio: 'inherit' });
}

async function install() {
  const { platform, arch } = getPlatform();
  const binaryName = getBinaryName(platform);
  const binDir = path.join(__dirname, '..', 'bin');

  // Ensure bin directory exists
  if (!fs.existsSync(binDir)) {
    fs.mkdirSync(binDir, { recursive: true });
  }

  const binaryPath = path.join(binDir, binaryName);

  // Check if binary already exists
  if (fs.existsSync(binaryPath)) {
    console.log('cssmark binary already exists');
    return;
  }

  const ext = platform === 'windows' ? 'zip' : 'tar.gz';
  const assetName = `cssmark_${VERSION}_${platform}_${arch}.${ext}`;
  const downloadUrl = `https://github.com/${REPO}/releases/download/v${VERSION}/${assetName}`;

  console.log(`Downloading cssmark v${VERSION} for ${platform}-${arch}...`);
  console.log(`URL: ${downloadUrl}`);

  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), 'cssmark-'));
  const archivePath = path.join(tmpDir, assetName);

  try {
    await downloadFile(downloadUrl, archivePath);

    console.log('Extracting...');
    if (ext === 'zip') {
      await extractZip(archivePath, tmpDir);
    } else {
      await extractTarGz(archivePath, tmpDir);
    }

    // Find and move binary
    const extractedBinary = path.join(tmpDir, binaryName);
    if (fs.existsSync(extractedBinary)) {
      fs.copyFileSync(extractedBinary, binaryPath);
      fs.chmodSync(binaryPath, 0o755);
      console.log(`Installed cssmark to ${binaryPath}`);
    } else {
      throw new Error(`Binary not found in archive: ${extractedBinary}`);
    }
  } finally {
    // Cleanup
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

install().catch((err) => {
  console.error('Failed to install cssmark:', err.message);
  process.exit(1);
});
