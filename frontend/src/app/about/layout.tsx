import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "About Trunc - Our Story & Mission",
  description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses. Discover our mission and team.",
  openGraph: {
    title: "About Trunc - Our Story & Mission",
    description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses.",
  },
  twitter: {
    card: "summary",
    title: "About Trunc - Our Story & Mission",
    description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses.",
  },
};

export default function AboutLayout({ children }: { children: React.ReactNode }) {
  return children;
}
