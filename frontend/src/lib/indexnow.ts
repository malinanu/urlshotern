const INDEXNOW_KEY = 'trunc-indexnow-2026';
const SITE_URL = 'https://trunc.site';

export async function submitToIndexNow(urls: string[]): Promise<void> {
  if (urls.length === 0) return;

  const body = {
    host: 'trunc.site',
    key: INDEXNOW_KEY,
    keyLocation: `${SITE_URL}/${INDEXNOW_KEY}.txt`,
    urlList: urls,
  };

  const res = await fetch('https://api.indexnow.org/indexnow', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json; charset=utf-8' },
    body: JSON.stringify(body),
  });

  if (!res.ok && res.status !== 202) {
    throw new Error(`IndexNow submission failed: ${res.status} ${res.statusText}`);
  }
}

// Submit all blog post URLs — call this after publishing or updating posts.
export async function submitBlogPosts(slugs: string[]): Promise<void> {
  const urls = slugs.map((slug) => `${SITE_URL}/blog/${slug}`);
  return submitToIndexNow(urls);
}
