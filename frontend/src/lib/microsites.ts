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
  | 'divider'
  | 'product_card'
  | 'product_carousel'
  | 'event_details'
  | 'countdown'
  | 'agenda'
  | 'map'
  | 'video'
  | 'audio'
  | 'form';

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

/** Template styles; "classic" is the original centered card. */
export type TemplateStyle = 'classic' | 'letterpress' | 'board' | 'flyer' | 'programme';

export interface MicrositeTheme {
  style?: TemplateStyle;
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
  can_use_forms: boolean;
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
  form_submits: number;
  calendar_adds: number;
  video_plays: number;
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

export interface MicrositeSubmission {
  id: number;
  block_id: string;
  kind: 'newsletter' | 'contact';
  email: string;
  name?: string;
  message?: string;
  created_at: string;
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

export type BlockGroup = 'basics' | 'contact' | 'shop' | 'event' | 'media' | 'forms';

export const BLOCK_GROUPS: { id: BlockGroup; name: string }[] = [
  { id: 'basics', name: 'Basics' },
  { id: 'contact', name: 'Contact' },
  { id: 'shop', name: 'Shop' },
  { id: 'event', name: 'Event' },
  { id: 'media', name: 'Video & audio' },
  { id: 'forms', name: 'Forms' },
];

export interface BlockCatalogEntry {
  type: BlockType;
  name: string;
  description: string;
  group: BlockGroup;
  /** Needs the forms entitlement (Pro and Ultra). */
  requiresForms?: boolean;
  create: () => BlockData;
}

/** The visitor's IANA time zone, e.g. "Asia/Colombo" (UTC if unknown). */
export function browserTimeZone(): string {
  try {
    return Intl.DateTimeFormat().resolvedOptions().timeZone || 'UTC';
  } catch {
    return 'UTC';
  }
}

/** A local "YYYY-MM-DDTHH:mm" value `days` from now at `hour`:00. */
export function localDateTime(days: number, hour: number): string {
  const d = new Date();
  d.setDate(d.getDate() + days);
  d.setHours(hour, 0, 0, 0);
  const p = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${p(d.getMonth() + 1)}-${p(d.getDate())}T${p(d.getHours())}:${p(d.getMinutes())}`;
}

const productData = (): BlockData => ({
  name: 'Product name',
  price: '',
  compare_price: '',
  badge: '',
  description: '',
  image_url: '',
  buy_url: '',
  button_text: 'Buy now',
});

/** A carousel item: product fields plus a stable item id. */
export const newProduct = (): BlockData => ({ id: newId('itm'), ...productData() });

export const BLOCK_CATALOG: BlockCatalogEntry[] = [
  {
    type: 'profile_header',
    name: 'Profile',
    description: 'Avatar, name, title and company',
    group: 'basics',
    create: () => ({ display_name: 'Your Name', title: '', company: '', bio: '' }),
  },
  {
    type: 'link_list',
    name: 'Links',
    description: 'A stack of link buttons',
    group: 'basics',
    create: () => ({ items: [{ id: newId('itm'), label: 'My website', url: '', highlight: false }] }),
  },
  { type: 'text', name: 'Text', description: 'A paragraph of plain text', group: 'basics', create: () => ({ text: '', align: 'center' }) },
  { type: 'image', name: 'Image', description: 'A picture, optionally linked', group: 'basics', create: () => ({ url: '', alt: '' }) },
  { type: 'divider', name: 'Divider', description: 'A line or spacing', group: 'basics', create: () => ({ style: 'line' }) },
  {
    type: 'vcard_action',
    name: 'Save Contact',
    description: 'One-tap .vcf download for iPhone and Android',
    group: 'contact',
    create: () => ({ button_text: 'Save Contact', first_name: 'Your', last_name: 'Name' }),
  },
  {
    type: 'contact_buttons',
    name: 'Contact buttons',
    description: 'Call, text, email and WhatsApp',
    group: 'contact',
    create: () => ({ items: [{ id: newId('itm'), kind: 'call', value: '', label: 'Call' }] }),
  },
  {
    type: 'social_icons',
    name: 'Social icons',
    description: 'LinkedIn, Instagram, GitHub and more',
    group: 'contact',
    create: () => ({ items: [{ id: newId('itm'), platform: 'linkedin', url: '' }] }),
  },
  {
    type: 'product_card',
    name: 'Product',
    description: 'Photo, price, badge and a buy button',
    group: 'shop',
    create: productData,
  },
  {
    type: 'product_carousel',
    name: 'Product carousel',
    description: 'A swipeable strip of products',
    group: 'shop',
    create: () => ({ title: 'This week', items: [newProduct(), newProduct()] }),
  },
  {
    type: 'event_details',
    name: 'Event details',
    description: 'Date, venue, tickets and add-to-calendar',
    group: 'event',
    create: () => ({
      title: 'Event name',
      start: localDateTime(30, 9),
      end: localDateTime(30, 17),
      timezone: browserTimeZone(),
      venue: '',
      address: '',
      ticket_url: '',
      ticket_text: 'Get tickets',
      show_add_to_calendar: true,
    }),
  },
  {
    type: 'countdown',
    name: 'Countdown',
    description: 'Days, hours and minutes to a moment',
    group: 'event',
    create: () => ({ target: localDateTime(30, 9), timezone: browserTimeZone(), label: 'Starts in', done_text: 'Happening now' }),
  },
  {
    type: 'agenda',
    name: 'Agenda',
    description: 'A timetable of sessions',
    group: 'event',
    create: () => ({ title: 'Programme', items: [{ id: newId('itm'), time: '09:00', title: 'Opening', speaker: '', track: '', detail: '' }] }),
  },
  {
    type: 'map',
    name: 'Map & directions',
    description: 'Address, directions link and a tap-to-load map',
    group: 'event',
    create: () => ({ label: 'Venue', address: '' }),
  },
  {
    type: 'video',
    name: 'Video',
    description: 'YouTube or Vimeo, loads when tapped',
    group: 'media',
    create: () => ({ url: '', title: '' }),
  },
  {
    type: 'audio',
    name: 'Audio',
    description: 'A Spotify track, album, playlist or episode',
    group: 'media',
    create: () => ({ url: '', title: '' }),
  },
  {
    type: 'form',
    name: 'Newsletter or contact form',
    description: 'Collect emails or messages in your Inbox',
    group: 'forms',
    requiresForms: true,
    create: () => ({ kind: 'newsletter', title: 'Get updates', description: '', button_text: 'Subscribe', notify: true }),
  },
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
  { id: 'schibsted', name: 'Schibsted Grotesk' },
  { id: 'martian', name: 'Martian Mono' },
  { id: 'barlow', name: 'Barlow Condensed' },
  { id: 'hanken', name: 'Hanken Grotesk' },
];

export interface TemplateStyleInfo {
  id: TemplateStyle;
  name: string;
  description: string;
  /** Applied when the owner switches to this style (colors stay editable). */
  theme: Partial<MicrositeTheme>;
}

/** Mirrors the styled themes in internal/services/microsite_templates.go. */
export const TEMPLATE_STYLES: TemplateStyleInfo[] = [
  { id: 'classic', name: 'Classic', description: 'A centered card on your own background', theme: { font: 'system' } },
  {
    id: 'letterpress',
    name: 'Letterpress card',
    description: 'A thick business card with painted edges',
    theme: { font: 'schibsted', accent: '#2340d6', card_fill: '#ffffff', text: '#14161c', button_text: '#ffffff' },
  },
  {
    id: 'board',
    name: 'Departure board',
    description: 'Links as departures, the newest boarding now',
    theme: { font: 'martian', accent: '#ffd400', card_fill: '#1b2a6b', text: '#ffffff', button_text: '#1b2a6b' },
  },
  {
    id: 'flyer',
    name: 'Weekly flyer',
    description: 'Big prices, deal badges and a bold masthead',
    theme: { font: 'barlow', accent: '#d7261e', card_fill: '#ffffff', text: '#17181c', button_text: '#ffffff' },
  },
  {
    id: 'programme',
    name: 'Programme',
    description: 'A printed event programme with a timetable',
    theme: { font: 'hanken', accent: '#0e7c66', card_fill: '#ffffff', text: '#111318', button_text: '#ffffff' },
  },
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
