/**
 * Microsites (micro-landing pages & digital vCards) — shared types, block
 * catalog and helpers for the /microsites builder. Shapes mirror
 * internal/models/microsite.go.
 */

export const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

/** Build-time switch; mirrors the backend ENABLE_MICROSITES flag. */
export const MICROSITES_ENABLED = process.env.NEXT_PUBLIC_ENABLE_MICROSITES === 'true';

export type BlockType =
  | 'profile_header'
  | 'vcard_action'
  | 'link_list'
  | 'contact_buttons'
  | 'social_icons'
  | 'text'
  | 'image'
  | 'divider';

/**
 * Block payloads are validated and normalized server-side (see
 * internal/services/microsite_blocks.go); the editor treats them as loose records.
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export type BlockData = Record<string, any>;

export interface MicrositeBlock {
  id: string;
  type: BlockType;
  schema_version?: number;
  hidden?: boolean;
  data: BlockData;
}

export interface MicrositeBackground {
  type: 'solid' | 'linear' | 'radial' | 'image';
  color?: string;
  stops?: string[];
  angle?: number;
  image_url?: string;
  overlay?: number;
}

export interface MicrositeTheme {
  background: MicrositeBackground;
  card_fill: string;
  accent: string;
  button_text: string;
  text: string;
  font: string;
  button_radius: 'sharp' | 'rounded' | 'pill';
  shadow: 'none' | 'soft' | 'strong';
  avatar_border: 'none' | 'thin' | 'thick';
}

export interface MicrositeSettings {
  hide_badge: boolean;
  allow_indexing: boolean;
  seo_title?: string;
  seo_description?: string;
  og_image_url?: string;
  favicon_url?: string;
}

export interface MicrositeContent {
  theme: MicrositeTheme;
  blocks: MicrositeBlock[];
  settings: MicrositeSettings;
}

export interface Microsite {
  id: number;
  user_id: number;
  slug: string;
  title: string;
  template: string;
  status: 'draft' | 'published';
  draft: MicrositeContent;
  published?: MicrositeContent;
  version: number;
  short_code?: string;
  published_at?: string;
  created_at: string;
  updated_at: string;
  public_url: string;
}

export interface MicrositeSummary {
  id: number;
  slug: string;
  title: string;
  template: string;
  status: 'draft' | 'published';
  short_code?: string;
  views_30d: number;
  published_at?: string;
  updated_at: string;
}

export interface MicrositeEntitlements {
  max_pages: number;
  can_remove_badge: boolean;
  analytics_days: number;
  used_pages: number;
  can_create_more: boolean;
}

export interface MicrositeTemplateInfo {
  id: string;
  name: string;
  description: string;
}

export interface MicrositeCount {
  label: string;
  count: number;
}

export interface MicrositeAnalytics {
  days: number;
  views: number;
  unique_visitors: number;
  link_clicks: number;
  vcard_downloads: number;
  contact_actions: Record<string, number>;
  ctr: number;
  contact_save_rate: number;
  daily: { date: string; views: number; unique_visitors: number; clicks: number; vcard_downloads: number }[];
  top_referrers: MicrositeCount[];
  referrer_sources: MicrositeCount[];
  devices: MicrositeCount[];
  operating_systems: MicrositeCount[];
  countries: MicrositeCount[];
  blocks: { block_id: string; item_id?: string; event_type: string; label: string; count: number }[];
}

export interface ApiError {
  status: number;
  error?: string;
  message: string;
  field?: string;
}

/** Turn an uploaded-asset path (/uploads/...) into a URL the dashboard can load. */
export function assetSrc(url?: string): string {
  if (!url) return '';
  return url.startsWith('/') ? `${API_BASE_URL}${url}` : url;
}

export function newId(prefix: 'blk' | 'itm'): string {
  const rand =
    typeof crypto !== 'undefined' && 'getRandomValues' in crypto
      ? Array.from(crypto.getRandomValues(new Uint8Array(6)), (b) => b.toString(16).padStart(2, '0')).join('')
      : Math.random().toString(16).slice(2, 14);
  return `${prefix}_${rand}`;
}

export interface BlockCatalogEntry {
  type: BlockType;
  name: string;
  description: string;
  create: () => BlockData;
}

export const BLOCK_CATALOG: BlockCatalogEntry[] = [
  {
    type: 'profile_header',
    name: 'Profile',
    description: 'Avatar, name, title and company',
    create: () => ({ display_name: 'Your Name', title: '', company: '', bio: '' }),
  },
  {
    type: 'vcard_action',
    name: 'Save Contact',
    description: 'One-tap .vcf download for iPhone and Android',
    create: () => ({ button_text: 'Save Contact', first_name: 'Your', last_name: 'Name' }),
  },
  {
    type: 'contact_buttons',
    name: 'Contact buttons',
    description: 'Call, text, email and WhatsApp',
    create: () => ({ items: [{ id: newId('itm'), kind: 'call', value: '', label: 'Call' }] }),
  },
  {
    type: 'link_list',
    name: 'Links',
    description: 'A stack of link buttons',
    create: () => ({ items: [{ id: newId('itm'), label: 'My website', url: '', highlight: false }] }),
  },
  {
    type: 'social_icons',
    name: 'Social icons',
    description: 'LinkedIn, Instagram, GitHub and more',
    create: () => ({ items: [{ id: newId('itm'), platform: 'linkedin', url: '' }] }),
  },
  { type: 'text', name: 'Text', description: 'A paragraph of plain text', create: () => ({ text: '', align: 'center' }) },
  { type: 'image', name: 'Image', description: 'A picture, optionally linked', create: () => ({ url: '', alt: '' }) },
  { type: 'divider', name: 'Divider', description: 'A line or spacing', create: () => ({ style: 'line' }) },
];

export const BLOCK_NAMES: Record<BlockType, string> = Object.fromEntries(
  BLOCK_CATALOG.map((b) => [b.type, b.name])
) as Record<BlockType, string>;

export const SOCIAL_PLATFORMS: { id: string; name: string }[] = [
  { id: 'linkedin', name: 'LinkedIn' },
  { id: 'github', name: 'GitHub' },
  { id: 'x', name: 'X' },
  { id: 'instagram', name: 'Instagram' },
  { id: 'facebook', name: 'Facebook' },
  { id: 'youtube', name: 'YouTube' },
  { id: 'tiktok', name: 'TikTok' },
  { id: 'threads', name: 'Threads' },
  { id: 'website', name: 'Website' },
];

export const FONTS: { id: string; name: string }[] = [
  { id: 'system', name: 'System' },
  { id: 'serif', name: 'Serif' },
  { id: 'mono', name: 'Monospace' },
  { id: 'rounded', name: 'Rounded' },
  { id: 'humanist', name: 'Humanist' },
  { id: 'geometric', name: 'Geometric' },
];

export interface ThemePreset {
  name: string;
  theme: Partial<MicrositeTheme>;
}

export const THEME_PRESETS: ThemePreset[] = [
  {
    name: 'Ocean',
    theme: {
      background: { type: 'linear', stops: ['#1e3c72', '#2a5298'], angle: 135 },
      card_fill: '#ffffff', accent: '#2563eb', button_text: '#ffffff', text: '#111827',
    },
  },
  {
    name: 'Sunset',
    theme: {
      background: { type: 'linear', stops: ['#f857a6', '#ff5858'], angle: 135 },
      card_fill: '#ffffff', accent: '#111827', button_text: '#ffffff', text: '#111827',
    },
  },
  {
    name: 'Forest',
    theme: {
      background: { type: 'radial', stops: ['#134e5e', '#71b280'] },
      card_fill: '#f8faf5', accent: '#166534', button_text: '#ffffff', text: '#14251b',
    },
  },
  {
    name: 'Midnight',
    theme: {
      background: { type: 'solid', color: '#0b1020' },
      card_fill: '#151b2e', accent: '#00d2ff', button_text: '#0b1020', text: '#e5e7eb',
    },
  },
  {
    name: 'Paper',
    theme: {
      background: { type: 'solid', color: '#f3f0e8' },
      card_fill: '#fffdf8', accent: '#3f3f46', button_text: '#ffffff', text: '#27272a',
    },
  },
];

/**
 * Turn a server validation field path (e.g. "blocks[2].data.items[0].url")
 * into something readable ("Links › item 1 › url: must be a valid web address").
 */
export function describeValidation(field: string | undefined, message: string, blocks: MicrositeBlock[]): string {
  if (!field) return message;
  const m = field.match(/^blocks\[(\d+)\](?:\.data)?(?:\.items\[(\d+)\])?(?:\.(\w+))?$/);
  if (m) {
    const block = blocks[Number(m[1])];
    const parts = [block ? BLOCK_NAMES[block.type] ?? block.type : 'Block'];
    if (m[2] !== undefined) parts.push(`item ${Number(m[2]) + 1}`);
    if (m[3] && m[3] !== 'data') parts.push(m[3].replace(/_/g, ' '));
    return `${parts.join(' › ')}: ${message}`;
  }
  if (field.startsWith('theme.')) return `Design › ${field.slice(6).replace(/[._]/g, ' ')}: ${message}`;
  if (field.startsWith('settings.')) return `Settings › ${field.slice(9).replace(/_/g, ' ')}: ${message}`;
  const names: Record<string, string> = { title: 'Page name', slug: 'Address' };
  return `${names[field] ?? field}: ${message}`;
}

/** Small fetch wrapper: JSON in/out, normalized errors. */
export async function apiFetch<T>(url: string, token: string | null, init: RequestInit = {}): Promise<T> {
  const headers: Record<string, string> = { ...(init.headers as Record<string, string>) };
  if (token) headers.Authorization = `Bearer ${token}`;
  if (init.body && !(init.body instanceof FormData)) headers['Content-Type'] = 'application/json';
  const res = await fetch(url, { ...init, headers });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err: ApiError = {
      status: res.status,
      error: data.error,
      message: data.message || `Request failed (${res.status})`,
      field: data.field,
    };
    throw err;
  }
  return data as T;
}

export function isApiError(e: unknown): e is ApiError {
  return typeof e === 'object' && e !== null && 'status' in e && 'message' in e;
}
