import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Contact Us - Trunc",
  description: "Get in touch with the Trunc team. We're here to help with questions about our URL shortener, plans, or technical support.",
  alternates: {
    canonical: 'https://trunc.site/contact',
  },
  openGraph: {
    title: "Contact Us - Trunc",
    description: "Get in touch with the Trunc team. We're here to help.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Contact Us - Trunc",
    description: "Get in touch with the Trunc team. We're here to help.",
  },
};

export default function ContactLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
