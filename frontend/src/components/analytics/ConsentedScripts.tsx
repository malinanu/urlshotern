"use client";

import { useEffect, useState } from "react";
import Script from "next/script";

/**
 * Loads non-essential analytics / session-recording scripts ONLY after the
 * visitor has actively granted "analytics" consent via the cookie banner.
 *
 * Why this exists: under GDPR/ePrivacy and the UK PECR, trackers that are not
 * strictly necessary must not run until the user opts in. Rendering these
 * <Script> tags conditionally — instead of placing them unconditionally in the
 * document <head> — is what makes the consent real rather than cosmetic.
 *
 * Consent is read from the same `cookie_consent` localStorage key written by
 * CookieConsent.tsx, and we react live to the `cookie-consent-updated` event so
 * granting (or withdrawing) consent takes effect without a hard reload.
 */
function hasAnalyticsConsent(): boolean {
  try {
    const raw = localStorage.getItem("cookie_consent");
    if (!raw) return false;
    const parsed = JSON.parse(raw);
    return parsed?.analytics === true;
  } catch {
    return false;
  }
}

export function ConsentedScripts() {
  const [analyticsAllowed, setAnalyticsAllowed] = useState(false);

  useEffect(() => {
    setAnalyticsAllowed(hasAnalyticsConsent());

    function onConsentUpdated(e: Event) {
      const detail = (e as CustomEvent).detail as { analytics?: boolean } | undefined;
      setAnalyticsAllowed(detail?.analytics === true || hasAnalyticsConsent());
    }

    window.addEventListener("cookie-consent-updated", onConsentUpdated);
    return () => window.removeEventListener("cookie-consent-updated", onConsentUpdated);
  }, []);

  // Note: scripts already injected cannot be "un-injected" without a reload.
  // Withdrawing consent stops them loading on the next page load; for an
  // immediate stop the user can reload after withdrawing.
  if (!analyticsAllowed) return null;

  return (
    <>
      {/* Google Analytics (GA4) */}
      <Script
        src="https://www.googletagmanager.com/gtag/js?id=G-KERP09FSLV"
        strategy="afterInteractive"
      />
      <Script id="gtag-init" strategy="afterInteractive">
        {`
          window.dataLayer = window.dataLayer || [];
          function gtag(){dataLayer.push(arguments);}
          gtag('js', new Date());
          gtag('config', 'G-KERP09FSLV');
        `}
      </Script>

      {/* Ahrefs Analytics */}
      <Script
        src="https://analytics.ahrefs.com/analytics.js"
        data-key="wh1b55CvCXjzUDz3SMYYWw"
        strategy="afterInteractive"
      />

      {/* ContentSquare session analytics */}
      <Script
        src="https://t.contentsquare.net/uxa/0ac352c074f83.js"
        strategy="afterInteractive"
      />
    </>
  );
}
