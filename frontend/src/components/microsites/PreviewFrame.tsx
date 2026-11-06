'use client';

import { useEffect, useRef, useState } from 'react';
import { ComputerDesktopIcon, DevicePhoneMobileIcon, DeviceTabletIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { apiFetch, describeValidation, type MicrositeContent } from '@/lib/microsites';

type Device = 'mobile' | 'tablet' | 'desktop';
const WIDTHS: Record<Device, number> = { mobile: 390, tablet: 768, desktop: 1280 };

interface Props {
  title: string;
  slug: string;
  content: MicrositeContent;
}

/**
 * Live preview: the draft is rendered by the same Go renderer that serves the
 * public page and shown in a sandboxed srcdoc iframe (no scripts, no same-origin
 * access), so preview and published output cannot drift and page styles can't
 * leak into the dashboard.
 */
export default function PreviewFrame({ title, slug, content }: Props) {
  const { getAccessToken } = useAuth();
  const [html, setHtml] = useState('');
  const [validation, setValidation] = useState<{ field: string; message: string } | null>(null);
  const [device, setDevice] = useState<Device>('mobile');
  const [box, setBox] = useState(0);
  const containerRef = useRef<HTMLDivElement>(null);
  const seq = useRef(0);

  useEffect(() => {
    const id = ++seq.current;
    const t = setTimeout(async () => {
      try {
        const res = await apiFetch<{ html: string; validation_error: { field: string; message: string } | null }>(
          API_ENDPOINTS.microsites.preview,
          getAccessToken(),
          { method: 'POST', body: JSON.stringify({ title, slug, content }) }
        );
        if (id === seq.current) {
          setHtml(res.html);
          setValidation(res.validation_error);
        }
      } catch {
        // keep the last good preview
      }
    }, 300);
    return () => clearTimeout(t);
  }, [title, slug, content, getAccessToken]);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const ro = new ResizeObserver(() => setBox(el.clientWidth));
    ro.observe(el);
    return () => ro.disconnect();
  }, []);

  const width = WIDTHS[device];
  const scale = box > 0 ? Math.min(1, (box - 16) / width) : 1;
  const height = device === 'mobile' ? 780 : 860;

  return (
    <div className="flex flex-col items-center gap-3 h-full">
      <div className="inline-flex rounded-lg border border-border p-0.5 bg-muted/50" role="group" aria-label="Preview device">
        {(
          [
            ['mobile', DevicePhoneMobileIcon, 'Phone'],
            ['tablet', DeviceTabletIcon, 'Tablet'],
            ['desktop', ComputerDesktopIcon, 'Desktop'],
          ] as const
        ).map(([d, Icon, label]) => (
          <button
            key={d}
            type="button"
            onClick={() => setDevice(d)}
            aria-pressed={device === d}
            aria-label={label}
            className={`p-1.5 rounded-md ${device === d ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
          >
            <Icon className="h-4 w-4" />
          </button>
        ))}
      </div>

      {validation && (
        <p role="status" className="w-full rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/30 px-3 py-2 text-xs text-amber-800 dark:text-amber-200">
          Not saved — {describeValidation(validation.field, validation.message, content.blocks)}
        </p>
      )}

      <div ref={containerRef} className="w-full flex justify-center overflow-hidden">
        <div style={{ width: width * scale, height: height * scale }}>
          <div
            className={`origin-top-left overflow-hidden bg-white ${device === 'mobile' ? 'rounded-[2.25rem] border-[10px] border-gray-900 shadow-2xl' : 'rounded-xl border border-border shadow-lg'}`}
            style={{ width, height, transform: `scale(${scale})` }}
          >
            {html ? (
              <iframe title="Page preview" srcDoc={html} sandbox="" className="w-full h-full border-0" />
            ) : (
              <div className="w-full h-full animate-pulse bg-muted" />
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
