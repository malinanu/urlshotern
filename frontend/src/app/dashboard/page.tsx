'use client';

import { useEffect, useState, useRef } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { useAuth } from '@/contexts/AuthContext';
import { QRCodeModal } from '@/components/QRCode';
import { getCached, setCached, prefetch } from '@/lib/prefetch-cache';
import {
  ChartBarIcon,
  LinkIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  CalendarIcon,
  GlobeAltIcon,
  EyeIcon,
  TrashIcon,
  PencilIcon,
  QrCodeIcon,
  SwatchIcon,
  ArrowRightIcon,
} from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface URLData {
  id: number;
  short_code: string;
  short_url: string;
  original_url: string;
  created_at: string;
  click_count: number;
  is_active: boolean;
  is_public: boolean;
  title?: string;
  description?: string;
  expires_at?: string;
  team_id?: number;
}

interface Team {
  id: number;
  name: string;
  display_name?: string;
}

interface DashboardStats {
  total_urls: number;
  total_clicks: number;
  active_urls: number;
  today_clicks: number;
  month_clicks: number;
}

export default function DashboardPage() {
  const { user, isAuthenticated, isLoading, getAccessToken, refreshToken } = useAuth();
  const [urls, setUrls] = useState<URLData[]>(() => {
    const cached = getCached<{ urls: URLData[] }>('dashboard:urls');
    return cached?.urls ?? [];
  });
  const [stats, setStats] = useState<DashboardStats | null>(() => getCached<DashboardStats>('dashboard:stats'));
  const [loadingUrls, setLoadingUrls] = useState(getCached('dashboard:urls') === null);
  const [loadingStats, setLoadingStats] = useState(getCached('dashboard:stats') === null);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [editingUrl, setEditingUrl] = useState<URLData | null>(null);
  const [editError, setEditError] = useState<string | null>(null);
  const [qrModalUrl, setQrModalUrl] = useState<URLData | null>(null);
  const [userTeams, setUserTeams] = useState<Team[]>([]);
  const router = useRouter();
  const abortControllerRef = useRef<AbortController | null>(null);

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchData = async (retryOnTokenRefresh = true) => {
    if (!isAuthenticated || !user) return;

    // Cancel previous request if still pending
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    abortControllerRef.current = new AbortController();
    const token = getAccessToken();
    if (!token) return;

    const makeAuthenticatedRequest = async (url: string) => {
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        signal: abortControllerRef.current?.signal,
      });

      // Handle token expiration
      if (response.status === 401 && retryOnTokenRefresh) {
        try {
          await refreshToken();
          return fetchData(false); // Retry once with new token
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError);
          // Could redirect to login or show error message
        }
      }

      return response;
    };

    try {
      if (getCached('dashboard:urls') === null) setLoadingUrls(true);
      if (getCached('dashboard:stats') === null) setLoadingStats(true);

      // Fetch user's URLs from backend
      const urlsResponse = await makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/my-urls`);

      if (urlsResponse && urlsResponse.ok) {
        const urlsData = await urlsResponse.json();
        setUrls(urlsData.urls || []);
        setCached('dashboard:urls', urlsData);
      } else if (urlsResponse) {
        console.error('Failed to fetch URLs:', urlsResponse.status);
      }

      // Fetch dashboard stats from backend
      const statsResponse = await makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/dashboard/stats`);

      if (statsResponse && statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
        setCached('dashboard:stats', statsData);
      } else if (statsResponse) {
        console.error('Failed to fetch stats:', statsResponse.status);
      }
    } catch (error: any) {
      if (error.name !== 'AbortError') {
        console.error('Error fetching dashboard data:', error);
      }
    } finally {
      setLoadingUrls(false);
      setLoadingStats(false);
    }
  };

  // Fetch user's URLs and stats
  useEffect(() => {
    fetchData();

    return () => {
      // Cancel pending requests on unmount
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [isAuthenticated, user, getAccessToken]);

  // Fetch user's teams for the edit modal
  useEffect(() => {
    if (!isAuthenticated) return;
    const token = getAccessToken();
    if (!token) return;
    fetch(`${API_BASE_URL}/api/v1/teams`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => r.ok ? r.json() : Promise.reject(r.status))
      .then(data => setUserTeams(data.teams ?? data ?? []))
      .catch(() => {});
  }, [isAuthenticated, getAccessToken]);

  // Callback for when a new URL is created
  const handleUrlCreated = async (newUrl: any) => {
    // Add the new URL to the existing list immediately for instant feedback
    setUrls(prevUrls => [newUrl, ...prevUrls]);

    // Update stats optimistically
    if (stats) {
      setStats(prevStats => prevStats ? {
        ...prevStats,
        total_urls: prevStats.total_urls + 1,
        active_urls: prevStats.active_urls + 1,
      } : null);
    }

    // Refresh all data from server to ensure accuracy
    // This happens in the background without disrupting the UI
    setTimeout(async () => {
      try {
        await fetchData();
      } catch (error: any) {
        if (error.name !== 'AbortError') {
          console.error('Failed to refresh dashboard after URL creation:', error);
        }
      }
    }, 1000); // Small delay to allow optimistic update to be visible
  };


  const copyToClipboard = async (url: string, id: number) => {
    try {
      await navigator.clipboard.writeText(url);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const deleteURL = async (shortCode: string) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/my-urls/${shortCode}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        // Remove the URL from the local state
        setUrls(urls.filter(url => url.short_code !== shortCode));

        // Update stats
        if (stats) {
          setStats({
            ...stats,
            total_urls: stats.total_urls - 1,
            active_urls: stats.active_urls - 1,
          });
        }
      } else {
        console.error('Failed to delete URL:', response.status);
        alert('Failed to delete URL. Please try again.');
      }
    } catch (error) {
      console.error('Error deleting URL:', error);
      alert('Error deleting URL. Please try again.');
    }
  };

  const updateURL = async (shortCode: string, updates: {
    original_url?: string;
    title?: string;
    description?: string;
    expires_at?: string | null;
    is_public?: boolean;
    team_id?: number | null;
  }) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/my-urls/${shortCode}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });

      if (response.ok) {
        const updatedUrl = await response.json();

        // Update the URL in the local state
        setUrls(urls.map(url =>
          url.short_code === shortCode
            ? { ...url, ...updatedUrl }
            : url
        ));

        setEditingUrl(null);
        setEditError(null);
      } else {
        const errData = await response.json().catch(() => null);
        setEditError(errData?.error || errData?.message || 'Failed to update URL. Please try again.');
      }
    } catch (error) {
      console.error('Error updating URL:', error);
      setEditError('Network error. Please try again.');
    }
  };

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center bg-background">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground">Dashboard</h1>
            <p className="mt-2 text-muted-foreground">
              Welcome back, {user?.name}! Manage your shortened URLs and view analytics.
            </p>
          </div>

          {/* Stats Cards */}
          {loadingStats ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              {[1, 2, 3].map((i) => (
                <div key={i} className="bg-card p-6 rounded-lg shadow-sm border border-border animate-pulse">
                  <div className="h-4 bg-muted rounded w-24 mb-2"></div>
                  <div className="h-8 bg-muted rounded w-16"></div>
                </div>
              ))}
            </div>
          ) : stats && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                <div className="flex items-center">
                  <LinkIcon className="h-8 w-8 text-primary" />
                  <div className="ml-4">
                    <p className="text-sm text-muted-foreground">Total URLs</p>
                    <p className="text-2xl font-bold text-foreground">{stats.total_urls}</p>
                  </div>
                </div>
              </div>

              <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                <div className="flex items-center">
                  <ChartBarIcon className="h-8 w-8 text-primary" />
                  <div className="ml-4">
                    <p className="text-sm text-muted-foreground">Total Clicks</p>
                    <p className="text-2xl font-bold text-foreground">{stats.total_clicks}</p>
                  </div>
                </div>
              </div>

              <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                <div className="flex items-center">
                  <GlobeAltIcon className="h-8 w-8 text-primary" />
                  <div className="ml-4">
                    <p className="text-sm text-muted-foreground">Active URLs</p>
                    <p className="text-2xl font-bold text-foreground">{stats.active_urls}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Quick Access Feature Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-8">
            <a
              href="/custom-domains"
              className="bg-card border border-border rounded-lg p-5 flex items-start space-x-4 hover:border-primary/50 hover:shadow-md transition-all group"
            >
              <div className="p-2 bg-primary/10 rounded-lg shrink-0 group-hover:bg-primary/20 transition-colors">
                <GlobeAltIcon className="h-6 w-6 text-primary" />
              </div>
              <div className="flex-1 min-w-0">
                <div className="flex items-center justify-between">
                  <h3 className="font-semibold text-foreground">Custom Domains</h3>
                  <ArrowRightIcon className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                </div>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Use your own branded domain for short links
                </p>
              </div>
            </a>

            {user?.account_type === 'enterprise' ? (
              <a
                href="/settings?tab=branding"
                className="bg-card border border-border rounded-lg p-5 flex items-start space-x-4 hover:border-primary/50 hover:shadow-md transition-all group"
              >
                <div className="p-2 bg-primary/10 rounded-lg shrink-0 group-hover:bg-primary/20 transition-colors">
                  <SwatchIcon className="h-6 w-6 text-primary" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-foreground">White-Label Branding</h3>
                    <ArrowRightIcon className="h-4 w-4 text-muted-foreground group-hover:text-primary transition-colors" />
                  </div>
                  <p className="text-sm text-muted-foreground mt-0.5">
                    Customize colors, logo, and company name
                  </p>
                </div>
              </a>
            ) : (
              <div className="bg-card border border-border rounded-lg p-5 flex items-start space-x-4 opacity-60">
                <div className="p-2 bg-muted rounded-lg shrink-0">
                  <SwatchIcon className="h-6 w-6 text-muted-foreground" />
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between">
                    <h3 className="font-semibold text-foreground">White-Label Branding</h3>
                    <span className="text-xs font-medium px-2 py-0.5 rounded-full bg-primary/10 text-primary">Ultra</span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-0.5">
                    Customize colors, logo, and company name
                  </p>
                </div>
              </div>
            )}
          </div>

          {/* Trunc */}
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-foreground mb-4">Create New Short URL</h2>
            <URLShortener onUrlCreated={handleUrlCreated} />
          </div>

          {/* URLs List */}
          <div className="bg-card rounded-lg shadow-sm border border-border">
            <div className="px-6 py-4 border-b border-border">
              <h2 className="text-xl font-semibold text-foreground">Your URLs</h2>
            </div>

            {loadingUrls ? (
              <div className="p-6">
                <div className="space-y-4">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="animate-pulse">
                      <div className="h-4 bg-muted rounded w-full mb-2"></div>
                      <div className="h-4 bg-muted rounded w-3/4"></div>
                    </div>
                  ))}
                </div>
              </div>
            ) : urls.length === 0 ? (
              <div className="p-6 text-center">
                <LinkIcon className="mx-auto h-12 w-12 text-muted-foreground" />
                <h3 className="mt-4 text-lg font-medium text-foreground">No URLs yet</h3>
                <p className="mt-2 text-muted-foreground">
                  Create your first short URL using the form above.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-muted">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Original URL
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Short URL
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Clicks
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Created
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-border">
                    {urls.map((url) => (
                      <tr key={url.id} className="hover:bg-muted/50 transition-colors">
                        <td className="px-6 py-4">
                          <div className="max-w-xs truncate">
                            <a
                              href={url.original_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary hover:text-primary/80 text-sm"
                            >
                              {url.original_url}
                            </a>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-2">
                            <span className="text-sm text-foreground font-mono">
                              {url.short_url}
                            </span>
                            <button
                              onClick={() => copyToClipboard(url.short_url, url.id)}
                              aria-label={copiedId === url.id ? "Copied" : "Copy short URL"}
                              className="p-1 text-muted-foreground hover:text-foreground transition-colors"
                            >
                              {copiedId === url.id ? (
                                <CheckIcon className="h-4 w-4 text-green-500" />
                              ) : (
                                <ClipboardDocumentIcon className="h-4 w-4" />
                              )}
                            </button>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-1">
                            <EyeIcon className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm text-foreground">{url.click_count}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-1">
                            <CalendarIcon className="h-4 w-4 text-muted-foreground" />
                            <span className="text-sm text-muted-foreground">{formatDate(url.created_at)}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-2">
                            <button
                              onClick={() => router.push(`/analytics?shortCode=${url.short_code}`)}
                              onMouseEnter={() => {
                                router.prefetch('/analytics');
                                const token = getAccessToken();
                                if (token) {
                                  prefetch(`analytics:url:${url.short_code}`, () =>
                                    fetch(`${API_BASE_URL}/api/v1/analytics/${url.short_code}`, {
                                      headers: { Authorization: `Bearer ${token}` },
                                    }).then(r => r.ok ? r.json() : Promise.reject(r.status))
                                  );
                                }
                              }}
                              className="text-primary hover:text-primary/80 text-sm font-medium"
                            >
                              Analytics
                            </button>
                            <button
                              onClick={() => setQrModalUrl(url)}
                              aria-label="Show QR code"
                              className="p-1 text-primary hover:text-primary/80 transition-colors"
                              title="Show QR Code"
                            >
                              <QrCodeIcon className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => setEditingUrl(url)}
                              aria-label="Edit URL"
                              className="p-1 text-primary hover:text-primary/80 transition-colors"
                              title="Edit URL"
                            >
                              <PencilIcon className="h-4 w-4" />
                            </button>
                            <button
                              onClick={() => {
                                if (confirm('Are you sure you want to delete this URL? This action cannot be undone.')) {
                                  deleteURL(url.short_code);
                                }
                              }}
                              aria-label="Delete URL"
                              className="p-1 text-destructive hover:text-destructive/80 transition-colors"
                              title="Delete URL"
                            >
                              <TrashIcon className="h-4 w-4" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Edit URL Modal */}
      {editingUrl && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold text-foreground mb-4">Edit URL</h3>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                const formData = new FormData(e.currentTarget);
                const teamIdRaw = formData.get('team_id') as string;
                const updates = {
                  title: formData.get('title') as string || undefined,
                  description: formData.get('description') as string || undefined,
                  is_public: formData.get('is_public') === 'on',
                  team_id: teamIdRaw === '' ? null : Number(teamIdRaw) === 0 ? 0 : Number(teamIdRaw),
                };
                updateURL(editingUrl.short_code, updates);
              }}
            >
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Title (optional)
                  </label>
                  <input
                    type="text"
                    name="title"
                    defaultValue={editingUrl.title || ''}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Enter a title for your URL"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-1">
                    Description (optional)
                  </label>
                  <textarea
                    name="description"
                    defaultValue={editingUrl.description || ''}
                    rows={3}
                    className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                    placeholder="Enter a description for your URL"
                  />
                </div>

                <div className="flex items-center">
                  <input
                    type="checkbox"
                    name="is_public"
                    defaultChecked={editingUrl.is_public}
                    className="mr-2 border-input text-primary bg-background focus:ring-primary"
                  />
                  <label className="text-sm text-foreground">
                    Make this URL publicly visible
                  </label>
                </div>

                {userTeams.length > 0 && (
                  <div>
                    <label className="block text-sm font-medium text-foreground mb-1">
                      Assign to team
                    </label>
                    <select
                      name="team_id"
                      defaultValue={editingUrl.team_id ?? ''}
                      className="w-full px-3 py-2 border border-input bg-background text-foreground rounded-md focus:outline-none focus:ring-2 focus:ring-primary"
                    >
                      <option value="0">No team</option>
                      {userTeams.map(t => (
                        <option key={t.id} value={t.id}>{t.display_name || t.name}</option>
                      ))}
                    </select>
                  </div>
                )}
              </div>

              <div className="flex space-x-3 mt-6">
                <button
                  type="button"
                  onClick={() => setEditingUrl(null)}
                  className="flex-1 px-4 py-2 border border-input text-foreground rounded-md hover:bg-muted"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
                >
                  Update
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* QR Code Modal */}
      {qrModalUrl && (
        <QRCodeModal
          isOpen={true}
          onClose={() => setQrModalUrl(null)}
          value={qrModalUrl.short_url}
          title="QR Code"
          description={`QR code for ${qrModalUrl.short_url}`}
          size={300}
        />
      )}
    </Layout>
  );
}