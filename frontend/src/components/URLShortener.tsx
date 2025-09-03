'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ClipboardDocumentIcon, CheckIcon, QrCodeIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';
import { useAuth } from '@/contexts/AuthContext';
import { QRCodeModal } from '@/components/QRCode';

const urlSchema = z.object({
  url: z
    .string()
    .min(1, 'Please enter a URL')
    .url('Please enter a valid URL'),
});

type URLFormData = z.infer<typeof urlSchema>;

interface ShortenedURL {
  short_code: string;
  short_url: string;
  original_url: string;
  created_at: string;
}

interface URLShortenerProps {
  onUrlCreated?: (url: ShortenedURL) => void;
}

export default function URLShortener({ onUrlCreated }: URLShortenerProps) {
  const [shortenedURL, setShortenedURL] = useState<ShortenedURL | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [showQRModal, setShowQRModal] = useState(false);
  const { getAccessToken, isAuthenticated, refreshToken } = useAuth();

  const {
    register,
    handleSubmit,
    formState: { errors },
    reset,
  } = useForm<URLFormData>({
    resolver: zodResolver(urlSchema),
  });

  const onSubmit = async (data: URLFormData) => {
    setIsLoading(true);
    setError(null);
    setShortenedURL(null);

    const attemptRequest = async (retryOnTokenRefresh = true): Promise<ShortenedURL> => {
      const token = getAccessToken();
      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      
      // Include Authorization header if user is authenticated
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      const response = await fetch('http://localhost:8080/api/v1/shorten', {
        method: 'POST',
        headers,
        body: JSON.stringify(data),
      });

      // Handle token expiration
      if (response.status === 401 && isAuthenticated && retryOnTokenRefresh) {
        try {
          await refreshToken();
          return attemptRequest(false); // Retry once with new token
        } catch (refreshError) {
          // If refresh fails, continue without authentication
          console.warn('Token refresh failed, proceeding without authentication');
          const headersWithoutAuth: Record<string, string> = {
            'Content-Type': 'application/json',
          };
          
          const retryResponse = await fetch('http://localhost:8080/api/v1/shorten', {
            method: 'POST',
            headers: headersWithoutAuth,
            body: JSON.stringify(data),
          });
          
          if (!retryResponse.ok) {
            const errorData = await retryResponse.json();
            throw new Error(errorData.message || 'Failed to shorten URL');
          }
          
          return retryResponse.json();
        }
      }

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URL');
      }

      return response.json();
    };

    try {
      const result = await attemptRequest();
      setShortenedURL(result);
      
      // Trigger callback to refresh dashboard if provided
      onUrlCreated?.(result);
      
      reset();
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  const copyToClipboard = async () => {
    if (!shortenedURL) return;

    try {
      await navigator.clipboard.writeText(shortenedURL.short_url);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  return (
    <div className="max-w-2xl mx-auto">
      {/* URL Shortening Form */}
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label
              htmlFor="url"
              className="block text-sm font-medium text-black mb-2"
            >
              Enter your long URL
            </label>
            <div className="flex flex-col sm:flex-row gap-3">
              <div className="flex-1">
                <input
                  {...register('url')}
                  type="url"
                  id="url"
                  placeholder="https://example.com/very-long-url"
                  className={clsx(
                    'w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors text-black placeholder-gray-500',
                    errors.url
                      ? 'border-red-300 focus:ring-red-500 focus:border-red-500'
                      : 'border-gray-300'
                  )}
                />
                {errors.url && (
                  <p className="mt-1 text-sm text-red-600">{errors.url.message}</p>
                )}
              </div>
              <button
                type="submit"
                disabled={isLoading}
                className="px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
              >
                {isLoading ? 'Shortening...' : 'Shorten URL'}
              </button>
            </div>
          </div>
        </form>

        {/* Error Display */}
        {error && (
          <div className="mt-4 p-4 bg-red-50 border border-red-200 rounded-lg">
            <p className="text-sm text-red-600">{error}</p>
          </div>
        )}

        {/* Success Result */}
        {shortenedURL && (
          <div className="mt-6 p-4 bg-green-50 border border-green-200 rounded-lg">
            <h3 className="text-sm font-medium text-green-800 mb-3">
              URL shortened successfully!
            </h3>
            <div className="space-y-3">
              <div>
                <label className="block text-xs font-medium text-green-700 mb-1">
                  Short URL
                </label>
                <div className="flex items-center gap-2">
                  <input
                    type="text"
                    value={shortenedURL.short_url}
                    readOnly
                    className="flex-1 px-3 py-2 text-sm bg-white border border-green-300 rounded focus:ring-2 focus:ring-green-500 text-black"
                  />
                  <button
                    onClick={copyToClipboard}
                    className="flex items-center gap-1 px-3 py-2 text-sm text-green-700 hover:text-green-800 hover:bg-green-100 rounded transition-colors"
                  >
                    {copied ? (
                      <>
                        <CheckIcon className="h-4 w-4" />
                        Copied!
                      </>
                    ) : (
                      <>
                        <ClipboardDocumentIcon className="h-4 w-4" />
                        Copy
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => setShowQRModal(true)}
                    className="flex items-center gap-1 px-3 py-2 text-sm text-green-700 hover:text-green-800 hover:bg-green-100 rounded transition-colors"
                  >
                    <QrCodeIcon className="h-4 w-4" />
                    QR Code
                  </button>
                </div>
              </div>
              <div>
                <label className="block text-xs font-medium text-green-700 mb-1">
                  Original URL
                </label>
                <p className="text-sm text-black break-all">
                  {shortenedURL.original_url}
                </p>
              </div>
              <div className="flex justify-between items-center text-xs text-green-600">
                <span>Short Code: {shortenedURL.short_code}</span>
                <span>
                  Created: {new Date(shortenedURL.created_at).toLocaleString()}
                </span>
              </div>
            </div>
          </div>
        )}
      </div>

      {/* QR Code Modal */}
      {shortenedURL && (
        <QRCodeModal
          isOpen={showQRModal}
          onClose={() => setShowQRModal(false)}
          value={shortenedURL.short_url}
          title="QR Code for Short URL"
          description={`QR code for ${shortenedURL.short_url}`}
          size={300}
        />
      )}
    </div>
  );
}