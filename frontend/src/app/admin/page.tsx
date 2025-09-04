'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { 
  UserGroupIcon,
  LinkIcon,
  ChartBarIcon,
  CogIcon,
  ExclamationTriangleIcon,
  CheckCircleIcon,
  XMarkIcon,
  EyeIcon,
  TrashIcon,
  ShieldCheckIcon,
  ServerIcon,
  ClockIcon,
  DocumentTextIcon,
  KeyIcon,
} from '@heroicons/react/24/outline';
import CMSManager from '@/components/cms/CMSManager';
import APIKeyManager from '@/components/api-keys/APIKeyManager';

interface SystemStats {
  total_users: number;
  active_users: number;
  total_urls: number;
  total_clicks: number;
  storage_used: string;
  daily_signups: number;
  daily_urls_created: number;
  system_status: 'healthy' | 'warning' | 'error';
}

interface User {
  id: number;
  name: string;
  email: string;
  is_active: boolean;
  is_email_verified: boolean;
  account_type: string;
  created_at: string;
  last_login_at?: string;
  urls_count: number;
  total_clicks: number;
}

interface RecentActivity {
  id: number;
  user_id: number;
  user_name: string;
  activity_type: string;
  description: string;
  created_at: string;
  ip_address?: string;
}

export default function AdminPage() {
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const [stats, setStats] = useState<SystemStats | null>(null);
  const [users, setUsers] = useState<User[]>([]);
  const [recentActivity, setRecentActivity] = useState<RecentActivity[]>([]);
  const [loadingStats, setLoadingStats] = useState(true);
  const [loadingUsers, setLoadingUsers] = useState(true);
  const [loadingActivity, setLoadingActivity] = useState(true);
  const [activeTab, setActiveTab] = useState('overview');
  const router = useRouter();

  // Check if user is admin
  const isAdmin = user?.account_type === 'admin' || user?.email === 'admin@urlshorter.com';

  // Redirect if not authenticated or not admin
  useEffect(() => {
    if (!isLoading && (!isAuthenticated || !isAdmin)) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, isLoading, isAdmin, router]);

  // Fetch admin data
  useEffect(() => {
    if (!isAuthenticated || !isAdmin) return;

    const fetchAdminData = async () => {
      const token = getAccessToken();
      if (!token) return;

      const headers = {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      };

      try {
        // Fetch system stats
        setLoadingStats(true);
        const statsResponse = await fetch('http://localhost:8080/api/v1/admin/stats', { headers });
        if (statsResponse.ok) {
          const statsData = await statsResponse.json();
          setStats(statsData);
        }
        setLoadingStats(false);

        // Fetch users
        setLoadingUsers(true);
        const usersResponse = await fetch('http://localhost:8080/api/v1/admin/users?limit=100', { headers });
        if (usersResponse.ok) {
          const usersData = await usersResponse.json();
          setUsers(usersData.users || []);
        }
        setLoadingUsers(false);

        // Fetch recent activity
        setLoadingActivity(true);
        const activityResponse = await fetch('http://localhost:8080/api/v1/admin/activity?limit=50', { headers });
        if (activityResponse.ok) {
          const activityData = await activityResponse.json();
          setRecentActivity(activityData.activities || []);
        }
        setLoadingActivity(false);

      } catch (error) {
        console.error('Error fetching admin data:', error);
        setLoadingStats(false);
        setLoadingUsers(false);
        setLoadingActivity(false);
      }
    };

    fetchAdminData();
  }, [isAuthenticated, isAdmin, getAccessToken]);

  const toggleUserStatus = async (userId: number, isActive: boolean) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/admin/users/${userId}/status`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ is_active: !isActive }),
      });

      if (response.ok) {
        // Update local state
        setUsers(users.map(user => 
          user.id === userId ? { ...user, is_active: !isActive } : user
        ));
      }
    } catch (error) {
      console.error('Error updating user status:', error);
    }
  };

  const deleteUser = async (userId: number) => {
    if (!confirm('Are you sure you want to delete this user? This action cannot be undone.')) {
      return;
    }

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/admin/users/${userId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        // Remove from local state
        setUsers(users.filter(user => user.id !== userId));
      }
    } catch (error) {
      console.error('Error deleting user:', error);
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

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'healthy': return 'text-green-600 bg-green-100';
      case 'warning': return 'text-yellow-600 bg-yellow-100';
      case 'error': return 'text-red-600 bg-red-100';
      default: return 'text-gray-600 bg-gray-100';
    }
  };

  const getActivityIcon = (activityType: string) => {
    switch (activityType) {
      case 'user_login': return <UserGroupIcon className="h-4 w-4" />;
      case 'url_created': return <LinkIcon className="h-4 w-4" />;
      case 'url_clicked': return <EyeIcon className="h-4 w-4" />;
      default: return <ClockIcon className="h-4 w-4" />;
    }
  };

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  if (!isAuthenticated || !isAdmin) {
    return null; // Will redirect
  }

  const tabs = [
    { id: 'overview', name: 'Overview', icon: ChartBarIcon },
    { id: 'users', name: 'Users', icon: UserGroupIcon },
    { id: 'cms', name: 'Content', icon: DocumentTextIcon },
    { id: 'api-keys', name: 'API Keys', icon: KeyIcon },
    { id: 'activity', name: 'Activity', icon: ClockIcon },
    { id: 'system', name: 'System', icon: ServerIcon },
  ];

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <ShieldCheckIcon className="h-8 w-8 text-red-600" />
              <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
            </div>
            <p className="text-gray-600">
              Manage system users, monitor activity, and view system statistics.
            </p>
          </div>

          {/* Navigation Tabs */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
            <div className="border-b border-gray-200">
              <nav className="flex space-x-8 px-6" aria-label="Tabs">
                {tabs.map((tab) => {
                  const Icon = tab.icon;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors ${
                        activeTab === tab.id
                          ? 'border-primary-500 text-primary-600'
                          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                      }`}
                    >
                      <Icon className="h-5 w-5" />
                      <span>{tab.name}</span>
                    </button>
                  );
                })}
              </nav>
            </div>
          </div>

          {/* Tab Content */}
          {activeTab === 'overview' && (
            <div className="space-y-6">
              {/* System Stats Cards */}
              {loadingStats ? (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                  {[1, 2, 3, 4].map((i) => (
                    <div key={i} className="bg-white p-6 rounded-lg shadow-sm border animate-pulse">
                      <div className="h-4 bg-gray-200 rounded w-24 mb-2"></div>
                      <div className="h-8 bg-gray-200 rounded w-16"></div>
                    </div>
                  ))}
                </div>
              ) : stats && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <div className="flex items-center">
                      <UserGroupIcon className="h-8 w-8 text-blue-600" />
                      <div className="ml-4">
                        <p className="text-sm text-gray-600">Total Users</p>
                        <p className="text-2xl font-bold text-black">{stats.total_users}</p>
                      </div>
                    </div>
                  </div>
                  
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <div className="flex items-center">
                      <LinkIcon className="h-8 w-8 text-green-600" />
                      <div className="ml-4">
                        <p className="text-sm text-gray-600">Total URLs</p>
                        <p className="text-2xl font-bold text-black">{stats.total_urls}</p>
                      </div>
                    </div>
                  </div>
                  
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <div className="flex items-center">
                      <ChartBarIcon className="h-8 w-8 text-purple-600" />
                      <div className="ml-4">
                        <p className="text-sm text-gray-600">Total Clicks</p>
                        <p className="text-2xl font-bold text-black">{stats.total_clicks}</p>
                      </div>
                    </div>
                  </div>
                  
                  <div className="bg-white p-6 rounded-lg shadow-sm border">
                    <div className="flex items-center">
                      <ServerIcon className="h-8 w-8 text-orange-600" />
                      <div className="ml-4">
                        <p className="text-sm text-gray-600">System Status</p>
                        <div className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium capitalize ${getStatusColor(stats.system_status)}`}>
                          {stats.system_status}
                        </div>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Quick Stats */}
              {stats && (
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                  <h3 className="text-lg font-medium text-gray-900 mb-4">Today's Activity</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="text-center">
                      <p className="text-2xl font-bold text-blue-600">{stats.daily_signups}</p>
                      <p className="text-sm text-gray-600">New Signups</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-green-600">{stats.daily_urls_created}</p>
                      <p className="text-sm text-gray-600">URLs Created</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-orange-600">{stats.active_users}</p>
                      <p className="text-sm text-gray-600">Active Users</p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'users' && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">User Management</h2>
              </div>
              
              {loadingUsers ? (
                <div className="p-6">
                  <div className="space-y-4">
                    {[1, 2, 3, 4, 5].map((i) => (
                      <div key={i} className="animate-pulse">
                        <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                        <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-gray-50">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          User
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Account
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          URLs/Clicks
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Last Login
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {users.map((user) => (
                        <tr key={user.id} className="hover:bg-gray-50">
                          <td className="px-6 py-4">
                            <div>
                              <div className="text-sm font-medium text-black">{user.name}</div>
                              <div className="text-sm text-gray-600">{user.email}</div>
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex items-center space-x-2">
                              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${
                                user.is_active ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'
                              }`}>
                                {user.is_active ? 'Active' : 'Inactive'}
                              </span>
                              {user.is_email_verified && (
                                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                              )}
                              <span className="text-xs text-gray-500 capitalize">{user.account_type}</span>
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-black">
                              {user.urls_count} URLs, {user.total_clicks} clicks
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-gray-600">
                              {user.last_login_at ? formatDate(user.last_login_at) : 'Never'}
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex items-center space-x-2">
                              <button
                                onClick={() => toggleUserStatus(user.id, user.is_active)}
                                className={`p-1 rounded transition-colors ${
                                  user.is_active 
                                    ? 'text-red-600 hover:text-red-800' 
                                    : 'text-green-600 hover:text-green-800'
                                }`}
                                title={user.is_active ? 'Deactivate user' : 'Activate user'}
                              >
                                {user.is_active ? (
                                  <XMarkIcon className="h-4 w-4" />
                                ) : (
                                  <CheckCircleIcon className="h-4 w-4" />
                                )}
                              </button>
                              <button
                                onClick={() => deleteUser(user.id)}
                                className="p-1 text-red-600 hover:text-red-800 transition-colors"
                                title="Delete user"
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
          )}

          {activeTab === 'activity' && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200">
              <div className="px-6 py-4 border-b">
                <h2 className="text-xl font-semibold text-black">Recent Activity</h2>
              </div>
              
              {loadingActivity ? (
                <div className="p-6">
                  <div className="space-y-4">
                    {[1, 2, 3, 4, 5].map((i) => (
                      <div key={i} className="animate-pulse">
                        <div className="h-4 bg-gray-200 rounded w-full mb-2"></div>
                        <div className="h-4 bg-gray-200 rounded w-3/4"></div>
                      </div>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="p-6">
                  <div className="space-y-4">
                    {recentActivity.map((activity) => (
                      <div key={activity.id} className="flex items-start space-x-3 p-3 bg-gray-50 rounded-lg">
                        <div className="flex-shrink-0 p-1 bg-white rounded-full">
                          {getActivityIcon(activity.activity_type)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-black">
                            <span className="font-medium">{activity.user_name}</span>{' '}
                            {activity.description}
                          </p>
                          <div className="flex items-center space-x-2 mt-1">
                            <p className="text-xs text-gray-500">
                              {formatDate(activity.created_at)}
                            </p>
                            {activity.ip_address && (
                              <span className="text-xs text-gray-400">
                                IP: {activity.ip_address}
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'cms' && (
            <CMSManager />
          )}

          {activeTab === 'api-keys' && (
            <APIKeyManager />
          )}

          {activeTab === 'system' && (
            <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
              <h2 className="text-xl font-semibold text-black mb-4">System Information</h2>
              
              {stats && (
                <div className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <h3 className="text-md font-medium text-black mb-3">System Status</h3>
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-600">Overall Status</span>
                          <span className={`text-sm font-medium capitalize ${
                            stats.system_status === 'healthy' ? 'text-green-600' :
                            stats.system_status === 'warning' ? 'text-yellow-600' : 'text-red-600'
                          }`}>
                            {stats.system_status}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-gray-600">Storage Used</span>
                          <span className="text-sm font-medium text-black">{stats.storage_used}</span>
                        </div>
                      </div>
                    </div>

                    <div>
                      <h3 className="text-md font-medium text-black mb-3">Quick Actions</h3>
                      <div className="space-y-2">
                        <button className="w-full text-left px-3 py-2 text-sm text-blue-600 hover:bg-blue-50 rounded">
                          Clear System Cache
                        </button>
                        <button className="w-full text-left px-3 py-2 text-sm text-green-600 hover:bg-green-50 rounded">
                          Export System Logs
                        </button>
                        <button className="w-full text-left px-3 py-2 text-sm text-orange-600 hover:bg-orange-50 rounded">
                          Run System Diagnostics
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}