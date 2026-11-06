import Link from 'next/link';
import { Sparkles, ShieldCheck, Zap, QrCode, Globe2 } from 'lucide-react';
import URLShortener from '@/components/URLShortener';
import { Reveal } from '@/components/motion/Reveal';
import HeroAnalyticsPreview from './HeroAnalyticsPreview';

/**
 * Split, asymmetric hero. Left column carries the value proposition and the
 * primary action (the URL shortener). Right column shows a real analytics
 * preview so the hero has a genuine product visual, not just a gradient.
 */
export default function Hero() {
  return (
    <section className="relative overflow-hidden pt-10 pb-14 sm:pt-16 sm:pb-20 lg:pt-24 lg:pb-28">
      {/* Decorative background layers */}
      <div className="pointer-events-none absolute inset-0 -z-10 bg-grid opacity-60" aria-hidden />
      <div className="pointer-events-none absolute inset-0 -z-10 bg-aurora animate-aurora" aria-hidden />

      <div className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-8 px-4 sm:gap-12 sm:px-6 lg:grid-cols-[1.05fr_0.95fr] lg:gap-8 lg:px-8">
        <div>
          <Reveal>
            <Link
              href="/features"
              className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3.5 py-1 text-xs sm:text-sm font-semibold text-primary transition-colors hover:bg-primary/15"
            >
              <Sparkles className="h-4 w-4" />
              The Modern Connections & Link Platform
            </Link>
          </Reveal>

          <Reveal delay={0.05}>
            <h1 className="mt-5 text-balance text-[2.25rem] font-extrabold tracking-tight text-foreground sm:mt-6 sm:text-5xl lg:text-6xl xl:text-7xl leading-[1.05]">
              Every connection starts with a{' '}
              <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
                smarter link
              </span>
            </h1>
          </Reveal>

          <Reveal delay={0.1}>
            <p className="mt-4 max-w-xl text-pretty text-base text-muted-foreground sm:mt-5 sm:text-lg leading-relaxed">
              Shorten links, generate dynamic QR codes, and create high-converting digital vCards. Powered by sub-15ms global edge routing and real-time telemetry.
            </p>
          </Reveal>

          {/* Value Badge Chips */}
          <Reveal delay={0.12}>
            <div className="mt-4 flex flex-wrap items-center gap-3 text-xs font-medium text-muted-foreground">
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted/60 border border-border">
                <Globe2 className="h-3.5 w-3.5 text-primary" />
                Custom Domains
              </span>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted/60 border border-border">
                <QrCode className="h-3.5 w-3.5 text-primary" />
                Dynamic QR
              </span>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted/60 border border-border">
                <Zap className="h-3.5 w-3.5 text-primary" />
                Sub-15ms Edge
              </span>
              <span className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-muted/60 border border-border">
                <ShieldCheck className="h-3.5 w-3.5 text-emerald-400" />
                VirusTotal Verified
              </span>
            </div>
          </Reveal>

          <Reveal delay={0.15}>
            <div className="mt-6 sm:mt-8">
              <URLShortener />
            </div>
          </Reveal>
        </div>

        <div className="lg:pl-4">
          <HeroAnalyticsPreview />
        </div>
      </div>
    </section>
  );
}
