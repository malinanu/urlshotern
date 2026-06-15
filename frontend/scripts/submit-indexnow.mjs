#!/usr/bin/env node
/**
 * Submit all blog posts and key static pages to IndexNow after a deploy.
 *
 * Blog posts are statically generated from src/app/blog/posts.ts, so they have no
 * runtime "publish" event (unlike CMS pages, which ping IndexNow on publish via the
 * notifyIndexNow server action). Run this once per deploy to tell Bing/Yandex/Naver
 * to re-crawl the static content:
 *
 *   npm run submit-indexnow
 *
 * Self-contained: extracts slugs from posts.ts with a regex so it needs no build step
 * or TypeScript loader.
 */
import { readFileSync } from 'node:fs';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

const INDEXNOW_KEY = 'trunc-indexnow-2026';
const SITE_URL = 'https://trunc.site';

const __dirname = dirname(fileURLToPath(import.meta.url));
const postsPath = join(__dirname, '..', 'src', 'app', 'blog', 'posts.ts');

function getBlogSlugs() {
  const src = readFileSync(postsPath, 'utf8');
  const slugs = [];
  const re = /slug:\s*'([^']+)'/g;
  let m;
  while ((m = re.exec(src))) slugs.push(m[1]);
  return [...new Set(slugs)];
}

const staticPaths = [
  '/',
  '/features',
  '/pricing',
  '/about',
  '/blog',
  '/api-docs',
  '/blog/authors/malin-jayaweera',
];

async function main() {
  const urls = [
    ...staticPaths.map((p) => `${SITE_URL}${p}`),
    ...getBlogSlugs().map((slug) => `${SITE_URL}/blog/${slug}`),
  ];

  console.log(`Submitting ${urls.length} URLs to IndexNow…`);

  const res = await fetch('https://api.indexnow.org/indexnow', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json; charset=utf-8' },
    body: JSON.stringify({
      host: 'trunc.site',
      key: INDEXNOW_KEY,
      keyLocation: `${SITE_URL}/${INDEXNOW_KEY}.txt`,
      urlList: urls,
    }),
  });

  if (res.ok || res.status === 202) {
    console.log(`✓ IndexNow accepted ${urls.length} URLs (HTTP ${res.status}).`);
  } else {
    console.error(`✗ IndexNow submission failed: HTTP ${res.status} ${res.statusText}`);
    process.exit(1);
  }
}

main().catch((err) => {
  console.error('✗ IndexNow submission error:', err.message);
  process.exit(1);
});
