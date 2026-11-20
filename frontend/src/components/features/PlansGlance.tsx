import Link from 'next/link';
import { ArrowRight, Check } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { PLAN_NAMES } from '@/lib/plan-utils';
import { cn } from '@/lib/utils';
import { SectionHeading, SIGNUP_HREF } from './shared';

interface PlanSummary {
  name: string;
  forWho: string;
  points: string[];
  highlight?: boolean;
}

// Limits from PredefinedPlans in internal/services/billing.go. Prices live on /pricing only.
const PLANS: PlanSummary[] = [
  {
    name: PLAN_NAMES.free,
    forWho: 'To try everything that matters',
    highlight: true,
    points: [
      '10 active short links',
      'QR codes and all 12 card designs',
      ...(MICROSITES_ENABLED ? ['1 page'] : []),
      'Click analytics and API access',
    ],
  },
  {
    name: PLAN_NAMES.premium,
    forWho: 'For professionals and small teams',
    points: [
      '1,000 short links',
      'Your own domain and back-halves',
      'Password protection and bulk links',
      'A year of analytics history',
      ...(MICROSITES_ENABLED ? ['10 pages with forms'] : []),
      'Teams of up to 5',
    ],
  },
  {
    name: PLAN_NAMES.growth,
    forWho: 'For teams that share a lot',
    points: [
      'Everything in Premium',
      'Unlimited links and clicks',
      'Analytics history with no time limit',
      ...(MICROSITES_ENABLED ? ['Unlimited pages, no Trunc badge'] : []),
    ],
  },
  {
    name: PLAN_NAMES.enterprise,
    forWho: 'For larger organisations',
    points: ['Everything in Growth', 'AI insights and forecasts', 'White-label branding', 'Unlimited team members'],
  },
];

export default function PlansGlance() {
  return (
    <section id="plans" className="scroll-mt-20 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
          <SectionHeading
            eyebrow="Plans"
            title="Start free. Upgrade when you need more."
            lead="The free plan is not a trial. Keep it as long as you like, and move up only when you need more links, a domain or a team."
          />
          <Button asChild variant="outline" size="lg" className="h-11 w-full shrink-0 sm:w-auto">
            <Link href="/pricing">
              See prices and compare plans
              <ArrowRight className="h-4 w-4" />
            </Link>
          </Button>
        </div>

        <Stagger className="mt-10 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4 lg:gap-5">
          {PLANS.map((p) => (
            <StaggerItem
              key={p.name}
              className={cn(
                'flex flex-col rounded-2xl border p-5 sm:p-6',
                p.highlight ? 'border-primary bg-primary/[0.05] ring-1 ring-primary/30' : 'border-border bg-card/50'
              )}
            >
              <h3 className="text-xl font-bold text-foreground">{p.name}</h3>
              <p className="mt-1 text-sm text-muted-foreground">{p.forWho}</p>
              <ul className="mt-5 space-y-2.5">
                {p.points.map((pt) => (
                  <li key={pt} className="flex items-start gap-2 text-sm text-foreground">
                    <Check className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600 dark:text-emerald-400" aria-hidden />
                    {pt}
                  </li>
                ))}
              </ul>
              {p.highlight && (
                <Button asChild className="mt-6 h-11 w-full font-semibold">
                  <Link href={SIGNUP_HREF}>Create your free account</Link>
                </Button>
              )}
            </StaggerItem>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
