import DOMPurify from 'isomorphic-dompurify';

/**
 * Sanitize an untrusted/CMS-authored HTML string before it is passed to
 * dangerouslySetInnerHTML. Strips scripts, event handlers, and other XSS
 * vectors while preserving the formatting markup CMS content relies on.
 *
 * isomorphic-dompurify runs in both the browser and Node (SSR), so this is
 * safe to call from server and client components alike.
 */
export function sanitizeHtml(dirty: string | null | undefined): string {
  if (!dirty) return '';
  return DOMPurify.sanitize(dirty, {
    // Disallow embedding of external/script content; allow target for links.
    ADD_ATTR: ['target'],
    FORBID_TAGS: ['script', 'style', 'iframe', 'object', 'embed', 'form'],
    FORBID_ATTR: ['onerror', 'onload', 'onclick', 'onmouseover'],
  });
}
