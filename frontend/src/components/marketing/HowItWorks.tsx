import { Stagger, StaggerItem } from '@/components/motion/Stagger';

const steps = [
  {
    n: '01',
    title: 'Paste your link',
    body: 'Drop any long URL into the box. Add a custom code or password if you want.',
  },
  {
    n: '02',
    title: 'Get a short link',
    body: 'We generate a fast, reliable short link and QR code in an instant.',
  },
  {
    n: '03',
    title: 'Track everything',
    body: 'Share it, then watch clicks, locations, and referrers roll in live.',
  },
];

/**
 * A numbered horizontal timeline. Deliberately a different layout family from
 * the bento grid and cards, so the page does not repeat a single pattern.
 */
export default function HowItWorks() {
  return (
    <section className="border-t border-border bg-card/40 py-14 sm:py-20 lg:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="max-w-2xl">
          <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl md:text-4xl">
            From long to tracked in seconds
          </h2>
          <p className="mt-3 text-base text-muted-foreground sm:mt-4 sm:text-lg">
            No setup, no learning curve. Three steps and you are done.
          </p>
        </div>

        <Stagger className="mt-10 grid grid-cols-1 gap-6 sm:mt-14 md:grid-cols-3 md:gap-10">
          {steps.map((step) => (
            <StaggerItem key={step.n} className="relative flex gap-4 md:block">
              <span className="font-mono text-3xl font-bold text-primary/25 sm:text-5xl">
                {step.n}
              </span>
              <div>
                <div className="hidden h-px w-full bg-border md:mt-4 md:block" />
                <h3 className="text-lg font-semibold text-foreground sm:text-xl md:mt-5">
                  {step.title}
                </h3>
                <p className="mt-1 text-sm text-muted-foreground sm:text-base md:mt-2">
                  {step.body}
                </p>
              </div>
            </StaggerItem>
          ))}
        </Stagger>
      </div>
    </section>
  );
}
