import type { Metadata } from "next";
import type { ReactNode } from "react";
import { JsonLd } from "@/components/JsonLd";

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: "URL Shortener Pricing: Free, Premium & Enterprise | Trunc",
  description: `Free URL shortener with no credit card required. Compare Trunc plans: Free, Premium ($9.99/mo), and Enterprise ($39.99/mo). Get analytics, custom domains, and QR codes. Cancel anytime.`,
  keywords: "free url shortener, url shortener pricing, free url shortener no sign up, best free url shortener, custom url shortener free",
  alternates: {
    canonical: `${SITE_URL}/pricing`,
  },
  openGraph: {
    type: 'website',
    url: `${SITE_URL}/pricing`,
    siteName: 'Trunc',
    title: "URL Shortener Pricing: Free, Premium & Enterprise | Trunc",
    description: "Free URL shortener with no credit card required. Get analytics, custom domains, and QR codes. Cancel anytime.",
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "URL Shortener Pricing: Free, Premium & Enterprise | Trunc",
    description: "Free URL shortener with no credit card required. Get analytics, custom domains, and QR codes. Cancel anytime.",
    images: [`${SITE_URL}/og-image.webp`],
  },
};

const breadcrumbSchema = {
  '@context': 'https://schema.org',
  '@type': 'BreadcrumbList',
  itemListElement: [
    { '@type': 'ListItem', position: 1, name: 'Home', item: SITE_URL },
    { '@type': 'ListItem', position: 2, name: 'Pricing', item: `${SITE_URL}/pricing` },
  ],
};

const pricingSchema = {
  '@context': 'https://schema.org',
  '@type': 'SoftwareApplication',
  name: 'Trunc URL Shortener',
  applicationCategory: 'UtilitiesApplication',
  operatingSystem: 'Web',
  url: SITE_URL,
  description: 'Free URL shortener with click analytics, custom domains, QR codes, A/B testing, and AI insights.',
  offers: [
    {
      '@type': 'Offer',
      name: 'Free',
      description: 'Perfect for getting started. 10 short links/month, basic analytics, QR codes, custom short codes.',
      price: '0',
      priceCurrency: 'USD',
      priceSpecification: {
        '@type': 'UnitPriceSpecification',
        price: '0',
        priceCurrency: 'USD',
        unitText: 'MONTH',
      },
      url: `${SITE_URL}/pricing`,
    },
    {
      '@type': 'Offer',
      name: 'Premium',
      description: '1,000 short links/month, advanced analytics, priority support, custom domains, team collaboration (5 users), API access, password protection.',
      price: '9.99',
      priceCurrency: 'USD',
      priceSpecification: {
        '@type': 'UnitPriceSpecification',
        price: '9.99',
        priceCurrency: 'USD',
        unitText: 'MONTH',
      },
      url: `${SITE_URL}/pricing`,
    },
    {
      '@type': 'Offer',
      name: 'Enterprise',
      description: 'Unlimited short links, advanced analytics, 24/7 premium support, unlimited team members, unlimited API access, custom domains, password protection.',
      price: '39.99',
      priceCurrency: 'USD',
      priceSpecification: {
        '@type': 'UnitPriceSpecification',
        price: '39.99',
        priceCurrency: 'USD',
        unitText: 'MONTH',
      },
      url: `${SITE_URL}/pricing`,
    },
  ],
};

const faqSchema = {
  "@context": "https://schema.org",
  "@type": "FAQPage",
  "mainEntity": [
    {
      "@type": "Question",
      "name": "Can I change plans at any time?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and billing is prorated."
      }
    },
    {
      "@type": "Question",
      "name": "What happens if I exceed my link limit?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "On the Free plan, you'll need to upgrade. On paid plans, we'll notify you and provide options to upgrade or purchase additional capacity."
      }
    },
    {
      "@type": "Question",
      "name": "Do you offer annual billing?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes! Annual billing comes with a 20% discount. Contact us for annual pricing details."
      }
    },
    {
      "@type": "Question",
      "name": "Is there a free trial for paid plans?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Yes, we offer a 14-day free trial for all paid plans. No credit card required to start."
      }
    },
    {
      "@type": "Question",
      "name": "Can I use my own domain?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Custom domains are available on Premium and Enterprise plans. You can use your own branded short domain."
      }
    },
    {
      "@type": "Question",
      "name": "What kind of support do you provide?",
      "acceptedAnswer": {
        "@type": "Answer",
        "text": "Free users get community support. Premium users get priority email support. Enterprise users get 24/7 phone and email support."
      }
    }
  ]
};

export default function PricingLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <JsonLd data={breadcrumbSchema} />
      <JsonLd data={pricingSchema} />
      <JsonLd data={faqSchema} />
      {children}
    </>
  );
}
