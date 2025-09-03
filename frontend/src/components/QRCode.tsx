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
      <div className={`flex items-center justify-center bg-gray-100 rounded-lg p-4 ${className}`}>
        <div className="text-center">
          <QrCodeIcon className="h-8 w-8 text-gray-400 mx-auto mb-2" />
          <p className="text-sm text-gray-500">{error}</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`relative group ${className}`}>
      <canvas 
        ref={canvasRef}
        className="rounded-lg shadow-sm border border-gray-200"
        style={{ maxWidth: '100%', height: 'auto' }}
      />
      <div className="absolute inset-0 bg-black bg-opacity-0 group-hover:bg-opacity-10 transition-all duration-200 rounded-lg flex items-center justify-center opacity-0 group-hover:opacity-100">
        <button
          onClick={downloadQR}
          className="bg-white text-gray-700 px-3 py-2 rounded-lg shadow-lg hover:bg-gray-50 transition-colors flex items-center gap-2 text-sm font-medium"
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
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="flex items-center justify-between p-6 border-b">
          <div>
            <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
            {description && (
              <p className="text-sm text-gray-500 mt-1">{description}</p>
            )}
          </div>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-gray-500 transition-colors"
          >
            <XMarkIcon className="h-6 w-6" />
          </button>
        </div>

        <div className="p-6">
          <div className="flex justify-center mb-6">
            <QRCode {...qrProps} className="max-w-full" />
          </div>

          <div className="text-center">
            <p className="text-sm text-gray-600 mb-4">
              Scan this QR code with your phone to quickly access the URL
            </p>
            
            <div className="flex flex-col sm:flex-row gap-3">
              <button
                onClick={downloadQR}
                className="flex-1 bg-primary-600 text-white px-4 py-2 rounded-lg hover:bg-primary-700 transition-colors flex items-center justify-center gap-2"
              >
                <ArrowDownTrayIcon className="h-4 w-4" />
                Download QR Code
              </button>
              
              <button
                onClick={onClose}
                className="flex-1 bg-gray-200 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-300 transition-colors"
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