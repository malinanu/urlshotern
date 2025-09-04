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
} from '@heroicons/react/24/outline';

interface APIKey {
  id: number;
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
  api_key_id: number;
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
      const response = await fetch('http://localhost:8080/api/v1/api-keys', {
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
      const response = await fetch('http://localhost:8080/api/v1/api-keys', {
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

  const handleDelete = async (keyId: number) => {
    if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/api-keys/${keyId}`, {
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

  const handleRevoke = async (keyId: number) => {
    if (!confirm('Are you sure you want to revoke this API key?')) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/api-keys/${keyId}/revoke`, {
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

  const fetchStats = async (keyId: number) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/api-keys/${keyId}/stats`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSelectedStats(data.stats);
      }
    } catch (error) {
      console.error('Error fetching API key stats:', error);
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
          <h2 className="text-2xl font-bold text-gray-900">API Keys</h2>
          <p className="text-gray-600">Manage API keys for programmatic access</p>
        </div>
        <button
          onClick={() => setShowForm(true)}
          className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center space-x-2"
        >
          <PlusIcon className="h-5 w-5" />
          <span>Create API Key</span>
        </button>
      </div>

      {/* API Keys List */}
      <div className="bg-white rounded-lg shadow border border-gray-200">
        {loading ? (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
            <p className="text-gray-600 mt-2">Loading API keys...</p>
          </div>
        ) : apiKeys.length === 0 ? (
          <div className="p-8 text-center">
            <KeyIcon className="h-12 w-12 text-gray-400 mx-auto" />
            <p className="text-gray-600 mt-2">No API keys found</p>
            <p className="text-gray-500 text-sm">Create your first API key to get started</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    API Key
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Last Used
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Rate Limit
                  </th>
                  <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Actions
                  </th>
                </tr>
              </thead>
              <tbody className="divide-y divide-gray-200">
                {apiKeys.map((apiKey) => (
                  <tr key={apiKey.id} className="hover:bg-gray-50">
                    <td className="px-6 py-4">
                      <div>
                        <div className="text-sm font-medium text-gray-900">{apiKey.name}</div>
                        <div className="text-sm text-gray-500 font-mono">{apiKey.key_prefix}</div>
                        {apiKey.description && (
                          <div className="text-xs text-gray-400 mt-1">{apiKey.description}</div>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center space-x-2">
                        <span
                          className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                            !apiKey.is_active
                              ? 'bg-red-100 text-red-800'
                              : isExpired(apiKey.expires_at)
                              ? 'bg-yellow-100 text-yellow-800'
                              : 'bg-green-100 text-green-800'
                          }`}
                        >
                          {!apiKey.is_active
                            ? 'Revoked'
                            : isExpired(apiKey.expires_at)
                            ? 'Expired'
                            : 'Active'}
                        </span>
                        {apiKey.expires_at && !isExpired(apiKey.expires_at) && (
                          <span className="text-xs text-gray-500">
                            Expires {formatDate(apiKey.expires_at)}
                          </span>
                        )}
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900">
                        {apiKey.last_used_at ? formatDate(apiKey.last_used_at) : 'Never'}
                      </div>
                      {apiKey.last_used_ip && (
                        <div className="text-xs text-gray-500">IP: {apiKey.last_used_ip}</div>
                      )}
                    </td>
                    <td className="px-6 py-4">
                      <div className="text-sm text-gray-900">
                        {apiKey.rate_limit.toLocaleString()}/hour
                      </div>
                    </td>
                    <td className="px-6 py-4">
                      <div className="flex items-center justify-end space-x-2">
                        <button
                          onClick={() => fetchStats(apiKey.id)}
                          className="p-2 text-blue-600 hover:text-blue-800 transition-colors"
                          title="View stats"
                        >
                          <ChartBarIcon className="h-4 w-4" />
                        </button>
                        {apiKey.is_active && (
                          <button
                            onClick={() => handleRevoke(apiKey.id)}
                            className="p-2 text-yellow-600 hover:text-yellow-800 transition-colors"
                            title="Revoke API key"
                          >
                            <EyeSlashIcon className="h-4 w-4" />
                          </button>
                        )}
                        <button
                          onClick={() => handleDelete(apiKey.id)}
                          className="p-2 text-red-600 hover:text-red-800 transition-colors"
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
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full m-4 max-h-[90vh] overflow-y-auto">
            <form onSubmit={handleSubmit} className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">Create New API Key</h3>
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <span className="sr-only">Close</span>
                  ✕
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label htmlFor="name" className="block text-sm font-medium text-gray-700">
                    Name *
                  </label>
                  <input
                    type="text"
                    id="name"
                    required
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    placeholder="My API Key"
                  />
                </div>

                <div>
                  <label htmlFor="description" className="block text-sm font-medium text-gray-700">
                    Description
                  </label>
                  <textarea
                    id="description"
                    rows={3}
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    placeholder="Optional description"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-3">
                    Permissions
                  </label>
                  <div className="space-y-2 max-h-40 overflow-y-auto border border-gray-200 rounded-md p-3">
                    {availablePermissions.map((permission) => (
                      <label key={permission.id} className="flex items-start space-x-3">
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
                          className="mt-1 rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
                        />
                        <div>
                          <div className="text-sm font-medium text-gray-700">{permission.name}</div>
                          <div className="text-xs text-gray-500">{permission.description}</div>
                        </div>
                      </label>
                    ))}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div>
                    <label htmlFor="rate_limit" className="block text-sm font-medium text-gray-700">
                      Rate Limit (requests/hour)
                    </label>
                    <input
                      type="number"
                      id="rate_limit"
                      min="1"
                      value={formData.rate_limit}
                      onChange={(e) => setFormData({ ...formData, rate_limit: parseInt(e.target.value) || 1000 })}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>

                  <div>
                    <label htmlFor="expires_at" className="block text-sm font-medium text-gray-700">
                      Expires At (optional)
                    </label>
                    <input
                      type="datetime-local"
                      id="expires_at"
                      value={formData.expires_at}
                      onChange={(e) => setFormData({ ...formData, expires_at: e.target.value })}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500"
                    />
                  </div>
                </div>
              </div>

              <div className="flex justify-end space-x-3 pt-6 border-t">
                <button
                  type="button"
                  onClick={() => setShowForm(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
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
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full m-4">
            <div className="p-6 space-y-6">
              <div className="text-center">
                <KeyIcon className="h-12 w-12 text-green-600 mx-auto" />
                <h3 className="text-lg font-medium text-gray-900 mt-4">API Key Created!</h3>
                <p className="text-gray-600 mt-2">
                  Copy your API key now. You won't be able to see it again.
                </p>
              </div>

              <div className="bg-gray-50 p-4 rounded-lg">
                <div className="flex items-center justify-between">
                  <code className="text-sm font-mono text-gray-800 break-all">{newApiKey}</code>
                  <button
                    onClick={() => copyToClipboard(newApiKey)}
                    className="ml-3 p-2 text-gray-500 hover:text-gray-700 transition-colors"
                    title="Copy to clipboard"
                  >
                    {copiedKey === newApiKey ? (
                      <CheckIcon className="h-5 w-5 text-green-600" />
                    ) : (
                      <ClipboardDocumentIcon className="h-5 w-5" />
                    )}
                  </button>
                </div>
              </div>

              <div className="bg-yellow-50 border border-yellow-200 p-4 rounded-lg">
                <p className="text-sm text-yellow-800">
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
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md shadow-sm hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
                >
                  I've Saved My Key
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Stats Modal */}
      {selectedStats && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full m-4">
            <div className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">API Key Statistics</h3>
                <button
                  onClick={() => setSelectedStats(null)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <span className="sr-only">Close</span>
                  ✕
                </button>
              </div>

              <div className="grid grid-cols-2 gap-4">
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-gray-900">
                    {selectedStats.total_requests.toLocaleString()}
                  </div>
                  <div className="text-sm text-gray-600">Total Requests</div>
                </div>
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-gray-900">
                    {selectedStats.today_requests.toLocaleString()}
                  </div>
                  <div className="text-sm text-gray-600">Today</div>
                </div>
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-gray-900">
                    {selectedStats.success_rate.toFixed(1)}%
                  </div>
                  <div className="text-sm text-gray-600">Success Rate</div>
                </div>
                <div className="bg-gray-50 p-4 rounded-lg">
                  <div className="text-2xl font-bold text-gray-900">
                    {selectedStats.avg_response_time.toFixed(0)}ms
                  </div>
                  <div className="text-sm text-gray-600">Avg Response Time</div>
                </div>
              </div>

              <div className="flex justify-end">
                <button
                  onClick={() => setSelectedStats(null)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
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