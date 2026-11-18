import { BarChart3, FlaskConical, Link2, ShieldCheck, Sparkles, Users, type LucideIcon } from 'lucide-react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { cn } from '@/lib/utils';
import { PlanTag, SectionHeading, type PlanTier } from './shared';

interface Row {
  label: string;
  tier?: PlanTier;
  /** Only true when Pages (Microsites) is enabled. */
  pages?: boolean;
}

interface Tile {
  Icon: LucideIcon;
  title: string;
  /** Set when every row is on the same plans; the rows then carry no tag of their own. */
  tier?: PlanTier;
  rows: Row[];
  tint?: boolean;
}

// Each plan label matches internal/services/billing.go and internal/middleware/plan_gating.go.
const TILES: Tile[] = [
  {
    Icon: Link2,
    title: 'Short links',
    tint: true,
    rows: [
      { label: 'Change where a link goes, any time', tier: 'all' },
      { label: 'Expiry dates and social previews', tier: 'all' },
      { label: 'Custom back-halves', tier: 'premium' },
      { label: 'Your own domain', tier: 'premium' },
      { label: 'Password protection', tier: 'premium' },
      { label: 'Bulk shortening', tier: 'premium' },
    ],
  },
  {
    Icon: BarChart3,
    title: 'Analytics',
    rows: [
      { label: 'Clicks over time, devices, referrers', tier: 'all' },
      { label: 'Countries and cities', tier: 'premium' },
      { label: 'Time-of-day heatmap', tier: 'premium' },
      { label: 'Real-time feed and live map', tier: 'premium' },
      { label: 'UTM campaign reports', tier: 'premium' },
      { label: 'Bot filtering and CSV export', tier: 'premium' },
    ],
  },
  {
    Icon: Sparkles,
    title: 'AI insights',
    tier: 'enterprise',
    rows: [
      { label: 'Plain-language traffic summaries' },
      { label: 'Alerts when traffic looks unusual' },
      { label: 'Recommendations for your links' },
      { label: 'Traffic forecasts' },
      { label: 'Audience insights' },
    ],
  },
  {
    Icon: FlaskConical,
    title: 'Experiments',
    tier: 'all',
    rows: [
      { label: 'A/B tests with significance checks' },
      { label: 'Sample-size calculator' },
      { label: 'Conversion goals' },
      { label: 'Multi-touch attribution' },
    ],
  },
  {
    Icon: Users,
    title: 'Teams and API',
    rows: [
      { label: 'REST API and API keys', tier: 'all' },
      { label: 'Teams with roles', tier: 'premium' },
      { label: 'No Trunc badge on your pages', tier: 'growth', pages: true },
      { label: 'Unlimited team members', tier: 'enterprise' },
      { label: 'White-label branding', tier: 'enterprise' },
    ],
  },
  {
    Icon: ShieldCheck,
    title: 'Privacy and trust',
    tint: true,
    rows: [
      { label: 'Only you and your team see your analytics', tier: 'all' },
      { label: 'Pages count visitors without cookies', tier: 'all', pages: true },
      { label: 'Spam-protected forms on pages', tier: 'premium', pages: true },
      { label: 'One-click unsubscribe from our emails', tier: 'all' },
      { label: 'Clear DMCA and abuse process', tier: 'all' },
    ],
  },
];

/** The rest of the platform, with the plan each feature is on. */
export default function PlatformGrid() {
  return (
    <section id="platform" className="scroll-mt-20 py-14 sm:py-20 lg:py-24">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <SectionHeading
          eyebrow="Links and analytics"
          title="Everything else, in the same account"
          lead="Your cards and pages sit next to Trunc’s link, analytics and team tools. Each feature shows the plans it is on."
        />

        <Stagger className="mt-10 grid grid-cols-1 gap-4 sm:mt-12 md:grid-cols-2 lg:grid-cols-3 lg:gap-5">
          {TILES.map(({ Icon, title, tier, rows, tint }) => (
            <StaggerItem
              key={title}
              className={cn(
                'flex flex-col rounded-2xl border p-5 sm:p-6',
                tint ? 'border-primary/20 bg-primary/[0.05]' : 'border-border bg-card/50'
              )}
            >
              <div className="flex items-center justify-between gap-3">
                <div className="flex items-center gap-3">
                  <span className="grid h-10 w-10 place-items-center rounded-xl bg-primary/10 text-primary">
                    <Icon className="h-5 w-5" aria-hidden />
                  </span>
                  <h3 className="text-lg font-semibold text-foreground">{title}</h3>
                </div>
                {tier && <PlanTag tier={tier} />}
              </div>
              <ul className="mt-4 divide-y divide-border/70">
                {rows.filter((r) => MICROSITES_ENABLED || !r.pages).map((r) => (
                  <li key={r.label} className="flex items-center justify-between gap-3 py-2.5 text-sm text-foreground">
                    <span>{r.label}</span>
                    {r.tier && <PlanTag tier={r.tier} />}
                  </li>
                ))}
              </ul>
            </StaggerItem>
          ))}
        </Stagger>

        <p className="mt-6 text-sm text-muted-foreground">
          <span className="font-semibold text-foreground">Premium+</span> means Premium, Growth and Enterprise.{' '}
          <span className="font-semibold text-foreground">Growth+</span> means Growth and Enterprise.
        </p>
      </div>
    </section>
  );
}
