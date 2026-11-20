'use client';

import React, { useEffect, useMemo, useRef, useState } from 'react';
import type { MicroPage, MicroLinkItem } from '@/types/micropage';
import { API_ENDPOINTS } from '@/lib/api-config';
import { CardRenderer } from './CardRenderer';
import { CardActions } from './CardActions';
import { getCardTemplate } from './catalog';
import { readCardConfig, toCardData } from './cardData';
import type { CardTemplateId } from './types';

interface CardPageViewProps {
  page: Partial<MicroPage>;
  isPreview?: boolean;
  onLinkClick?: (linkId: string, url: string) => void;
  onSaveContact?: () => void;
  className?: string;
}

/**
 * The page a visitor sees for a digital visiting card: the card artwork at its own proportions,
 * the actions (save contact, download PNG/JPG, download QR) and any extra links underneath.
 */
export function CardPageView({ page, isPreview = false, onLinkClick, onSaveContact, className = '' }: CardPageViewProps) {
  const templateId = page.theme as CardTemplateId;
  const tpl = getCardTemplate(templateId);
  const cardRef = useRef<HTMLElement>(null);

  // The QR code points at the page's own address; only known in the browser.
  const [origin, setOrigin] = useState('');
  useEffect(() => setOrigin(window.location.origin), []);
  const pageUrl = `${origin}/m/${page.slug || ''}`;

  const cfg = readCardConfig(page);
  const data = useMemo(
    () => (tpl ? toCardData(page, tpl, { pageUrl }) : null),
    [page, tpl, pageUrl]
  );
  if (!tpl || !data) return null;

  const links: MicroLinkItem[] = (page.links || []).filter((l) => l.is_active !== false && l.url);

  return (
    <div className={`w-full min-h-full flex flex-col items-center px-4 py-6 gap-5 ${className}`}>
      <div className="w-full max-w-[440px] rounded-[18px] overflow-hidden shadow-[0_24px_60px_-14px_rgba(0,0,0,0.6)]">
        <CardRenderer
          ref={cardRef}
          templateId={templateId}
          data={data}
          paletteId={cfg.palette}
          accent={cfg.accent}
          logoPlate={cfg.logo_plate}
          interactive={!isPreview}
          ctaHref={page.slug ? API_ENDPOINTS.micropages.vcard(page.slug) : undefined}
          onCtaClick={(e) => {
            // With no custom CTA link, the button saves the contact.
            if (!data.ctaHref && onSaveContact) {
              e.preventDefault();
              onSaveContact();
            }
          }}
        />
      </div>

      {!isPreview && (
        <div className="w-full max-w-[440px]">
          <CardActions
            getCard={() => cardRef.current}
            slug={page.slug || 'card'}
            qrValue={data.qrValue}
            pageUrl={pageUrl}
            onSaveContact={onSaveContact}
          />
        </div>
      )}

      {links.length > 0 && (
        <ul className="w-full max-w-[440px] space-y-2.5">
          {links.map((l) => (
            <li key={l.id}>
              <a
                href={isPreview ? '#' : l.url}
                target={isPreview ? '_self' : '_blank'}
                rel="noopener noreferrer"
                onClick={(e) => {
                  if (isPreview) e.preventDefault();
                  onLinkClick?.(l.id, l.url);
                }}
                className="flex items-center justify-between gap-3 w-full px-4 py-3.5 rounded-2xl bg-white/10 hover:bg-white/15 border border-white/15 text-white text-sm font-semibold transition-all"
              >
                <span className="flex items-center gap-3 min-w-0">
                  {l.icon && <span className="text-lg flex-shrink-0">{l.icon}</span>}
                  <span className="truncate">{l.title}</span>
                </span>
                <span aria-hidden="true" className="opacity-60">›</span>
              </a>
            </li>
          ))}
        </ul>
      )}

      {!isPreview && (
        <a href="/" target="_blank" rel="noopener noreferrer" className="text-[11px] text-white/50 hover:text-white/80">
          Created with <span className="font-semibold">Trunc vCard</span>
        </a>
      )}
    </div>
  );
}

export default CardPageView;
