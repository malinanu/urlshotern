import type { MicroPage } from '@/types/micropage';
import { API_ENDPOINTS } from '@/lib/api-config';
import type { CardConfig, CardData, CardSocial, CardTemplateDefinition } from './types';

/** The card settings the builder stores under `config.card`. */
export function readCardConfig(page: Partial<MicroPage> | null | undefined): CardConfig {
  const cfg = (page?.config || {}) as Record<string, unknown>;
  const card = cfg.card;
  return card && typeof card === 'object' ? (card as CardConfig) : {};
}

/** Uploaded images are stored as "/uploads/mp/..." paths; this points them at the API host. */
export function assetUrl(u: string | undefined | null): string {
  const v = (u || '').trim();
  if (!v.startsWith('/') || v.startsWith('//')) return v;
  return API_ENDPOINTS.micropages.assets.replace(/\/api\/v1\/micropages\/assets$/, '') + v;
}

/** An image the card may show: an http(s) URL or one of our own upload paths. */
export function imageSrc(u: string | undefined | null): string {
  const v = (u || '').trim();
  if (v.startsWith('/uploads/')) return assetUrl(v);
  return safeHttpUrl(v);
}

/** Returns the URL when it is http(s), otherwise an empty string. */
export function safeHttpUrl(url: string | undefined | null): string {
  const v = (url || '').trim();
  if (!v) return '';
  try {
    const u = new URL(v);
    return u.protocol === 'http:' || u.protocol === 'https:' ? u.toString() : '';
  } catch {
    return '';
  }
}

/** Accepts "example.com/path" as well as full URLs. */
export function websiteHref(website: string): string {
  const v = (website || '').trim();
  if (!v) return '';
  if (/^[a-z][a-z0-9+.-]*:/i.test(v)) return safeHttpUrl(v);
  return safeHttpUrl('https://' + v);
}

export function telHref(phone: string): string {
  const digits = (phone || '').replace(/[^\d+]/g, '');
  return digits ? `tel:${digits}` : '';
}

export function whatsappHref(num: string): string {
  const digits = (num || '').replace(/\D/g, '');
  return digits ? `https://wa.me/${digits}` : '';
}

export function mapsHref(address: string): string {
  const v = (address || '').trim();
  return v ? `https://www.google.com/maps/search/?api=1&query=${encodeURIComponent(v)}` : '';
}

/** Step used by the stylesheets to shrink long names instead of clipping them. */
export function nameLengthClass(name: string): 's' | 'm' | 'l' | 'xl' {
  const n = [...(name || '')].length;
  if (n <= 14) return 's';
  if (n <= 22) return 'm';
  if (n <= 32) return 'l';
  return 'xl';
}

export function initials(name: string): string {
  const parts = (name || '').trim().split(/\s+/).filter(Boolean);
  return parts
    .slice(0, 2)
    .map((p) => [...p][0] || '')
    .join('')
    .toUpperCase();
}

const CONTACT_PLATFORMS = new Set(['email', 'mail', 'phone', 'website', 'whatsapp']);

/**
 * One mapping from a micro-page to the card. Every template reads this same shape, so a
 * template never has user data written into it.
 */
export function toCardData(
  page: Partial<MicroPage>,
  tpl: CardTemplateDefinition,
  opts: { pageUrl: string }
): CardData {
  const card = readCardConfig(page);
  const vc = page.vcard;
  const social: CardSocial[] = (page.social_links || []).filter(
    (s) => s && s.url && !CONTACT_PLATFORMS.has(String(s.platform).toLowerCase())
  );

  // WhatsApp: the card setting, else a whatsapp social link (wa.me/<number>).
  let whatsapp = (card.whatsapp || '').trim();
  if (!whatsapp) {
    const wa = (page.social_links || []).find((s) => String(s.platform).toLowerCase() === 'whatsapp');
    whatsapp = wa ? (wa.url.match(/(\d[\d\s-]{5,})/)?.[1] || '').trim() : '';
  }

  const address = [vc?.address, vc?.city, vc?.state, vc?.postal_code, vc?.country]
    .map((p) => (p || '').trim())
    .filter(Boolean)
    .join(', ');

  const id = Number(page.id || 0);
  return {
    name: (page.title || '').trim() || 'Your Name',
    role: (vc?.title || '').trim(),
    company: (vc?.organization || '').trim(),
    about: (card.about ?? page.bio ?? '').trim(),
    photo: imageSrc(page.avatar_url),
    logo: imageSrc(card.logo_url),
    phone: (vc?.phone || vc?.cell_phone || vc?.work_phone || '').trim(),
    whatsapp,
    email: (vc?.email || '').trim(),
    website: (vc?.website || '').trim(),
    address,
    social,
    tags: (card.tags || []).map((t) => String(t).trim()).filter(Boolean).slice(0, 4),
    ctaLabel: (card.cta_text || '').trim() || tpl.cta,
    ctaHref: safeHttpUrl(card.cta_url),
    qrValue: safeHttpUrl(card.qr_url) || opts.pageUrl,
    idNumber: `TC-${String(id).padStart(6, '0')}`,
  };
}

/** Sample content for the gallery thumbnails and previews. Not a real person. */
export const SAMPLE_CARD: CardData = {
  name: 'Alexandra Reyes',
  role: 'Chief Executive Officer',
  company: 'Meridian Group',
  about: 'Building resilient, human-centred businesses across 14 markets for more than two decades.',
  photo: '',
  logo: '',
  phone: '+1 (415) 555-0142',
  whatsapp: '+1 415 555 0142',
  email: 'alexandra@meridian.example',
  website: 'meridiangroup.example',
  address: '88 Market Street, San Francisco, CA',
  social: [
    { platform: 'linkedin', url: 'https://linkedin.com/' },
    { platform: 'instagram', url: 'https://instagram.com/' },
    { platform: 'x', url: 'https://x.com/' },
    { platform: 'youtube', url: 'https://youtube.com/' },
  ],
  tags: ['Branding', 'Strategy', 'Identity'],
  ctaLabel: '',
  ctaHref: '',
  qrValue: 'https://trunc.site/m/alexandra-reyes',
  idNumber: 'TC-004217',
};

/** A neutral portrait placeholder (inline SVG), so previews never show a broken image. */
export const SAMPLE_PHOTO =
  'data:image/svg+xml;utf8,' +
  encodeURIComponent(
    `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 400 480"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#6b7a99"/><stop offset="1" stop-color="#2a2f45"/></linearGradient></defs><rect width="400" height="480" fill="url(#g)"/><circle cx="200" cy="190" r="86" fill="#e8c9ac"/><path d="M104 150c8-70 64-100 112-92 52 8 84 56 80 100-24-40-64-52-96-48-40 4-76 20-96 40z" fill="#2b2420"/><path d="M40 480c6-92 70-130 160-130s154 38 160 130z" fill="#1f2a3a"/><path d="M160 350l40 56 40-56" fill="#f3f3f3"/></svg>`
  );
