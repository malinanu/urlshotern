'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import {
  KeyIcon,
  PlusIcon,
  EyeIcon,
  EyeSlashIcon,
  TrashIcon,
  ChartBarIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  XMarkIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface APIKey {
  id: string;
  key_prefix: string;
  name: string;
  description?: string;
  permissions: string[];
  rate_limit: number;
  last_used_at?: string;
  last_used_ip?: string;
  expires_at?: string;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface CreateAPIKeyRequest {
  name: string;
  description?: string;
  permissions: string[];
  rate_limit: number;
  expires_at?: string;
}

interface APIKeyStats {
  api_key_id: string;
  total_requests: number;
  today_requests: number;
  week_requests: number;
  month_requests: number;
  success_rate: number;
  avg_response_time: number;
}

export default function APIKeyManager() {
  const { getAccessToken } = useAuth();
  const [apiKeys, setAPIKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [newApiKey, setNewApiKey] = useState<string>('');
  const [showNewKey, setShowNewKey] = useState(false);
  const [copiedKey, setCopiedKey] = useState<string>('');
  const [formData, setFormData] = useState<CreateAPIKeyRequest>({
    name: '',
    description: '',
    permissions: ['url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read'],
    rate_limit: 1000,
    expires_at: '',
  });
  const [selectedStats, setSelectedStats] = useState<APIKeyStats | null>(null);
  const [statsLoading, setStatsLoading] = useState<string | null>(null); // Track which key's stats are loading
  const [statsError, setStatsError] = useState<string | null>(null);

  const availablePermissions = [
    { id: 'url.create', name: 'Create URLs', description: 'Create short URLs' },
    { id: 'url.read', name: 'Read URLs', description: 'View URL information' },
    { id: 'url.update', name: 'Update URLs', description: 'Modify existing URLs' },
    { id: 'url.delete', name: 'Delete URLs', description: 'Remove URLs' },
    { id: 'analytics.read', name: 'Read Analytics', description: 'View click analytics' },
    { id: 'domain.create', name: 'Create Custom Domains', description: 'Add custom domains' },
    { id: 'domain.read', name: 'Read Custom Domains', description: 'View custom domains' },
    { id: 'domain.update', name: 'Update Custom Domains', description: 'Modify custom domains' },
    { id: 'domain.delete', name: 'Delete Custom Domains', description: 'Remove custom domains' },
  ];

  useEffect(() => {
    fetchAPIKeys();
  }, []);

  const fetchAPIKeys = async () => {
    const token = getAccessToken();
    if (!token) return;

    try {
      setLoading(true);
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setAPIKeys(data.data || []);
      }
    } catch (error) {
      console.error('Error fetching API keys:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          ...formData,
          expires_at: formData.expires_at ? new Date(formData.expires_at).toISOString() : null,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        setNewApiKey(data.plain_key);
        setShowNewKey(true);
        await fetchAPIKeys();
        setShowForm(false);
        setFormData({
          name: '',
          description: '',
          permissions: ['url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read'],
          rate_limit: 1000,
          expires_at: '',
        });
      } else {
        const error = await response.json();
        alert(error.error || 'Failed to create API key');
      }
    } catch (error) {
      console.error('Error creating API key:', error);
    }
  };

  const handleDelete = async (keyId: string) => {
    if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys/${keyId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        await fetchAPIKeys();
      } else {
        alert('Failed to delete API key');
      }
    } catch (error) {
      console.error('Error deleting API key:', error);
    }
  };

  const handleRevoke = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key?')) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys/${keyId}/revoke`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        await fetchAPIKeys();
      } else {
        alert('Failed to revoke API key');
      }
    } catch (error) {
      console.error('Error revoking API key:', error);
    }
  };

  const fetchStats = async (keyId: string) => {
    const token = getAccessToken();
    if (!token) {
      setStatsError('Please log in to view statistics');
      return;
    }

    setStatsLoading(keyId);
    setStatsError(null);

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/api-keys/${keyId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      // Validate response structure
      if (!data || typeof data !== 'object') {
        throw new Error('Invalid response format from server');
      }

      // Ensure all required fields exist with defaults
      const validatedStats: APIKeyStats = {
        api_key_id: data.api_key_id || keyId,
        total_requests: data.total_requests ?? 0,
        today_requests: data.today_requests ?? 0,
        week_requests: data.week_requests ?? 0,
        month_requests: data.month_requests ?? 0,
        success_rate: data.success_rate ?? 0,
        avg_response_time: data.avg_response_time ?? 0,
      };

      setSelectedStats(validatedStats);
    } catch (error) {
      console.error('Error fetching API key stats:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to load statistics';
      setStatsError(errorMessage);

      // Still show modal with error message
      setSelectedStats(null);
    } finally {
      setStatsLoading(null);
    }
  };

  const copyToClipboard = async (text: string) => {
    try {
      await navigator.clipboard.writeText(text);
      setCopiedKey(text);
      setTimeout(() => setCopiedKey(''), 2000);
    } catch (err) {
      console.error('Failed to copy text: ', err);
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

  const isExpired = (expiresAt?: string) => {
    if (!expiresAt) return false;
    return new Date(expiresAt) < new Date();
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-foreground">API Keys</h2>
          <p className="text-muted-foreground">Manage API keys for programmatic access</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors"
        >
          <PlusIcon className="h-5 w-5" />
          <span>Create API Key</span>
        </button>
      </div>

      {/* API Keys List */}
      <div className="bg-card rounded-lg shadow-sm border border-border">
        {loading ? (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
            <p className="text-muted-foreground mt-2">Loading API keys...</p>
          </div>
        ) : apiKeys.length === 0 ? (
          <div className="p-8 text-center">
            <KeyIcon className="h-12 w-12 text-muted-foreground mx-auto" />
            <p className="text-muted-foreground mt-2">No API keys found</p>
            <p className="text-muted-foreground/80 text-sm">Create your first API key to get started</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-muted">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    API Key
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Last Used
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Rate Limit
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-muted-foreground uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border">
                {apiKeys.map((apiKey) => (
                  <tr key={apiKey.id} className="hover:bg-muted/50 transition-colors">
                    <td className="px-6 py-4">
                      <div>
                        <div className="text-sm font-medium text-foreground">{apiKey.name}</div>
                        <div className="text-sm text-muted-foreground font-mono">{apiKey.key_prefix}</div>
                        {apiKey.description && (
                          <div className="text-xs text-muted-foreground/80 mt-1">{apiKey.description}</div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center space-x-2">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${!apiKey.is_active
                              ? 'bg-destructive/10 text-destructive'
                              : isExpired(apiKey.expires_at)
                                ? 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-500'
                                : 'bg-green-500/10 text-green-600 dark:text-green-500'
                            }`}
                        >
                          {!apiKey.is_active
                            ? 'Revoked'
                            : isExpired(apiKey.expires_at)
                              ? 'Expired'
                              : 'Active'}
                        </span>
                        {apiKey.expires_at && !isExpired(apiKey.expires_at) && (
                          <span className="text-xs text-muted-foreground">
                            Expires {formatDate(apiKey.expires_at)}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-foreground">
                        {apiKey.last_used_at ? formatDate(apiKey.last_used_at) : 'Never'}
                      </div>
                      {apiKey.last_used_ip && (
                        <div className="text-xs text-muted-foreground">IP: {apiKey.last_used_ip}</div>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-foreground">
                        {apiKey.rate_limit.toLocaleString()}/hour
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => fetchStats(apiKey.id)}
                          disabled={statsLoading === apiKey.id}
                          className={`p-2 transition-colors ${statsLoading === apiKey.id
                              ? 'text-muted-foreground cursor-wait'
                              : 'text-primary hover:text-primary/80'
                            }`}
                          title={statsLoading === apiKey.id ? "Loading..." : "View stats"}
                        >
                          {statsLoading === apiKey.id ? (
                            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
                            </svg>
                          ) : (
                            <ChartBarIcon className="h-4 w-4" />
                          )}
                        </button>
                        {apiKey.is_active && (
                          <button
                            onClick={() => handleRevoke(apiKey.id)}
                            className="p-2 text-yellow-600 dark:text-yellow-500 hover:text-yellow-700 dark:hover:text-yellow-400 transition-colors"
                            title="Revoke API key"
                          >
                            <EyeSlashIcon className="h-4 w-4" />
                          </button>
                        )}
                        <button
                          onClick={() => handleDelete(apiKey.id)}
                          className="p-2 text-destructive hover:text-destructive/80 transition-colors"
                          title="Delete API key"
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

      {/* Create API Key Modal */}
      {showForm && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card rounded-lg shadow-xl max-w-2xl w-full m-4 max-h-[90vh] overflow-y-auto border border-border">
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-foreground">Create New API Key</h3>
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                >
                  <span className="sr-only">Close</span>
                  ✕
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-foreground">
                    Name *
                  </label>
                  <input
                    type="text"
                    id="name"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 bg-background border border-border rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary text-foreground"
                    placeholder="My API Key"
                  />
                </div>

                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-foreground">
                    Description
                  </label>
                  <textarea
                    id="description"
                    rows={3}
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 bg-background border border-border rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary text-foreground"
                    placeholder="Optional description"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-foreground mb-3">
                    Permissions
                  </label>
                  <div className="space-y-2 max-h-40 overflow-y-auto border border-border rounded-md p-3">
                    {availablePermissions.map((permission) => (
                      <label key={permission.id} className="flex items-start space-x-3 cursor-pointer">
                        <input
                          type="checkbox"
                          checked={formData.permissions.includes(permission.id)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setFormData({
                                ...formData,
                                permissions: [...formData.permissions, permission.id],
                              });
                            } else {
                              setFormData({
                                ...formData,
                                permissions: formData.permissions.filter((p) => p !== permission.id),
                              });
                            }
                          }}
                          className="mt-1 flex-shrink-0 h-4 w-4 text-primary bg-background border-border rounded focus:ring-primary focus:ring-offset-background"
                        />
                        <div>
                          <div className="text-sm font-medium text-foreground">{permission.name}</div>
                          <div className="text-xs text-muted-foreground">{permission.description}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="rate_limit" className="block text-sm font-medium text-foreground">
                      Rate Limit (requests/hour)
                    </label>
                    <input
                      type="number"
                      id="rate_limit"
                      min="1"
                      value={formData.rate_limit}
                      onChange={(e) => setFormData({ ...formData, rate_limit: parseInt(e.target.value) || 1000 })}
                      className="mt-1 block w-full px-3 py-2 bg-background border border-border rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary text-foreground"
                    />
                  </div>

                  <div>
                    <label htmlFor="expires_at" className="block text-sm font-medium text-foreground">
                      Expires At (optional)
                    </label>
                    <input
                      type="datetime-local"
                      id="expires_at"
                      value={formData.expires_at}
                      onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
                      className="mt-1 block w-full px-3 py-2 bg-background border border-border rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary text-foreground"
                    />
                  </div>
                </div>
              </div>

              <div className="flex justify-end space-x-3 pt-6 border-t border-border">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm font-medium text-foreground bg-background border border-border rounded-md shadow-sm hover:bg-muted focus:outline-none transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-md shadow-sm hover:bg-primary/90 focus:outline-none transition-colors"
                >
                  Create API Key
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* New API Key Display Modal */}
      {showNewKey && newApiKey && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card rounded-lg shadow-xl max-w-2xl w-full m-4 border border-border">
            <div className="p-6 space-y-6">
              <div className="text-center">
                <KeyIcon className="h-12 w-12 text-green-600 dark:text-green-500 mx-auto" />
                <h3 className="text-lg font-medium text-foreground mt-4">API Key Created!</h3>
                <p className="text-muted-foreground mt-2">
                  Copy your API key now. You won't be able to see it again.
                </p>
              </div>

              <div className="bg-muted p-4 rounded-lg border border-border">
                <div className="flex items-center justify-between">
                  <code className="text-sm font-mono text-foreground break-all">{newApiKey}</code>
                  <button
                    onClick={() => copyToClipboard(newApiKey)}
                    className="ml-3 p-2 text-muted-foreground hover:text-foreground transition-colors"
                    title="Copy to clipboard"
                  >
                    {copiedKey === newApiKey ? (
                      <CheckIcon className="h-5 w-5 text-green-600 dark:text-green-500" />
                    ) : (
                      <ClipboardDocumentIcon className="h-5 w-5" />
                    )}
                  </button>
                </div>
              </div>

              <div className="bg-yellow-500/10 border border-yellow-500/20 p-4 rounded-lg">
                <p className="text-sm text-yellow-600 dark:text-yellow-500">
                  <strong>Important:</strong> Store this API key securely. For security reasons,
                  we cannot show it to you again. If you lose it, you'll need to create a new one.
                </p>
              </div>

              <div className="flex justify-end">
                <button
                  onClick={() => {
                    setShowNewKey(false);
                    setNewApiKey('');
                  }}
                  className="px-4 py-2 text-sm font-medium text-primary-foreground bg-primary border border-transparent rounded-md shadow-sm hover:bg-primary/90 focus:outline-none transition-colors"
                >
                  I've Saved My Key
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Stats Modal */}
      {(selectedStats || statsError) && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card rounded-lg shadow-xl max-w-2xl w-full m-4 border border-border">
            <div className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-foreground">API Key Statistics</h3>
                <button
                  onClick={() => {
                    setSelectedStats(null);
                    setStatsError(null);
                  }}
                  className="text-muted-foreground hover:text-foreground transition-colors"
                >
                  <XMarkIcon className="h-6 w-6" />
                </button>
              </div>

              {statsError ? (
                <div className="bg-destructive/10 border border-destructive/20 rounded-lg p-4">
                  <div className="flex items-center">
                    <ExclamationTriangleIcon className="h-5 w-5 text-destructive mr-2 flex-shrink-0" />
                    <p className="text-destructive font-medium">{statsError}</p>
                  </div>
                </div>
              ) : selectedStats ? (
                <>
                  {selectedStats.total_requests === 0 && (
                    <div className="bg-primary/10 border border-primary/20 rounded-lg p-4 mb-4">
                      <p className="text-primary font-medium">
                        This API key hasn't been used yet. Statistics will appear once you start making requests.
                      </p>
                    </div>
                  )}

                  <div className="grid grid-cols-2 gap-4">
                    <div className="bg-muted p-4 rounded-lg border border-border">
                      <div className="text-2xl font-bold text-foreground">
                        {selectedStats.total_requests.toLocaleString()}
                      </div>
                      <div className="text-sm text-muted-foreground">Total Requests</div>
                    </div>
                    <div className="bg-muted p-4 rounded-lg border border-border">
                      <div className="text-2xl font-bold text-foreground">
                        {selectedStats.today_requests.toLocaleString()}
                      </div>
                      <div className="text-sm text-muted-foreground">Today</div>
                    </div>
                    <div className="bg-muted p-4 rounded-lg border border-border">
                      <div className="text-2xl font-bold text-foreground">
                        {selectedStats.success_rate.toFixed(1)}%
                      </div>
                      <div className="text-sm text-muted-foreground">Success Rate</div>
                    </div>
                    <div className="bg-muted p-4 rounded-lg border border-border">
                      <div className="text-2xl font-bold text-foreground">
                        {selectedStats.avg_response_time.toFixed(0)}ms
                      </div>
                      <div className="text-sm text-muted-foreground">Avg Response Time</div>
                    </div>
                  </div>
                </>
              ) : null}

              <div className="flex justify-end">
                <button
                  onClick={() => {
                    setSelectedStats(null);
                    setStatsError(null);
                  }}
                  className="px-4 py-2 text-sm font-medium text-foreground bg-background border border-border rounded-md shadow-sm hover:bg-muted focus:outline-none transition-colors"
                >
                  Close
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}