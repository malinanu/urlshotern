import { API_ENDPOINTS } from '@/lib/api-config';
import { getAccessToken } from '@/utils/tokenValidation';

/**
 * Uploads a profile photo or company logo. The server sniffs the real file type (JPEG or PNG
 * only), re-encodes it and returns a path such as /uploads/mp/<user>/<hash>.jpg.
 */
export async function uploadCardImage(file: File): Promise<string> {
  const form = new FormData();
  form.append('file', file);
  const res = await fetch(API_ENDPOINTS.micropages.assets, {
    method: 'POST',
    headers: { Authorization: `Bearer ${getAccessToken()}` },
    body: form,
  });
  const body = (await res.json().catch(() => ({}))) as { url?: string; message?: string };
  if (!res.ok || !body.url) {
    throw new Error(body.message || 'Upload failed. Use a JPEG or PNG under 2 MB.');
  }
  return body.url;
}
