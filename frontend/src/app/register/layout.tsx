import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Get Started Free - Trunc",
  description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes. No credit card required.",
  openGraph: {
    title: "Get Started Free - Trunc",
    description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes.",
  },
  twitter: {
    card: "summary",
    title: "Get Started Free - Trunc",
    description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes.",
  },
};

export default function RegisterLayout({ children }: { children: React.ReactNode }) {
  return children;
}
