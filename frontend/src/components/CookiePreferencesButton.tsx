"use client";

/**
 * Re-opens the cookie consent banner so a visitor can review or withdraw their
 * choice at any time. GDPR requires withdrawing consent to be as easy as giving
 * it — this button is the standing entry point (placed in the footer).
 */
export function CookiePreferencesButton({ className }: { className?: string }) {
  return (
    <button
      type="button"
      onClick={() => window.dispatchEvent(new CustomEvent("open-cookie-preferences"))}
      className={
        className ??
        "text-base text-muted-foreground hover:text-foreground transition-colors"
      }
    >
      Cookie Preferences
    </button>
  );
}
