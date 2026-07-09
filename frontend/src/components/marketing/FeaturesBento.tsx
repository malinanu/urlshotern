import {
  BarChart3,
  Globe,
  QrCode,
  ShieldCheck,
  Sparkles,
  Zap,
} from 'lucide-react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { cn } from '@/lib/utils';

interface Cell {
  title: string;
  body: string;
  Icon: typeof Zap;
  tint?: boolean;
}

// Six features on an even, always-filling grid (no col/row spanning that could
// leave empty cells). Two cells are tinted for visual variety.
const cells: Cell[] = [
  {
    title: 'Real time analytics',
    body: 'See clicks, locations, devices, and referrers the moment they happen. Every link ships with a live dashboard.',
    Icon: BarChart3,
    tint: true,
  },
  {
    title: 'Lightning fast redirects',
    body: 'Global edge infrastructure keeps every redirect under a blink, everywhere in the world.',
    Icon: Zap,
  },
  {
    title: 'Custom domains',
    body: 'Brand every link with your own domain so your audience always sees you, not us.',
    Icon: Globe,
  },
  {
    title: 'QR codes built in',
    body: 'Generate a scannable QR code for any link instantly, ready for print or screen.',
    Icon: QrCode,
  },
  {
    title: 'AI insights',
    body: 'Plain language summaries and anomaly alerts that tell you what your traffic is doing.',
    Icon: Sparkles,
    tint: true,
  },
  {
    title: 'Secure by default',
    body: 'Password protection, link expiry, and enterprise grade reliability on every plan.',
    Icon: ShieldCheck,
  },
];

export default function FeaturesBento() {
  return (
    <section className="py-14 sm:py-20 lg:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl">
          <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-4xl lg:text-5xl">
            More than a shortener
          </h2>
          <p className="mt-3 text-base text-muted-foreground sm:mt-4 sm:text-lg">
            Everything you need to shorten, brand, and understand your links, in
            one place.
          </p>
        </div>

        <Stagger className="mt-8 grid grid-cols-2 gap-3 sm:mt-12 sm:gap-4 lg:grid-cols-3">
          {cells.map((cell) => (
            <StaggerItem
              key={cell.title}
              className={cn(
                'group h-full rounded-[calc(var(--radius))] border border-border p-4 transition-all duration-300 hover:-translate-y-1 sm:p-6',
                cell.tint
                  ? 'bg-gradient-to-br from-primary/10 via-card to-card'
                  : 'bg-card hover:border-primary/30',
              )}
            >
              <div className="mb-3 inline-flex h-10 w-10 items-center justify-center rounded-xl bg-primary/10 text-primary transition-transform duration-300 group-hover:scale-110 sm:mb-5 sm:h-11 sm:w-11">
                <cell.Icon className="h-5 w-5" />
              </div>
              <h3 className="text-base font-semibold text-foreground sm:text-lg">
                {cell.title}
              </h3>
              <p className="mt-1 text-xs leading-relaxed text-muted-foreground sm:mt-2 sm:text-sm">
                {cell.body}
              </p>
            </StaggerItem>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
