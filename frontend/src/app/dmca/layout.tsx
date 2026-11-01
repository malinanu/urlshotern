import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "DMCA & Copyright Policy - Trunc",
  description: "How to report copyright infringement on Trunc short links, contact our Designated Agent, and submit a DMCA takedown or counter-notice.",
  alternates: {
    canonical: 'https://trunc.site/dmca',
  },
};

export default function DMCALayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
