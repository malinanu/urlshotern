import { toJpeg, toPng } from 'html-to-image';
import { qrPngDataUrl } from './qr';

/** The card's design size. Exports are rendered at this size times the scale. */
export const CARD_W = 600;
export const CARD_H = 1050;

export type CardImageFormat = 'png' | 'jpeg';

const isSafari = () =>
  typeof navigator !== 'undefined' && /^((?!chrome|android).)*safari/i.test(navigator.userAgent);

async function settle(root: HTMLElement) {
  if (typeof document !== 'undefined' && 'fonts' in document) {
    try {
      await document.fonts.ready;
    } catch {
      /* fonts API unavailable: carry on */
    }
  }
  const imgs = Array.from(root.querySelectorAll('img'));
  await Promise.all(
    imgs.map((img) =>
      img.complete
        ? Promise.resolve()
        : new Promise<void>((res) => {
            img.onload = () => res();
            img.onerror = () => res();
          })
    )
  );
}

/**
 * Renders the card to an image at its design size (600 x 1050) times `scale`.
 * A copy is mounted off-screen at exactly 600px wide, so the card's em-based sizing resolves
 * the same way every time, whatever size it is shown at on the page.
 */
export async function renderCardImage(card: HTMLElement, format: CardImageFormat, scale = 2): Promise<string> {
  const host = document.createElement('div');
  host.setAttribute('aria-hidden', 'true');
  host.style.cssText = `position:fixed;left:-10000px;top:0;width:${CARD_W}px;pointer-events:none;`;
  const clone = card.cloneNode(true) as HTMLElement;
  clone.style.width = `${CARD_W}px`;
  clone.style.maxWidth = 'none';
  host.appendChild(clone);
  document.body.appendChild(host);
  try {
    await settle(clone);
    const opts = {
      width: CARD_W,
      height: CARD_H,
      pixelRatio: scale,
      cacheBust: false,
      quality: 0.95,
      backgroundColor: format === 'jpeg' ? '#ffffff' : undefined,
    };
    const render = format === 'png' ? toPng : toJpeg;
    // Safari can drop fonts and images on the first pass.
    if (isSafari()) await render(clone, opts);
    return await render(clone, opts);
  } finally {
    host.remove();
  }
}

export function downloadDataUrl(dataUrl: string, filename: string) {
  const a = document.createElement('a');
  a.href = dataUrl;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

export async function downloadCardImage(card: HTMLElement, format: CardImageFormat, slug: string, scale = 2) {
  const url = await renderCardImage(card, format, scale);
  downloadDataUrl(url, `${slug || 'card'}.${format === 'png' ? 'png' : 'jpg'}`);
}

export async function downloadQrImage(value: string, slug: string) {
  const url = await qrPngDataUrl(value, 1024);
  downloadDataUrl(url, `${slug || 'card'}-qr.png`);
}
