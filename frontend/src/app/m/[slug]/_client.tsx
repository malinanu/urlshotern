'use client';

import React, { useState, useEffect } from 'react';
import { MicroPage } from '@/types/micropage';
import { API_ENDPOINTS } from '@/lib/api-config';
import { MicroPageRenderer } from '@/components/micropages/MicroPageRenderer';
import { QRCodeModal } from '@/components/QRCode';
import {
  QrCodeIcon,
  ShareIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  ArrowDownTrayIcon,
} from '@heroicons/react/24/outline';

interface PublicMicroPageClientProps {
  slug: string;
  initialData?: MicroPage | null;
}

export default function PublicMicroPageClient({ slug, initialData }: PublicMicroPageClientProps) {
  const [page, setPage] = useState<MicroPage | null>(initialData || null);
  const [loading, setLoading] = useState(!initialData);
  const [notFound, setNotFound] = useState(false);
  const [qrOpen, setQrOpen] = useState(false);
  const [copied, setCopied] = useState(false);

  useEffect(() => {
    if (!initialData) {
      const fetchPage = async () => {
        try {
          setLoading(true);
          const res = await fetch(API_ENDPOINTS.micropages.public(slug));
          if (!res.ok) {
            setNotFound(true);
            return;
          }
          const data = await res.json();
          setPage(data);
        } catch {
          setNotFound(true);
        } finally {
          setLoading(false);
        }
      };

      fetchPage();
    }
  }, [slug, initialData]);

  const handleLinkClick = (linkId: string, url: string) => {
    // Fire-and-forget click metric tracking
    try {
      const clickEndpoint = API_ENDPOINTS.micropages.click(slug, linkId);
      if (typeof navigator !== 'undefined' && navigator.sendBeacon) {
        navigator.sendBeacon(clickEndpoint);
      } else {
        fetch(clickEndpoint, { method: 'POST', keepalive: true }).catch(() => {});
      }
    } catch {
      // Ignore click track failure
    }
  };

  const handleSaveContact = () => {
    // Trigger vCard download endpoint
    const vcardUrl = API_ENDPOINTS.micropages.vcard(slug);
    const link = document.createElement('a');
    link.href = vcardUrl;
    link.setAttribute('download', `${slug}.vcf`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const handleCopyLink = () => {
    if (typeof window !== 'undefined') {
      navigator.clipboard.writeText(window.location.href);
      setCopied(true);
      setTimeout(() => setCopied(false), 2500);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-4">
        <div className="w-10 h-10 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mb-4" />
        <p className="text-slate-400 text-sm">Loading vCard...</p>
      </div>
    );
  }

  if (notFound || !page) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center p-6 text-center">
        <div className="w-16 h-16 rounded-full bg-slate-900 border border-slate-800 flex items-center justify-center text-slate-500 text-2xl mb-4">
          📇
        </div>
        <h1 className="text-2xl font-bold text-white">Micro-Page Not Found</h1>
        <p className="mt-2 text-sm text-slate-400 max-w-sm">
          The link you visited might have been moved, removed, or is set to private.
        </p>
        <a
          href="/"
          className="mt-6 px-5 py-2.5 rounded-xl text-xs font-semibold bg-blue-600 hover:bg-blue-500 text-white shadow-lg shadow-blue-500/25 transition-all"
        >
          Create Your Own vCard
        </a>
      </div>
    );
  }

  const currentUrl = typeof window !== 'undefined' ? window.location.href : '';

  return (
    <main className="min-h-screen flex flex-col items-center justify-start bg-slate-950 text-white relative">
      {/* Floating Share / Action Top Bar */}
      <div className="fixed top-4 right-4 z-40 flex items-center gap-2">
        <button
          type="button"
          onClick={handleCopyLink}
          title="Copy Link"
          className="p-2.5 rounded-full bg-black/60 hover:bg-black/90 backdrop-blur-md text-white border border-white/20 shadow-lg transition-all active:scale-95"
        >
          {copied ? <CheckIcon className="w-4 h-4 text-emerald-400" /> : <ClipboardDocumentIcon className="w-4 h-4" />}
        </button>

        <button
          type="button"
          onClick={() => setQrOpen(true)}
          title="Show QR Code"
          className="p-2.5 rounded-full bg-black/60 hover:bg-black/90 backdrop-blur-md text-white border border-white/20 shadow-lg transition-all active:scale-95"
        >
          <QrCodeIcon className="w-4 h-4" />
        </button>

        <button
          type="button"
          onClick={handleSaveContact}
          title="Save Contact (.vcf)"
          className="p-2.5 rounded-full bg-black/60 hover:bg-black/90 backdrop-blur-md text-white border border-white/20 shadow-lg transition-all active:scale-95"
        >
          <ArrowDownTrayIcon className="w-4 h-4 text-blue-400" />
        </button>
      </div>

      {/* Main Renderer Container */}
      <div className="w-full max-w-md min-h-screen shadow-2xl flex flex-col">
        <MicroPageRenderer
          page={page}
          isPreview={false}
          onLinkClick={handleLinkClick}
          onSaveContact={handleSaveContact}
        />
      </div>

      {/* QR Code Modal */}
      {qrOpen && (
        <QRCodeModal
          isOpen={true}
          onClose={() => setQrOpen(false)}
          value={currentUrl}
          title={`Scan to open /m/${page.slug}`}
        />
      )}
    </main>
  );
}
