'use client';

import React, { useRef, useState } from 'react';
import { ArrowUpTrayIcon, TrashIcon } from '@heroicons/react/24/outline';
import type { MicroPage } from '@/types/micropage';
import { CardActions } from './CardActions';
import { getCardTemplate, isHexColor, paletteAccent, paletteSwatch, suggestLogoPlate } from './catalog';
import { logoTone } from './logoTone';
import { imageSrc, readCardConfig, toCardData } from './cardData';
import { uploadCardImage } from './upload';
import type { CardConfig } from './types';

interface CardSettingsPanelProps {
  page: MicroPage;
  onChange: (page: MicroPage) => void;
  /** Returns the live card in the preview, for the download buttons. */
  getCard: () => HTMLElement | null;
}

const input =
  'w-full bg-slate-800/80 border border-slate-700 focus:border-blue-500 focus:outline-none rounded-xl px-4 py-2.5 text-white text-sm placeholder:text-slate-500';
const label = 'block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1.5';
const hint = 'mt-1 text-[11px] text-slate-500';

/** Everything on the card that is not part of the vCard tab: logo, description, CTA, QR, colours, export. */
export function CardSettingsPanel({ page, onChange, getCard }: CardSettingsPanelProps) {
  const tpl = getCardTemplate(page.theme);
  const cfg = readCardConfig(page);
  const logoInput = useRef<HTMLInputElement>(null);
  const [uploading, setUploading] = useState(false);
  const [uploadError, setUploadError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [tagsText, setTagsText] = useState((cfg.tags || []).join(', '));
  if (!tpl) return null;

  const setCard = (patch: Partial<CardConfig>) =>
    onChange({ ...page, config: { ...(page.config || {}), card: { ...cfg, ...patch } } });

  const origin = typeof window !== 'undefined' ? window.location.origin : '';
  const pageUrl = `${origin}/m/${page.slug}`;
  const accent = isHexColor(cfg.accent) ? cfg.accent : paletteAccent(tpl, cfg.palette);
  const data = toCardData(page, tpl, { pageUrl });

  const onLogoFile = async (file?: File) => {
    if (!file) return;
    setUploading(true);
    setUploadError(null);
    try {
      const [url, tone] = await Promise.all([uploadCardImage(file), logoTone(file).catch(() => 'mixed' as const)]);
      // A logo that would disappear into the card gets a plate in the opposite tone.
      setCard({ logo_url: url, logo_plate: suggestLogoPlate(tpl.id, tone) });
    } catch (e) {
      setUploadError((e as Error).message);
    } finally {
      setUploading(false);
      if (logoInput.current) logoInput.current.value = '';
    }
  };

  return (
    <div className="bg-slate-900/70 border border-slate-800 rounded-2xl p-6 space-y-7">
      <div>
        <h3 className="text-base font-bold text-white">Digital card: {tpl.name}</h3>
        <p className="text-xs text-slate-400 mt-0.5">
          Branding, call to action, QR code and colours. Your name, title, phone, email and address are on the vCard tab; your
          photo is on the Profile tab.
        </p>
      </div>

      {/* Company logo */}
      <div>
        <span className={label}>Company logo</span>
        <div className="flex items-center gap-3">
          <div className="w-20 h-12 rounded-lg bg-white/90 border border-slate-700 flex items-center justify-center overflow-hidden flex-shrink-0">
            {cfg.logo_url ? (
              // eslint-disable-next-line @next/next/no-img-element
              <img src={imageSrc(cfg.logo_url)} alt="Company logo" className="max-w-full max-h-full object-contain" />
            ) : (
              <span className="text-[10px] text-slate-500">No logo</span>
            )}
          </div>
          <input
            ref={logoInput}
            type="file"
            accept="image/png,image/jpeg"
            className="sr-only"
            id="card-logo-file"
            onChange={(e) => onLogoFile(e.target.files?.[0])}
          />
          <label
            htmlFor="card-logo-file"
            className="inline-flex items-center gap-2 px-3.5 py-2.5 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white cursor-pointer"
          >
            <ArrowUpTrayIcon className="w-4 h-4" />
            {uploading ? 'Uploading…' : cfg.logo_url ? 'Replace logo' : 'Upload logo'}
          </label>
          {cfg.logo_url && (
            <button
              type="button"
              onClick={() => setCard({ logo_url: '' })}
              aria-label="Remove logo"
              className="p-2 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-800"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          )}
        </div>
        <input
          type="text"
          value={cfg.logo_url || ''}
          onChange={(e) => setCard({ logo_url: e.target.value })}
          placeholder="or paste an image link (https://…)"
          className={`${input} mt-2`}
        />
        {uploadError && (
          <p role="alert" className="mt-1.5 text-[11px] text-rose-400">
            {uploadError}
          </p>
        )}
        <p className={hint}>JPEG or PNG, up to 2 MB. Uploaded images are re-encoded and always export correctly.</p>
        <div className="mt-3 flex items-center gap-3">
          <label htmlFor="card-logo-plate" className="text-xs text-slate-300">
            Logo background
          </label>
          <select
            id="card-logo-plate"
            value={cfg.logo_plate || 'auto'}
            onChange={(e) => setCard({ logo_plate: e.target.value as CardConfig['logo_plate'] })}
            className="bg-slate-800 border border-slate-700 rounded-lg px-3 py-1.5 text-xs text-white"
          >
            <option value="auto">Automatic</option>
            <option value="light">Light plate (for dark logos)</option>
            <option value="dark">Dark plate (for white logos)</option>
            <option value="none">None</option>
          </select>
        </div>
      </div>

      {/* Company description */}
      <div>
        <label htmlFor="card-about" className={label}>
          Company description
        </label>
        <textarea
          id="card-about"
          rows={3}
          maxLength={220}
          value={cfg.about ?? page.bio ?? ''}
          onChange={(e) => setCard({ about: e.target.value })}
          placeholder="One or two lines about you or your company"
          className={`${input} resize-none`}
        />
        <p className={hint}>{(cfg.about ?? page.bio ?? '').length}/220. Long text is trimmed to fit the card.</p>
      </div>

      <div className="grid sm:grid-cols-2 gap-4">
        <div>
          <label htmlFor="card-wa" className={label}>
            WhatsApp number
          </label>
          <input
            id="card-wa"
            type="tel"
            value={cfg.whatsapp || ''}
            onChange={(e) => setCard({ whatsapp: e.target.value })}
            placeholder="+1 415 555 0142"
            className={input}
          />
          <p className={hint}>Include the country code.</p>
        </div>
        <div>
          <label htmlFor="card-tags" className={label}>
            Specialisation labels
          </label>
          <input
            id="card-tags"
            type="text"
            value={tagsText}
            onChange={(e) => {
              setTagsText(e.target.value);
              setCard({
                tags: e.target.value
                  .split(',')
                  .map((t) => t.trim())
                  .filter(Boolean)
                  .slice(0, 4),
              });
            }}
            placeholder="Branding, Strategy, Identity"
            className={input}
          />
          <p className={hint}>Up to 4, separated by commas. Shown on templates that have a label row.</p>
        </div>
      </div>

      {/* CTA + QR */}
      <div className="grid sm:grid-cols-2 gap-4">
        <div>
          <label htmlFor="card-cta" className={label}>
            Button text
          </label>
          <input
            id="card-cta"
            type="text"
            maxLength={28}
            value={cfg.cta_text || ''}
            onChange={(e) => setCard({ cta_text: e.target.value })}
            placeholder={tpl.cta}
            className={input}
          />
        </div>
        <div>
          <label htmlFor="card-cta-url" className={label}>
            Button link
          </label>
          <input
            id="card-cta-url"
            type="url"
            value={cfg.cta_url || ''}
            onChange={(e) => setCard({ cta_url: e.target.value })}
            placeholder="Leave empty to save your contact"
            className={input}
          />
        </div>
        <div className="sm:col-span-2">
          <label htmlFor="card-qr" className={label}>
            QR code destination
          </label>
          <input
            id="card-qr"
            type="url"
            value={cfg.qr_url || ''}
            onChange={(e) => setCard({ qr_url: e.target.value })}
            placeholder={pageUrl}
            className={input}
          />
          <p className={hint}>Defaults to this card&apos;s own address. The QR code is generated from it live.</p>
        </div>
      </div>

      {/* Colours */}
      <div>
        <span className={label}>Colours</span>
        <div className="flex flex-wrap gap-2">
          {tpl.palettes.map((p) => {
            const sw = paletteSwatch(tpl, p);
            const on = (cfg.palette || tpl.palettes[0].id) === p.id;
            return (
              <button
                key={p.id}
                type="button"
                aria-pressed={on}
                onClick={() => setCard({ palette: p.id, accent: undefined })}
                className={`flex items-center gap-2 pl-1.5 pr-3 py-1.5 rounded-full border text-xs font-semibold ${
                  on ? 'border-blue-500 bg-blue-500/10 text-white' : 'border-slate-700 text-slate-300 hover:border-slate-600'
                }`}
              >
                <span
                  className="w-5 h-5 rounded-full border border-white/20"
                  style={{ background: `linear-gradient(135deg, ${sw.bg} 50%, ${sw.accent} 50%)` }}
                />
                {p.name}
              </button>
            );
          })}
        </div>
        <div className="mt-3 flex items-center gap-3">
          <label htmlFor="card-accent" className="text-xs text-slate-300">
            Accent colour
          </label>
          <input
            id="card-accent"
            type="color"
            value={accent}
            onChange={(e) => setCard({ accent: e.target.value })}
            className="w-10 h-8 rounded border border-slate-700 bg-transparent cursor-pointer"
          />
          <span className="text-[11px] font-mono text-slate-400">{accent}</span>
          {isHexColor(cfg.accent) && (
            <button
              type="button"
              onClick={() => setCard({ accent: undefined })}
              className="text-[11px] text-blue-400 hover:text-blue-300"
            >
              Reset
            </button>
          )}
        </div>
        <p className={hint}>
          The layout and typography are part of the template and stay as designed. Colours are limited to the palettes and one
          accent so the card stays readable.
        </p>
      </div>

      {/* Share + export */}
      <div>
        <span className={label}>Share and download</span>
        <div className="flex items-center gap-2 mb-3">
          <input readOnly value={pageUrl} aria-label="Card address" className={`${input} font-mono text-xs`} />
          <button
            type="button"
            onClick={() => {
              navigator.clipboard?.writeText(pageUrl);
              setCopied(true);
              setTimeout(() => setCopied(false), 2000);
            }}
            className="px-3.5 py-2.5 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white flex-shrink-0"
          >
            {copied ? 'Copied' : 'Copy link'}
          </button>
        </div>
        <CardActions getCard={getCard} slug={page.slug || 'card'} qrValue={data.qrValue} />
        <p className={hint}>Save your changes first so the downloads match what visitors see.</p>
      </div>
    </div>
  );
}

export default CardSettingsPanel;
