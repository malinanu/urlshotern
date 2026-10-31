'use client';

import { useEffect, useId, useRef, useState } from 'react';
import { ArrowUpTrayIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { apiFetch, assetSrc, isApiError } from '@/lib/microsites';

const inputClass =
  'w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40';

export function TextField({
  label,
  value,
  onChange,
  placeholder,
  maxLength,
  type = 'text',
  hint,
}: {
  label: string;
  value?: string;
  onChange: (v: string) => void;
  placeholder?: string;
  maxLength?: number;
  type?: string;
  hint?: string;
}) {
  const id = useId();
  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <input
        id={id}
        type={type}
        value={value ?? ''}
        placeholder={placeholder}
        maxLength={maxLength}
        onChange={(e) => onChange(e.target.value)}
        className={inputClass}
      />
      {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

export function TextAreaField({
  label,
  value,
  onChange,
  maxLength,
  rows = 3,
}: {
  label: string;
  value?: string;
  onChange: (v: string) => void;
  maxLength?: number;
  rows?: number;
}) {
  const id = useId();
  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <textarea
        id={id}
        value={value ?? ''}
        rows={rows}
        maxLength={maxLength}
        onChange={(e) => onChange(e.target.value)}
        className={inputClass}
      />
      {maxLength && (
        <p className="mt-1 text-right text-xs text-muted-foreground">
          {(value ?? '').length}/{maxLength}
        </p>
      )}
    </div>
  );
}

export function SelectField({
  label,
  value,
  onChange,
  options,
}: {
  label: string;
  value?: string;
  onChange: (v: string) => void;
  options: { value: string; label: string }[];
}) {
  const id = useId();
  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <select id={id} value={value ?? ''} onChange={(e) => onChange(e.target.value)} className={inputClass}>
        {options.map((o) => (
          <option key={o.value} value={o.value}>
            {o.label}
          </option>
        ))}
      </select>
    </div>
  );
}

export function ToggleField({
  label,
  checked,
  onChange,
  disabled,
  hint,
}: {
  label: string;
  checked?: boolean;
  onChange: (v: boolean) => void;
  disabled?: boolean;
  hint?: string;
}) {
  const id = useId();
  return (
    <div className="flex items-start gap-3">
      <input
        id={id}
        type="checkbox"
        checked={!!checked}
        disabled={disabled}
        onChange={(e) => onChange(e.target.checked)}
        className="mt-0.5 h-4 w-4 rounded border-border"
      />
      <label htmlFor={id} className={`text-sm ${disabled ? 'text-muted-foreground' : 'text-foreground'}`}>
        {label}
        {hint && <span className="block text-xs text-muted-foreground">{hint}</span>}
      </label>
    </div>
  );
}

let zoneCache: string[] | null = null;
function timeZones(): string[] {
  if (zoneCache) return zoneCache;
  try {
    const intl = Intl as unknown as { supportedValuesOf?: (key: string) => string[] };
    zoneCache = intl.supportedValuesOf ? intl.supportedValuesOf('timeZone') : [];
  } catch {
    zoneCache = [];
  }
  if (!zoneCache.includes('UTC')) zoneCache = ['UTC', ...zoneCache];
  return zoneCache;
}

/** IANA time zone picker (free text with suggestions). */
export function TimeZoneField({ label, value, onChange }: { label: string; value?: string; onChange: (v: string) => void }) {
  const id = useId();
  const listId = `${id}-zones`;
  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <input
        id={id}
        list={listId}
        value={value ?? ''}
        onChange={(e) => onChange(e.target.value)}
        placeholder="Europe/London"
        className={inputClass}
        autoComplete="off"
        spellCheck={false}
      />
      <datalist id={listId}>
        {timeZones().map((z) => (
          <option key={z} value={z} />
        ))}
      </datalist>
    </div>
  );
}

export function ColorField({ label, value, onChange }: { label: string; value?: string; onChange: (v: string) => void }) {
  const id = useId();
  const [text, setText] = useState(value ?? '#000000');
  useEffect(() => {
    setText(value ?? '#000000');
  }, [value]);
  const shown = /^#[0-9a-fA-F]{6}$/.test(text) ? text : value ?? '#000000';
  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <div className="flex items-center gap-2">
        <input
          type="color"
          aria-label={`${label} picker`}
          value={shown}
          onChange={(e) => {
            setText(e.target.value);
            onChange(e.target.value);
          }}
          className="h-9 w-10 rounded border border-border bg-background p-0.5 cursor-pointer"
        />
        <input
          id={id}
          value={text}
          maxLength={7}
          onChange={(e) => {
            setText(e.target.value);
            if (/^#[0-9a-fA-F]{6}$/.test(e.target.value)) onChange(e.target.value.toLowerCase());
          }}
          onBlur={() => setText(value ?? '#000000')}
          className={`${inputClass} font-mono`}
        />
      </div>
    </div>
  );
}

/** Image input: paste an https URL or upload a JPEG/PNG to the page's asset store. */
export function ImageField({
  label,
  value,
  onChange,
  micrositeId,
}: {
  label: string;
  value?: string;
  onChange: (v: string) => void;
  micrositeId: number;
}) {
  const id = useId();
  const { getAccessToken } = useAuth();
  const fileRef = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [error, setError] = useState('');

  const upload = async (file: File) => {
    setError('');
    setUploading(true);
    try {
      const form = new FormData();
      form.append('file', file);
      const asset = await apiFetch<{ url: string }>(API_ENDPOINTS.microsites.upload(micrositeId), getAccessToken(), {
        method: 'POST',
        body: form,
      });
      onChange(asset.url);
    } catch (e) {
      setError(isApiError(e) ? e.message : 'Upload failed');
    } finally {
      setUploading(false);
      if (fileRef.current) fileRef.current.value = '';
    }
  };

  return (
    <div>
      <label htmlFor={id} className="block text-xs font-medium text-muted-foreground mb-1">
        {label}
      </label>
      <div className="flex items-center gap-2">
        {value ? (
          // eslint-disable-next-line @next/next/no-img-element
          <img src={assetSrc(value)} alt="" className="h-9 w-9 rounded object-cover border border-border shrink-0" />
        ) : null}
        <input
          id={id}
          value={value ?? ''}
          placeholder="https://… or upload"
          onChange={(e) => onChange(e.target.value)}
          className={inputClass}
        />
        <button
          type="button"
          onClick={() => fileRef.current?.click()}
          disabled={uploading}
          className="shrink-0 inline-flex items-center gap-1 rounded-lg border border-border px-2.5 py-2 text-xs font-medium text-foreground hover:bg-muted disabled:opacity-50"
        >
          <ArrowUpTrayIcon className="h-4 w-4" />
          {uploading ? 'Uploading…' : 'Upload'}
        </button>
        <input
          ref={fileRef}
          type="file"
          accept="image/png,image/jpeg"
          className="hidden"
          onChange={(e) => e.target.files?.[0] && upload(e.target.files[0])}
        />
      </div>
      {error && <p role="alert" className="mt-1 text-xs text-red-600">{error}</p>}
      {value && (
        <button type="button" onClick={() => onChange('')} className="mt-1 text-xs text-muted-foreground hover:text-foreground">
          Remove image
        </button>
      )}
    </div>
  );
}
