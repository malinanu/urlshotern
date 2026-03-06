import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Contact Us - Trunc",
  description: "Get in touch with the Trunc team. We're here to help with questions about our URL shortener, plans, or technical support.",
  openGraph: {
    title: "Contact Us - Trunc",
    description: "Get in touch with the Trunc team. We're here to help.",
  },
  twitter: {
    card: "summary",
    title: "Contact Us - Trunc",
    description: "Get in touch with the Trunc team. We're here to help.",
  },
};

export default function ContactLayout({ children }: { children: React.ReactNode }) {
  return children;
}
