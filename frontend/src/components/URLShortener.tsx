'use client';

import { useState, useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { ClipboardDocumentIcon, CheckIcon, QrCodeIcon } from '@heroicons/react/24/outline';
import clsx from 'clsx';
import { useAuth } from '@/contexts/AuthContext';
import { QRCodeModal } from '@/components/QRCode';
import { API_ENDPOINTS } from '@/lib/api-config';

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

interface CustomDomain {
  id: number;
  domain: string;
  is_verified: boolean;
  is_active: boolean;
}

interface Team {
  id: number;
  name: string;
  display_name?: string;
}

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export default function URLShortener({ onUrlCreated }: URLShortenerProps) {
  const [shortenedURL, setShortenedURL] = useState<ShortenedURL | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);
  const [showQRModal, setShowQRModal] = useState(false);
  const [customCode, setCustomCode] = useState('');
  const [password, setPassword] = useState('');
  const [showAdvanced, setShowAdvanced] = useState(false);
  const [ogTitle, setOgTitle] = useState('');
  const [ogDescription, setOgDescription] = useState('');
  const [ogImage, setOgImage] = useState('');
  const [domains, setDomains] = useState<CustomDomain[]>([]);
  const [selectedDomainID, setSelectedDomainID] = useState<number | null>(null);
  const [teams, setTeams] = useState<Team[]>([]);
  const [selectedTeamID, setSelectedTeamID] = useState<number | null>(null);
  const { user, getAccessToken, isAuthenticated, refreshToken } = useAuth();

  const isPro = user?.account_type === 'premium' || user?.account_type === 'enterprise';

  useEffect(() => {
    if (!isPro) return;
    const token = getAccessToken();
    const headers: Record<string, string> = {};
    if (token) headers.Authorization = `Bearer ${token}`;
    // Fetch verified custom domains
    fetch(`${API_BASE_URL}/api/v1/custom-domains/`, { headers })
      .then(r => r.ok ? r.json() : null)
      .then(data => {
        if (data?.domains) {
          setDomains((data.domains as CustomDomain[]).filter(d => d.is_verified && d.is_active));
        }
      })
      .catch(() => {});
    // Fetch teams
    fetch(`${API_BASE_URL}/api/v1/teams`, { headers })
      .then(r => r.ok ? r.json() : null)
      .then(data => { if (data?.teams) setTeams(data.teams as Team[]); })
      .catch(() => {});
  }, [isPro, getAccessToken]);

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

      const body: Record<string, unknown> = { url: data.url };
      if (isPro && customCode.trim()) body.custom_code = customCode.trim();
      if (isPro && password.trim()) body.password = password.trim();
      if (isPro && ogTitle.trim()) body.og_title = ogTitle.trim();
      if (isPro && ogDescription.trim()) body.og_description = ogDescription.trim();
      if (isPro && ogImage.trim()) body.og_image = ogImage.trim();
      if (isPro && selectedDomainID) body.domain_id = selectedDomainID;
      if (isPro && selectedTeamID) body.team_id = selectedTeamID;

      const response = await fetch(API_ENDPOINTS.urls.shorten, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
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

          const retryResponse = await fetch(API_ENDPOINTS.urls.shorten, {
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
        if (errorData.error === 'anonymous_limit_reached') {
          throw new Error('You\'ve used all 5 free URLs. Sign up for a free account to create unlimited short URLs!');
        }
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
      setCustomCode('');
      setPassword('');
      setOgTitle('');
      setOgDescription('');
      setOgImage('');
      setSelectedDomainID(null);
      setSelectedTeamID(null);
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
      <div className="bg-card rounded-lg shadow-sm border border-border p-6">
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div>
            <label
              htmlFor="url"
              className="block text-sm font-medium text-foreground mb-2"
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
                    'w-full px-4 py-3 border rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 transition-colors bg-background text-foreground placeholder-muted-foreground',
                    errors.url
                      ? 'border-destructive focus:ring-destructive focus:border-destructive'
                      : 'border-input hover:border-border'
                  )}
                />
                {errors.url && (
                  <p className="mt-1 text-sm text-red-600">{errors.url.message}</p>
                )}
              </div>
              <button
                type="submit"
                disabled={isLoading}
                className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 focus:ring-2 focus:ring-primary focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
              >
                {isLoading ? 'Shortening...' : 'Shorten URL'}
              </button>
            </div>
          </div>

          {/* Pro Options: Custom short code + Password */}
          {isPro && (
            <div>
              <button
                type="button"
                onClick={() => setShowAdvanced(!showAdvanced)}
                className="text-xs text-muted-foreground hover:text-foreground transition-colors"
              >
                {showAdvanced ? '▲ Hide options' : '▼ Advanced options (custom code, password)'}
              </button>
              {showAdvanced && (
                <div className="mt-3 space-y-3">
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">
                        Custom short code <span className="text-muted-foreground">(optional)</span>
                      </label>
                      <input
                        type="text"
                        value={customCode}
                        onChange={(e) => setCustomCode(e.target.value)}
                        placeholder="e.g. my-link"
                        maxLength={20}
                        className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">
                        Password protect <span className="text-muted-foreground">(optional)</span>
                      </label>
                      <input
                        type="password"
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        placeholder="Set a password"
                        className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                    </div>
                    {domains.length > 0 && (
                      <div>
                        <label className="block text-xs font-medium text-foreground mb-1">
                          Custom domain <span className="text-muted-foreground">(optional)</span>
                        </label>
                        <select
                          value={selectedDomainID ?? ''}
                          onChange={(e) => setSelectedDomainID(e.target.value ? Number(e.target.value) : null)}
                          className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                        >
                          <option value="">Default (trunc.site)</option>
                          {domains.map(d => (
                            <option key={d.id} value={d.id}>{d.domain}</option>
                          ))}
                        </select>
                      </div>
                    )}
                    {teams.length > 0 && (
                      <div>
                        <label className="block text-xs font-medium text-foreground mb-1">
                          Assign to team <span className="text-muted-foreground">(optional)</span>
                        </label>
                        <select
                          value={selectedTeamID ?? ''}
                          onChange={(e) => setSelectedTeamID(e.target.value ? Number(e.target.value) : null)}
                          className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                        >
                          <option value="">No team</option>
                          {teams.map(t => (
                            <option key={t.id} value={t.id}>{t.display_name || t.name}</option>
                          ))}
                        </select>
                      </div>
                    )}
                  </div>
                  <p className="text-xs text-muted-foreground font-medium">Open Graph / Social preview</p>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">
                        OG Title <span className="text-muted-foreground">(optional)</span>
                      </label>
                      <input
                        type="text"
                        value={ogTitle}
                        onChange={(e) => setOgTitle(e.target.value)}
                        placeholder="Custom preview title"
                        className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-medium text-foreground mb-1">
                        OG Image URL <span className="text-muted-foreground">(optional)</span>
                      </label>
                      <input
                        type="url"
                        value={ogImage}
                        onChange={(e) => setOgImage(e.target.value)}
                        placeholder="https://..."
                        className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                    </div>
                    <div className="sm:col-span-2">
                      <label className="block text-xs font-medium text-foreground mb-1">
                        OG Description <span className="text-muted-foreground">(optional)</span>
                      </label>
                      <input
                        type="text"
                        value={ogDescription}
                        onChange={(e) => setOgDescription(e.target.value)}
                        placeholder="Custom preview description"
                        className="w-full px-3 py-2 text-sm border border-input bg-background text-foreground rounded-lg focus:outline-none focus:ring-2 focus:ring-primary/50"
                      />
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </form>

        {/* Error Display */}
        {error && (
          <div className="mt-4 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
            <p className="text-sm text-destructive">{error}</p>
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
                    className="flex-1 px-3 py-2 text-sm bg-background border border-border rounded focus:ring-2 focus:ring-primary text-foreground"
                  />
                  <button
                    onClick={copyToClipboard}
                    aria-label={copied ? "Copied to clipboard" : "Copy short URL to clipboard"}
                    className="flex items-center gap-1 px-3 py-2 text-sm text-primary hover:text-primary/80 hover:bg-primary/10 rounded transition-colors"
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
                    aria-label="Show QR code for this short URL"
                    className="flex items-center gap-1 px-3 py-2 text-sm text-primary hover:text-primary/80 hover:bg-primary/10 rounded transition-colors"
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
                <p className="text-sm text-foreground break-all">
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