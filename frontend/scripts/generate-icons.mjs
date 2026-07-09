// Renders the branded PWA/favicon icon set from scripts/icon-source.svg into
// public/ and public/icons/. Run once (or whenever the source changes):
//
//   docker run --rm -v "$PWD/frontend":/w -w /w node:22-bookworm-slim \
//     sh -c "npm i --no-save sharp && node scripts/generate-icons.mjs"
//
// sharp is intentionally NOT a project dependency — the committed PNGs are the
// artifact; this script only regenerates them.
import sharp from 'sharp';
import { mkdir } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';

const here = path.dirname(fileURLToPath(import.meta.url));
const frontend = path.resolve(here, '..');
const src = path.join(here, 'icon-source.svg');
const publicDir = path.join(frontend, 'public');
const iconsDir = path.join(publicDir, 'icons');

// [outputPath, size]
const targets = [
  // root favicons / platform icons referenced by src/app/layout.tsx
  ['favicon-16x16.png', 16],
  ['favicon-32x32.png', 32],
  ['apple-touch-icon.png', 180],
  ['android-chrome-192x192.png', 192],
  ['android-chrome-512x512.png', 512],
  // manifest icons
  ['icons/icon-72x72.png', 72],
  ['icons/icon-96x96.png', 96],
  ['icons/icon-128x128.png', 128],
  ['icons/icon-144x144.png', 144],
  ['icons/icon-152x152.png', 152],
  ['icons/icon-192x192.png', 192],
  ['icons/icon-384x384.png', 384],
  ['icons/icon-512x512.png', 512],
  // manifest shortcut icons
  ['icons/shortcut-shorten.png', 96],
  ['icons/shortcut-analytics.png', 96],
];

await mkdir(iconsDir, { recursive: true });

for (const [rel, size] of targets) {
  const out = path.join(publicDir, rel);
  await sharp(src, { density: 384 })
    .resize(size, size, { fit: 'cover' })
    .png()
    .toFile(out);
  console.log(`wrote ${rel} (${size}x${size})`);
}

console.log('done');
