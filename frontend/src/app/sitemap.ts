import type { MetadataRoute } from 'next';

export const dynamic = 'force-static';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
const SITE_URL = process.env.NEXT_PUBLIC_SITE_URL || 'https://trunc.site';

async function fetchCMSPages(): Promise<Array<{ slug: string; updated_at?: string }>> {
  try {
    // Uses the internal API key or server-side access if configured
    const apiKey = process.env.SITEMAP_API_KEY;
    const headers: Record<string, string> = {};
    if (apiKey) headers['X-API-Key'] = apiKey;

    const res = await fetch(`${API_BASE_URL}/api/v1/admin/cms/pages?limit=500&status=published`, {
      headers,
      next: { revalidate: 3600 },
    });
    if (!res.ok) return [];
    const data = await res.json();
    const pages = data.pages ?? data ?? [];
    return Array.isArray(pages) ? pages : [];
  } catch {
    return [];
  }
}

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const staticRoutes: MetadataRoute.Sitemap = [
    { url: `${SITE_URL}/`, lastModified: new Date(), changeFrequency: 'weekly', priority: 1.0 },
    { url: `${SITE_URL}/features`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.8 },
    { url: `${SITE_URL}/pricing`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.8 },
    { url: `${SITE_URL}/about`, lastModified: new Date(), changeFrequency: 'monthly', priority: 0.6 },
    { url: `${SITE_URL}/contact`, lastModified: new Date(), changeFrequency: 'yearly', priority: 0.5 },
    { url: `${SITE_URL}/privacy`, lastModified: new Date(), changeFrequency: 'yearly', priority: 0.3 },
    { url: `${SITE_URL}/terms`, lastModified: new Date(), changeFrequency: 'yearly', priority: 0.3 },
  ];

  const cmsPages = await fetchCMSPages();
  const dynamicRoutes: MetadataRoute.Sitemap = cmsPages
    .filter((p) => p.slug && p.slug !== '_placeholder')
    .map((p) => ({
      url: `${SITE_URL}/pages/${p.slug}`,
      lastModified: p.updated_at ? new Date(p.updated_at) : new Date(),
      changeFrequency: 'monthly' as const,
      priority: 0.6,
    }));

  return [...staticRoutes, ...dynamicRoutes];
}
