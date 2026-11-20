'use client';

import Link from 'next/link';
import { ArrowRight, Check, Link2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Reveal } from '@/components/motion/Reveal';
import { CardRenderer } from '@/components/micropages/cards/CardRenderer';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { DEMO_CARD, SIGNUP_HREF } from './shared';

const JUMP = [
  { href: '#cards', label: 'Digital cards' },
  ...(MICROSITES_ENABLED ? [{ href: '#pages', label: 'Pages' }] : []),
  { href: '#qr', label: 'QR codes' },
  { href: '#platform', label: 'Links & analytics' },
  { href: '#plans', label: 'Plans' },
];

// Seven days of an example week, as bar heights in percent. Labelled "Example" on the page.
const WEEK = [38, 52, 46, 70, 64, 88, 76];

/**
 * Split hero: the promise and one sign-up action on the left, a real card from the
 * template library on the right, with an example of what its owner sees.
 */
export default function FeaturesHero() {
  return (
    <section className="relative overflow-hidden pt-10 pb-10 sm:pt-16 sm:pb-14 lg:pt-20">
      <div className="pointer-events-none absolute inset-0 -z-10 bg-grid opacity-60" aria-hidden />
      <div className="pointer-events-none absolute inset-0 -z-10 bg-aurora animate-aurora" aria-hidden />

      <div className="mx-auto grid max-w-7xl grid-cols-1 items-center gap-12 px-4 sm:px-6 lg:grid-cols-[1.08fr_0.92fr] lg:gap-10 lg:px-8">
        <div>
          <Reveal>
            <a
              href="#cards"
              className="inline-flex max-w-full items-center gap-2 rounded-full border border-primary/20 bg-primary/10 py-1 pl-1 pr-3 text-xs font-semibold text-primary transition-colors hover:bg-primary/15 sm:text-sm"
            >
              <span className="rounded-full bg-primary px-2 py-0.5 text-[11px] font-bold uppercase tracking-wide text-primary-foreground">
                New
              </span>
              <span className="truncate">12 digital business card designs</span>
              <ArrowRight className="h-3.5 w-3.5 shrink-0" />
            </a>
          </Reveal>

          <Reveal delay={0.05}>
            <h1 className="mt-5 text-balance text-[2.3rem] font-extrabold leading-[1.05] tracking-tight text-foreground sm:mt-6 sm:text-5xl lg:text-6xl">
              Your card, your links,{' '}
              <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
                your numbers
              </span>
            </h1>
          </Reveal>

          <Reveal delay={0.1}>
            <p className="mt-4 max-w-xl text-pretty text-base leading-relaxed text-muted-foreground sm:mt-5 sm:text-lg">
              Trunc gives you a digital business card, short links, QR codes
              {MICROSITES_ENABLED ? ' and mobile pages' : ''} in one account, and shows you how often each one is opened.
              Make it once, share it everywhere, and keep it up to date without reprinting a thing.
            </p>
          </Reveal>

          <Reveal delay={0.15}>
            <div className="mt-7 flex flex-col gap-3 sm:flex-row sm:items-center">
              <Button asChild size="lg" className="h-12 w-full px-7 text-base font-semibold sm:w-auto">
                <Link href={SIGNUP_HREF}>
                  Create your free account
                  <ArrowRight className="h-4 w-4" />
                </Link>
              </Button>
              <Button asChild size="lg" variant="outline" className="h-12 w-full px-6 text-base sm:w-auto">
                <a href="#new">See what&apos;s new</a>
              </Button>
            </div>
            <ul className="mt-4 flex flex-wrap gap-x-5 gap-y-1.5 text-sm text-muted-foreground">
              {['Free plan', 'No credit card'].map((t) => (
                <li key={t} className="inline-flex items-center gap-1.5">
                  <Check className="h-4 w-4 text-emerald-600 dark:text-emerald-400" aria-hidden />
                  {t}
                </li>
              ))}
            </ul>
          </Reveal>
        </div>

        <HeroVisual />
      </div>

      <nav aria-label="On this page" className="mx-auto mt-12 max-w-7xl px-4 sm:mt-16 sm:px-6 lg:px-8">
        <ul className="-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 [scrollbar-width:none] sm:mx-0 sm:flex-wrap sm:justify-center sm:overflow-visible sm:px-0 [&::-webkit-scrollbar]:hidden">
          {JUMP.map((j) => (
            <li key={j.href} className="shrink-0">
              <a
                href={j.href}
                className="inline-flex items-center rounded-full border border-border bg-background/80 px-4 py-2 text-sm font-medium text-foreground shadow-sm transition-colors hover:border-primary/40 hover:text-primary"
              >
                {j.label}
              </a>
            </li>
          ))}
        </ul>
      </nav>
    </section>
  );
}

function HeroVisual() {
  return (
    <Reveal
      delay={0.1}
      direction="left"
      className="relative mx-auto w-full max-w-[25rem] pt-11 lg:mr-0 lg:max-w-[33rem] lg:pt-4"
    >
      <div className="pointer-events-none absolute inset-x-6 top-12 bottom-12 -z-10 rounded-[3rem] bg-gradient-to-br from-primary/30 via-indigo-400/20 to-fuchsia-400/20 blur-3xl" aria-hidden />

      <div
        className="mx-auto w-[64%] -rotate-2 drop-shadow-[0_30px_45px_rgba(15,23,42,0.28)] lg:ml-[4%] lg:w-[56%]"
        role="img"
        aria-label="A sample digital business card in the Pulse design, with contact buttons and a QR code"
      >
        <div aria-hidden>
          <CardRenderer templateId="pulse" data={DEMO_CARD} interactive={false} />
        </div>
      </div>

      {/* The card's own short address */}
      {/* Phones: above the card. Desktop: beside it, so neither covers the card. */}
      <div className="absolute left-0 top-0 flex items-center gap-2 rounded-full border border-border bg-background/95 py-1.5 pl-1.5 pr-3 text-xs font-medium text-foreground shadow-lg backdrop-blur sm:text-sm lg:left-auto lg:right-0 lg:top-[16%]">
        <span className="grid h-6 w-6 place-items-center rounded-full bg-primary text-primary-foreground">
          <Link2 className="h-3.5 w-3.5" aria-hidden />
        </span>
        <span className="font-mono">trunc.site/m/alexandra</span>
      </div>

      {/* What the owner sees. Example numbers, labelled as such. */}
      <div className="relative z-10 -mt-8 ml-auto w-[12rem] rounded-2xl border border-border bg-background/95 p-3.5 shadow-xl backdrop-blur sm:w-52 lg:absolute lg:right-0 lg:top-[36%] lg:mt-0">
        <div className="flex items-center justify-between gap-2">
          <span className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">Card views</span>
          <span className="rounded-full bg-muted px-1.5 py-0.5 text-[10px] font-medium text-muted-foreground">Example</span>
        </div>
        <div className="mt-1 flex items-baseline gap-2">
          <span className="text-2xl font-bold tabular-nums text-foreground">1,284</span>
          <span className="text-xs font-semibold text-emerald-600 dark:text-emerald-400">+18%</span>
        </div>
        <div className="mt-2 flex h-10 items-end gap-1" aria-hidden>
          {WEEK.map((h, i) => (
            <span
              key={i}
              className={i === WEEK.length - 1 ? 'flex-1 rounded-sm bg-primary' : 'flex-1 rounded-sm bg-primary/25'}
              style={{ height: `${h}%` }}
            />
          ))}
        </div>
        <p className="mt-2 text-[11px] text-muted-foreground">
          <span className="font-semibold text-foreground">212</span> saved the contact
        </p>
      </div>
    </Reveal>
  );
}
