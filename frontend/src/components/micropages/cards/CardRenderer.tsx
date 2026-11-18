'use client';

import React, { forwardRef, useMemo } from 'react';
import './css/index.css';
import { ContactIcon, SocialGlyph, type ContactKind } from './icons';
import { qrShape } from './qr';
import { cardStyleVars, getCardTemplate } from './catalog';
import { initials, mapsHref, nameLengthClass, safeHttpUrl, telHref, websiteHref, whatsappHref } from './cardData';
import type { CardData, CardTemplateId } from './types';

export interface CardRendererProps {
  templateId: CardTemplateId;
  data: CardData;
  /** Palette preset id from the catalog. */
  paletteId?: string;
  /** Custom accent colour (#rrggbb). Changes the accent only, never the layout. */
  accent?: string;
  /** Background behind the logo. 'auto' (or unset) uses the template's default. */
  logoPlate?: 'auto' | 'light' | 'dark' | 'none';
  /** When false (previews, thumbnails) links are inert. */
  interactive?: boolean;
  /** Used when the card has no custom CTA link: the page supplies "Save Contact". */
  ctaHref?: string;
  onCtaClick?: (e: React.MouseEvent<HTMLAnchorElement>) => void;
  className?: string;
}

interface Row {
  kind: ContactKind;
  label: string;
  value: string;
  href: string;
}

function buildRows(d: CardData): Row[] {
  const rows: Row[] = [];
  if (d.phone) rows.push({ kind: 'phone', label: 'Phone', value: d.phone, href: telHref(d.phone) });
  if (d.whatsapp) rows.push({ kind: 'whatsapp', label: 'WhatsApp', value: d.whatsapp, href: whatsappHref(d.whatsapp) });
  if (d.email) rows.push({ kind: 'email', label: 'Email', value: d.email, href: `mailto:${d.email}` });
  if (d.website) rows.push({ kind: 'website', label: 'Web', value: d.website, href: websiteHref(d.website) });
  if (d.address) rows.push({ kind: 'address', label: 'Address', value: d.address, href: mapsHref(d.address) });
  return rows;
}

/**
 * One renderer for all twelve card templates. The markup is the same for each; the template's
 * stylesheet (css/<id>.css) decides how it looks. The Go renderer for Microsites emits this
 * same markup, so both engines share the stylesheets.
 */
export const CardRenderer = forwardRef<HTMLElement, CardRendererProps>(function CardRenderer(
  { templateId, data, paletteId, accent, logoPlate, interactive = true, ctaHref, onCtaClick, className = '' },
  ref
) {
  const tpl = getCardTemplate(templateId);
  const qr = useMemo(() => qrShape(data.qrValue), [data.qrValue]);
  const rows = useMemo(() => buildRows(data), [data]);
  const styleVars = useMemo(() => (tpl ? cardStyleVars(tpl, paletteId, accent) : {}), [tpl, paletteId, accent]);
  if (!tpl) return null;

  const inert = (e: React.MouseEvent) => {
    if (!interactive) e.preventDefault();
  };
  const link = (href: string) => (interactive && href ? href : '#');
  const ext = interactive ? { target: '_blank' as const, rel: 'noopener noreferrer' } : {};
  const cta = data.ctaLabel || tpl.cta;
  const ctaTarget = data.ctaHref || ctaHref || '#';

  return (
    <article
      ref={ref}
      className={`vc vc-${templateId} ${className}`.trim()}
      data-vc={templateId}
      data-nl={nameLengthClass(data.name)}
      data-logo={!logoPlate || logoPlate === 'auto' ? tpl.logo : logoPlate}
      style={styleVars as React.CSSProperties}
    >
      <div className="vc-in">
        <i className="vc-d vc-d1" aria-hidden="true" />
        <i className="vc-d vc-d2" aria-hidden="true" />
        <i className="vc-d vc-d3" aria-hidden="true" />
        <i className="vc-d vc-d4" aria-hidden="true" />

        <header className="vc-top">
          {data.logo && (
            <span className="vc-logo">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={data.logo} alt={data.company ? `${data.company} logo` : 'Company logo'} />
            </span>
          )}
          {data.company && <span className="vc-brand">{data.company}</span>}
          <span className="vc-tag">DIGITAL ID · {data.idNumber}</span>
        </header>

        <figure className="vc-photo">
          {data.photo ? (
            // eslint-disable-next-line @next/next/no-img-element
            <img src={data.photo} alt={data.name} />
          ) : (
            <span className="vc-mono" aria-hidden="true">
              {initials(data.name)}
            </span>
          )}
        </figure>

        <div className="vc-who">
          <h1 className="vc-name">{data.name}</h1>
          {data.role && <p className="vc-role">{data.role}</p>}
          {data.company && <p className="vc-co">{data.company}</p>}
        </div>

        {data.about && <p className="vc-about">{data.about}</p>}

        {data.tags.length > 0 && (
          <ul className="vc-tags">
            {data.tags.map((t) => (
              <li key={t}>{t}</li>
            ))}
          </ul>
        )}

        {rows.length > 0 && (
          <ul className="vc-list">
            {rows.map((r) => (
              <li key={r.kind}>
                <a
                  className={`vc-row vc-row-${r.kind}`}
                  href={link(r.href)}
                  onClick={inert}
                  tabIndex={interactive ? undefined : -1}
                  {...(r.kind === 'website' || r.kind === 'address' || r.kind === 'whatsapp' ? ext : {})}
                >
                  <span className="vc-ico">
                    <ContactIcon kind={r.kind} />
                  </span>
                  <span className="vc-lab">{r.label}</span>
                  <span className="vc-val">{r.value}</span>
                </a>
              </li>
            ))}
          </ul>
        )}

        {data.social.length > 0 && (
          <ul className="vc-soc">
            {data.social.map((s, i) => (
              <li key={`${s.platform}-${i}`}>
                <a
                  className="vc-s"
                  href={link(safeHttpUrl(s.url))}
                  onClick={inert}
                  title={s.label || s.platform}
                  aria-label={s.label || s.platform}
                  tabIndex={interactive ? undefined : -1}
                  {...ext}
                >
                  <SocialGlyph platform={s.platform} />
                </a>
              </li>
            ))}
          </ul>
        )}

        <footer className="vc-foot">
          <div className="vc-qr" role="img" aria-label="QR code for this card">
            {qr && (
              <svg viewBox={qr.viewBox} shapeRendering="crispEdges" aria-hidden="true">
                <path d={qr.d} />
              </svg>
            )}
          </div>
          <p className="vc-scan">Scan to connect</p>
          <a
            className="vc-cta"
            href={link(ctaTarget)}
            onClick={(e) => {
              if (!interactive) {
                e.preventDefault();
                return;
              }
              onCtaClick?.(e);
            }}
            tabIndex={interactive ? undefined : -1}
            {...(data.ctaHref ? ext : {})}
          >
            {cta}
            <ContactIcon kind="arrow" />
          </a>
        </footer>
      </div>
    </article>
  );
});

export default CardRenderer;
