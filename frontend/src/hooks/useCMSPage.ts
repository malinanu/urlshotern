import { useState, useEffect } from 'react';

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
 * Custom hook to fetch CMS page content with fallback support
 *
 * @param slug - The page slug to fetch (e.g., 'home', 'about', 'contact')
 * @param fallbackContent - Fallback content to use if CMS fetch fails
 * @returns Object containing content, loading state, error, and whether CMS was used
 *
 * @example
 * ```tsx
 * const { content, loading, usedCMS } = useCMSPage('home', {
 *   slug: 'home',
 *   title: 'Fallback Title',
 *   content: 'Fallback content...'
 * });
 *
 * if (loading) return <div>Loading...</div>;
 *
 * return (
 *   <div>
 *     <h1>{content?.title}</h1>
 *     {usedCMS && <div className="text-xs text-gray-500">Content from CMS</div>}
 *   </div>
 * );
 * ```
 */
export function useCMSPage(
  slug: string,
  fallbackContent: Partial<CMSPage> = {}
): UseCMSPageResult {
  const [content, setContent] = useState<CMSPage | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<Error | null>(null);
  const [usedCMS, setUsedCMS] = useState<boolean>(false);

  useEffect(() => {
    let isMounted = true;

    const fetchPage = async () => {
      try {
        setLoading(true);
        setError(null);

        const response = await fetch(`/api/v1/pages/${slug}`);

        if (!response.ok) {
          throw new Error(`Failed to fetch page: ${response.status} ${response.statusText}`);
        }

        const data = await response.json();

        // Handle both direct page object and wrapped response
        const pageData = data.page || data;

        if (isMounted && pageData) {
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

          // Use fallback content on error
          setContent({
            slug,
            title: '',
            content: '',
            ...fallbackContent,
          } as CMSPage);

          // Log error in development
          if (process.env.NODE_ENV === 'development') {
            console.warn(`[useCMSPage] Failed to fetch CMS content for "${slug}", using fallback:`, errorObj);
          }
        }
      } finally {
        if (isMounted) {
          setLoading(false);
        }
      }
    };

    fetchPage();

    return () => {
      isMounted = false;
    };
  }, [slug, fallbackContent]);

  return { content, loading, error, usedCMS };
}

/**
 * Helper function to safely access nested content blocks
 *
 * @param page - CMS page object
 * @param path - Dot-notation path to the content block (e.g., 'hero.title')
 * @param defaultValue - Default value if path not found
 * @returns The content at the specified path or default value
 *
 * @example
 * ```tsx
 * const heroTitle = getContentBlock(page, 'hero.title', 'Default Title');
 * const stats = getContentBlock(page, 'social_proof.stats', []);
 * ```
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
 * Helper function to get custom field value
 *
 * @param page - CMS page object
 * @param key - Field key
 * @param defaultValue - Default value if key not found
 * @returns The field value or default value
 *
 * @example
 * ```tsx
 * const inquiryTypes = getCustomField(page, 'inquiry_types', []);
 * ```
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
