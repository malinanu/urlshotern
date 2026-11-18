import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Reveal } from '@/components/motion/Reveal';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { MICROSITES_ENABLED } from '@/lib/microsites';
import { SectionHeading, SIGNUP_HREF } from './shared';

const STEPS = [
  {
    n: '01',
    title: 'Create your free account',
    body: 'Sign up with your email. No credit card, no trial clock.',
  },
  {
    n: '02',
    title: MICROSITES_ENABLED ? 'Make a card, page or link' : 'Make a card or a link',
    body: 'Start from a design or paste a long link. You can change any of it later.',
  },
  {
    n: '03',
    title: 'Share it and watch the numbers',
    body: 'Every visit shows up in your dashboard, so you can see what works and do more of it.',
  },
];

/** Three steps, then the one action the whole page asks for. */
export default function FeaturesClosing() {
  return (
    <>
      <section className="border-t border-border bg-card/40 py-14 sm:py-20 lg:py-24">
        <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
          <SectionHeading title="Up and running in three steps" />
          <Stagger className="mt-10 grid grid-cols-1 gap-6 sm:mt-12 md:grid-cols-3 md:gap-10">
            {STEPS.map((s) => (
              <StaggerItem key={s.n} className="flex gap-4 md:block">
                <span className="font-mono text-3xl font-bold text-primary/30 sm:text-5xl">{s.n}</span>
                <div>
                  <div className="hidden h-px w-full bg-border md:mt-4 md:block" />
                  <h3 className="text-lg font-semibold text-foreground sm:text-xl md:mt-5">{s.title}</h3>
                  <p className="mt-1 text-sm text-muted-foreground sm:text-base md:mt-2">{s.body}</p>
                </div>
              </StaggerItem>
            ))}
          </Stagger>
        </div>
      </section>

      <section className="relative overflow-hidden bg-primary py-14 sm:py-20 lg:py-24">
        <div className="pointer-events-none absolute -right-24 -top-24 h-80 w-80 rounded-full bg-white/10 blur-3xl" aria-hidden />
        <div className="pointer-events-none absolute -bottom-24 -left-24 h-80 w-80 rounded-full bg-white/10 blur-3xl" aria-hidden />
        <Reveal className="relative mx-auto max-w-3xl px-4 text-center sm:px-6 lg:px-8">
          <h2 className="text-balance text-[1.75rem] font-bold leading-tight tracking-tight text-primary-foreground sm:text-4xl lg:text-5xl">
            Your card, your links, your numbers. Start free.
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-base text-primary-foreground sm:mt-4 sm:text-lg">
            Make your card in a few minutes, share it today, and see how many people opened it.
          </p>
          <div className="mt-7 flex flex-col items-center justify-center gap-3 sm:mt-8 sm:flex-row">
            <Button asChild size="lg" className="h-12 w-full bg-background px-7 text-base font-semibold text-primary hover:bg-background/90 sm:w-auto">
              <Link href={SIGNUP_HREF}>
                Create your free account
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <Button
              asChild
              size="lg"
              variant="outline"
              className="h-12 w-full border-primary-foreground/40 bg-transparent px-6 text-base text-primary-foreground hover:bg-primary-foreground/10 hover:text-primary-foreground sm:w-auto"
            >
              <Link href="/pricing">See pricing</Link>
            </Button>
          </div>
        </Reveal>
      </section>
    </>
  );
}
