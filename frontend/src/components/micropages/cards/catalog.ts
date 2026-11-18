import catalogJson from './catalog.json';
import type {
  CardCollection,
  CardCollectionId,
  CardPalette,
  CardPaletteVars,
  CardTemplateDefinition,
  CardTemplateId,
} from './types';

const catalog = catalogJson as {
  collections: CardCollection[];
  templates: CardTemplateDefinition[];
};

export const CARD_COLLECTIONS: CardCollection[] = catalog.collections;
export const CARD_TEMPLATES: CardTemplateDefinition[] = catalog.templates;

const BY_ID = new Map<string, CardTemplateDefinition>(CARD_TEMPLATES.map((t) => [t.id, t]));

export function isCardTemplateId(id: string | undefined | null): id is CardTemplateId {
  return !!id && BY_ID.has(id);
}

export function getCardTemplate(id: string | undefined | null): CardTemplateDefinition | undefined {
  return id ? BY_ID.get(id) : undefined;
}

export function templatesInCollection(id: CardCollectionId): CardTemplateDefinition[] {
  return CARD_TEMPLATES.filter((t) => t.collection === id);
}

export function findPalette(tpl: CardTemplateDefinition, paletteId?: string): CardPalette {
  return tpl.palettes.find((p) => p.id === paletteId) || tpl.palettes[0];
}

const VAR_NAMES: Record<keyof CardPaletteVars, string> = {
  bg: '--vc-bg',
  ink: '--vc-ink',
  mute: '--vc-mute',
  accent: '--vc-accent',
  accent2: '--vc-accent2',
  surface: '--vc-surface',
  line: '--vc-line',
  onAccent: '--vc-on-accent',
  plate: '--vc-plate',
  qr: '--vc-qr',
  g2: '--vc-g2',
  g3: '--vc-g3',
  glow: '--vc-glow',
  sky1: '--vc-sky1',
  sky2: '--vc-sky2',
};

const HEX = /^#([0-9a-f]{6})$/i;

export function isHexColor(v: string | undefined | null): v is string {
  return !!v && HEX.test(v);
}

/** Black or near-black text for light accents, white for dark ones (WCAG relative luminance). */
export function readableOn(hex: string): string {
  const m = HEX.exec(hex);
  if (!m) return '#ffffff';
  const n = parseInt(m[1], 16);
  const lin = (c: number) => {
    const s = c / 255;
    return s <= 0.03928 ? s / 12.92 : Math.pow((s + 0.055) / 1.055, 2.4);
  };
  const L = 0.2126 * lin((n >> 16) & 255) + 0.7152 * lin((n >> 8) & 255) + 0.0722 * lin(n & 255);
  return L > 0.4 ? '#14110c' : '#ffffff';
}

/**
 * Inline CSS variables for a card: the chosen palette, then an optional custom accent.
 * A custom accent changes only the accent colour; the layout and type stay the template's own.
 */
export function cardStyleVars(
  tpl: CardTemplateDefinition,
  paletteId?: string,
  accent?: string
): Record<string, string> {
  const out: Record<string, string> = {};
  const vars = findPalette(tpl, paletteId).vars;
  (Object.keys(vars) as (keyof CardPaletteVars)[]).forEach((k) => {
    const v = vars[k];
    if (v) out[VAR_NAMES[k]] = v;
  });
  if (isHexColor(accent)) {
    out['--vc-accent'] = accent;
    out['--vc-on-accent'] = readableOn(accent);
  }
  return out;
}

/** Each template's default colours (they match the stylesheet defaults), for swatches and the colour picker. */
const DEFAULTS: Record<CardTemplateId, { bg: string; accent: string; ink: string; onAccent: string }> = {
  zenith: { bg: '#0b0b0d', accent: '#c9aa6a', ink: '#f4eee2', onAccent: '#17130a' },
  nova: { bg: '#f7f2e9', accent: '#a98b5b', ink: '#2a2723', onAccent: '#fbf9f4' },
  atlas: { bg: '#f5ecdc', accent: '#b48a3c', ink: '#3a2e22', onAccent: '#fffaf0' },
  axiom: { bg: '#090c13', accent: '#22d3ee', ink: '#e8f1f8', onAccent: '#031318' },
  vectra: { bg: '#121418', accent: '#c9ced6', ink: '#eceff3', onAccent: '#101216' },
  pulse: { bg: '#0a0f2c', accent: '#3d6bff', ink: '#f1f3ff', onAccent: '#ffffff' },
  horizon: { bg: '#ffffff', accent: '#1f5fd6', ink: '#0f2748', onAccent: '#ffffff' },
  verdant: { bg: '#0e2a1e', accent: '#c8a45e', ink: '#f2eee1', onAccent: '#0e2a1e' },
  skyway: { bg: '#f3f9ff', accent: '#1672c4', ink: '#0d2c4f', onAccent: '#ffffff' },
  creo: { bg: '#1b0836', accent: '#ff7a59', ink: '#fdf7ff', onAccent: '#2a0a1a' },
  bravo: { bg: '#f6eedb', accent: '#ff6a13', ink: '#111111', onAccent: '#111111' },
  imagine: { bg: '#17181c', accent: '#ffd60a', ink: '#f7f4ea', onAccent: '#17181c' },
};

/** The accent colour a palette resolves to (before any custom accent). */
export function paletteAccent(tpl: CardTemplateDefinition, paletteId?: string): string {
  return findPalette(tpl, paletteId).vars.accent || DEFAULTS[tpl.id].accent;
}

/** Two colours that represent a palette in a swatch. */
export function paletteSwatch(tpl: CardTemplateDefinition, palette: CardPalette): { bg: string; accent: string } {
  return {
    bg: palette.vars.bg || DEFAULTS[tpl.id].bg,
    accent: palette.vars.accent || DEFAULTS[tpl.id].accent,
  };
}

/** Page colours a card template starts with (Microsites stores them in the page theme). */
export function cardThemeDefaults(id: CardTemplateId): { bg: string; accent: string; ink: string; onAccent: string } {
  return DEFAULTS[id];
}
