import Link from 'next/link';
import { ArrowDownTrayIcon, ArrowTopRightOnSquareIcon, CodeBracketIcon, KeyIcon } from '@heroicons/react/24/outline';
import Layout from '@/components/layout/Layout';
import { INTEGRATIONS, installAction, installSteps } from '@/lib/integrations';

/** Every plugin, how to install it today, and the API for everything else. */
export default function IntegrationsPage() {
  return (
    <Layout>
      <div className="mx-auto max-w-5xl px-4 py-12 sm:px-6 sm:py-16 lg:px-8">
        <p className="text-sm font-semibold uppercase tracking-wider text-primary">Integrations</p>
        <h1 className="mt-2 text-balance text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
          Use Trunc where you already work
        </h1>
        <p className="mt-3 max-w-2xl text-base text-muted-foreground sm:text-lg">
          Each plugin signs in with your Trunc account (Sign in with Trunc) or an API key, and every link it makes shows up in
          your dashboard with its clicks.
        </p>

        <div className="mt-10 grid grid-cols-1 gap-5 md:grid-cols-2">
          {INTEGRATIONS.map((i) => {
            const action = installAction(i);
            return (
              <section key={i.id} id={i.id} className="flex scroll-mt-24 flex-col rounded-2xl border border-border bg-card p-6">
                <div className="flex items-center gap-3">
                  <span className={`grid h-11 w-11 place-items-center rounded-xl border text-xs font-bold ${i.tone}`}>{i.badge}</span>
                  <h2 className="text-lg font-semibold text-foreground">{i.name}</h2>
                </div>
                <p className="mt-3 text-sm leading-relaxed text-muted-foreground">{i.summary}</p>

                {action ? (
                  <ol className="mt-4 list-decimal space-y-1.5 pl-5 text-sm text-foreground">
                    {installSteps(i).map((s) => (
                      <li key={s}>{s}</li>
                    ))}
                  </ol>
                ) : null}

                <div className="mt-auto pt-6">
                  {action ? (
                    <a
                      href={action.href}
                      {...(action.external ? { target: '_blank', rel: 'noopener noreferrer' } : { download: '' })}
                      className="inline-flex h-11 items-center gap-2 rounded-xl bg-primary px-5 text-sm font-semibold text-primary-foreground hover:bg-primary/90"
                    >
                      {action.external ? (
                        <ArrowTopRightOnSquareIcon className="h-4 w-4" aria-hidden />
                      ) : (
                        <ArrowDownTrayIcon className="h-4 w-4" aria-hidden />
                      )}
                      {action.label}
                    </a>
                  ) : (
                    <span className="inline-flex h-11 items-center rounded-xl border border-dashed border-border px-5 text-sm font-medium text-muted-foreground">
                      {i.id === 'shopify' ? 'Coming to the Shopify App Store' : 'Coming to Canva Apps'}
                    </span>
                  )}
                </div>
              </section>
            );
          })}
        </div>

        <section className="mt-12 rounded-2xl border border-border bg-muted/30 p-6">
          <h2 className="text-lg font-semibold text-foreground">Build your own</h2>
          <p className="mt-1 text-sm text-muted-foreground">
            Everything the plugins do is available in the REST API, on every plan.
          </p>
          <div className="mt-4 flex flex-wrap gap-3">
            <Link
              href="/api-keys"
              className="inline-flex h-10 items-center gap-2 rounded-xl border border-border bg-background px-4 text-sm font-semibold text-foreground hover:bg-muted"
            >
              <KeyIcon className="h-4 w-4" aria-hidden /> API keys
            </Link>
            <Link
              href="/api-docs"
              className="inline-flex h-10 items-center gap-2 rounded-xl border border-border bg-background px-4 text-sm font-semibold text-foreground hover:bg-muted"
            >
              <CodeBracketIcon className="h-4 w-4" aria-hidden /> API docs
            </Link>
          </div>
        </section>
      </div>
    </Layout>
  );
}
