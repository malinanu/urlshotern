import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Sign In - Trunc",
  description: "Sign in to your Trunc account to manage your shortened links, view analytics, and access your dashboard.",
  robots: { index: false, follow: false },
};

export default function LoginLayout({ children }: { children: React.ReactNode }) {
  return children;
}
