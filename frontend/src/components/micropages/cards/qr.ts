import QRCode from 'qrcode';

export interface QrShape {
  /** SVG viewBox including a 4-module quiet zone, as the QR specification asks for. */
  viewBox: string;
  /** One path of horizontal runs of dark modules. */
  d: string;
}

/**
 * Builds a real QR code as a vector path. It is drawn as SVG so it stays sharp at any size
 * and exports cleanly. Error correction M, dark modules on a light plate.
 */
export function qrShape(text: string): QrShape | null {
  if (!text) return null;
  try {
    const q = QRCode.create(text, { errorCorrectionLevel: 'M' });
    const n = q.modules.size;
    const data = q.modules.data;
    let d = '';
    for (let y = 0; y < n; y++) {
      let x = 0;
      while (x < n) {
        if (!data[y * n + x]) {
          x++;
          continue;
        }
        let w = 1;
        while (x + w < n && data[y * n + x + w]) w++;
        d += `M${x} ${y}h${w}v1h-${w}z`;
        x += w;
      }
    }
    return { viewBox: `-4 -4 ${n + 8} ${n + 8}`, d };
  } catch {
    return null;
  }
}

/** A standalone PNG of the QR code, for the "download QR" action. */
export async function qrPngDataUrl(text: string, size = 1024): Promise<string> {
  return QRCode.toDataURL(text, {
    errorCorrectionLevel: 'M',
    margin: 4,
    width: size,
    color: { dark: '#000000', light: '#ffffff' },
  });
}
