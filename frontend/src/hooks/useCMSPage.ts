import { useState, useEffect, useRef } from 'react';

export interface CMSPage {
  id?: number;
  slug: string;
  title: string;
  content: string;
  content_blocks?: Record<string, any>;
  custom_fields?: Record<string, any>;
  meta_title?: string;
  meta_description?: string;
  meta_keywords?: string;
  is_published?: boolean;
  created_at?: string;
  updated_at?: string;
}

export interface UseCMSPageResult {
  content: CMSPage | null;
  loading: boolean;
  error: Error | null;
  usedCMS: boolean;
}

/**
 * Custom hook to fetch CMS page content with fallback support.
 *
 * Renders fallback content immediately (loading: false) so the static export
 * always bakes in real content for crawlers. The CMS fetch runs client-side
 * after hydration and updates the content silently if it succeeds.
 */
export function useCMSPage(
  slug: string,
  fallbackContent: Partial<CMSPage> = {}
): UseCMSPageResult {
  const buildFallback = (): CMSPage => ({
    slug,
    title: '',
    content: '',
    ...fallbackContent,
  } as CMSPage);

  // Start with fallback content and loading:false so the pre-rendered HTML
  // always contains real content (H1, body copy) rather than a spinner.
  const [content, setContent] = useState<CMSPage | null>(buildFallback);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<Error | null>(null);
  const [usedCMS, setUsedCMS] = useState<boolean>(false);

  // Keep fallback stable across renders so the effect dep array doesn't change.
  const fallbackRef = useRef(fallbackContent);

  useEffect(() => {
    let isMounted = true;

    const fetchPage = async () => {
      try {
        const response = await fetch(`/api/v1/pages/${slug}`);

        if (!response.ok) {
          throw new Error(`Failed to fetch page: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();
        const pageData = data.page || data;

        if (isMounted && pageData) {
          if (pageData.content_blocks && typeof pageData.content_blocks === 'string') {
            try {
              pageData.content_blocks = JSON.parse(pageData.content_blocks);
            } catch (e) {
              console.warn('[useCMSPage] Failed to parse content_blocks JSON:', e);
            }
          }
          setContent(pageData);
          setUsedCMS(true);
        } else {
          throw new Error('Page data not found in response');
        }
      } catch (err) {
        if (isMounted) {
          const errorObj = err instanceof Error ? err : new Error('Unknown error');
          setError(errorObj);
          setUsedCMS(false);
          if (process.env.NODE_ENV === 'development') {
            console.warn(`[useCMSPage] Failed to fetch CMS content for "${slug}", using fallback:`, errorObj);
          }
        }
      }
    };

    fetchPage();

    return () => {
      isMounted = false;
    };
  }, [slug]);

  return { content, loading, error, usedCMS };
}

/**
 * Helper function to safely access nested content blocks.
 */
export function getContentBlock<T = any>(
  page: CMSPage | null,
  path: string,
  defaultValue: T
): T {
  if (!page || !page.content_blocks) {
    return defaultValue;
  }

  const keys = path.split('.');
  let current: any = page.content_blocks;

  for (const key of keys) {
    if (current === null || current === undefined || typeof current !== 'object') {
      return defaultValue;
    }
    current = current[key];
  }

  return current !== undefined ? current : defaultValue;
}

/**
 * Helper function to get custom field value.
 */
export function getCustomField<T = any>(
  page: CMSPage | null,
  key: string,
  defaultValue: T
): T {
  if (!page || !page.custom_fields) {
    return defaultValue;
  }

  const value = page.custom_fields[key];
  return value !== undefined ? value : defaultValue;
}
