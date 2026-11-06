export type MicroPageTheme =
  | 'executive-obsidian'
  | 'california-minimalist'
  | 'cyber-neon'
  | 'nordic-sage'
  | 'swiss-grid'
  | 'sunset-glass'
  | 'cobalt-steel'
  | 'editorial-monocle'
  | 'neo-brutalist'
  | 'aura-holographic';

export interface VCardContact {
  first_name: string;
  last_name: string;
  organization: string;
  title: string;
  email: string;
  phone: string;
  cell_phone?: string;
  work_phone?: string;
  address?: string;
  city?: string;
  state?: string;
  postal_code?: string;
  country?: string;
  website?: string;
  note?: string;
}

export type SocialPlatform =
  | 'twitter'
  | 'linkedin'
  | 'github'
  | 'instagram'
  | 'youtube'
  | 'whatsapp'
  | 'telegram'
  | 'email'
  | 'phone'
  | 'website'
  | 'discord'
  | 'tiktok';

export interface MicroSocialLink {
  platform: SocialPlatform | string;
  url: string;
  label?: string;
}

export interface MicroLinkItem {
  id: string;
  title: string;
  url: string;
  icon?: string;
  description?: string;
  is_highlighted?: boolean;
  is_active: boolean;
  sort_order: number;
  click_count?: number;
}

export interface MicroPage {
  id: number;
  user_id: number;
  slug: string;
  title: string;
  bio: string;
  avatar_url: string;
  cover_url: string;
  theme: MicroPageTheme;
  custom_css: string;
  config: Record<string, any>;
  vcard: VCardContact | null;
  social_links: MicroSocialLink[];
  links: MicroLinkItem[];
  is_published: boolean;
  view_count: number;
  created_at: string;
  updated_at: string;
}

export interface CreateMicroPagePayload {
  slug: string;
  title: string;
  bio?: string;
  avatar_url?: string;
  cover_url?: string;
  theme?: MicroPageTheme;
  custom_css?: string;
  config?: Record<string, any>;
  vcard?: VCardContact;
  social_links?: MicroSocialLink[];
  links?: MicroLinkItem[];
  is_published?: boolean;
}

export interface UpdateMicroPagePayload {
  slug?: string;
  title?: string;
  bio?: string;
  avatar_url?: string;
  cover_url?: string;
  theme?: MicroPageTheme;
  custom_css?: string;
  config?: Record<string, any>;
  vcard?: VCardContact;
  social_links?: MicroSocialLink[];
  links?: MicroLinkItem[];
  is_published?: boolean;
}

export interface TemplateDefinition {
  id: MicroPageTheme;
  name: string;
  description: string;
  tag: string;
  previewBg: string;
  accentColor: string;
}
