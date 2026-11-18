export type CardTemplateId =
  | 'zenith'
  | 'nova'
  | 'atlas'
  | 'axiom'
  | 'vectra'
  | 'pulse'
  | 'horizon'
  | 'verdant'
  | 'skyway'
  | 'creo'
  | 'bravo'
  | 'imagine';

export type CardCollectionId =
  | 'premium-minimal'
  | 'modern-futuristic'
  | 'elegant-corporate'
  | 'creative-colourful';

/** Palette variables. Keys map to CSS custom properties: onAccent -> --vc-on-accent, etc. */
export type CardPaletteVars = Partial<
  Record<
    | 'bg'
    | 'ink'
    | 'mute'
    | 'accent'
    | 'accent2'
    | 'surface'
    | 'line'
    | 'onAccent'
    | 'plate'
    | 'qr'
    | 'g2'
    | 'g3'
    | 'glow'
    | 'sky1'
    | 'sky2',
    string
  >
>;

export interface CardPalette {
  id: string;
  name: string;
  vars: CardPaletteVars;
}

export interface CardCollection {
  id: CardCollectionId;
  name: string;
  description: string;
}

export interface CardTemplateDefinition {
  id: CardTemplateId;
  name: string;
  collection: CardCollectionId;
  category: string;
  description: string;
  audience: string;
  cta: string;
  /** Default background behind the company logo: a light plate, a dark plate, or none. */
  logo: 'light' | 'dark' | 'none';
  fonts: string[];
  palettes: CardPalette[];
}

export interface CardSocial {
  platform: string;
  url: string;
  label?: string;
}

/** Everything a card template can show. Every template renders from this one shape. */
export interface CardData {
  name: string;
  role: string;
  company: string;
  about: string;
  photo: string;
  logo: string;
  phone: string;
  whatsapp: string;
  email: string;
  website: string;
  address: string;
  social: CardSocial[];
  tags: string[];
  ctaLabel: string;
  /** Where the call to action goes. Empty means the page supplies Save Contact. */
  ctaHref: string;
  qrValue: string;
  idNumber: string;
}

/** What the builder stores in micro-page `config.card`. All optional. */
export interface CardConfig {
  logo_url?: string;
  about?: string;
  whatsapp?: string;
  cta_text?: string;
  cta_url?: string;
  qr_url?: string;
  tags?: string[];
  palette?: string;
  accent?: string;
  /** Background behind the logo. 'auto' uses the template's default. */
  logo_plate?: 'auto' | 'light' | 'dark' | 'none';
}
