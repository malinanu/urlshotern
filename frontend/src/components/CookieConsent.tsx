"use client";

import { useState, useEffect } from "react";
import Link from "next/link";
import { CONSENT_OPEN_EVENT, getConsent, saveConsent, type ConsentChoice } from "@/lib/consent";

const NONE: ConsentChoice = { analytics: false, replay: false, marketing: false };
const ALL: ConsentChoice = { analytics: true, replay: true, marketing: true };

function Toggle({ checked, onChange, label }: { checked: boolean; onChange: () => void; label: string }) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      aria-label={label}
      onClick={onChange}
      className={`relative flex h-6 w-11 flex-shrink-0 items-center rounded-full transition-colors ${
        checked ? "bg-primary" : "bg-muted"
      }`}
    >
      <span
        className={`h-5 w-5 rounded-full bg-white shadow transition-transform ${
          checked ? "translate-x-5" : "translate-x-0.5"
        }`}
      />
    </button>
  );
}

export function CookieConsent() {
  const [visible, setVisible] = useState(false);
  const [expanded, setExpanded] = useState(false);
  // Optional categories start switched off: nothing beyond essentials without an opt-in
  const [choice, setChoice] = useState<ConsentChoice>(NONE);

  useEffect(() => {
    if (!getConsent()) {
      setVisible(true);
    }

    // Footer "Cookie settings" re-opens the banner with the current choices
    const open = () => {
      const current = getConsent();
      setChoice(current ? { analytics: current.analytics, replay: current.replay, marketing: current.marketing } : NONE);
      setExpanded(true);
      setVisible(true);
    };
    window.addEventListener(CONSENT_OPEN_EVENT, open);
    return () => window.removeEventListener(CONSENT_OPEN_EVENT, open);
  }, []);

  if (!visible) return null;

  function save(next: ConsentChoice) {
    saveConsent(next);
    setVisible(false);
  }

  const toggle = (key: keyof ConsentChoice) => setChoice((c) => ({ ...c, [key]: !c[key] }));

  return (
    <div
      role="dialog"
      aria-label="Cookie preferences"
      className="fixed bottom-0 left-0 right-0 z-50 border-t border-border bg-background shadow-lg"
    >
      <div className="mx-auto max-w-7xl px-4 py-4 sm:px-6">
        {/* Main bar */}
        <div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
          <p className="text-sm text-muted-foreground">
            We use essential cookies to run the site. With your permission we&apos;d also use analytics
            and session replay, which records how you interact with our pages (form fields are masked).
            Nothing optional runs unless you opt in.{" "}
            <Link href="/privacy" className="underline hover:text-foreground">
              Cookie Policy
            </Link>
          </p>
          <div className="flex flex-shrink-0 flex-wrap gap-2">
            <button
              onClick={() => save(NONE)}
              className="rounded-md border border-border px-3 py-1.5 text-sm font-medium text-foreground hover:bg-muted transition-colors"
            >
              Reject Non-Essential
            </button>
            <button
              onClick={() => save(ALL)}
              className="rounded-md bg-primary px-3 py-1.5 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors"
            >
              Accept All
            </button>
            <button
              onClick={() => setExpanded((v) => !v)}
              aria-expanded={expanded}
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
                    Aggregate usage statistics (Google Analytics, Ahrefs) to improve the service.
                  </p>
                </div>
                <Toggle checked={choice.analytics} onChange={() => toggle("analytics")} label="Analytics cookies" />
              </div>

              {/* Session replay */}
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-foreground">Session Analytics &amp; Replay</p>
                  <p className="text-xs text-muted-foreground">
                    Records mouse movements, clicks, scrolling and page changes (Contentsquare) so we can
                    see where the site is confusing. Everything you type into form fields is masked and
                    never recorded.
                  </p>
                </div>
                <Toggle checked={choice.replay} onChange={() => toggle("replay")} label="Session analytics and replay" />
              </div>

              {/* Marketing */}
              <div className="flex items-start justify-between gap-4">
                <div>
                  <p className="text-sm font-medium text-foreground">Marketing Cookies</p>
                  <p className="text-xs text-muted-foreground">
                    Marketing emails and promotional content preferences.
                  </p>
                </div>
                <Toggle checked={choice.marketing} onChange={() => toggle("marketing")} label="Marketing cookies" />
              </div>
            </div>

            <div className="mt-4 flex items-center justify-between">
              <Link href="/privacy" className="text-xs text-muted-foreground underline hover:text-foreground">
                Cookie Policy
              </Link>
              <button
                onClick={() => save(choice)}
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
