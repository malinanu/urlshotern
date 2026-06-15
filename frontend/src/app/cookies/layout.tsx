import type { Metadata } from "next";
import type { ReactNode } from "react";

const SITE_URL = "https://trunc.site";

export const metadata: Metadata = {
  title: "Cookie Policy | Trunc",
  description:
    "How Trunc uses cookies and similar technologies, the cookies we set, and how to manage or withdraw your consent.",
  alternates: {
    canonical: `${SITE_URL}/cookies`,
  },
  openGraph: {
    type: "website",
    url: `${SITE_URL}/cookies`,
    siteName: "Trunc",
    title: "Cookie Policy | Trunc",
    description:
      "How Trunc uses cookies and similar technologies, and how to manage or withdraw consent.",
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "Cookie Policy | Trunc",
    description: "How Trunc uses cookies and how to manage your consent.",
    images: [`${SITE_URL}/og-image.webp`],
  },
};

export default function CookiesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
