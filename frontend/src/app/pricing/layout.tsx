import type { Metadata } from "next";
import type { ReactNode } from "react";
import { PLAN_NAMES } from '@/lib/plan-utils';
import { JsonLd } from "@/components/JsonLd";

export const metadata: Metadata = {
  title: "Pricing — Free URL Shortener Plans | Trunc",
  description: `Free URL shortener with no credit card required. Compare Trunc plans: ${PLAN_NAMES.free}, ${PLAN_NAMES.premium}, and ${PLAN_NAMES.enterprise}. Get analytics, custom domains, and QR codes. Cancel anytime.`,
  keywords: "free url shortener, url shortener pricing, free url shortener no sign up, best free url shortener, custom url shortener free",
  alternates: {
    canonical: 'https://trunc.site/pricing',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/pricing',
    siteName: 'Trunc',
    title: "Pricing — Free URL Shortener Plans | Trunc",
    description: "Free URL shortener with no credit card required. Get analytics, custom domains, and QR codes. Cancel anytime.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Pricing — Free URL Shortener Plans | Trunc",
    description: "Free URL shortener with no credit card required. Get analytics, custom domains, and QR codes. Cancel anytime.",
  },
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
      <JsonLd data={faqSchema} />
      {children}
    </>
  );
}
