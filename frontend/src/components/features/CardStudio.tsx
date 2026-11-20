'use client';

import Link from 'next/link';
import { useId, useMemo, useRef, useState, type KeyboardEvent } from 'react';
import { ArrowRight, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Reveal } from '@/components/motion/Reveal';
import { CardRenderer } from '@/components/micropages/cards/CardRenderer';
import { CARD_COLLECTIONS, paletteSwatch, templatesInCollection } from '@/components/micropages/cards/catalog';
import type { CardCollectionId, CardData } from '@/components/micropages/cards/types';
import { cn } from '@/lib/utils';
import { DEMO_CARD, SectionHeading, SIGNUP_HREF } from './shared';

const PERKS = ['Live QR code', 'One-tap save to contacts', 'WhatsApp, maps and socials', 'PNG, JPG and QR downloads'];

/** Lower-case letters and digits only, for the sample email and website on the card. */
function slugPart(s: string): string {
  return s
    .normalize('NFKD')
    .toLowerCase()
    .replace(/[^a-z0-9]/g, '');
}

/** The sample card with whatever the visitor typed. Empty fields fall back to the sample. */
function personalise(name: string, role: string, company: string): CardData {
  const n = name.trim();
  const r = role.trim();
  const c = company.trim();
  if (!n && !r && !c) return DEMO_CARD;
  const first = slugPart(n.split(/\s+/)[0] || '') || 'hello';
  const domain = (slugPart(c) || slugPart(DEMO_CARD.company)) + '.example';
  return {
    ...DEMO_CARD,
    name: n || DEMO_CARD.name,
    role: r || DEMO_CARD.role,
    company: c || DEMO_CARD.company,
    email: n || c ? `${first}@${domain}` : DEMO_CARD.email,
    website: c ? domain : DEMO_CARD.website,
  };
}

/**
 * "Try it with your name": the visitor types, and the real card templates update live.
 * Nothing is saved or sent anywhere; the button takes them to sign up to keep it.
 */
export default function CardStudio() {
  const uid = useId();
  const [name, setName] = useState('');
  const [role, setRole] = useState('');
  const [company, setCompany] = useState('');
  const [collection, setCollection] = useState<CardCollectionId>(CARD_COLLECTIONS[0].id);
  const [palettes, setPalettes] = useState<Record<string, string>>({});
  const tabRefs = useRef<(HTMLButtonElement | null)[]>([]);

  const data = useMemo(() => personalise(name, role, company), [name, role, company]);
  const templates = templatesInCollection(collection);
  const active = CARD_COLLECTIONS.find((c) => c.id === collection) ?? CARD_COLLECTIONS[0];

  const onTabKey = (e: KeyboardEvent<HTMLButtonElement>, i: number) => {
    const last = CARD_COLLECTIONS.length - 1;
    let next = -1;
    if (e.key === 'ArrowRight') next = i === last ? 0 : i + 1;
    else if (e.key === 'ArrowLeft') next = i === 0 ? last : i - 1;
    else if (e.key === 'Home') next = 0;
    else if (e.key === 'End') next = last;
    if (next < 0) return;
    e.preventDefault();
    setCollection(CARD_COLLECTIONS[next].id);
    tabRefs.current[next]?.focus();
  };

  const field =
    'h-11 w-full rounded-xl border border-border bg-background px-3.5 text-[15px] text-foreground shadow-sm placeholder:text-muted-foreground/70 focus:border-primary focus:outline-none focus:ring-2 focus:ring-primary/25';

  return (
    <section id="cards" className="scroll-mt-20 border-y border-border bg-card/40 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-5xl px-4 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Digital business cards"
          title="Try a card with your name on it"
          lead="Type your name and pick a design. This is the same card you get in your account: people scan the QR code or open your link, and save you to their contacts in one tap."
        />

        <Reveal delay={0.05} className="mt-8 grid gap-3 sm:grid-cols-3">
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-foreground">Your name</span>
            <input
              className={field}
              value={name}
              maxLength={40}
              autoComplete="off"
              placeholder={DEMO_CARD.name}
              onChange={(e) => setName(e.target.value)}
            />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-foreground">Job title</span>
            <input
              className={field}
              value={role}
              maxLength={44}
              autoComplete="off"
              placeholder={DEMO_CARD.role}
              onChange={(e) => setRole(e.target.value)}
            />
          </label>
          <label className="block">
            <span className="mb-1.5 block text-sm font-medium text-foreground">Company</span>
            <input
              className={field}
              value={company}
              maxLength={40}
              autoComplete="off"
              placeholder={DEMO_CARD.company}
              onChange={(e) => setCompany(e.target.value)}
            />
          </label>
        </Reveal>

        <div
          role="tablist"
          aria-label="Card collections"
          className="-mx-4 mt-8 flex gap-2 overflow-x-auto px-4 pb-1 [scrollbar-width:none] sm:mx-0 sm:flex-wrap sm:overflow-visible sm:px-0 [&::-webkit-scrollbar]:hidden"
        >
          {CARD_COLLECTIONS.map((c, i) => {
            const on = c.id === collection;
            return (
              <button
                key={c.id}
                ref={(el) => {
                  tabRefs.current[i] = el;
                }}
                type="button"
                role="tab"
                id={`${uid}-tab-${c.id}`}
                aria-selected={on}
                aria-controls={`${uid}-panel`}
                tabIndex={on ? 0 : -1}
                onClick={() => setCollection(c.id)}
                onKeyDown={(e) => onTabKey(e, i)}
                className={cn(
                  'shrink-0 rounded-full border px-4 py-2 text-sm font-semibold transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary',
                  on
                    ? 'border-foreground bg-foreground text-background'
                    : 'border-border bg-background text-foreground hover:border-foreground/40'
                )}
              >
                {c.name}
              </button>
            );
          })}
        </div>
        <p className="mt-3 text-sm text-muted-foreground">{active.description}</p>

        <div
          role="tabpanel"
          id={`${uid}-panel`}
          aria-labelledby={`${uid}-tab-${collection}`}
          className="-mx-4 mt-6 flex snap-x snap-mandatory gap-4 overflow-x-auto px-4 pb-4 [scrollbar-width:none] sm:gap-6 md:mx-0 md:grid md:grid-cols-3 md:overflow-visible md:px-0 md:pb-0 [&::-webkit-scrollbar]:hidden"
        >
          {templates.map((tpl) => {
            const paletteId = palettes[tpl.id] || tpl.palettes[0].id;
            return (
              <figure key={tpl.id} className="w-[72%] shrink-0 snap-center sm:w-[44%] md:w-auto">
                <div
                  role="img"
                  aria-label={`${tpl.name} card preview with ${data.name}`}
                  className="drop-shadow-[0_18px_30px_rgba(15,23,42,0.18)]"
                >
                  <div aria-hidden>
                    <CardRenderer templateId={tpl.id} data={data} paletteId={paletteId} interactive={false} />
                  </div>
                </div>
                <figcaption className="mt-4 flex items-start justify-between gap-3">
                  <span className="min-w-0">
                    <span className="block font-semibold text-foreground">{tpl.name}</span>
                    <span className="block truncate text-sm text-muted-foreground">{tpl.category}</span>
                  </span>
                  <span className="flex shrink-0 gap-1.5" role="group" aria-label={`${tpl.name} colours`}>
                    {tpl.palettes.map((p) => {
                      const sw = paletteSwatch(tpl, p);
                      const on = p.id === paletteId;
                      return (
                        <button
                          key={p.id}
                          type="button"
                          title={p.name}
                          aria-label={p.name}
                          aria-pressed={on}
                          onClick={() => setPalettes((s) => ({ ...s, [tpl.id]: p.id }))}
                          className={cn(
                            'h-7 w-7 rounded-full border-2 transition-transform focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary focus-visible:ring-offset-2',
                            on ? 'scale-110 border-foreground' : 'border-background shadow-[0_0_0_1px_hsl(var(--border))] hover:scale-105'
                          )}
                          style={{ background: `linear-gradient(135deg, ${sw.bg} 50%, ${sw.accent} 50%)` }}
                        />
                      );
                    })}
                  </span>
                </figcaption>
              </figure>
            );
          })}
        </div>

        <Reveal className="mt-10 flex flex-col items-start gap-6 rounded-2xl border border-border bg-background p-5 sm:p-6 lg:flex-row lg:items-center lg:justify-between">
          <ul className="grid grid-cols-1 gap-x-6 gap-y-2 text-sm text-foreground sm:grid-cols-2">
            {PERKS.map((p) => (
              <li key={p} className="inline-flex items-center gap-2">
                <Check className="h-4 w-4 shrink-0 text-emerald-600 dark:text-emerald-400" aria-hidden />
                {p}
              </li>
            ))}
          </ul>
          <div className="flex w-full flex-col gap-2 lg:w-auto lg:items-end">
            <Button asChild size="lg" className="h-12 w-full px-7 text-base font-semibold lg:w-auto">
              <Link href={SIGNUP_HREF}>
                Make this my card
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <span className="text-xs text-muted-foreground">Free plan. Add your photo, logo and details after you sign up.</span>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
