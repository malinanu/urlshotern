'use client';

import { openConsentSettings } from '@/lib/consent';

/** Re-opens the cookie banner so visitors can change or withdraw consent at any time. */
export function CookieSettingsButton({ className }: { className?: string }) {
  return (
    <button type="button" onClick={openConsentSettings} className={className}>
      Cookie settings
    </button>
  );
}
