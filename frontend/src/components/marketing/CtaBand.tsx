import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Reveal } from '@/components/motion/Reveal';

/**
 * Full bleed closing CTA. Single primary intent (start free) plus a secondary
 * link to pricing. No duplicate CTA intent elsewhere on the page.
 */
export default function CtaBand() {
  return (
    <section className="relative overflow-hidden bg-primary py-14 sm:py-20 lg:py-28">
      <div className="pointer-events-none absolute -right-24 -top-24 h-80 w-80 rounded-full bg-white/10 blur-3xl" aria-hidden />
      <div className="pointer-events-none absolute -bottom-24 -left-24 h-80 w-80 rounded-full bg-white/10 blur-3xl" aria-hidden />

      <div className="relative mx-auto max-w-3xl px-4 text-center sm:px-6 lg:px-8">
        <Reveal>
          <h2 className="text-balance text-2xl font-bold tracking-tight text-primary-foreground sm:text-4xl lg:text-5xl">
            Start shortening in seconds
          </h2>
          <p className="mx-auto mt-3 max-w-xl text-base text-primary-foreground/80 sm:mt-4 sm:text-lg">
            Join thousands of businesses tracking and optimizing their links with
            Trunc. Free to start, no credit card.
          </p>
          <div className="mt-6 flex flex-col items-center justify-center gap-3 sm:mt-8 sm:flex-row">
            <Button
              asChild
              size="lg"
              className="w-full bg-background text-primary hover:bg-background/90 sm:w-auto"
            >
              <Link href="/register">
                Get started free
                <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
            <Button
              asChild
              size="lg"
              variant="outline"
              className="w-full border-primary-foreground/30 bg-transparent text-primary-foreground hover:bg-primary-foreground/10 hover:text-primary-foreground sm:w-auto"
            >
              <Link href="/pricing">View pricing</Link>
            </Button>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
