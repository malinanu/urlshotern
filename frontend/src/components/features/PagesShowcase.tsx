import Link from 'next/link';
import { ArrowRight, CalendarPlus, Inbox, Blocks } from 'lucide-react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { Reveal } from '@/components/motion/Reveal';
import { PlanTag, SectionHeading, SIGNUP_HREF } from './shared';

const SHOTS = [
  {
    src: '/features/page-card.webp',
    title: 'Business card',
    body: 'Save Contact up front, then call, WhatsApp and email.',
    alt: 'A business card page: a white card with a blue painted edge, the owner’s name and a Save Contact button',
  },
  {
    src: '/features/page-links.webp',
    title: 'Link in bio',
    body: 'Your newest link first, the rest in a tidy list.',
    alt: 'A link-in-bio page styled like a station departure board, with the newest link highlighted in yellow',
  },
  {
    src: '/features/page-shop.webp',
    title: 'Shop and offers',
    body: 'Prices, deal badges and Buy buttons that sell.',
    alt: 'A shop page styled like a weekly flyer, with large prices, a starburst badge and Buy Now buttons',
  },
  {
    src: '/features/page-event.webp',
    title: 'Event',
    body: 'Date, venue, countdown, tickets and the programme.',
    alt: 'An event page with the date, time and venue, a countdown, Get tickets and Add to calendar buttons, and the programme',
  },
];

const DETAILS = [
  {
    Icon: Blocks,
    title: 'Blocks for everything',
    body: 'Product cards and carousels, countdowns, agendas, maps with directions, and YouTube, Vimeo and Spotify players that load when tapped.',
    tier: 'all' as const,
  },
  {
    Icon: CalendarPlus,
    title: 'Saved in one tap',
    body: 'Visitors add your event to their calendar or save your contact to their phone, and you see how many did.',
    tier: 'all' as const,
  },
  {
    Icon: Inbox,
    title: 'Forms with an inbox',
    body: 'Newsletter and contact forms with spam protection. Replies land in your inbox with an email alert, and export to CSV.',
    tier: 'premium' as const,
  },
];

/** Mobile pages (Microsites). Only rendered when the Pages feature is enabled. */
export default function PagesShowcase() {
  return (
    <section id="pages" className="scroll-mt-20 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
          <SectionHeading
            eyebrow="Pages"
            title="A mobile page for whatever you are sharing"
            lead="Start from a design made for the job, change the words and colours, and publish at trunc.site/m/your-name. Each page has its own analytics."
          />
          <Reveal className="shrink-0">
            <Link href={SIGNUP_HREF} className="inline-flex items-center gap-1.5 text-sm font-semibold text-primary hover:underline">
              Build your first page free
              <ArrowRight className="h-4 w-4" aria-hidden />
            </Link>
          </Reveal>
        </div>

        <Stagger className="-mx-4 mt-10 flex snap-x snap-mandatory gap-5 overflow-x-auto px-4 pb-4 [scrollbar-width:none] sm:mx-0 sm:grid sm:grid-cols-2 sm:gap-8 sm:overflow-visible sm:px-0 lg:grid-cols-4 [&::-webkit-scrollbar]:hidden">
          {SHOTS.map((s) => (
            <StaggerItem key={s.src} className="w-[62%] shrink-0 snap-center sm:w-auto">
              <figure>
                <div className="mx-auto max-w-[15rem] rounded-[2.2rem] border border-white/10 bg-slate-900 p-2 shadow-[0_24px_48px_-12px_rgba(15,23,42,0.35)]">
                  <div className="overflow-hidden rounded-[1.75rem] bg-background">
                    {/* eslint-disable-next-line @next/next/no-img-element -- static export, pre-sized WebP */}
                    <img
                      src={s.src}
                      alt={s.alt}
                      width={480}
                      height={985}
                      loading="lazy"
                      decoding="async"
                      className="block h-auto w-full"
                    />
                  </div>
                </div>
                <figcaption className="mx-auto mt-4 max-w-[15rem]">
                  <span className="block font-semibold text-foreground">{s.title}</span>
                  <span className="mt-0.5 block text-sm text-muted-foreground">{s.body}</span>
                </figcaption>
              </figure>
            </StaggerItem>
          ))}
        </Stagger>

        <div className="mt-12 grid gap-4 md:grid-cols-3">
          {DETAILS.map(({ Icon, title, body, tier }) => (
            <Reveal key={title} className="rounded-2xl border border-border bg-card/50 p-5 sm:p-6">
              <div className="flex items-start justify-between gap-3">
                <span className="grid h-10 w-10 place-items-center rounded-xl bg-primary/10 text-primary">
                  <Icon className="h-5 w-5" aria-hidden />
                </span>
                <PlanTag tier={tier} />
              </div>
              <h3 className="mt-4 font-semibold text-foreground">{title}</h3>
              <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">{body}</p>
            </Reveal>
          ))}
        </div>

        <p className="mt-6 text-sm text-muted-foreground">
          Free includes 1 page, Premium 10, and Growth and Enterprise have no limit.
        </p>
      </div>
    </section>
  );
}
