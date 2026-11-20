'use client';

import Link from 'next/link';
import { useDeferredValue, useMemo, useState } from 'react';
import { ArrowRight, Download, RefreshCw, ScanLine } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Reveal } from '@/components/motion/Reveal';
import { qrPngDataUrl, qrShape } from '@/components/micropages/cards/qr';
import { SectionHeading, SIGNUP_HREF } from './shared';

const DEFAULT_URL = 'https://trunc.site';

/** What the visitor typed, as a web address. A bare domain gets https:// in front. */
function toUrl(raw: string): string {
  const v = raw.trim();
  if (!v) return DEFAULT_URL;
  if (/^[a-z][a-z0-9+.-]*:/i.test(v)) return v;
  return `https://${v}`;
}

const POINTS = [
  {
    Icon: RefreshCw,
    title: 'Change it after you print it',
    body: 'In your account a QR code points at your short link, so you can send it somewhere new without reprinting the menu, flyer or box.',
  },
  {
    Icon: ScanLine,
    title: 'See every visit',
    body: 'Each scan is a visit to your short link, so it shows up in your analytics by day, device and source.',
  },
];

/** A real, scannable QR code generated in the browser from whatever the visitor types. */
export default function QrPlayground() {
  const [raw, setRaw] = useState('');
  const value = toUrl(useDeferredValue(raw));
  const shape = useMemo(() => qrShape(value), [value]);
  const [busy, setBusy] = useState(false);

  const download = async () => {
    setBusy(true);
    try {
      const href = await qrPngDataUrl(value, 1024);
      const a = document.createElement('a');
      a.href = href;
      a.download = 'qr-code.png';
      a.click();
    } finally {
      setBusy(false);
    }
  };

  return (
    <section id="qr" className="scroll-mt-20 border-y border-border bg-card/40 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto grid max-w-6xl items-center gap-10 px-4 sm:px-6 lg:grid-cols-[1.1fr_0.9fr] lg:gap-16 lg:px-8">
        <div>
          <SectionHeading
            eyebrow="QR codes"
            title="Make a QR code right now"
            lead="Type any web address and point your phone’s camera at the code. It is a real code, made in your browser."
          />

          <Reveal delay={0.05} className="mt-7">
            <label htmlFor="qr-url" className="mb-1.5 block text-sm font-medium text-foreground">
              Web address
            </label>
            <div className="flex flex-col gap-2 sm:flex-row">
              <input
                id="qr-url"
                type="url"
                inputMode="url"
                autoComplete="off"
                spellCheck={false}
                value={raw}
                maxLength={300}
                placeholder="yourwebsite.com/menu"
                onChange={(e) => setRaw(e.target.value)}
                className="h-12 w-full min-w-0 rounded-xl sm:flex-1 border border-border bg-background px-4 font-mono text-[15px] text-foreground shadow-sm placeholder:font-sans placeholder:text-muted-foreground/70 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/25"
              />
              <Button type="button" variant="outline" size="lg" className="h-12 px-5" onClick={download} disabled={!shape || busy}>
                <Download className="h-4 w-4" aria-hidden />
                Download PNG
              </Button>
            </div>
          </Reveal>

          <div className="mt-8 grid gap-5 sm:grid-cols-2">
            {POINTS.map(({ Icon, title, body }) => (
              <Reveal key={title}>
                <Icon className="h-5 w-5 text-primary" aria-hidden />
                <h3 className="mt-2 font-semibold text-foreground">{title}</h3>
                <p className="mt-1 text-sm leading-relaxed text-muted-foreground">{body}</p>
              </Reveal>
            ))}
          </div>

          <Reveal className="mt-7">
            <Link href={SIGNUP_HREF} className="inline-flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline">
              Get trackable QR codes free
              <ArrowRight className="h-4 w-4" aria-hidden />
            </Link>
          </Reveal>
        </div>

        <Reveal direction="left" className="mx-auto w-full max-w-[22rem]">
          <div className="rounded-[1.75rem] border border-border bg-background p-5 shadow-[0_24px_48px_-16px_rgba(15,23,42,0.25)] sm:p-6">
            <div className="rounded-2xl bg-white p-3">
              {shape ? (
                <svg
                  viewBox={shape.viewBox}
                  className="block h-auto w-full"
                  shapeRendering="crispEdges"
                  role="img"
                  aria-label={`QR code that opens ${value}`}
                  data-qr-value={value}
                >
                  <rect x="-4" y="-4" width="100%" height="100%" fill="#ffffff" />
                  <path d={shape.d} fill="#0b1220" />
                </svg>
              ) : (
                <div className="grid aspect-square place-items-center text-center text-sm text-slate-500">
                  That address is too long for a QR code.
                </div>
              )}
            </div>
            <p className="mt-4 truncate text-center font-mono text-xs text-muted-foreground" title={value}>
              {value}
            </p>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
