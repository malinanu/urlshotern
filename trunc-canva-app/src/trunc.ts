// Calls the Trunc API with the token Canva holds for the signed-in user
// (Sign in with Trunc, client trunc-canva) and makes QR codes.
import QRCode from "qrcode";

/** The Trunc server (CANVA_BACKEND_HOST), without a trailing slash. */
export const truncBaseUrl = (): string => BACKEND_HOST.replace(/\/+$/, "");

/** The only scope Trunc grants: create and read the user's links. */
export const TRUNC_SCOPE = new Set(["links"]);

export class TruncError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "TruncError";
  }
}

export type ShortLink = { shortUrl: string; shortCode: string };

export function isWebAddress(value: string): boolean {
  try {
    const u = new URL(value.trim());
    return u.protocol === "https:" || u.protocol === "http:";
  } catch {
    return false;
  }
}

export async function shortenUrl(token: string, url: string): Promise<ShortLink> {
  let res: Response;
  try {
    res = await fetch(`${truncBaseUrl()}/api/v1/shorten`, {
      method: "POST",
      headers: { Authorization: `Bearer ${token}`, "Content-Type": "application/json" },
      body: JSON.stringify({ url: url.trim() }),
    });
  } catch {
    throw new TruncError("network", 0);
  }
  const data = (await res.json().catch(() => ({}))) as Record<string, unknown>;
  if (!res.ok) {
    const msg = [data.message, data.error_description, data.error].find((v) => typeof v === "string" && v);
    throw new TruncError(typeof msg === "string" ? msg : `status ${res.status}`, res.status);
  }
  return { shortUrl: String(data.short_url), shortCode: String(data.short_code) };
}

/** A 1024 px PNG QR code with a quiet zone, as a data URL. */
export function qrDataUrl(text: string): Promise<string> {
  return QRCode.toDataURL(text, { errorCorrectionLevel: "M", margin: 4, width: 1024 });
}
