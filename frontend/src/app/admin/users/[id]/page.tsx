'use client';

import { useEffect, useState } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { adminAuthService } from '@/lib/admin-auth-service';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  ArrowLeftIcon,
  UserCircleIcon,
  LinkIcon,
  CursorArrowRaysIcon,
  CalendarIcon,
  ClockIcon,
  ChartBarIcon,
  ExclamationCircleIcon,
} from '@heroicons/react/24/outline';
import { CheckCircleIcon } from '@heroicons/react/24/solid';

interface User {
  id: number;
  name: string;
  email: string;
  account_type: string;
  is_active: boolean;
  email_verified: boolean;
  phone_verified: boolean;
  created_at: string;
  last_login_at?: string;
}

interface UserStats {
  total_urls: number;
  total_clicks: number;
  avg_clicks_per_url: number;
  active_urls: number;
  expired_urls: number;
}

interface URLWithClicks {
  id: number;
  short_code: string;
  original_url: string;
  click_count: number;
  created_at: string;
  expires_at?: string;
}

interface ActivityLog {
  id: number;
  action: string;
  resource_type?: string;
  created_at: string;
  ip_address?: string;
}

interface TimeSeriesDataPoint {
  date: string;
  count: number;
}

interface UserDetailedStats {
  user: User;
  stats: UserStats;
  top_urls: URLWithClicks[];
  recent_activity: ActivityLog[];
  urls_over_time: TimeSeriesDataPoint[];
  clicks_over_time: TimeSeriesDataPoint[];
}

export default function UserDetailPage() {
  const params = useParams();
  const router = useRouter();
  const userId = params.id as string;

  const [userStats, setUserStats] = useState<UserDetailedStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchUserStats = async () => {
      const token = adminAuthService.getAdminToken();
      if (!token) {
        router.push('/admin/login');
        return;
      }

      try {
        setLoading(true);
        setError(null);

        const response = await fetch(`${API_ENDPOINTS.admin.users}/${userId}/stats`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (!response.ok) {
          throw new Error(`Failed to fetch user stats: ${response.statusText}`);
        }

        const data = await response.json();
        setUserStats(data);
      } catch (err) {
        console.error('Error fetching user stats:', err);
        setError(err instanceof Error ? err.message : 'An error occurred');
      } finally {
        setLoading(false);
      }
    };

    if (userId) {
      fetchUserStats();
    }
  }, [userId, router]);

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const formatShortDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
    });
  };

  const truncateURL = (url: string, maxLength: number = 50) => {
    if (url.length <= maxLength) return url;
    return url.substring(0, maxLength) + '...';
  };

  if (loading) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
          <div className="text-center">
            <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600 mx-auto"></div>
            <p className="mt-4 text-gray-600">Loading user details...</p>
          </div>
        </div>
      </Layout>
    );
  }

  if (error || !userStats) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
          <div className="text-center">
            <ExclamationCircleIcon className="h-12 w-12 text-red-600 mx-auto" />
            <p className="mt-4 text-gray-900 font-medium">Failed to load user details</p>
            <p className="text-gray-600">{error}</p>
            <button
              onClick={() => router.push('/admin')}
              className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700"
            >
              Back to Admin Dashboard
            </button>
          </div>
        </div>
      </Layout>
    );
  }

  const { user, stats, top_urls, recent_activity, urls_over_time, clicks_over_time } = userStats;

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header with Back Button */}
          <div className="mb-6">
            <button
              onClick={() => router.push('/admin')}
              className="flex items-center text-gray-600 hover:text-gray-900 mb-4"
            >
              <ArrowLeftIcon className="h-5 w-5 mr-2" />
              Back to Admin Dashboard
            </button>

            <div className="flex items-center justify-between">
              <div className="flex items-center space-x-4">
                <UserCircleIcon className="h-12 w-12 text-gray-400" />
                <div>
                  <h1 className="text-3xl font-bold text-gray-900">{user.name}</h1>
                  <p className="text-gray-600">{user.email}</p>
                </div>
              </div>

              <div className="flex items-center space-x-2">
                <span className={`px-3 py-1 rounded-full text-sm font-medium ${
                  user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                }`}>
                  {user.is_active ? 'Active' : 'Inactive'}
                </span>
                {user.email_verified && (
                  <CheckCircleIcon className="h-5 w-5 text-green-600" title="Email Verified" />
                )}
                <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium capitalize">
                  {user.account_type}
                </span>
              </div>
            </div>

            <div className="mt-4 grid grid-cols-2 gap-4 text-sm text-gray-600">
              <div className="flex items-center">
                <CalendarIcon className="h-4 w-4 mr-2" />
                <span>Joined: {formatDate(user.created_at)}</span>
              </div>
              <div className="flex items-center">
                <ClockIcon className="h-4 w-4 mr-2" />
                <span>Last Login: {user.last_login_at ? formatDate(user.last_login_at) : 'Never'}</span>
              </div>
            </div>
          </div>

          {/* Stats Cards */}
          <div className="grid grid-cols-1 md:grid-cols-4 gap-6 mb-8">
            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <LinkIcon className="h-8 w-8 text-blue-600" />
                <div className="ml-4">
                  <p className="text-sm text-gray-600">Total URLs</p>
                  <p className="text-2xl font-bold text-black">{stats.total_urls}</p>
                </div>
              </div>
            </div>

            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <CursorArrowRaysIcon className="h-8 w-8 text-green-600" />
                <div className="ml-4">
                  <p className="text-sm text-gray-600">Total Clicks</p>
                  <p className="text-2xl font-bold text-black">{stats.total_clicks}</p>
                </div>
              </div>
            </div>

            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <ChartBarIcon className="h-8 w-8 text-purple-600" />
                <div className="ml-4">
                  <p className="text-sm text-gray-600">Avg Clicks/URL</p>
                  <p className="text-2xl font-bold text-black">{stats.avg_clicks_per_url.toFixed(1)}</p>
                </div>
              </div>
            </div>

            <div className="bg-white p-6 rounded-lg shadow-sm border">
              <div className="flex items-center">
                <CheckCircleIcon className="h-8 w-8 text-orange-600" />
                <div className="ml-4">
                  <p className="text-sm text-gray-600">Active URLs</p>
                  <p className="text-2xl font-bold text-black">{stats.active_urls}</p>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-8">
            {/* Top URLs */}
            <div className="bg-white rounded-lg shadow-sm border">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">Top URLs by Clicks</h2>
              </div>
              <div className="p-6">
                {top_urls.length > 0 ? (
                  <div className="space-y-3">
                    {top_urls.map((url) => (
                      <div key={url.id} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-black truncate">/{url.short_code}</p>
                          <p className="text-xs text-gray-600 truncate">{truncateURL(url.original_url, 60)}</p>
                        </div>
                        <div className="ml-4 text-right">
                          <p className="text-lg font-bold text-primary-600">{url.click_count}</p>
                          <p className="text-xs text-gray-500">clicks</p>
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-center text-gray-500 py-8">No URLs created yet</p>
                )}
              </div>
            </div>

            {/* Recent Activity */}
            <div className="bg-white rounded-lg shadow-sm border">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">Recent Activity</h2>
              </div>
              <div className="p-6">
                {recent_activity.length > 0 ? (
                  <div className="space-y-3">
                    {recent_activity.map((activity) => (
                      <div key={activity.id} className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                        <div className="flex-1 min-w-0">
                          <p className="text-sm font-medium text-black capitalize">{activity.action.replace(/_/g, ' ')}</p>
                          <p className="text-xs text-gray-600">{formatDate(activity.created_at)}</p>
                          {activity.ip_address && (
                            <p className="text-xs text-gray-500">IP: {activity.ip_address}</p>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-center text-gray-500 py-8">No recent activity</p>
                )}
              </div>
            </div>
          </div>

          {/* Time Series Charts */}
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* URLs Created Over Time */}
            <div className="bg-white rounded-lg shadow-sm border">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">URLs Created (Last 30 Days)</h2>
              </div>
              <div className="p-6">
                {urls_over_time.length > 0 ? (
                  <div className="space-y-2">
                    {urls_over_time.map((dataPoint, index) => (
                      <div key={index} className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">{formatShortDate(dataPoint.date)}</span>
                        <div className="flex items-center flex-1 mx-4">
                          <div className="flex-1 bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-blue-600 h-2 rounded-full"
                              style={{
                                width: `${Math.min(100, (dataPoint.count / Math.max(...urls_over_time.map(d => d.count))) * 100)}%`
                              }}
                            />
                          </div>
                        </div>
                        <span className="text-sm font-medium text-black w-8 text-right">{dataPoint.count}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-center text-gray-500 py-8">No data available</p>
                )}
              </div>
            </div>

            {/* Clicks Over Time */}
            <div className="bg-white rounded-lg shadow-sm border">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">Clicks Over Time (Last 30 Days)</h2>
              </div>
              <div className="p-6">
                {clicks_over_time.length > 0 ? (
                  <div className="space-y-2">
                    {clicks_over_time.map((dataPoint, index) => (
                      <div key={index} className="flex items-center justify-between">
                        <span className="text-sm text-gray-600">{formatShortDate(dataPoint.date)}</span>
                        <div className="flex items-center flex-1 mx-4">
                          <div className="flex-1 bg-gray-200 rounded-full h-2">
                            <div
                              className="bg-green-600 h-2 rounded-full"
                              style={{
                                width: `${Math.min(100, (dataPoint.count / Math.max(...clicks_over_time.map(d => d.count))) * 100)}%`
                              }}
                            />
                          </div>
                        </div>
                        <span className="text-sm font-medium text-black w-8 text-right">{dataPoint.count}</span>
                      </div>
                    ))}
                  </div>
                ) : (
                  <p className="text-center text-gray-500 py-8">No data available</p>
                )}
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}
