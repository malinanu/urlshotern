'use client';

import Link from 'next/link';
import { LockClosedIcon } from '@heroicons/react/24/outline';
import { PLAN_NAMES } from '@/lib/plan-utils';
import {
  FONTS,
  THEME_PRESETS,
  type MicrositeBackground,
  type MicrositeEntitlements,
  type MicrositeSettings,
  type MicrositeTheme,
} from '@/lib/microsites';
import { ColorField, ImageField, SelectField, TextAreaField, TextField, ToggleField } from './fields';

interface Props {
  theme: MicrositeTheme;
  settings: MicrositeSettings;
  entitlements: MicrositeEntitlements | null;
  micrositeId: number;
  onTheme: (theme: MicrositeTheme) => void;
  onSettings: (settings: MicrositeSettings) => void;
}

function Segmented<T extends string>({
  label,
  value,
  options,
  onChange,
}: {
  label: string;
  value: T;
  options: { value: T; label: string }[];
  onChange: (v: T) => void;
}) {
  return (
    <fieldset>
      <legend className="block text-xs font-medium text-muted-foreground mb-1">{label}</legend>
      <div className="inline-flex rounded-lg border border-border p-0.5 bg-muted/50">
        {options.map((o) => (
          <button
            key={o.value}
            type="button"
            onClick={() => onChange(o.value)}
            aria-pressed={value === o.value}
            className={`px-3 py-1 text-xs font-medium rounded-md ${
              value === o.value ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'
            }`}
          >
            {o.label}
          </button>
        ))}
      </div>
    </fieldset>
  );
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="grid gap-3">
      <h3 className="text-sm font-semibold text-foreground">{title}</h3>
      {children}
    </section>
  );
}

function presetSwatch(bg: MicrositeBackground): string {
  if (bg.type === 'solid') return bg.color ?? '#000';
  const stops = (bg.stops ?? []).join(', ');
  return bg.type === 'radial' ? `radial-gradient(circle at 50% 0%, ${stops})` : `linear-gradient(${bg.angle ?? 135}deg, ${stops})`;
}

export default function DesignPanel({ theme, settings, entitlements, micrositeId, onTheme, onSettings }: Props) {
  const bg = theme.background;
  const setBg = (patch: Partial<MicrositeBackground>) => onTheme({ ...theme, background: { ...bg, ...patch } });
  const setTheme = (patch: Partial<MicrositeTheme>) => onTheme({ ...theme, ...patch });
  const setSettings = (patch: Partial<MicrositeSettings>) => onSettings({ ...settings, ...patch });

  const changeBgType = (type: MicrositeBackground['type']) => {
    const first = bg.stops?.[0] ?? bg.color ?? '#1e3c72';
    if (type === 'solid') setBg({ type, color: first });
    else if (type === 'image') setBg({ type, color: bg.color ?? '#111827', overlay: bg.overlay ?? 30 });
    else setBg({ type, stops: bg.stops?.length ? bg.stops : [first, '#2a5298'], angle: bg.angle ?? 135 });
  };

  return (
    <div className="grid gap-6">
      <Section title="Presets">
        <div className="grid grid-cols-5 gap-2">
          {THEME_PRESETS.map((p) => (
            <button
              key={p.name}
              type="button"
              onClick={() => onTheme({ ...theme, ...p.theme, background: { ...(p.theme.background as MicrositeBackground) } })}
              className="group grid gap-1 text-center"
              title={p.name}
            >
              <span
                className="h-12 rounded-lg border border-border group-hover:ring-2 group-hover:ring-primary/40"
                style={{ background: presetSwatch(p.theme.background as MicrositeBackground) }}
              />
              <span className="text-[11px] text-muted-foreground">{p.name}</span>
            </button>
          ))}
        </div>
      </Section>

      <Section title="Background">
        <Segmented
          label="Style"
          value={bg.type}
          onChange={changeBgType}
          options={[
            { value: 'solid', label: 'Flat' },
            { value: 'linear', label: 'Linear' },
            { value: 'radial', label: 'Radial' },
            { value: 'image', label: 'Image' },
          ]}
        />
        {bg.type === 'solid' && <ColorField label="Color" value={bg.color} onChange={(v) => setBg({ color: v })} />}
        {(bg.type === 'linear' || bg.type === 'radial') && (
          <>
            <div className="grid grid-cols-2 gap-3">
              {(bg.stops ?? []).map((s, i) => (
                <ColorField
                  key={i}
                  label={`Color ${i + 1}`}
                  value={s}
                  onChange={(v) => setBg({ stops: (bg.stops ?? []).map((x, k) => (k === i ? v : x)) })}
                />
              ))}
            </div>
            <div className="flex gap-2">
              {(bg.stops ?? []).length < 3 ? (
                <button type="button" className="text-xs text-primary hover:underline" onClick={() => setBg({ stops: [...(bg.stops ?? []), '#ffffff'] })}>
                  Add a third color
                </button>
              ) : (
                <button type="button" className="text-xs text-muted-foreground hover:underline" onClick={() => setBg({ stops: (bg.stops ?? []).slice(0, 2) })}>
                  Use two colors
                </button>
              )}
            </div>
            {bg.type === 'linear' && (
              <div>
                <label htmlFor="bg-angle" className="block text-xs font-medium text-muted-foreground mb-1">
                  Angle: {bg.angle ?? 0}°
                </label>
                <input
                  id="bg-angle"
                  type="range"
                  min={0}
                  max={360}
                  value={bg.angle ?? 0}
                  onChange={(e) => setBg({ angle: Number(e.target.value) })}
                  className="w-full"
                />
              </div>
            )}
          </>
        )}
        {bg.type === 'image' && (
          <>
            <ImageField label="Background image" value={bg.image_url} onChange={(v) => setBg({ image_url: v })} micrositeId={micrositeId} />
            <div>
              <label htmlFor="bg-overlay" className="block text-xs font-medium text-muted-foreground mb-1">
                Darken: {bg.overlay ?? 0}%
              </label>
              <input
                id="bg-overlay"
                type="range"
                min={0}
                max={80}
                value={bg.overlay ?? 0}
                onChange={(e) => setBg({ overlay: Number(e.target.value) })}
                className="w-full"
              />
            </div>
            <ColorField label="Fallback color" value={bg.color} onChange={(v) => setBg({ color: v })} />
          </>
        )}
      </Section>

      <Section title="Colors">
        <div className="grid grid-cols-2 gap-3">
          <ColorField label="Card" value={theme.card_fill} onChange={(v) => setTheme({ card_fill: v })} />
          <ColorField label="Text" value={theme.text} onChange={(v) => setTheme({ text: v })} />
          <ColorField label="Accent" value={theme.accent} onChange={(v) => setTheme({ accent: v })} />
          <ColorField label="Button text" value={theme.button_text} onChange={(v) => setTheme({ button_text: v })} />
        </div>
      </Section>

      <Section title="Typography & shapes">
        <SelectField label="Font" value={theme.font} onChange={(v) => setTheme({ font: v })} options={FONTS.map((f) => ({ value: f.id, label: f.name }))} />
        <Segmented
          label="Button corners"
          value={theme.button_radius}
          onChange={(v) => setTheme({ button_radius: v })}
          options={[
            { value: 'sharp', label: 'Sharp' },
            { value: 'rounded', label: 'Rounded' },
            { value: 'pill', label: 'Pill' },
          ]}
        />
        <Segmented
          label="Card shadow"
          value={theme.shadow}
          onChange={(v) => setTheme({ shadow: v })}
          options={[
            { value: 'none', label: 'None' },
            { value: 'soft', label: 'Soft' },
            { value: 'strong', label: 'Strong' },
          ]}
        />
        <Segmented
          label="Avatar border"
          value={theme.avatar_border}
          onChange={(v) => setTheme({ avatar_border: v })}
          options={[
            { value: 'none', label: 'None' },
            { value: 'thin', label: 'Thin' },
            { value: 'thick', label: 'Thick' },
          ]}
        />
      </Section>

      <Section title="Sharing & search">
        <TextField label="Page title (browser tab & link previews)" value={settings.seo_title} onChange={(v) => setSettings({ seo_title: v })} maxLength={70} />
        <TextAreaField label="Description" value={settings.seo_description} onChange={(v) => setSettings({ seo_description: v })} maxLength={200} rows={2} />
        <ImageField label="Link preview image" value={settings.og_image_url} onChange={(v) => setSettings({ og_image_url: v })} micrositeId={micrositeId} />
        <ImageField label="Favicon" value={settings.favicon_url} onChange={(v) => setSettings({ favicon_url: v })} micrositeId={micrositeId} />
        <ToggleField
          label="Allow search engines to index this page"
          checked={settings.allow_indexing}
          onChange={(v) => setSettings({ allow_indexing: v })}
        />
      </Section>

      <Section title="Branding">
        <ToggleField
          label="Hide the “Made with Trunc” badge"
          checked={settings.hide_badge}
          disabled={!entitlements?.can_remove_badge}
          onChange={(v) => setSettings({ hide_badge: v })}
          hint={entitlements?.can_remove_badge ? undefined : `Available on the ${PLAN_NAMES.enterprise} plan.`}
        />
        {!entitlements?.can_remove_badge && (
          <Link href="/billing" className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:underline">
            <LockClosedIcon className="h-3.5 w-3.5" />
            Upgrade to remove branding
          </Link>
        )}
      </Section>
    </div>
  );
}
