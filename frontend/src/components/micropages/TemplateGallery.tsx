'use client';

import React, { useEffect, useMemo, useRef, useState } from 'react';
import { CheckIcon, EyeIcon, XMarkIcon } from '@heroicons/react/24/outline';
import type { MicroPage, MicroPageTheme } from '@/types/micropage';
import { TEMPLATES } from './templates';
import { MicroPageRenderer } from './MicroPageRenderer';
import { CardRenderer } from './cards/CardRenderer';
import { CARD_COLLECTIONS, CARD_TEMPLATES, isCardTemplateId, paletteSwatch, templatesInCollection } from './cards/catalog';
import { SAMPLE_CARD, SAMPLE_PHOTO } from './cards/cardData';
import type { CardCollectionId, CardData, CardTemplateDefinition } from './cards/types';

type Filter = 'all' | CardCollectionId | 'classic';

interface TemplateGalleryProps {
  /** The selected theme id. */
  value: string;
  /** Called when the user picks a template ("Use Template"). */
  onSelect: (theme: string, paletteId?: string) => void;
  /** Prefix added to card template ids (Microsites uses "card_"). */
  idPrefix?: string;
  /** Show the ten classic MicroPages themes. */
  showClassic?: boolean;
  /** Show the palette chooser in the preview. */
  showPalettes?: boolean;
  /** The user's own details, shown inside the previews. Falls back to a sample person. */
  previewData?: CardData;
  /** Limits the grid height (it scrolls inside). */
  scrollClassName?: string;
}

const SAMPLE_PAGE: Partial<MicroPage> = {
  title: 'Alexandra Reyes',
  bio: 'Building resilient, human-centred businesses across 14 markets.',
  vcard: {
    first_name: 'Alexandra',
    last_name: 'Reyes',
    organization: 'Meridian Group',
    title: 'Chief Executive Officer',
    email: 'alexandra@meridian.example',
    phone: '+1 (415) 555-0142',
    city: 'San Francisco',
  },
  social_links: [
    { platform: 'linkedin', url: 'https://linkedin.com' },
    { platform: 'instagram', url: 'https://instagram.com' },
  ],
  links: [
    { id: 'l1', title: 'Book a meeting', url: 'https://example.com', icon: '📅', is_active: true, sort_order: 0, is_highlighted: true },
    { id: 'l2', title: 'Our latest work', url: 'https://example.com', icon: '✨', is_active: true, sort_order: 1 },
  ],
};

function sampleFor(data?: CardData): CardData {
  const base = data || { ...SAMPLE_CARD, photo: SAMPLE_PHOTO };
  // Each template shows its own default call to action.
  return { ...base, ctaLabel: '', ctaHref: '', photo: base.photo || SAMPLE_PHOTO };
}

export function TemplateGallery({
  value,
  onSelect,
  idPrefix = '',
  showClassic = true,
  showPalettes = true,
  previewData,
  scrollClassName = 'max-h-[58vh]',
}: TemplateGalleryProps) {
  const [filter, setFilter] = useState<Filter>('all');
  const [preview, setPreview] = useState<string | null>(null);
  const data = useMemo(() => sampleFor(previewData), [previewData]);

  const filters: { id: Filter; label: string }[] = [
    { id: 'all', label: 'All' },
    ...CARD_COLLECTIONS.map((c) => ({ id: c.id as Filter, label: c.name })),
    ...(showClassic ? [{ id: 'classic' as Filter, label: 'Classic pages' }] : []),
  ];

  const showCollection = (id: CardCollectionId) => filter === 'all' || filter === id;

  return (
    <div>
      <div role="tablist" aria-label="Template collections" className="flex flex-wrap gap-1.5 mb-3">
        {filters.map((f) => (
          <button
            key={f.id}
            type="button"
            role="tab"
            aria-selected={filter === f.id}
            onClick={() => setFilter(f.id)}
            className={`px-3 py-1.5 rounded-full text-[11px] font-semibold border transition-colors ${
              filter === f.id
                ? 'bg-blue-600 border-blue-500 text-white'
                : 'bg-slate-800/60 border-slate-700 text-slate-300 hover:text-white hover:border-slate-600'
            }`}
          >
            {f.label}
          </button>
        ))}
      </div>

      <div className={`${scrollClassName} overflow-y-auto pr-1 space-y-7`}>
        {CARD_COLLECTIONS.filter((c) => showCollection(c.id)).map((c) => (
          <section key={c.id} aria-labelledby={`coll-${c.id}`}>
            <div className="mb-3">
              <h4 id={`coll-${c.id}`} className="text-sm font-bold text-white">
                {c.name}
              </h4>
              <p className="text-[11px] text-slate-400">{c.description}</p>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              {templatesInCollection(c.id).map((t) => (
                <CardTile
                  key={t.id}
                  tpl={t}
                  data={data}
                  selected={value === idPrefix + t.id}
                  onPreview={() => setPreview(t.id)}
                  onUse={() => onSelect(idPrefix + t.id, t.palettes[0].id)}
                />
              ))}
            </div>
          </section>
        ))}

        {showClassic && (filter === 'all' || filter === 'classic') && (
          <section aria-labelledby="coll-classic">
            <div className="mb-3">
              <h4 id="coll-classic" className="text-sm font-bold text-white">
                Classic pages
              </h4>
              <p className="text-[11px] text-slate-400">Scrolling link-in-bio pages with a Save Contact button.</p>
            </div>
            <div className="grid grid-cols-2 sm:grid-cols-3 gap-3">
              {TEMPLATES.map((t) => (
                <div
                  key={t.id}
                  className={`rounded-2xl border p-3 flex flex-col gap-3 ${
                    value === t.id ? 'border-blue-500 ring-1 ring-blue-500 bg-blue-500/5' : 'border-slate-800 bg-slate-900/60'
                  }`}
                >
                  <button
                    type="button"
                    onClick={() => setPreview(t.id)}
                    aria-label={`Preview ${t.name}`}
                    className={`relative w-full aspect-[4/7] rounded-xl overflow-hidden shadow-lg ${t.previewBg}`}
                  >
                    <span
                      className="absolute left-0 right-0 bottom-0 h-1.5"
                      style={{ backgroundColor: t.accentColor }}
                    />
                    <span className="absolute inset-0 flex flex-col items-center justify-center gap-2 p-3">
                      <span
                        className="w-12 h-12 rounded-full border-2"
                        style={{ borderColor: t.accentColor, backgroundColor: `${t.accentColor}22` }}
                      />
                      <span className="w-3/4 h-2 rounded-full bg-current opacity-30" />
                      <span className="w-1/2 h-2 rounded-full bg-current opacity-20" />
                      <span className="mt-3 w-4/5 h-7 rounded-lg" style={{ backgroundColor: t.accentColor, opacity: 0.85 }} />
                    </span>
                  </button>
                  <TileMeta name={t.name} tag={t.tag} description={t.description} />
                  <TileButtons
                    selected={value === t.id}
                    onPreview={() => setPreview(t.id)}
                    onUse={() => onSelect(t.id)}
                  />
                </div>
              ))}
            </div>
          </section>
        )}
      </div>

      {preview && (
        <PreviewModal
          id={preview}
          data={data}
          selected={value === idPrefix + preview}
          showPalettes={showPalettes}
          onClose={() => setPreview(null)}
          onUse={(paletteId) => {
            onSelect(idPrefix + preview, paletteId);
            setPreview(null);
          }}
        />
      )}
    </div>
  );
}

function TileMeta({ name, tag, description }: { name: string; tag: string; description: string }) {
  return (
    <div className="min-h-[5.2rem]">
      <div className="flex items-center justify-between gap-2">
        <h5 className="text-sm font-bold text-white truncate">{name}</h5>
        <span className="text-[9px] uppercase tracking-wider font-semibold text-slate-400 bg-slate-800 rounded-full px-2 py-0.5 flex-shrink-0">
          {tag}
        </span>
      </div>
      <p className="mt-1 text-[11px] text-slate-400 leading-snug line-clamp-3">{description}</p>
    </div>
  );
}

function TileButtons({ selected, onPreview, onUse }: { selected: boolean; onPreview: () => void; onUse: () => void }) {
  return (
    <div className="grid grid-cols-2 gap-2 mt-auto">
      <button
        type="button"
        onClick={onPreview}
        className="flex items-center justify-center gap-1.5 px-2 py-2 rounded-lg text-[11px] font-semibold bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white"
      >
        <EyeIcon className="w-3.5 h-3.5" />
        Preview
      </button>
      <button
        type="button"
        onClick={onUse}
        aria-pressed={selected}
        className={`flex items-center justify-center gap-1.5 px-2 py-2 rounded-lg text-[11px] font-semibold border ${
          selected
            ? 'bg-emerald-600/20 border-emerald-500/60 text-emerald-300'
            : 'bg-blue-600 hover:bg-blue-500 border-blue-500 text-white'
        }`}
      >
        {selected ? (
          <>
            <CheckIcon className="w-3.5 h-3.5" />
            Selected
          </>
        ) : (
          'Use Template'
        )}
      </button>
    </div>
  );
}

function CardTile({
  tpl,
  data,
  selected,
  onPreview,
  onUse,
}: {
  tpl: CardTemplateDefinition;
  data: CardData;
  selected: boolean;
  onPreview: () => void;
  onUse: () => void;
}) {
  return (
    <div
      className={`rounded-2xl border p-3 flex flex-col gap-3 ${
        selected ? 'border-blue-500 ring-1 ring-blue-500 bg-blue-500/5' : 'border-slate-800 bg-slate-900/60'
      }`}
    >
      <button
        type="button"
        onClick={onPreview}
        aria-label={`Preview ${tpl.name}`}
        className="block w-full rounded-xl overflow-hidden shadow-lg ring-1 ring-black/30 hover:ring-blue-500/60 transition"
      >
        <div className="pointer-events-none" aria-hidden="true">
          <CardRenderer templateId={tpl.id} data={data} interactive={false} />
        </div>
      </button>
      <TileMeta name={tpl.name} tag={tpl.category} description={tpl.description} />
      <TileButtons selected={selected} onPreview={onPreview} onUse={onUse} />
    </div>
  );
}

function PreviewModal({
  id,
  data,
  selected,
  showPalettes,
  onClose,
  onUse,
}: {
  id: string;
  data: CardData;
  selected: boolean;
  showPalettes: boolean;
  onClose: () => void;
  onUse: (paletteId?: string) => void;
}) {
  const tpl = CARD_TEMPLATES.find((t) => t.id === id);
  const classic = TEMPLATES.find((t) => t.id === id);
  const [paletteId, setPaletteId] = useState(tpl?.palettes[0].id);
  const closeRef = useRef<HTMLButtonElement>(null);

  useEffect(() => {
    closeRef.current?.focus();
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    document.addEventListener('keydown', onKey);
    const prev = document.body.style.overflow;
    document.body.style.overflow = 'hidden';
    return () => {
      document.removeEventListener('keydown', onKey);
      document.body.style.overflow = prev;
    };
  }, [onClose]);

  const name = tpl?.name || classic?.name || '';
  return (
    <div
      className="fixed inset-0 z-[60] flex items-center justify-center p-3 sm:p-6 bg-black/80 backdrop-blur-sm"
      onClick={onClose}
    >
      <div
        role="dialog"
        aria-modal="true"
        aria-label={`${name} preview`}
        onClick={(e) => e.stopPropagation()}
        className="relative w-full max-w-3xl max-h-[92vh] overflow-y-auto bg-slate-900 border border-slate-800 rounded-3xl shadow-2xl p-5 sm:p-7"
      >
        <button
          ref={closeRef}
          type="button"
          onClick={onClose}
          aria-label="Close preview"
          className="absolute top-4 right-4 p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800"
        >
          <XMarkIcon className="w-5 h-5" />
        </button>

        <div className="grid md:grid-cols-[minmax(0,340px)_1fr] gap-6 items-start">
          <div className="w-full max-w-[340px] mx-auto rounded-2xl overflow-hidden shadow-[0_20px_50px_-12px_rgba(0,0,0,0.7)]">
            {tpl ? (
              <CardRenderer templateId={tpl.id} data={data} paletteId={paletteId} interactive={false} />
            ) : (
              <div className="h-[560px] overflow-y-auto bg-slate-950">
                <MicroPageRenderer page={{ ...SAMPLE_PAGE, theme: id as MicroPageTheme }} isPreview />
              </div>
            )}
          </div>

          <div className="space-y-4">
            <div>
              <p className="text-[10px] uppercase tracking-wider font-semibold text-blue-400">
                {tpl ? tpl.category : 'Classic page'}
              </p>
              <h3 className="text-2xl font-bold text-white mt-0.5">{name}</h3>
              <p className="mt-2 text-sm text-slate-300 leading-relaxed">{tpl?.description || classic?.description}</p>
            </div>

            {tpl && (
              <>
                <p className="text-xs text-slate-400">
                  <span className="font-semibold text-slate-300">Best for:</span> {tpl.audience}
                </p>
                {showPalettes && (
                <div>
                  <p className="text-[10px] uppercase tracking-wider font-semibold text-slate-400 mb-2">Colour palette</p>
                  <div className="flex flex-wrap gap-2">
                    {tpl.palettes.map((p) => {
                      const sw = paletteSwatch(tpl, p);
                      const on = paletteId === p.id;
                      return (
                        <button
                          key={p.id}
                          type="button"
                          onClick={() => setPaletteId(p.id)}
                          aria-pressed={on}
                          className={`flex items-center gap-2 pl-1.5 pr-3 py-1.5 rounded-full border text-[11px] font-semibold ${
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
                </div>
                )}
                <p className="text-[11px] text-slate-500">
                  This preview uses {data.name === SAMPLE_CARD.name ? 'a sample person' : 'your details'}. The QR code points to
                  your card once it is created.
                </p>
              </>
            )}

            <div className="flex gap-2.5 pt-2">
              <button
                type="button"
                onClick={() => onUse(paletteId)}
                className="flex-1 px-4 py-3 rounded-xl text-sm font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/25"
              >
                {selected ? 'Keep this template' : 'Use Template'}
              </button>
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-3 rounded-xl text-sm font-semibold bg-slate-800 hover:bg-slate-700 border border-slate-700 text-white"
              >
                Close
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

export { isCardTemplateId };
export default TemplateGallery;
