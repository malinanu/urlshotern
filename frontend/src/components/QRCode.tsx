'use client';

import { useState, useEffect, useRef } from 'react';
import QRCodeLib from 'qrcode';
import { QrCodeIcon, ArrowDownTrayIcon, XMarkIcon } from '@heroicons/react/24/outline';

interface QRCodeProps {
  value: string;
  size?: number;
  level?: 'L' | 'M' | 'Q' | 'H';
  includeMargin?: boolean;
  color?: {
    dark?: string;
    light?: string;
  };
  className?: string;
}

interface QRCodeModalProps extends QRCodeProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  description?: string;
}

export function QRCode({ 
  value, 
  size = 256, 
  level = 'M',
  includeMargin = true,
  color = { dark: '#000000', light: '#FFFFFF' },
  className = ''
}: QRCodeProps) {
  const [qrDataURL, setQrDataURL] = useState<string>('');
  const [error, setError] = useState<string>('');
  const canvasRef = useRef<HTMLCanvasElement>(null);

  useEffect(() => {
    const generateQR = async () => {
      if (!value) {
        setError('No value provided for QR code');
        return;
      }

      try {
        const canvas = canvasRef.current;
        if (!canvas) return;

        await QRCodeLib.toCanvas(canvas, value, {
          errorCorrectionLevel: level,
          width: size,
          margin: includeMargin ? 4 : 0,
          color: color,
        });

        // Also generate data URL for downloading
        const dataURL = await QRCodeLib.toDataURL(value, {
          errorCorrectionLevel: level,
          width: size,
          margin: includeMargin ? 4 : 0,
          color: color,
        });
        setQrDataURL(dataURL);
        setError('');
      } catch (err) {
        console.error('Error generating QR code:', err);
        setError('Failed to generate QR code');
      }
    };

    generateQR();
  }, [value, size, level, includeMargin, color]);

  const downloadQR = () => {
    if (!qrDataURL) return;

    const link = document.createElement('a');
    link.href = qrDataURL;
    link.download = `qr-code-${Date.now()}.png`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  if (error) {
    return (
      <div className={`flex items-center justify-center bg-surface-container dark:bg-zinc-900 rounded-xl p-4 border border-outline-variant/30 ${className}`}>
        <div className="text-center">
          <QrCodeIcon className="h-8 w-8 text-muted-foreground mx-auto mb-2" />
          <p className="text-xs text-muted-foreground font-mono">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`relative group ${className}`}>
      <canvas 
        ref={canvasRef}
        className="rounded-xl shadow-sm border border-outline-variant/30"
        style={{ maxWidth: '100%', height: 'auto' }}
      />
      <div className="absolute inset-0 bg-black/10 group-hover:bg-black/20 transition-all duration-200 rounded-xl flex items-center justify-center opacity-0 group-hover:opacity-100">
        <button
          onClick={downloadQR}
          aria-label="Download QR code"
          className="bg-surface-container-lowest dark:bg-zinc-800 text-foreground px-3 py-2 rounded-xl shadow-lg border border-outline-variant/30 hover:bg-surface-container transition-all flex items-center gap-2 text-xs font-semibold"
        >
          <ArrowDownTrayIcon className="h-4 w-4" />
          Download
        </button>
      </div>
    </div>
  );
}

export function QRCodeModal({ 
  isOpen, 
  onClose, 
  title = 'QR Code',
  description,
  ...qrProps 
}: QRCodeModalProps) {
  useEffect(() => {
    const handleEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        onClose();
      }
    };

    if (isOpen) {
      document.addEventListener('keydown', handleEscape);
      document.body.style.overflow = 'hidden';
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.body.style.overflow = 'unset';
    };
  }, [isOpen, onClose]);

  const downloadQR = async () => {
    try {
      const dataURL = await QRCodeLib.toDataURL(qrProps.value, {
        errorCorrectionLevel: qrProps.level || 'M',
        width: qrProps.size || 256,
        margin: qrProps.includeMargin ? 4 : 0,
        color: qrProps.color || { dark: '#000000', light: '#FFFFFF' },
      });

      const link = document.createElement('a');
      link.href = dataURL;
      link.download = `qr-code-${Date.now()}.png`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
    } catch (err) {
      console.error('Error downloading QR code:', err);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-in fade-in duration-150">
      <div className="bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl shadow-2xl max-w-md w-full mx-4 overflow-hidden text-on-surface">
        <div className="flex items-center justify-between p-5 border-b border-outline-variant/20">
          <div>
            <h3 className="text-base font-bold text-foreground font-hanken">{title}</h3>
            {description && (
              <p className="text-xs text-muted-foreground mt-0.5 font-mono truncate max-w-xs">{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            aria-label="Close QR code modal"
            className="p-1 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors"
          >
            <XMarkIcon className="h-5 w-5" />
          </button>
        </div>

        <div className="p-6">
          <div className="flex justify-center mb-6">
            <div className="p-4 bg-white rounded-2xl border border-outline-variant/20 shadow-inner flex items-center justify-center">
              <QRCode {...qrProps} className="max-w-full" />
            </div>
          </div>

          <div className="text-center">
            <p className="text-xs text-muted-foreground mb-5">
              Scan this interactive dynamic QR code to instantly route to destination
            </p>
            
            <div className="flex flex-col sm:flex-row gap-2.5">
              <button
                onClick={downloadQR}
                className="flex-1 bg-primary hover:bg-primary-600 active:scale-[0.98] text-white px-4 py-2.5 rounded-xl text-xs font-semibold shadow-sm transition-all flex items-center justify-center gap-2"
              >
                <ArrowDownTrayIcon className="h-4 w-4 stroke-[2.5]" />
                Download PNG
              </button>
              
              <button
                onClick={onClose}
                className="flex-1 bg-surface-container dark:bg-zinc-900 hover:bg-surface-container-high text-foreground border border-outline-variant/30 px-4 py-2.5 rounded-xl text-xs font-semibold transition-all"
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

export default QRCode;