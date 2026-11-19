/**
 * The Trunc plugins. Each one links to its store listing once the matching
 * NEXT_PUBLIC_*_URL is set at build time; until then Chrome and WordPress
 * offer the .zip built by scripts/build-plugins.sh, and Shopify and Canva
 * (which can only be installed from their stores) say they are coming.
 */
export interface Integration {
  id: 'chrome' | 'wordpress' | 'shopify' | 'canva';
  name: string;
  badge: string;
  tone: string;
  summary: string;
  /** Store listing, when published. */
  storeUrl?: string;
  storeLabel: string;
  /** Direct download, for the ones users can install from a file. */
  download?: string;
  steps: string[];
  /** Steps when installing from the store instead of a file. */
  storeSteps?: string[];
}

const env = (v: string | undefined) => (v && /^https:\/\//.test(v) ? v : undefined);

export const INTEGRATIONS: Integration[] = [
  {
    id: 'chrome',
    name: 'Chrome extension',
    badge: 'CR',
    tone: 'bg-amber-500/10 text-amber-500 border-amber-500/20',
    summary: 'Shorten the page you are on in one click, copy it, show its QR code, and see your recent links.',
    storeUrl: env(process.env.NEXT_PUBLIC_CHROME_EXTENSION_URL),
    storeLabel: 'Add to Chrome',
    download: '/downloads/trunc-chrome-extension.zip',
    steps: [
      'Unzip the file.',
      'Open chrome://extensions and turn on Developer mode.',
      'Choose Load unpacked and select the unzipped folder.',
      'Open the extension and choose Sign in with Trunc.',
    ],
    storeSteps: ['Add the extension from the Chrome Web Store.', 'Open it and choose Sign in with Trunc.'],
  },
  {
    id: 'wordpress',
    name: 'WordPress plugin',
    badge: 'WP',
    tone: 'bg-sky-500/10 text-sky-600 border-sky-500/20',
    summary: 'Shorten links from your WordPress admin, and add them to posts with the [trunc] shortcode or the Trunc block.',
    storeUrl: env(process.env.NEXT_PUBLIC_WORDPRESS_PLUGIN_URL),
    storeLabel: 'Get it on WordPress.org',
    download: '/downloads/trunc-wordpress-plugin.zip',
    steps: [
      'In WordPress, go to Plugins → Add New → Upload Plugin.',
      'Choose the .zip file, install and activate it.',
      'Go to Trunc Links → Settings and choose Connect with Trunc.',
    ],
    storeSteps: ['In WordPress, go to Plugins → Add New and search for Trunc.', 'Activate it, then go to Trunc Links → Settings and choose Connect with Trunc.'],
  },
  {
    id: 'shopify',
    name: 'Shopify app',
    badge: 'SP',
    tone: 'bg-emerald-500/10 text-emerald-600 border-emerald-500/20',
    summary: 'Short links for products, collections and pages, with click counts, inside your Shopify admin.',
    storeUrl: env(process.env.NEXT_PUBLIC_SHOPIFY_APP_URL),
    storeLabel: 'Install on Shopify',
    steps: ['Install the app from the Shopify App Store.', 'In the app, choose Sign in with Trunc.'],
  },
  {
    id: 'canva',
    name: 'Canva app',
    badge: 'CV',
    tone: 'bg-cyan-500/10 text-cyan-600 border-cyan-500/20',
    summary: 'Make a short link and drop its QR code straight into your Canva design.',
    storeUrl: env(process.env.NEXT_PUBLIC_CANVA_APP_URL),
    storeLabel: 'Open in Canva',
    steps: ['Open the Trunc app from Apps in the Canva editor.', 'Choose Sign in with Trunc.'],
  },
];

/** Where the main button goes, and what it says. Null when not installable yet. */
export function installAction(i: Integration): { href: string; label: string; external: boolean } | null {
  if (i.storeUrl) return { href: i.storeUrl, label: i.storeLabel, external: true };
  if (i.download) return { href: i.download, label: 'Download .zip', external: false };
  return null;
}

/** The install steps that match the current install action. */
export function installSteps(i: Integration): string[] {
  return i.storeUrl ? i.storeSteps ?? i.steps : i.steps;
}
