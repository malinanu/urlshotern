import type { Metadata } from 'next';
import type { ReactNode } from 'react';

export const metadata: Metadata = {
  title: 'Integrations — Chrome, WordPress, Shopify and Canva | Trunc',
  description:
    'Use Trunc where you already work: a Chrome extension, a WordPress plugin, a Shopify app, a Canva app, and a REST API with API keys.',
  alternates: { canonical: 'https://trunc.site/integrations' },
};

export default function IntegrationsLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
