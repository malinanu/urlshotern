import { Contact, Download, ImagePlus, LayoutTemplate, type LucideIcon } from 'lucide-react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { SectionHeading } from './shared';

interface Item {
  Icon: LucideIcon;
  title: string;
  body: string;
  href: string;
}

const ITEMS: Item[] = [
  {
    Icon: Contact,
    title: '12 digital business card designs',
    body: 'Four collections, from black-and-gold executive to bold colour. Each card has a live QR code, one-tap save to contacts, WhatsApp, maps and your social links.',
    href: '#cards',
  },
  {
    Icon: Download,
    title: 'Your card as an image',
    body: 'Download a sharp 1200 × 2100 PNG or JPG for email signatures and slides, and a 1024 px QR code for print.',
    href: '#cards',
  },
  {
    Icon: ImagePlus,
    title: 'Photo and logo that stay readable',
    body: 'Upload a photo and your company logo. Trunc checks the logo’s colours and puts it on a light or dark plate so it never disappears into the card.',
    href: '#cards',
  },
  ...(MICROSITES_ENABLED
    ? [
        {
          Icon: LayoutTemplate,
          title: 'Pages for shops and events',
          body: 'Product prices, countdowns, programmes, maps, Add to calendar and contact forms with an inbox, on one mobile page.',
          href: '#pages',
        },
      ]
    : []),
];

/** The new things, in one scannable row, each pointing at the section that shows it. */
export default function WhatsNew() {
  return (
    <section id="new" className="scroll-mt-20 py-14 sm:py-20">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Just shipped"
          title="New in Trunc"
          lead="Everything here is live in your account today, and the card designs are on the free plan."
        />
        {/* gap-px over a border-coloured grid draws the dividers; the column counts never leave an empty cell. */}
        <Stagger
          className={`mt-10 grid grid-cols-1 gap-px overflow-hidden rounded-2xl border border-border bg-border ${
            ITEMS.length === 4 ? 'sm:grid-cols-2 lg:grid-cols-4' : 'sm:grid-cols-3'
          }`}
        >
          {ITEMS.map(({ Icon, title, body, href }) => (
            <StaggerItem key={title} className="bg-background p-6">
              <a href={href} className="group block h-full rounded-lg outline-none focus-visible:ring-2 focus-visible:ring-primary">
                <span className="grid h-10 w-10 place-items-center rounded-xl bg-primary/10 text-primary">
                  <Icon className="h-5 w-5" aria-hidden />
                </span>
                <h3 className="mt-4 text-base font-semibold text-foreground group-hover:text-primary">{title}</h3>
                <p className="mt-1.5 text-sm leading-relaxed text-muted-foreground">{body}</p>
              </a>
            </StaggerItem>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
