'use client';

import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import {
  Link2,
  QrCode,
  Contact,
  BarChart3,
  ShieldCheck,
  Zap,
  Globe2,
  Lock,
  Calendar,
  Sparkles,
  ArrowRight,
  CheckCircle2,
  Layers,
  Smartphone,
} from 'lucide-react';
import { useCMSPage, getContentBlock } from '@/hooks/useCMSPage';

export default function FeaturesPage() {
  const { content: page } = useCMSPage('features', {
    slug: 'features',
    title: 'Features',
    content: 'Powerful Features for Smart URL Management',
  });

  const hero = getContentBlock(page, 'hero', {
    title: 'The Modern Connections Platform',
    title_highlight: 'Built for Scale and Speed',
    tagline:
      'Everything you need to create branded short links, dynamic QR codes, digital vCards, and real-time analytics. Powered by global edge routing.',
  });

  return (
    <Layout>
      <div className="bg-background text-foreground min-h-screen">
        
        {/* Hero Section */}
        <section className="relative overflow-hidden pt-16 pb-12 sm:pt-24 sm:pb-16 border-b border-border bg-gradient-to-b from-primary/10 via-card/30 to-background">
          <div className="pointer-events-none absolute inset-0 bg-grid opacity-40" aria-hidden />
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-3.5 py-1 rounded-full border border-primary/20 bg-primary/10 text-primary text-xs font-semibold uppercase tracking-wider mb-6">
              <Sparkles className="h-3.5 w-3.5" />
              Complete Product Breakdown
            </div>
            
            <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight text-foreground max-w-4xl mx-auto leading-tight">
              Powerful tools for{' '}
              <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
                modern link infrastructure
              </span>
            </h1>

            <p className="mt-5 text-base sm:text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
              {hero.tagline}
            </p>

            {/* Anchor Pill Jump Bar */}
            <div className="mt-10 flex flex-wrap items-center justify-center gap-2">
              {[
                { id: 'shortener', label: 'Short Links & Domains', icon: Link2 },
                { id: 'qrcodes', label: 'Dynamic QR Codes', icon: QrCode },
                { id: 'micropages', label: 'vCards & Bio-Pages', icon: Contact },
                { id: 'analytics', label: 'Telemetry & AI', icon: BarChart3 },
                { id: 'security', label: 'Security & APIs', icon: ShieldCheck },
                { id: 'matrix', label: 'Comparison Matrix', icon: Layers },
              ].map((pill) => {
                const Icon = pill.icon;
                return (
                  <a
                    key={pill.id}
                    href={`#${pill.id}`}
                    className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-full text-xs font-medium border border-border bg-card/60 hover:bg-muted text-foreground transition-all"
                  >
                    <Icon className="h-3.5 w-3.5 text-primary" />
                    <span>{pill.label}</span>
                  </a>
                );
              })}
            </div>
          </div>
        </section>

        {/* Pillar 1: Short Links & Branded Domains */}
        <section id="shortener" className="py-16 sm:py-24 border-b border-border">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
              
              <div className="lg:col-span-6 space-y-6">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20">
                  <Link2 className="h-3.5 w-3.5" /> Pillar 01: Core Redirection
                </div>
                <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                  Branded Short Links with Sub-15ms Edge Latency
                </h2>
                <p className="text-muted-foreground text-base leading-relaxed">
                  Turn unwieldy destination URLs into concise, trustworthy links that match your brand identity. With custom vanity domains and custom back-halves, audience trust and click-through rates rise dramatically.
                </p>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pt-2">
                  <div className="p-4 rounded-xl bg-card border border-border">
                    <div className="font-bold text-foreground text-sm flex items-center gap-2">
                      <Globe2 className="h-4 w-4 text-primary" /> Custom Domains
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Use links.yourbrand.com with automatic SSL certificates.</p>
                  </div>
                  <div className="p-4 rounded-xl bg-card border border-border">
                    <div className="font-bold text-foreground text-sm flex items-center gap-2">
                      <Lock className="h-4 w-4 text-primary" /> Password Protection
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Gate sensitive documents and downloads with AES passwords.</p>
                  </div>
                  <div className="p-4 rounded-xl bg-card border border-border">
                    <div className="font-bold text-foreground text-sm flex items-center gap-2">
                      <Calendar className="h-4 w-4 text-primary" /> Link Expiration
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Set date or click thresholds for time-limited promotions.</p>
                  </div>
                  <div className="p-4 rounded-xl bg-card border border-border">
                    <div className="font-bold text-foreground text-sm flex items-center gap-2">
                      <Zap className="h-4 w-4 text-primary" /> Anycast Edge Caching
                    </div>
                    <p className="text-xs text-muted-foreground mt-1">Sub-15ms redirects across 280+ points of presence worldwide.</p>
                  </div>
                </div>
              </div>

              {/* Interactive Edge Routing Box */}
              <div className="lg:col-span-6">
                <div className="rounded-2xl border border-border bg-card p-6 shadow-xl font-mono text-xs">
                  <div className="flex items-center justify-between pb-3 border-b border-border">
                    <span className="text-muted-foreground">ANYCAST EDGE ROUTING SIMULATION</span>
                    <span className="inline-flex items-center gap-1.5 text-emerald-400 font-semibold">
                      <span className="w-2 h-2 rounded-full bg-emerald-400 animate-ping"></span> 99.99% HIT
                    </span>
                  </div>

                  <div className="my-6 space-y-4 font-sans">
                    <div className="p-3 rounded-xl bg-background border border-border flex items-center justify-between">
                      <div>
                        <div className="text-[11px] text-muted-foreground uppercase font-bold">Incoming Request</div>
                        <div className="text-sm font-bold text-foreground font-mono mt-0.5">https://trunc.site/q2-launch</div>
                      </div>
                      <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary">
                        Client: London, UK
                      </span>
                    </div>

                    <div className="flex items-center justify-center gap-3 text-xs text-muted-foreground font-mono">
                      <span>Anycast DNS</span>
                      <span>→</span>
                      <span className="text-emerald-400 font-bold">Edge Cache (12ms)</span>
                      <span>→</span>
                      <span>Origin Redis</span>
                    </div>

                    <div className="p-3 rounded-xl bg-background border border-border flex items-center justify-between">
                      <div>
                        <div className="text-[11px] text-muted-foreground uppercase font-bold">Final Destination (301)</div>
                        <div className="text-sm font-semibold text-emerald-400 truncate max-w-xs font-mono mt-0.5">
                          https://example.com/products/release-2026?src=app
                        </div>
                      </div>
                      <span className="text-xs font-bold text-muted-foreground">HTTP 301</span>
                    </div>
                  </div>

                  <div className="pt-3 border-t border-border flex items-center justify-between text-[11px] text-muted-foreground">
                    <span>TLS 1.3 Handshake: 4ms</span>
                    <span>Origin Lookup: 8ms</span>
                    <span className="text-foreground font-bold">Total: 12ms</span>
                  </div>
                </div>
              </div>

            </div>
          </div>
        </section>

        {/* Pillar 2: Dynamic QR Code Engine */}
        <section id="qrcodes" className="py-16 sm:py-24 border-b border-border bg-card/20">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
              
              {/* QR Mockup */}
              <div className="lg:col-span-6 order-2 lg:order-1">
                <div className="rounded-2xl border border-border bg-card p-8 shadow-xl text-center">
                  <div className="inline-block p-4 rounded-2xl bg-white border border-border shadow-md">
                    <div className="w-44 h-44 border-4 border-zinc-900 rounded-lg p-2.5 grid grid-cols-6 grid-rows-6 gap-1 bg-white">
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

                  <div className="mt-6 flex flex-wrap justify-center gap-3">
                    <span className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-muted border border-border text-foreground">
                      SVG Vector Export
                    </span>
                    <span className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-muted border border-border text-foreground">
                      PNG High-Res (300 DPI)
                    </span>
                    <span className="px-3.5 py-1.5 rounded-lg text-xs font-semibold bg-primary text-primary-foreground">
                      Update Link Anytime
                    </span>
                  </div>
                </div>
              </div>

              {/* QR Features Description */}
              <div className="lg:col-span-6 order-1 lg:order-2 space-y-6">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                  <QrCode className="h-3.5 w-3.5" /> Pillar 02: Dynamic Physical Codes
                </div>
                <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                  Dynamic QR Codes that You Never Have to Reprint
                </h2>
                <p className="text-muted-foreground text-base leading-relaxed">
                  Printed a brochure, billboard, or product label with a QR code? With Trunc Dynamic QR Codes, you can change the target URL in real time without altering the printed physical code.
                </p>

                <ul className="space-y-3 pt-2">
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>Live Destination Swapping:</strong> Repoint your QR code to seasonal campaigns or new product pages anytime.</span>
                  </li>
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>Full Scan Telemetry:</strong> Track physical engagement, scan locations, operating systems, and peak scan hours.</span>
                  </li>
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>Print-Ready Vectors:</strong> Export crisp SVGs ready for enterprise billboard signage, packaging, and business cards.</span>
                  </li>
                </ul>
              </div>

            </div>
          </div>
        </section>

        {/* Pillar 3: Micro-Pages & Digital vCards */}
        <section id="micropages" className="py-16 sm:py-24 border-b border-border">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
              
              <div className="lg:col-span-6 space-y-6">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20">
                  <Contact className="h-3.5 w-3.5" /> Pillar 03: Digital vCards & Bio-Hubs
                </div>
                <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                  Digital Business Cards with 1-Click Phone Contact Saves
                </h2>
                <p className="text-muted-foreground text-base leading-relaxed">
                  Replace outdated paper business cards with interactive digital landing pages. Allow clients and partners to save your complete contact details directly into their iPhone or Android address book with a single tap.
                </p>

                <div className="space-y-3 pt-2">
                  <div className="flex items-start gap-3 p-3.5 rounded-xl bg-card border border-border">
                    <Smartphone className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <div>
                      <div className="text-sm font-bold text-foreground">One-Click .vcf Phone Contact Download</div>
                      <p className="text-xs text-muted-foreground mt-0.5">Generates clean vCard 3.0 files compatible with iOS Contacts and Google Contacts.</p>
                    </div>
                  </div>

                  <div className="flex items-start gap-3 p-3.5 rounded-xl bg-card border border-border">
                    <Layers className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <div>
                      <div className="text-sm font-bold text-foreground">10 Handcrafted Visual Themes</div>
                      <p className="text-xs text-muted-foreground mt-0.5">Themes designed for executives (Executive Obsidian), engineers (Developer Terminal), and creatives (Creative Aurora).</p>
                    </div>
                  </div>

                  <div className="flex items-start gap-3 p-3.5 rounded-xl bg-card border border-border">
                    <Zap className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <div>
                      <div className="text-sm font-bold text-foreground">NFC Badge & Physical Card Sync</div>
                      <p className="text-xs text-muted-foreground mt-0.5">Program any standard NFC chip or smart card to point directly to your Trunc micro-page.</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Mobile Phone Mockup */}
              <div className="lg:col-span-6 flex justify-center">
                <div className="w-full max-w-sm rounded-[2.5rem] bg-zinc-950 border-4 border-zinc-800 p-6 shadow-2xl relative overflow-hidden">
                  <div className="w-24 h-4 bg-zinc-800 rounded-full mx-auto mb-6"></div>
                  
                  <div className="text-center">
                    <div className="w-16 h-16 rounded-full bg-gradient-to-r from-purple-500 to-indigo-500 mx-auto p-0.5 mb-3 shadow-md">
                      <div className="w-full h-full bg-zinc-900 rounded-full flex items-center justify-center font-bold text-lg text-white">
                        MJ
                      </div>
                    </div>
                    <div className="text-base font-bold text-white">Malin Jayaweera</div>
                    <div className="text-xs text-zinc-400">Founder & Managing Director</div>
                    <span className="inline-block mt-2 px-2.5 py-0.5 rounded-full text-[10px] font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20">
                      Executive Obsidian
                    </span>
                  </div>

                  <div className="mt-6 space-y-2.5 text-xs font-semibold">
                    <div className="p-3 rounded-xl bg-emerald-500 text-zinc-950 flex items-center justify-center gap-2 shadow-md">
                      <span>📥 Save Contact to Phone</span>
                    </div>
                    <div className="p-2.5 rounded-xl bg-zinc-900 border border-zinc-800 text-white text-center">
                      ✉️ Email: malin@threelogiq.com
                    </div>
                    <div className="p-2.5 rounded-xl bg-zinc-900 border border-zinc-800 text-white text-center">
                      🌐 Portfolio & Company Site
                    </div>
                    <div className="p-2.5 rounded-xl bg-zinc-900 border border-zinc-800 text-white text-center">
                      💼 LinkedIn Profile
                    </div>
                  </div>

                  <div className="mt-6 pt-4 border-t border-zinc-900 text-center text-[10px] text-zinc-500">
                    Powered by Trunc Micro-Landing Pages
                  </div>
                </div>
              </div>

            </div>
          </div>
        </section>

        {/* Pillar 4: Telemetry & AI Traffic Insights */}
        <section id="analytics" className="py-16 sm:py-24 border-b border-border bg-card/20">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-12 items-center">
              
              {/* Telemetry Visual Card */}
              <div className="lg:col-span-6">
                <div className="rounded-2xl border border-border bg-card p-6 shadow-xl font-sans">
                  <div className="flex items-center justify-between pb-4 border-b border-border">
                    <div>
                      <div className="text-xs text-muted-foreground uppercase font-bold">Audience Telemetry</div>
                      <div className="text-2xl font-black text-foreground tabular-nums mt-0.5">148,290 Clicks</div>
                    </div>
                    <span className="px-2.5 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                      +28.4% vs last week
                    </span>
                  </div>

                  {/* Sparkline Visual */}
                  <div className="my-6">
                    <div className="flex items-end gap-2 h-24 pt-4">
                      {[40, 55, 38, 70, 85, 60, 95].map((h, i) => (
                        <div key={i} className="flex-1 bg-muted rounded-t hover:bg-primary transition-all relative group" style={{ height: `${h}%` }}>
                          <div className="opacity-0 group-hover:opacity-100 absolute -top-7 left-1/2 -translate-x-1/2 px-1.5 py-0.5 rounded text-[10px] bg-foreground text-background font-mono transition-opacity whitespace-nowrap">
                            {h * 240}
                          </div>
                        </div>
                      ))}
                    </div>
                    <div className="flex justify-between text-[11px] text-muted-foreground font-mono mt-2">
                      <span>Mon</span><span>Tue</span><span>Wed</span><span>Thu</span><span>Fri</span><span>Sat</span><span>Sun</span>
                    </div>
                  </div>

                  {/* Top Countries Bar */}
                  <div className="space-y-2 border-t border-border pt-4">
                    <div className="flex justify-between text-xs font-medium">
                      <span className="text-foreground">United States</span>
                      <span className="text-muted-foreground font-mono">68,200 (46%)</span>
                    </div>
                    <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                      <div className="h-full bg-primary rounded-full" style={{ width: '46%' }}></div>
                    </div>

                    <div className="flex justify-between text-xs font-medium pt-2">
                      <span className="text-foreground">United Kingdom</span>
                      <span className="text-muted-foreground font-mono">34,100 (23%)</span>
                    </div>
                    <div className="w-full h-1.5 bg-muted rounded-full overflow-hidden">
                      <div className="h-full bg-sky-400 rounded-full" style={{ width: '23%' }}></div>
                    </div>
                  </div>
                </div>
              </div>

              {/* Analytics Description */}
              <div className="lg:col-span-6 space-y-6">
                <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  <BarChart3 className="h-3.5 w-3.5" /> Pillar 04: Real-Time Intelligence
                </div>
                <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                  Telemetry that Uncovers Your Audience Behavior
                </h2>
                <p className="text-muted-foreground text-base leading-relaxed">
                  Do not fly blind. Every click on a Trunc short link or scan on a QR code streams live telemetry to your dashboard. Know precisely where traffic originates, which social channels convert, and when your audience is active.
                </p>

                <ul className="space-y-3 pt-2">
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>Geographic Heatmaps:</strong> Identify country and city-level engagement patterns to optimize local ad campaigns.</span>
                  </li>
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>Device & Referrer Attribution:</strong> Distinguish between mobile, desktop, iOS, Android, and direct social referrers.</span>
                  </li>
                  <li className="flex items-start gap-2.5 text-sm text-foreground">
                    <CheckCircle2 className="h-5 w-5 text-primary shrink-0 mt-0.5" />
                    <span><strong>AI Traffic Summaries:</strong> Automated natural language digests that explain why a campaign is spiking or dropping.</span>
                  </li>
                </ul>
              </div>

            </div>
          </div>
        </section>

        {/* Pillar 5: Security & Infrastructure */}
        <section id="security" className="py-16 sm:py-24 border-b border-border">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="mx-auto max-w-3xl text-center mb-16">
              <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-primary/10 text-primary border border-primary/20 mb-3">
                <ShieldCheck className="h-3.5 w-3.5" /> Pillar 05: Enterprise Foundation
              </div>
              <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                Zero-Trust Reliability, Security & Developer APIs
              </h2>
              <p className="text-muted-foreground text-base mt-3">
                Engineered for critical marketing workflows, developers, and global brands requiring 99.99% uptime.
              </p>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="p-6 rounded-2xl bg-card border border-border shadow-sm">
                <div className="w-10 h-10 rounded-xl bg-emerald-500/10 text-emerald-400 flex items-center justify-center font-bold mb-4">
                  🛡️
                </div>
                <h3 className="font-bold text-foreground text-lg">VirusTotal Spam Engine</h3>
                <p className="text-sm text-muted-foreground mt-2 leading-relaxed">
                  Every shortened URL is scanned against VirusTotal databases. Malicious phishing links and spam are blocked before they ever resolve.
                </p>
              </div>

              <div className="p-6 rounded-2xl bg-card border border-border shadow-sm">
                <div className="w-10 h-10 rounded-xl bg-blue-500/10 text-blue-400 flex items-center justify-center font-bold mb-4">
                  ⚡
                </div>
                <h3 className="font-bold text-foreground text-lg">REST API & Webhooks</h3>
                <p className="text-sm text-muted-foreground mt-2 leading-relaxed">
                  Programmatically shorten links, generate QR codes, and trigger real-time webhooks on every click for marketing automation.
                </p>
              </div>

              <div className="p-6 rounded-2xl bg-card border border-border shadow-sm">
                <div className="w-10 h-10 rounded-xl bg-purple-500/10 text-purple-400 flex items-center justify-center font-bold mb-4">
                  🔒
                </div>
                <h3 className="font-bold text-foreground text-lg">Compliance & Privacy</h3>
                <p className="text-sm text-muted-foreground mt-2 leading-relaxed">
                  GDPR compliant telemetry, COPPA age gating, CAN-SPAM email suppression, and DMCA takedown automation built in.
                </p>
              </div>
            </div>
          </div>
        </section>

        {/* Feature Comparison Matrix */}
        <section id="matrix" className="py-16 sm:py-24 border-b border-border bg-card/20">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="mx-auto max-w-3xl text-center mb-16">
              <h2 className="text-3xl sm:text-4xl font-extrabold tracking-tight text-foreground">
                Compare Plan Capabilities
              </h2>
              <p className="text-muted-foreground text-base mt-2">
                Pick the tier that fits your volume and scaling requirements.
              </p>
            </div>

            <div className="overflow-x-auto rounded-2xl border border-border bg-card">
              <table className="w-full text-sm text-left">
                <thead className="bg-muted/60 border-b border-border text-foreground">
                  <tr>
                    <th className="p-4 sm:p-5 font-bold">Feature</th>
                    <th className="p-4 sm:p-5 font-bold text-center">Free</th>
                    <th className="p-4 sm:p-5 font-bold text-center text-primary">Starter / Pro</th>
                    <th className="p-4 sm:p-5 font-bold text-center">Growth / Team</th>
                    <th className="p-4 sm:p-5 font-bold text-center">Enterprise</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="p-4 font-medium text-foreground">Short Links / Month</td>
                    <td className="p-4 text-center text-muted-foreground">10</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Custom Vanity Domains</td>
                    <td className="p-4 text-center text-muted-foreground">—</td>
                    <td className="p-4 text-center font-bold text-primary">1 Domain</td>
                    <td className="p-4 text-center font-bold text-foreground">5 Domains</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Dynamic QR Codes</td>
                    <td className="p-4 text-center text-muted-foreground">Standard</td>
                    <td className="p-4 text-center font-bold text-foreground">Dynamic Vector SVG</td>
                    <td className="p-4 text-center font-bold text-foreground">Dynamic Vector SVG</td>
                    <td className="p-4 text-center font-bold text-foreground">Custom Branded</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Micro-Pages / vCards</td>
                    <td className="p-4 text-center text-muted-foreground">1 Basic</td>
                    <td className="p-4 text-center font-bold text-foreground">5 Themes</td>
                    <td className="p-4 text-center font-bold text-foreground">All 10 Themes</td>
                    <td className="p-4 text-center font-bold text-foreground">White-Label Hubs</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Analytics History Retention</td>
                    <td className="p-4 text-center text-muted-foreground">7 Days</td>
                    <td className="p-4 text-center font-bold text-foreground">365 Days</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited</td>
                    <td className="p-4 text-center font-bold text-foreground">Unlimited + Warehouse</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Password & Expiration Gates</td>
                    <td className="p-4 text-center text-muted-foreground">—</td>
                    <td className="p-4 text-center text-emerald-400 font-bold">✓</td>
                    <td className="p-4 text-center text-emerald-400 font-bold">✓</td>
                    <td className="p-4 text-center text-emerald-400 font-bold">✓</td>
                  </tr>
                  <tr>
                    <td className="p-4 font-medium text-foreground">Team Collaboration</td>
                    <td className="p-4 text-center text-muted-foreground">1 User</td>
                    <td className="p-4 text-center text-muted-foreground">1 User</td>
                    <td className="p-4 text-center font-bold text-foreground">5 Seats</td>
                    <td className="p-4 text-center font-bold text-foreground">Custom SSO/SAML</td>
                  </tr>
                </tbody>
              </table>
            </div>

            <div className="mt-8 text-center">
              <Link
                href="/pricing"
                className="inline-flex items-center gap-2 text-sm font-semibold text-primary hover:underline"
              >
                <span>View Full Pricing & Checkout Options</span>
                <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </div>
        </section>

        {/* Bottom CTA Band */}
        <section className="py-20 bg-gradient-to-br from-primary/15 via-background to-card border-t border-border text-center">
          <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
            <h2 className="text-3xl sm:text-5xl font-extrabold tracking-tight text-foreground">
              Ready to Upgrade Your Link Infrastructure?
            </h2>
            <p className="mt-4 text-base sm:text-lg text-muted-foreground max-w-xl mx-auto">
              Join thousands of creators, marketers, and businesses using Trunc to power their digital connections.
            </p>
            <div className="mt-8 flex flex-col sm:flex-row gap-4 justify-center">
              <Link
                href="/"
                className="px-8 py-3.5 rounded-xl font-semibold bg-primary text-primary-foreground hover:bg-primary/90 shadow-lg shadow-primary/20 transition-all text-sm"
              >
                Start for Free Today
              </Link>
              <Link
                href="/pricing"
                className="px-8 py-3.5 rounded-xl font-semibold border border-border bg-card hover:bg-muted text-foreground transition-all text-sm"
              >
                Explore Pricing Plans
              </Link>
            </div>
          </div>
        </section>

      </div>
    </Layout>
  );
}
