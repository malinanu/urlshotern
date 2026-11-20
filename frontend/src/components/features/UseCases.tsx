import { Briefcase, CalendarDays, Clapperboard, Store, type LucideIcon } from 'lucide-react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { SectionHeading } from './shared';

interface UseCase {
  Icon: LucideIcon;
  who: string;
  title: string;
  body: string;
  uses: string[];
}

const CASES: UseCase[] = [
  {
    Icon: Briefcase,
    who: 'Founders and sales',
    title: 'A card at every meeting',
    body: 'Show your QR code or send your link. They save you to their phone in one tap, and you see how many people did.',
    uses: ['Digital card', 'QR code', 'Analytics'],
  },
  {
    Icon: Clapperboard,
    who: 'Creators',
    title: 'One link in your bio, every week',
    body: 'Put the newest video, shop and newsletter on one page, then check which link people actually tap.',
    uses: [MICROSITES_ENABLED ? 'Link-in-bio page' : 'Bio page', 'Short links', 'UTM reports'],
  },
  {
    Icon: Store,
    who: 'Shops and restaurants',
    title: 'QR codes on menus and packaging',
    body: 'Print the code once. When the menu or the offer changes, point the link somewhere new and the print stays right.',
    uses: ['QR code', 'Editable short link', ...(MICROSITES_ENABLED ? ['Shop page'] : [])],
  },
  {
    Icon: CalendarDays,
    who: 'Events',
    title: 'A poster that stays up to date',
    body: MICROSITES_ENABLED
      ? 'Date, venue, countdown, programme and tickets on one page. Guests add it to their calendar, and a time change never needs a reprint.'
      : 'Put one QR code on the poster. If the venue or the ticket link changes, update the link and the poster stays right.',
    uses: MICROSITES_ENABLED ? ['Event page', 'QR code', 'Add to calendar'] : ['QR code', 'Editable short link'],
  },
];

export default function UseCases() {
  return (
    <section className="border-y border-border bg-card/40 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Built into your week"
          title="What people use Trunc for"
          lead="Start with one card or one link. Come back to update it when things change, and to see what is working."
        />
        <Stagger className="mt-10 grid grid-cols-1 gap-4 sm:mt-12 md:grid-cols-2 lg:gap-5">
          {CASES.map(({ Icon, who, title, body, uses }) => (
            <StaggerItem key={who} className="flex gap-4 rounded-2xl border border-border bg-background p-5 sm:p-6">
              <span className="grid h-11 w-11 shrink-0 place-items-center rounded-xl bg-foreground text-background">
                <Icon className="h-5 w-5" aria-hidden />
              </span>
              <div className="min-w-0">
                <p className="text-xs font-semibold uppercase tracking-wider text-primary">{who}</p>
                <h3 className="mt-1 text-lg font-semibold text-foreground">{title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">{body}</p>
                <ul className="mt-3 flex flex-wrap gap-1.5" aria-label="What they use">
                  {uses.map((u) => (
                    <li key={u} className="rounded-full border border-border bg-card px-2.5 py-0.5 text-xs font-medium text-foreground">
                      {u}
                    </li>
                  ))}
                </ul>
              </div>
            </StaggerItem>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
