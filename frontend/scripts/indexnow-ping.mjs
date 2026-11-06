// Notify IndexNow (Bing, Yandex, and partners) that URLs changed, so they get
// (re)crawled quickly. Run after each production deploy:
//
//   node scripts/indexnow-ping.mjs
//
// The key file public/<KEY>.txt must be deployed and reachable at
// https://trunc.site/<KEY>.txt for the submission to be accepted.
const HOST = 'trunc.site';
const KEY = '7f3c1e9a4b2d48f6a1c05e8d2b6f9a30';
const SITE = `https://${HOST}`;

// Public, indexable URLs. Keep in sync with sitemap.ts / blog posts.
const staticPaths = [
  '/',
  '/features',
  '/pricing',
  '/about',
  '/contact',
  '/api-docs',
  '/privacy',
  '/terms',
  '/dmca',
  '/blog',
];

async function main() {
  // Pull blog slugs from the compiled posts list if available; otherwise just
  // submit the static paths.
  let urlList = staticPaths.map((p) => `${SITE}${p}`);
  try {
    const { posts } = await import('../src/app/blog/posts.ts');
    urlList = urlList.concat(posts.map((p) => `${SITE}/blog/${p.slug}`));
  } catch {
    // posts.ts is TS; if this script is run outside a TS loader, skip it.
  }

  const body = { host: HOST, key: KEY, keyLocation: `${SITE}/${KEY}.txt`, urlList };
  const res = await fetch('https://api.indexnow.org/indexnow', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json; charset=utf-8' },
    body: JSON.stringify(body),
  });
  console.log(`IndexNow: ${res.status} ${res.statusText} (${urlList.length} URLs)`);
}

main().catch((e) => {
  console.error(e);
  process.exit(1);
});
