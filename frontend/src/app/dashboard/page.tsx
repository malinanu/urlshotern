'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { useAuth } from '@/contexts/AuthContext';
import { QRCodeModal } from '@/components/QRCode';
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
} from '@heroicons/react/24/outline';

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
  const [urls, setUrls] = useState<URLData[]>([]);
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loadingUrls, setLoadingUrls] = useState(true);
  const [loadingStats, setLoadingStats] = useState(true);
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [editingUrl, setEditingUrl] = useState<URLData | null>(null);
  const [qrModalUrl, setQrModalUrl] = useState<URLData | null>(null);
  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchData = async (retryOnTokenRefresh = true) => {
    if (!isAuthenticated || !user) return;

    const token = getAccessToken();
    if (!token) return;

    const makeAuthenticatedRequest = async (url: string) => {
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
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
      setLoadingUrls(true);
      setLoadingStats(true);

      // Fetch user's URLs from backend
      const urlsResponse = await makeAuthenticatedRequest('http://localhost:8080/api/v1/my-urls');
      
      if (urlsResponse && urlsResponse.ok) {
        const urlsData = await urlsResponse.json();
        setUrls(urlsData.urls || []);
      } else if (urlsResponse) {
        console.error('Failed to fetch URLs:', urlsResponse.status);
      }

      // Fetch dashboard stats from backend
      const statsResponse = await makeAuthenticatedRequest('http://localhost:8080/api/v1/dashboard/stats');
      
      if (statsResponse && statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      } else if (statsResponse) {
        console.error('Failed to fetch stats:', statsResponse.status);
      }
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoadingUrls(false);
      setLoadingStats(false);
    }
  };

  // Fetch user's URLs and stats
  useEffect(() => {
    fetchData();
  }, [isAuthenticated, user, getAccessToken]);

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
      } catch (error) {
        console.error('Failed to refresh dashboard after URL creation:', error);
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
      const response = await fetch(`http://localhost:8080/api/v1/my-urls/${shortCode}`, {
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

  const updateURL = async (shortCode: string, updates: { title?: string; description?: string; is_public?: boolean }) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/my-urls/${shortCode}`, {
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
      } else {
        console.error('Failed to update URL:', response.status);
        alert('Failed to update URL. Please try again.');
      }
    } catch (error) {
      console.error('Error updating URL:', error);
      alert('Error updating URL. Please try again.');
    }
  };

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-black">Dashboard</h1>
            <p className="mt-2 text-gray-600">
              Welcome back, {user?.name}! Manage your shortened URLs and view analytics.
            </p>
          </div>

          {/* Stats Cards */}
          {loadingStats ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              {[1, 2, 3].map((i) => (
                <div key={i} className="bg-white p-6 rounded-lg shadow-sm border animate-pulse">
                  <div className="h-4 bg-gray-200 rounded w-24 mb-2"></div>
                  <div className="h-8 bg-gray-200 rounded w-16"></div>
                </div>
              ))}
            </div>
          ) : stats && (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <LinkIcon className="h-8 w-8 text-primary-600" />
                  <div className="ml-4">
                    <p className="text-sm text-gray-600">Total URLs</p>
                    <p className="text-2xl font-bold text-black">{stats.total_urls}</p>
                  </div>
                </div>
              </div>
              
              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <ChartBarIcon className="h-8 w-8 text-green-600" />
                  <div className="ml-4">
                    <p className="text-sm text-gray-600">Total Clicks</p>
                    <p className="text-2xl font-bold text-black">{stats.total_clicks}</p>
                  </div>
                </div>
              </div>
              
              <div className="bg-white p-6 rounded-lg shadow-sm border">
                <div className="flex items-center">
                  <GlobeAltIcon className="h-8 w-8 text-blue-600" />
                  <div className="ml-4">
                    <p className="text-sm text-gray-600">Active URLs</p>
                    <p className="text-2xl font-bold text-black">{stats.active_urls}</p>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* URL Shortener */}
          <div className="mb-8">
            <h2 className="text-xl font-semibold text-black mb-4">Create New Short URL</h2>
            <URLShortener onUrlCreated={handleUrlCreated} />
          </div>

          {/* URLs List */}
          <div className="bg-white rounded-lg shadow-sm border">
            <div className="px-6 py-4 border-b">
              <h2 className="text-xl font-semibold text-black">Your URLs</h2>
            </div>
            
            {loadingUrls ? (
              <div className="p-6">
                <div className="space-y-4">
                  {[1, 2, 3].map((i) => (
                    <div key={i} className="animate-pulse">
                      <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                      <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                    </div>
                  ))}
                </div>
              </div>
            ) : urls.length === 0 ? (
              <div className="p-6 text-center">
                <LinkIcon className="mx-auto h-12 w-12 text-gray-400" />
                <h3 className="mt-4 text-lg font-medium text-black">No URLs yet</h3>
                <p className="mt-2 text-gray-600">
                  Create your first short URL using the form above.
                </p>
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Original URL
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Short URL
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Clicks
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Created
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Actions
                      </th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-gray-200">
                    {urls.map((url) => (
                      <tr key={url.id} className="hover:bg-gray-50">
                        <td className="px-6 py-4">
                          <div className="max-w-xs truncate">
                            <a
                              href={url.original_url}
                              target="_blank"
                              rel="noopener noreferrer"
                              className="text-primary-600 hover:text-primary-800 text-sm"
                            >
                              {url.original_url}
                            </a>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-2">
                            <span className="text-sm text-black font-mono">
                              {url.short_url}
                            </span>
                            <button
                              onClick={() => copyToClipboard(url.short_url, url.id)}
                              className="p-1 text-gray-400 hover:text-gray-600 transition-colors"
                            >
                              {copiedId === url.id ? (
                                <CheckIcon className="h-4 w-4 text-green-600" />
                              ) : (
                                <ClipboardDocumentIcon className="h-4 w-4" />
                              )}
                            </button>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-1">
                            <EyeIcon className="h-4 w-4 text-gray-400" />
                            <span className="text-sm text-black">{url.click_count}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-1">
                            <CalendarIcon className="h-4 w-4 text-gray-400" />
                            <span className="text-sm text-gray-600">{formatDate(url.created_at)}</span>
                          </div>
                        </td>
                        <td className="px-6 py-4">
                          <div className="flex items-center space-x-2">
                            <button 
                              onClick={() => router.push(`/analytics?shortCode=${url.short_code}`)}
                              className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                            >
                              Analytics
                            </button>
                            <button 
                              onClick={() => setQrModalUrl(url)}
                              className="p-1 text-purple-600 hover:text-purple-800 transition-colors"
                              title="Show QR Code"
                            >
                              <QrCodeIcon className="h-4 w-4" />
                            </button>
                            <button 
                              onClick={() => setEditingUrl(url)}
                              className="p-1 text-blue-600 hover:text-blue-800 transition-colors"
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
                              className="p-1 text-red-600 hover:text-red-800 transition-colors"
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
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md mx-4">
            <h3 className="text-lg font-semibold mb-4">Edit URL</h3>
            <form
              onSubmit={(e) => {
                e.preventDefault();
                const formData = new FormData(e.currentTarget);
                const updates = {
                  title: formData.get('title') as string || undefined,
                  description: formData.get('description') as string || undefined,
                  is_public: formData.get('is_public') === 'on',
                };
                updateURL(editingUrl.short_code, updates);
              }}
            >
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Title (optional)
                  </label>
                  <input
                    type="text"
                    name="title"
                    defaultValue={editingUrl.title || ''}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                    placeholder="Enter a title for your URL"
                  />
                </div>
                
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Description (optional)
                  </label>
                  <textarea
                    name="description"
                    defaultValue={editingUrl.description || ''}
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-primary-500"
                    placeholder="Enter a description for your URL"
                  />
                </div>
                
                <div className="flex items-center">
                  <input
                    type="checkbox"
                    name="is_public"
                    defaultChecked={editingUrl.is_public}
                    className="mr-2"
                  />
                  <label className="text-sm text-gray-700">
                    Make this URL publicly visible
                  </label>
                </div>
              </div>
              
              <div className="flex space-x-3 mt-6">
                <button
                  type="button"
                  onClick={() => setEditingUrl(null)}
                  className="flex-1 px-4 py-2 border border-gray-300 text-gray-700 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
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