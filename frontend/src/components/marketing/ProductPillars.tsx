'use client';

import { useState } from 'react';
import Link from 'next/link';
import {
  Link2,
  QrCode,
  Contact,
  BarChart3,
  ArrowRight,
  ShieldCheck,
  CheckCircle2,
  Globe2,
  Smartphone,
  Sparkles,
  Zap,
} from 'lucide-react';
import { cn } from '@/lib/utils';

interface Pillar {
  id: string;
  name: string;
  badge: string;
  icon: typeof Link2;
  title: string;
  description: string;
  bullets: string[];
  ctaText: string;
  ctaHref: string;
}

const pillars: Pillar[] = [
  {
    id: 'links',
    name: 'Short Links',
    badge: 'Core Redirection',
    icon: Link2,
    title: 'Short links that build trust and drive clicks',
    description:
      'Turn long, messy URLs into clean, memorable, branded links. Boost click-through rates by up to 34% with your own custom domain, custom slugs, and password-protected gateways.',
    bullets: [
      'Vanity domains (links.yourbrand.com)',
      'Sub-15ms Anycast edge redirection worldwide',
      'Password protection & expiration scheduling',
      'UTM parameter campaign tagging built in',
    ],
    ctaText: 'Explore Link Shortening',
    ctaHref: '/features#shortener',
  },
  {
    id: 'qrcodes',
    name: 'Dynamic QR Codes',
    badge: 'Offline to Online',
    icon: QrCode,
    title: 'Dynamic QR codes connected to live analytics',
    description:
      'Bridge your print, packaging, signage, and events with your digital ecosystem. Update destination URLs on the fly without ever having to reprint a physical code.',
    bullets: [
      'Dynamic redirects (change destination anytime)',
      'Scan volume, device, and location tracking',
      'High-resolution vector SVG and PNG downloads',
      'Custom colors, frames, and brand styling',
    ],
    ctaText: 'Explore QR Codes',
    ctaHref: '/features#qrcodes',
  },
  {
    id: 'micropages',
    name: 'vCards & Bio-Pages',
    badge: 'Mobile Presence',
    icon: Contact,
    title: 'Digital business cards and bio-links in seconds',
    description:
      'Give every team member, creator, and executive an instant mobile landing page. Enable one-tap phone contact saves (.vcf) and showcase your essential links anywhere.',
    bullets: [
      '10 handcrafted themes (Executive, Developer, Aurora)',
      'One-tap vCard (.vcf) download to Apple & Android contacts',
      'NFC badge & digital business card compatible',
      'Unified link-in-bio hub with social action buttons',
    ],
    ctaText: 'Explore Micro-Pages',
    ctaHref: '/features#micropages',
  },
  {
    id: 'analytics',
    name: 'Real-Time Telemetry',
    badge: 'Data & Insights',
    icon: BarChart3,
    title: 'Telemetry and AI insights on every interaction',
    description:
      'Know who is clicking, from which country, on what device, and at what hour. Get plain-language AI summaries that alert you to traffic spikes and campaign anomalies.',
    bullets: [
      'Real-time click and scan velocity streams',
      'Geographic country maps and city-level accuracy',
      'Device, OS, and browser breakdown',
      'Automated AI traffic insights and anomaly alerts',
    ],
    ctaText: 'Explore Telemetry',
    ctaHref: '/features#analytics',
  },
];

export default function ProductPillars() {
  const [activeTab, setActiveTab] = useState<string>('links');
  const activePillar = pillars.find((p) => p.id === activeTab) || pillars[0];

  return (
    <section className="relative overflow-hidden py-16 sm:py-24 lg:py-32 bg-card/20 border-b border-border">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        
        {/* Section Header */}
        <div className="mx-auto max-w-3xl text-center mb-12 sm:mb-16">
          <div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/10 px-3 py-1 text-xs font-semibold text-primary uppercase tracking-wider mb-4">
            <Sparkles className="h-3.5 w-3.5" />
            The Complete Connections Platform
          </div>
          <h2 className="text-3xl font-extrabold tracking-tight text-foreground sm:text-5xl">
            Everything your links need to <br className="hidden sm:inline" />
            <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
              perform, connect, and convert
            </span>
          </h2>
          <p className="mt-4 text-base sm:text-lg text-muted-foreground">
            From single short links to company-wide custom domains, dynamic QR campaigns, and instant digital vCards.
          </p>
        </div>

        {/* Tab Navigation Strip */}
        <div className="flex justify-center mb-12 overflow-x-auto pb-2">
          <div className="inline-flex p-1.5 rounded-2xl bg-muted/50 border border-border">
            {pillars.map((pillar) => {
              const Icon = pillar.icon;
              const isActive = activeTab === pillar.id;
              return (
                <button
                  key={pillar.id}
                  onClick={() => setActiveTab(pillar.id)}
                  className={cn(
                    'flex items-center gap-2 px-4 py-2.5 rounded-xl text-xs sm:text-sm font-semibold transition-all whitespace-nowrap',
                    isActive
                      ? 'bg-background text-foreground shadow-sm border border-border'
                      : 'text-muted-foreground hover:text-foreground'
                  )}
                >
                  <Icon className={cn('h-4 w-4', isActive ? 'text-primary' : 'text-muted-foreground')} />
                  <span>{pillar.name}</span>
                </button>
              );
            })}
          </div>
        </div>

        {/* Tab Content Display */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center bg-card rounded-3xl border border-border p-6 sm:p-10 shadow-xl backdrop-blur-sm">
          
          {/* Left Column: Description & Bullets */}
          <div className="lg:col-span-6 space-y-6">
            <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20">
              {activePillar.badge}
            </div>
            
            <h3 className="text-2xl sm:text-3xl font-bold tracking-tight text-foreground">
              {activePillar.title}
            </h3>

            <p className="text-muted-foreground text-sm sm:text-base leading-relaxed">
              {activePillar.description}
            </p>

            <ul className="space-y-3 pt-2">
              {activePillar.bullets.map((bullet, idx) => (
                <li key={idx} className="flex items-start gap-2.5 text-sm text-foreground">
                  <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                  <span>{bullet}</span>
                </li>
              ))}
            </ul>

            <div className="pt-4">
              <Link
                href={activePillar.ctaHref}
                className="inline-flex items-center gap-2 text-sm font-semibold text-primary hover:text-primary/80 transition-colors group"
              >
                <span>{activePillar.ctaText}</span>
                <ArrowRight className="h-4 w-4 transition-transform group-hover:translate-x-1" />
              </Link>
            </div>
          </div>

          {/* Right Column: Visual Product Mockup */}
          <div className="lg:col-span-6">
            
            {/* Visual for Short Links */}
            {activePillar.id === 'links' && (
              <div className="rounded-2xl border border-border bg-background/95 p-6 shadow-md font-sans">
                <div className="flex items-center justify-between pb-4 border-b border-border">
                  <div className="flex items-center gap-2">
                    <Globe2 className="h-4 w-4 text-primary" />
                    <span className="text-xs font-bold text-foreground">Branded Short Link</span>
                  </div>
                  <span className="px-2 py-0.5 rounded text-[11px] font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-semibold">
                    12ms Edge
                  </span>
                </div>
                
                <div className="mt-4 p-4 rounded-xl bg-muted/40 border border-border">
                  <div className="text-[11px] text-muted-foreground uppercase tracking-wider font-semibold">Destination URL</div>
                  <div className="text-xs text-foreground/80 font-mono truncate mt-0.5">https://yourcompany.com/products/summer-sale?utm_source=email</div>
                  
                  <div className="mt-3 pt-3 border-t border-border flex items-center justify-between">
                    <div>
                      <div className="text-[11px] text-muted-foreground uppercase tracking-wider font-semibold">Short Link</div>
                      <div className="text-sm font-bold text-primary font-mono mt-0.5">trunc.site/summer-sale</div>
                    </div>
                    <span className="px-3 py-1 rounded-lg text-xs font-semibold bg-primary text-primary-foreground">
                      Copy Link
                    </span>
                  </div>
                </div>

                <div className="mt-4 grid grid-cols-3 gap-2 text-center text-xs">
                  <div className="p-2.5 rounded-lg bg-card border border-border">
                    <div className="text-[11px] text-muted-foreground">Clicks</div>
                    <div className="font-bold text-foreground text-sm mt-0.5">14,290</div>
                  </div>
                  <div className="p-2.5 rounded-lg bg-card border border-border">
                    <div className="text-[11px] text-muted-foreground">Protected</div>
                    <div className="font-bold text-emerald-400 text-sm mt-0.5">AES-256</div>
                  </div>
                  <div className="p-2.5 rounded-lg bg-card border border-border">
                    <div className="text-[11px] text-muted-foreground">Malware</div>
                    <div className="font-bold text-primary text-sm mt-0.5">Clean (0)</div>
                  </div>
                </div>
              </div>
            )}

            {/* Visual for QR Codes */}
            {activePillar.id === 'qrcodes' && (
              <div className="rounded-2xl border border-border bg-background/95 p-6 shadow-md text-center">
                <div className="flex items-center justify-between pb-4 border-b border-border text-left">
                  <span className="text-xs font-bold text-foreground">Dynamic Vector QR</span>
                  <span className="px-2 py-0.5 rounded text-[11px] bg-primary/10 text-primary border border-primary/20 font-semibold">
                    Dynamic Sync
                  </span>
                </div>

                <div className="my-6 inline-block p-4 rounded-2xl bg-white border border-border shadow-inner">
                  {/* Stylized QR Code Visual */}
                  <div className="w-36 h-36 border-4 border-zinc-900 rounded-lg p-2 grid grid-cols-6 grid-rows-6 gap-1 bg-white">
                    <div className="col-span-2 row-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 row-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 row-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 row-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 bg-zinc-900 rounded-sm"></div>
                    <div className="col-span-2 bg-zinc-900 rounded-sm"></div>
                  </div>
                </div>

                <div className="flex justify-center gap-2">
                  <span className="px-3 py-1.5 rounded-lg text-xs font-semibold border border-border bg-muted text-foreground">
                    Download SVG
                  </span>
                  <span className="px-3 py-1.5 rounded-lg text-xs font-semibold bg-primary text-primary-foreground">
                    Track Scans (1,842)
                  </span>
                </div>
              </div>
            )}

            {/* Visual for Micro-Pages */}
            {activePillar.id === 'micropages' && (
              <div className="rounded-2xl border border-border bg-background/95 p-6 shadow-md">
                <div className="flex items-center justify-between pb-3 border-b border-border">
                  <span className="text-xs font-bold text-foreground flex items-center gap-1.5">
                    <Smartphone className="h-4 w-4 text-primary" />
                    Mobile vCard Preview
                  </span>
                  <span className="text-[11px] font-mono text-muted-foreground">trunc.site/m/alex</span>
                </div>

                {/* Mobile Mockup Card */}
                <div className="mt-4 mx-auto max-w-xs rounded-2xl bg-zinc-950 border border-zinc-800 p-4 text-center text-white shadow-lg">
                  <div className="w-14 h-14 rounded-full bg-gradient-to-r from-purple-500 to-indigo-500 mx-auto p-0.5 mb-2 shadow-md">
                    <div className="w-full h-full bg-zinc-900 rounded-full flex items-center justify-center font-bold text-base">
                      AR
                    </div>
                  </div>
                  <div className="text-sm font-bold">Alex Rivera</div>
                  <div className="text-[11px] text-zinc-400">Chief Marketing Officer · Acme Corp</div>
                  
                  <div className="mt-4 space-y-2 text-xs">
                    <div className="p-2 rounded-lg bg-zinc-900 border border-zinc-800 font-semibold text-emerald-400 flex items-center justify-center gap-1.5">
                      📥 Save Contact to Phone (.vcf)
                    </div>
                    <div className="p-2 rounded-lg bg-zinc-900/60 border border-zinc-800 text-zinc-300">
                      🌐 Official Company Website
                    </div>
                    <div className="p-2 rounded-lg bg-zinc-900/60 border border-zinc-800 text-zinc-300">
                      📅 Book a 30-Min Consultation
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* Visual for Analytics */}
            {activePillar.id === 'analytics' && (
              <div className="rounded-2xl border border-border bg-background/95 p-6 shadow-md font-sans">
                <div className="flex items-center justify-between pb-3 border-b border-border">
                  <span className="text-xs font-bold text-foreground flex items-center gap-1.5">
                    <Zap className="h-4 w-4 text-primary" />
                    Live Telemetry Stream
                  </span>
                  <span className="inline-flex items-center gap-1 text-[11px] text-emerald-400 font-semibold">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping"></span> Live
                  </span>
                </div>

                <div className="mt-4 space-y-2.5">
                  <div className="flex items-center justify-between p-2.5 rounded-lg bg-muted/40 border border-border text-xs">
                    <span className="font-semibold text-foreground flex items-center gap-2">
                      🇺🇸 United States
                    </span>
                    <span className="text-muted-foreground font-mono">18,420 clicks (48%)</span>
                  </div>
                  <div className="flex items-center justify-between p-2.5 rounded-lg bg-muted/40 border border-border text-xs">
                    <span className="font-semibold text-foreground flex items-center gap-2">
                      🇬🇧 United Kingdom
                    </span>
                    <span className="text-muted-foreground font-mono">8,910 clicks (23%)</span>
                  </div>
                  <div className="flex items-center justify-between p-2.5 rounded-lg bg-muted/40 border border-border text-xs">
                    <span className="font-semibold text-foreground flex items-center gap-2">
                      🇩🇪 Germany
                    </span>
                    <span className="text-muted-foreground font-mono">4,120 clicks (11%)</span>
                  </div>
                </div>

                <div className="mt-4 pt-3 border-t border-border flex items-center justify-between text-xs text-muted-foreground">
                  <span>AI Insight: <span className="text-foreground font-medium">Traffic +34% spike via LinkedIn campaign</span></span>
                </div>
              </div>
            )}

          </div>

        </div>

      </div>
    </section>
  );
}
