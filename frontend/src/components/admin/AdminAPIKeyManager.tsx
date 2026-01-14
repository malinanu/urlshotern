'use client';

import { useState, useEffect } from 'react';
import { adminAuthService } from '@/lib/admin-auth-service';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  KeyIcon,
  ChartBarIcon,
  TrashIcon,
  EyeSlashIcon,
  XMarkIcon,
  ExclamationTriangleIcon,
  UserCircleIcon,
  ClockIcon,
  CheckCircleIcon,
  XCircleIcon,
} from '@heroicons/react/24/outline';

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
  user_id: number;
  user_name?: string;
  user_email?: string;
}

interface APIKeyUsage {
  api_key_id: string;
  total_requests: number;
  today_requests: number;
  week_requests: number;
  month_requests: number;
  successful_requests: number;
  failed_requests: number;
  success_rate: number;
  avg_response_time: number;
  last_request_at?: string;
}

export default function AdminAPIKeyManager() {
  const [apiKeys, setAPIKeys] = useState<APIKey[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedKeyUsage, setSelectedKeyUsage] = useState<APIKeyUsage | null>(null);
  const [usageLoading, setUsageLoading] = useState<string | null>(null);
  const [usageError, setUsageError] = useState<string | null>(null);

  useEffect(() => {
    fetchAPIKeys();
  }, []);

  const fetchAPIKeys = async () => {
    const token = adminAuthService.getAdminToken();
    if (!token) return;

    try {
      setLoading(true);
      const response = await fetch(`${API_ENDPOINTS.admin.dashboard}/api-keys`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        // Ensure we always set an array, regardless of API response format
        const apiKeysData = Array.isArray(data?.data)
          ? data.data
          : (Array.isArray(data) ? data : []);
        setAPIKeys(apiKeysData);
      } else {
        console.error('Failed to fetch API keys:', response.statusText);
        setAPIKeys([]); // Set empty array on error to prevent crashes
      }
    } catch (error) {
      console.error('Error fetching API keys:', error);
      setAPIKeys([]); // Set empty array on error to prevent crashes
    } finally {
      setLoading(false);
    }
  };

  const fetchKeyUsage = async (keyId: string) => {
    const token = adminAuthService.getAdminToken();
    if (!token) {
      setUsageError('Admin authentication required');
      return;
    }

    setUsageLoading(keyId);
    setUsageError(null);

    try {
      const response = await fetch(`${API_ENDPOINTS.admin.dashboard}/api-keys/${keyId}/usage`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();

      // Validate and set defaults
      const validatedUsage: APIKeyUsage = {
        api_key_id: data.api_key_id || keyId,
        total_requests: data.total_requests ?? 0,
        today_requests: data.today_requests ?? 0,
        week_requests: data.week_requests ?? 0,
        month_requests: data.month_requests ?? 0,
        successful_requests: data.successful_requests ?? 0,
        failed_requests: data.failed_requests ?? 0,
        success_rate: data.success_rate ?? 0,
        avg_response_time: data.avg_response_time ?? 0,
        last_request_at: data.last_request_at,
      };

      setSelectedKeyUsage(validatedUsage);
    } catch (error) {
      console.error('Error fetching API key usage:', error);
      const errorMessage = error instanceof Error ? error.message : 'Failed to load usage statistics';
      setUsageError(errorMessage);
      setSelectedKeyUsage(null);
    } finally {
      setUsageLoading(null);
    }
  };

  const handleRevoke = async (keyId: string) => {
    if (!confirm('Are you sure you want to revoke this API key?')) return;

    const token = adminAuthService.getAdminToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_ENDPOINTS.admin.dashboard}/api-keys/${keyId}/revoke`, {
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
      alert('Error revoking API key');
    }
  };

  const handleDelete = async (keyId: string) => {
    if (!confirm('Are you sure you want to delete this API key? This action cannot be undone.')) return;

    const token = adminAuthService.getAdminToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_ENDPOINTS.admin.dashboard}/api-keys/${keyId}`, {
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
      alert('Error deleting API key');
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
      <div>
        <h2 className="text-2xl font-bold text-gray-900">API Keys Management</h2>
        <p className="text-gray-600">View and manage all API keys across all users</p>
      </div>

      {/* API Keys Table */}
      <div className="bg-white rounded-lg shadow border border-gray-200">
        {loading ? (
          <div className="p-8 text-center">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600 mx-auto"></div>
            <p className="text-gray-600 mt-2">Loading API keys...</p>
          </div>
        ) : apiKeys.length === 0 ? (
          <div className="p-8 text-center">
            <KeyIcon className="h-12 w-12 text-gray-400 mx-auto" />
            <p className="text-gray-600 mt-2">No API keys found</p>
            <p className="text-gray-500 text-sm">No users have created API keys yet</p>
          </div>
        ) : (
          <div className="overflow-x-auto">
            <table className="w-full">
              <thead className="bg-gray-50">
                <tr>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Key Name
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    User
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Status
                  </th>
                  <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Created
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
                {(Array.isArray(apiKeys) ? apiKeys : []).map((apiKey) => (
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
                      <div className="flex items-center">
                        <UserCircleIcon className="h-5 w-5 text-gray-400 mr-2" />
                        <div>
                          <div className="text-sm font-medium text-gray-900">
                            {apiKey.user_name || `User #${apiKey.user_id}`}
                          </div>
                          {apiKey.user_email && (
                            <div className="text-xs text-gray-500">{apiKey.user_email}</div>
                          )}
                        </div>
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
                          {!apiKey.is_active ? (
                            <>
                              <XCircleIcon className="h-3 w-3 mr-1" />
                              Revoked
                            </>
                          ) : isExpired(apiKey.expires_at) ? (
                            <>
                              <ClockIcon className="h-3 w-3 mr-1" />
                              Expired
                            </>
                          ) : (
                            <>
                              <CheckCircleIcon className="h-3 w-3 mr-1" />
                              Active
                            </>
                          )}
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
                        {formatDate(apiKey.created_at)}
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
                          onClick={() => fetchKeyUsage(apiKey.id)}
                          disabled={usageLoading === apiKey.id}
                          className={`p-2 transition-colors ${
                            usageLoading === apiKey.id
                              ? 'text-gray-400 cursor-wait'
                              : 'text-blue-600 hover:text-blue-800'
                          }`}
                          title={usageLoading === apiKey.id ? 'Loading...' : 'View usage statistics'}
                        >
                          {usageLoading === apiKey.id ? (
                            <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                              <circle
                                className="opacity-25"
                                cx="12"
                                cy="12"
                                r="10"
                                stroke="currentColor"
                                strokeWidth="4"
                                fill="none"
                              />
                              <path
                                className="opacity-75"
                                fill="currentColor"
                                d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                              />
                            </svg>
                          ) : (
                            <ChartBarIcon className="h-4 w-4" />
                          )}
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

      {/* Usage Statistics Modal */}
      {(selectedKeyUsage || usageError) && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-3xl w-full m-4">
            <div className="p-6 space-y-6">
              <div className="flex items-center justify-between">
                <h3 className="text-lg font-medium text-gray-900">API Key Usage Statistics</h3>
                <button
                  onClick={() => {
                    setSelectedKeyUsage(null);
                    setUsageError(null);
                  }}
                  className="text-gray-500 hover:text-gray-700"
                >
                  <XMarkIcon className="h-6 w-6" />
                </button>
              </div>

              {usageError ? (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                  <div className="flex items-center">
                    <ExclamationTriangleIcon className="h-5 w-5 text-red-500 mr-2 flex-shrink-0" />
                    <p className="text-red-700">{usageError}</p>
                  </div>
                </div>
              ) : selectedKeyUsage ? (
                <>
                  {selectedKeyUsage.total_requests === 0 && (
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-4">
                      <p className="text-blue-700">
                        This API key hasn't been used yet. Statistics will appear once requests are made.
                      </p>
                    </div>
                  )}

                  {/* Request Statistics */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-3">Request Statistics</h4>
                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
                      <div className="bg-gray-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-gray-900">
                          {selectedKeyUsage.total_requests.toLocaleString()}
                        </div>
                        <div className="text-sm text-gray-600">Total Requests</div>
                      </div>
                      <div className="bg-blue-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-blue-900">
                          {selectedKeyUsage.today_requests.toLocaleString()}
                        </div>
                        <div className="text-sm text-blue-700">Today</div>
                      </div>
                      <div className="bg-purple-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-purple-900">
                          {selectedKeyUsage.week_requests.toLocaleString()}
                        </div>
                        <div className="text-sm text-purple-700">This Week</div>
                      </div>
                      <div className="bg-indigo-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-indigo-900">
                          {selectedKeyUsage.month_requests.toLocaleString()}
                        </div>
                        <div className="text-sm text-indigo-700">This Month</div>
                      </div>
                    </div>
                  </div>

                  {/* Performance Metrics */}
                  <div>
                    <h4 className="text-sm font-medium text-gray-700 mb-3">Performance Metrics</h4>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                      <div className="bg-green-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-green-900">
                          {selectedKeyUsage.success_rate.toFixed(1)}%
                        </div>
                        <div className="text-sm text-green-700">Success Rate</div>
                        <div className="text-xs text-green-600 mt-1">
                          {selectedKeyUsage.successful_requests.toLocaleString()} / {selectedKeyUsage.total_requests.toLocaleString()}
                        </div>
                      </div>
                      <div className="bg-red-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-red-900">
                          {selectedKeyUsage.failed_requests.toLocaleString()}
                        </div>
                        <div className="text-sm text-red-700">Failed Requests</div>
                        <div className="text-xs text-red-600 mt-1">
                          {((selectedKeyUsage.failed_requests / Math.max(selectedKeyUsage.total_requests, 1)) * 100).toFixed(1)}% failure rate
                        </div>
                      </div>
                      <div className="bg-yellow-50 p-4 rounded-lg">
                        <div className="text-2xl font-bold text-yellow-900">
                          {selectedKeyUsage.avg_response_time.toFixed(0)}ms
                        </div>
                        <div className="text-sm text-yellow-700">Avg Response Time</div>
                      </div>
                    </div>
                  </div>

                  {/* Last Request */}
                  {selectedKeyUsage.last_request_at && (
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <div className="text-sm text-gray-600">Last Request</div>
                      <div className="text-lg font-medium text-gray-900">
                        {formatDate(selectedKeyUsage.last_request_at)}
                      </div>
                    </div>
                  )}
                </>
              ) : null}

              <div className="flex justify-end">
                <button
                  onClick={() => {
                    setSelectedKeyUsage(null);
                    setUsageError(null);
                  }}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md shadow-sm hover:bg-gray-50 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-primary-500"
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
