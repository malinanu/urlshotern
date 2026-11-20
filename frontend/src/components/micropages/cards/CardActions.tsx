'use client';

import React, { useState } from 'react';
import { ArrowDownTrayIcon, CheckIcon, ClipboardDocumentIcon, PhotoIcon, QrCodeIcon, UserPlusIcon } from '@heroicons/react/24/outline';
import { downloadCardImage, downloadQrImage } from './export';

interface CardActionsProps {
  /** Returns the live card element to export. */
  getCard: () => HTMLElement | null;
  slug: string;
  qrValue: string;
  /** The card's public address, for the Copy link button. */
  pageUrl?: string;
  onSaveContact?: () => void;
  className?: string;
}

const btn =
  'flex items-center justify-center gap-2 px-3.5 py-2.5 rounded-xl text-xs font-semibold border transition-all active:scale-[0.98] disabled:opacity-50 disabled:cursor-wait';

/** Save contact, download the card as PNG or JPG, and download the QR code. */
export function CardActions({ getCard, slug, qrValue, pageUrl, onSaveContact, className = '' }: CardActionsProps) {
  const [busy, setBusy] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const run = async (key: string, fn: () => Promise<void>) => {
    setBusy(key);
    setError(null);
    try {
      await fn();
    } catch (e) {
      console.error(e);
      setError('Could not create the image. If you used an image link from another site, upload the file instead.');
    } finally {
      setBusy(null);
    }
  };

  const card = () => {
    const el = getCard();
    if (!el) throw new Error('Card is not ready');
    return el;
  };

  return (
    <div className={`w-full ${className}`}>
      <div className="grid grid-cols-2 gap-2">
        {onSaveContact && (
          <button
            type="button"
            onClick={onSaveContact}
            className={`${btn} col-span-2 bg-blue-600 hover:bg-blue-500 border-blue-500 text-white`}
          >
            <UserPlusIcon className="w-4 h-4" />
            Save contact to phone
          </button>
        )}
        <button
          type="button"
          disabled={busy !== null}
          onClick={() => run('png', () => downloadCardImage(card(), 'png', slug))}
          className={`${btn} bg-slate-800 hover:bg-slate-700 border-slate-700 text-white`}
        >
          <PhotoIcon className="w-4 h-4" />
          {busy === 'png' ? 'Preparing…' : 'Download PNG'}
        </button>
        <button
          type="button"
          disabled={busy !== null}
          onClick={() => run('jpg', () => downloadCardImage(card(), 'jpeg', slug))}
          className={`${btn} bg-slate-800 hover:bg-slate-700 border-slate-700 text-white`}
        >
          <ArrowDownTrayIcon className="w-4 h-4" />
          {busy === 'jpg' ? 'Preparing…' : 'Download JPG'}
        </button>
        <button
          type="button"
          disabled={busy !== null || !qrValue}
          onClick={() => run('qr', () => downloadQrImage(qrValue, slug))}
          className={`${btn} col-span-2 bg-slate-800 hover:bg-slate-700 border-slate-700 text-white`}
        >
          <QrCodeIcon className="w-4 h-4" />
          Download QR code
        </button>
        {pageUrl && (
          <button
            type="button"
            onClick={() => {
              navigator.clipboard?.writeText(pageUrl);
              setCopied(true);
              setTimeout(() => setCopied(false), 2000);
            }}
            className={`${btn} col-span-2 bg-slate-800 hover:bg-slate-700 border-slate-700 text-white`}
          >
            {copied ? <CheckIcon className="w-4 h-4 text-emerald-400" /> : <ClipboardDocumentIcon className="w-4 h-4" />}
            {copied ? 'Link copied' : 'Copy card link'}
          </button>
        )}
      </div>
      {error && (
        <p role="alert" className="mt-2 text-[11px] text-rose-400">
          {error}
        </p>
      )}
    </div>
  );
}

export default CardActions;
