"use client";

import { useState, useEffect } from "react";
import Link from "next/link";

interface CookieConsentData {
  essential: true;
  analytics: boolean;
  marketing: boolean;
  savedAt: string;
}

function saveConsent(analytics: boolean, marketing: boolean): void {
  const data: CookieConsentData = {
    essential: true,
    analytics,
    marketing,
    savedAt: new Date().toISOString(),
  };
  localStorage.setItem("cookie_consent", JSON.stringify(data));
  window.dispatchEvent(new CustomEvent("cookie-consent-updated", { detail: data }));
}

export function CookieConsent() {
  const [visible, setVisible] = useState(false);
  const [expanded, setExpanded] = useState(false);
  const [analytics, setAnalytics] = useState(true);
  const [marketing, setMarketing] = useState(true);

  useEffect(() => {
    const stored = localStorage.getItem("cookie_consent");
    if (!stored) {
      setVisible(true);
    }
  }, []);

  if (!visible) return null;

  function acceptAll() {
    saveConsent(true, true);
    setVisible(false);
  }

  function rejectNonEssential() {
    saveConsent(false, false);
    setVisible(false);
  }

  function savePreferences() {
    saveConsent(analytics, marketing);
    setVisible(false);
  }

  return (
    <div className="fixed bottom-0 left-0 right-0 z-50 border-t border-border bg-background shadow-lg">
      <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6">
        {/* Main bar */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            We use cookies to improve your experience. Essential cookies are always active.{" "}
            <Link href="/privacy" className="underline hover:text-foreground">
              Cookie Policy
            </Link>
          </p>
          <div className="flex flex-shrink-0 flex-wrap gap-2">
            <button
              onClick={rejectNonEssential}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted transition-colors"
            >
              Reject Non-Essential
            </button>
            <button
              onClick={acceptAll}
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              Accept All
            </button>
            <button
              onClick={() => setExpanded((v) => !v)}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted transition-colors"
            >
              Manage {expanded ? "▲" : "▾"}
            </button>
          </div>
        </div>

        {/* Expanded preferences panel */}
        {expanded && (
          <div className="mt-4 rounded-lg border border-border bg-muted/30 p-4">
            <div className="space-y-3">
              {/* Essential */}
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-foreground">Essential Cookies</p>
                  <p className="text-xs text-muted-foreground">
                    Auth session, theme preferences. Always active.
                  </p>
                </div>
                <div className="flex h-6 w-11 flex-shrink-0 cursor-not-allowed items-center rounded-full bg-primary opacity-60">
                  <span className="ml-auto mr-0.5 h-5 w-5 rounded-full bg-white shadow" />
                </div>
              </div>

              {/* Analytics */}
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-foreground">Analytics Cookies</p>
                  <p className="text-xs text-muted-foreground">
                    Performance tracking, usage analytics to improve the service.
                  </p>
                </div>
                <button
                  role="switch"
                  aria-checked={analytics}
                  onClick={() => setAnalytics((v) => !v)}
                  className={`relative flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors ${
                    analytics ? "bg-primary" : "bg-muted"
                  }`}
                >
                  <span
                    className={`h-5 w-5 rounded-full bg-white shadow transition-transform ${
                      analytics ? "translate-x-5" : "translate-x-0.5"
                    }`}
                  />
                </button>
              </div>

              {/* Marketing */}
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-foreground">Marketing Cookies</p>
                  <p className="text-xs text-muted-foreground">
                    Marketing emails and promotional content preferences.
                  </p>
                </div>
                <button
                  role="switch"
                  aria-checked={marketing}
                  onClick={() => setMarketing((v) => !v)}
                  className={`relative flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors ${
                    marketing ? "bg-primary" : "bg-muted"
                  }`}
                >
                  <span
                    className={`h-5 w-5 rounded-full bg-white shadow transition-transform ${
                      marketing ? "translate-x-5" : "translate-x-0.5"
                    }`}
                  />
                </button>
              </div>
            </div>

            <div className="mt-4 flex items-center justify-between">
              <Link href="/privacy" className="text-xs text-muted-foreground underline hover:text-foreground">
                Cookie Policy
              </Link>
              <button
                onClick={savePreferences}
                className="rounded-md bg-primary px-4 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
              >
                Save Preferences
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
