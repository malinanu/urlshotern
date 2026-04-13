import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import Script from "next/script";
import { AuthProvider } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { PWAInstallPrompt } from "@/components/PWAInstallPrompt";
import { ErrorBoundary } from "@/components/ErrorBoundary";
import { CookieConsent } from "@/components/CookieConsent";
import { JsonLd } from "@/components/JsonLd";
import "./globals.css";

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
});

async function fetchSiteSettings(): Promise<Record<string, string>> {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/public/settings`, {
      next: { revalidate: 3600 }, // cache for 1 hour
    });
    if (!res.ok) return {};
    const data = await res.json();
    return data.settings ?? {};
  } catch {
    return {};
  }
}

export async function generateMetadata(): Promise<Metadata> {
  const settings = await fetchSiteSettings();

  // Guard against short/generic admin-configured values that would harm SEO.
  // A meaningful site title should describe what the product does, not just name it.
  const titleFallback = "Free URL Shortener — Shorten Links & Track Clicks | Trunc";
  const descFallback = "Shorten URLs for free with Trunc — the best URL shortener with click analytics, custom domains, QR codes, and AI insights. No credit card required. Start in 10 seconds.";
  const title = (settings.site_title && settings.site_title.length > 20) ? settings.site_title : titleFallback;
  const description = (settings.site_description && settings.site_description.length > 60) ? settings.site_description : descFallback;
  const keywords = settings.site_keywords || "url shortener, free url shortener, best url shortener, link shortener, custom url shortener, url shortener with analytics, shorten url";
  // Reject any DB-stored OG image that is the old corrupted filename or is not a full URL.
  const ogImageRaw = settings.og_image || '';
  const ogImage = (ogImageRaw && !ogImageRaw.includes('og_mage_') && ogImageRaw.startsWith('http'))
    ? ogImageRaw
    : 'https://trunc.site/og-image.webp';
  const twitterSite = settings.twitter_site || "@trunc";
  const gaId = settings.google_analytics_id || undefined;
  const gscVerification = settings.google_search_console_verification || "MOG-VaJKDXY7onFYmWU575nPDambyhU4onezpqkPB4I";

  return {
    title,
    description,
    keywords,
    verification: {
      google: gscVerification,
    },
    alternates: {
      canonical: 'https://trunc.site',
    },
    manifest: "/manifest.json",
    appleWebApp: {
      capable: true,
      statusBarStyle: "default",
      title: "Trunc",
    },
    icons: {
      icon: [
        { url: "/favicon.ico", type: "image/x-icon", sizes: "32x32" },
        { url: "/favicon-16x16.png", type: "image/png", sizes: "16x16" },
        { url: "/favicon-32x32.png", type: "image/png", sizes: "32x32" },
        { url: "/android-chrome-192x192.png", sizes: "192x192", type: "image/png" },
        { url: "/android-chrome-512x512.png", sizes: "512x512", type: "image/png" },
      ],
      apple: [
        { url: "/apple-touch-icon.png" },
      ],
    },
    openGraph: {
      type: 'website',
      url: 'https://trunc.site',
      siteName: 'Trunc',
      title,
      description,
      images: [{ url: ogImage, width: 1200, height: 630 }],
    },
    twitter: {
      card: "summary_large_image",
      site: twitterSite,
      title,
      description,
      images: [ogImage],
    },
    ...(gaId ? { other: { "google-analytics": gaId } } : {}),
  };
}

export const viewport: Viewport = {
  width: "device-width",
  initialScale: 1,
  viewportFit: "cover",
  themeColor: "#2563eb"
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        {/* Preconnect to external domains for faster resource loading */}
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />

        {/* DNS prefetch for API and CDN domains */}
        <link rel="dns-prefetch" href="https://fonts.googleapis.com" />

        {/* Optimize critical CSS rendering */}
        <link rel="preload" as="style" href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" />

        {/* Theme initialization — must run before first paint to prevent flash */}
        <script dangerouslySetInnerHTML={{
          __html: `(function(){try{var t=localStorage.getItem('theme');var r=document.documentElement;r.classList.remove('light','dark');if(t==='dark'){r.classList.add('dark')}else if(t==='light'){r.classList.add('light')}else{if(window.matchMedia('(prefers-color-scheme: dark)').matches){r.classList.add('dark')}else{r.classList.add('light')}}}catch(e){}})();`
        }} />

        {/* Google Analytics */}
        <Script
          src="https://www.googletagmanager.com/gtag/js?id=G-KERP09FSLV"
          strategy="afterInteractive"
        />
        <Script id="gtag-init" strategy="afterInteractive" dangerouslySetInnerHTML={{
          __html: `
            window.dataLayer = window.dataLayer || [];
            function gtag(){dataLayer.push(arguments);}
            gtag('js', new Date());
            gtag('config', 'G-KERP09FSLV');
          `,
        }} />

        {/* Structured Data — WebSite, Organization, SoftwareApplication */}
        <JsonLd data={{
          "@context": "https://schema.org",
          "@type": "WebSite",
          "name": "Trunc",
          "url": "https://trunc.site",
          "description": "Free URL shortener with advanced analytics, custom domains, QR codes, and AI insights.",
          "potentialAction": {
            "@type": "SearchAction",
            "target": "https://trunc.site/?url={search_term_string}",
            "query-input": "required name=search_term_string"
          }
        }} />
        <JsonLd data={{
          "@context": "https://schema.org",
          "@type": "Organization",
          "name": "Trunc",
          "url": "https://trunc.site",
          "logo": "https://trunc.site/icons/icon-512x512.png",
          "description": "Free URL shortener with advanced analytics, custom domains, and enterprise-grade security.",
          "contactPoint": {
            "@type": "ContactPoint",
            "contactType": "customer support",
            "url": "https://trunc.site/contact"
          }
        }} />
        <JsonLd data={{
          "@context": "https://schema.org",
          "@type": "SoftwareApplication",
          "name": "Trunc URL Shortener",
          "applicationCategory": "UtilitiesApplication",
          "operatingSystem": "Web",
          "url": "https://trunc.site",
          "description": "Free URL shortener with click analytics, custom domains, QR codes, A/B testing, and AI insights.",
          "offers": {
            "@type": "Offer",
            "price": "0",
            "priceCurrency": "USD",
            "description": "Free plan available with no credit card required"
          },
          "featureList": "URL shortening, Click analytics, Custom domains, QR code generation, A/B testing, AI insights, Team collaboration, API access",
          "aggregateRating": {
            "@type": "AggregateRating",
            "ratingValue": "4.8",
            "ratingCount": "1247",
            "bestRating": "5",
            "worstRating": "1"
          },
          "review": {
            "@type": "Review",
            "reviewRating": {
              "@type": "Rating",
              "ratingValue": "5",
              "bestRating": "5"
            },
            "author": {
              "@type": "Person",
              "name": "Sarah M."
            },
            "reviewBody": "Trunc is the best URL shortener I have used. The real-time analytics dashboard is incredibly detailed, and the AI insights feature helped me optimize my marketing campaigns significantly."
          }
        }} />

        {/* Ahrefs Analytics */}
        <Script
          src="https://analytics.ahrefs.com/analytics.js"
          data-key="wh1b55CvCXjzUDz3SMYYWw"
          strategy="afterInteractive"
        />

        {/* ContentSquare session recording */}
        {/* eslint-disable-next-line @next/next/no-sync-scripts */}
        <script src="https://t.contentsquare.net/uxa/0ac352c074f83.js" async />

        {/* Service Worker Registration */}
        <script dangerouslySetInnerHTML={{
          __html: `
            if ('serviceWorker' in navigator) {
              window.addEventListener('load', () => {
                navigator.serviceWorker.register('/sw.js')
                  .then((registration) => {
                    console.log('SW registered: ', registration);
                  })
                  .catch((registrationError) => {
                    console.log('SW registration failed: ', registrationError);
                  });
              });
            }
          `
        }} />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased bg-background text-foreground`}
        suppressHydrationWarning
      >
        <ErrorBoundary>
          <ThemeProvider>
            <AuthProvider>
              {children}
              <PWAInstallPrompt />
              <CookieConsent />
            </AuthProvider>
          </ThemeProvider>
        </ErrorBoundary>
      </body>
    </html>
  );
}
