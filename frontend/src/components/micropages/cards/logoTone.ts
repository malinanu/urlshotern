export type LogoTone = 'light' | 'dark' | 'mixed';

/**
 * Looks at an uploaded logo and says whether it is mostly light, mostly dark or neither.
 * It reads the file in the browser; nothing is uploaded for this. Transparent pixels are ignored.
 */
export async function logoTone(file: File): Promise<LogoTone> {
  const url = URL.createObjectURL(file);
  try {
    const img = await new Promise<HTMLImageElement>((resolve, reject) => {
      const i = new Image();
      i.onload = () => resolve(i);
      i.onerror = () => reject(new Error('Could not read the logo'));
      i.src = url;
    });
    const scale = Math.min(1, 96 / Math.max(img.width, img.height));
    const c = document.createElement('canvas');
    c.width = Math.max(1, Math.round(img.width * scale));
    c.height = Math.max(1, Math.round(img.height * scale));
    const g = c.getContext('2d');
    if (!g) return 'mixed';
    g.drawImage(img, 0, 0, c.width, c.height);
    const px = g.getImageData(0, 0, c.width, c.height).data;
    let n = 0;
    let light = 0;
    let dark = 0;
    for (let i = 0; i < px.length; i += 4) {
      if (px[i + 3] < 128) continue;
      n++;
      const l = (0.2126 * px[i] + 0.7152 * px[i + 1] + 0.0722 * px[i + 2]) / 255;
      if (l > 0.75) light++;
      else if (l < 0.35) dark++;
    }
    if (n < 20) return 'mixed';
    if (light / n > 0.6) return 'light';
    if (dark / n > 0.6) return 'dark';
    return 'mixed';
  } finally {
    URL.revokeObjectURL(url);
  }
}
