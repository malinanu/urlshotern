'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS, buildApiUrl } from '@/lib/api-config';
import { StatsSkeleton, UserTableSkeleton, ActivitySkeleton } from '@/components/skeletons/AdminSkeletons';
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
  MagnifyingGlassIcon,
  CubeIcon,
  PlusIcon,
  TicketIcon,
  UsersIcon,
  GlobeAltIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import CMSManager from '@/components/cms/CMSManager';
import APIKeyManager from '@/components/api-keys/APIKeyManager';
import { SEOToolsManager } from '@/components/seo/SEOToolsManager';
import { VisualPageBuilder } from '@/components/builder/VisualPageBuilder';
import { CreateUserModal } from '@/components/admin/CreateUserModal';
import { UserRolesModal } from '@/components/admin/UserRolesModal';
import { PLAN_NAMES } from '@/lib/plan-utils';

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
  is_admin: boolean;
  is_email_verified: boolean;
  account_type: string;
  created_at: string;
  last_login_at?: string;
  urls_count: number;
  total_clicks: number;
  roles: string[];
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

interface AdminDomain {
  id: number;
  domain: string;
  is_active: boolean;
  is_verified: boolean;
  ssl_enabled: boolean;
  dns_status: string;
  health_status: string;
  created_at: string;
}

interface DomainStats {
  total_domains: number;
  active_domains: number;
  verified_domains: number;
  ssl_domains: number;
  pending_domains: number;
  failed_domains: number;
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
  const [adminDomains, setAdminDomains] = useState<AdminDomain[]>([]);
  const [domainStats, setDomainStats] = useState<DomainStats | null>(null);
  const [loadingDomains, setLoadingDomains] = useState(false);
  const [showCreateUserModal, setShowCreateUserModal] = useState(false);
  const [selectedUserId, setSelectedUserId] = useState<number | null>(null);
  const [selectedUserName, setSelectedUserName] = useState<string>('');
  const [showRolesModal, setShowRolesModal] = useState(false);
  const [modalMode, setModalMode] = useState<'view' | 'manage'>('view');
  const [builderPageId, setBuilderPageId] = useState<number | null>(null);
  const router = useRouter();

  // Check if user is admin
  const isAdmin = user?.is_admin === true;

  // Redirect if not authenticated or not admin
  useEffect(() => {
    if (!isLoading && (!isAuthenticated || !isAdmin)) {
      router.push('/dashboard');
    }
  }, [isAuthenticated, isLoading, isAdmin, router]);

  // Fetch admin data function (can be called from modals)
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
      const statsResponse = await fetch(API_ENDPOINTS.admin.stats, { headers });
      if (statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      }
      setLoadingStats(false);

      // Fetch users
      setLoadingUsers(true);
      const usersResponse = await fetch(buildApiUrl(API_ENDPOINTS.admin.users, { limit: 100 }), { headers });
      if (usersResponse.ok) {
        const usersData = await usersResponse.json();
        setUsers(usersData.users || []);
      }
      setLoadingUsers(false);

      // Fetch recent activity
      setLoadingActivity(true);
      const activityResponse = await fetch(buildApiUrl(API_ENDPOINTS.admin.activity, { limit: 50 }), { headers });
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

  // Fetch admin data on mount
  useEffect(() => {
    if (!isAuthenticated || !isAdmin) return;
    fetchAdminData();
  }, [isAuthenticated, isAdmin]);

  const fetchDomainData = async () => {
    const token = getAccessToken();
    if (!token) return;
    setLoadingDomains(true);
    try {
      const headers = { 'Authorization': `Bearer ${token}`, 'Content-Type': 'application/json' };
      const [domainsRes, statsRes] = await Promise.all([
        fetch(API_ENDPOINTS.admin.domains.list, { headers }),
        fetch(API_ENDPOINTS.admin.domains.stats, { headers }),
      ]);
      if (domainsRes.ok) {
        const data = await domainsRes.json();
        setAdminDomains(data.domains ?? []);
      }
      if (statsRes.ok) {
        const data = await statsRes.json();
        setDomainStats(data.stats ?? null);
      }
    } catch (error) {
      console.error('Error fetching domain data:', error);
    } finally {
      setLoadingDomains(false);
    }
  };

  useEffect(() => {
    if (activeTab === 'domains' && isAuthenticated && isAdmin) {
      fetchDomainData();
    }
  }, [activeTab, isAuthenticated, isAdmin]);

  const toggleUserStatus = async (userId: number, isActive: boolean) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(API_ENDPOINTS.admin.userStatus(userId.toString()), {
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
      const response = await fetch(API_ENDPOINTS.admin.deleteUser(userId.toString()), {
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
      case 'healthy': return 'text-green-600 dark:text-green-500 bg-green-500/10';
      case 'warning': return 'text-amber-600 dark:text-amber-500 bg-amber-500/10';
      case 'error': return 'text-destructive bg-destructive/10';
      default: return 'text-muted-foreground bg-secondary';
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
        <div className="min-h-[80vh] flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
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
    { id: 'analytics', name: 'Analytics', icon: LinkIcon },
    { id: 'domains', name: 'Domains', icon: GlobeAltIcon },
    { id: 'cms', name: 'Content', icon: DocumentTextIcon },
    { id: 'builder', name: 'Page Builder', icon: CubeIcon },
    { id: 'seo', name: 'SEO Tools', icon: MagnifyingGlassIcon },
    { id: 'api-keys', name: 'API Keys', icon: KeyIcon },
    { id: 'coupons', name: 'Coupons', icon: TicketIcon },
    { id: 'activity', name: 'Activity', icon: ClockIcon },
    { id: 'system', name: 'System', icon: ServerIcon },
  ];

  return (
    <Layout>
      <div className="min-h-[80vh] bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <ShieldCheckIcon className="h-8 w-8 text-destructive" />
              <h1 className="text-3xl font-bold text-foreground">Admin Dashboard</h1>
            </div>
            <p className="text-muted-foreground">
              Manage system users, monitor activity, and view system statistics.
            </p>
          </div>

          {/* Navigation Tabs */}
          <div className="bg-card rounded-lg shadow-sm border border-border mb-6">
            <div className="border-b border-border">
              <nav className="flex space-x-8 px-6" aria-label="Tabs">
                {tabs.map((tab) => {
                  const Icon = tab.icon;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={`flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors ${activeTab === tab.id
                        ? 'border-primary text-primary'
                        : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
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
                <StatsSkeleton />
              ) : stats && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-6">
                  <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                    <div className="flex items-center">
                      <UserGroupIcon className="h-8 w-8 text-blue-600 dark:text-blue-500" />
                      <div className="ml-4">
                        <p className="text-sm text-muted-foreground">Total Users</p>
                        <p className="text-2xl font-bold text-foreground">{stats.total_users}</p>
                      </div>
                    </div>
                  </div>

                  <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                    <div className="flex items-center">
                      <LinkIcon className="h-8 w-8 text-green-600 dark:text-green-500" />
                      <div className="ml-4">
                        <p className="text-sm text-muted-foreground">Total URLs</p>
                        <p className="text-2xl font-bold text-foreground">{stats.total_urls}</p>
                      </div>
                    </div>
                  </div>

                  <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                    <div className="flex items-center">
                      <ChartBarIcon className="h-8 w-8 text-purple-600 dark:text-purple-500" />
                      <div className="ml-4">
                        <p className="text-sm text-muted-foreground">Total Clicks</p>
                        <p className="text-2xl font-bold text-foreground">{stats.total_clicks}</p>
                      </div>
                    </div>
                  </div>

                  <div className="bg-card p-6 rounded-lg shadow-sm border border-border">
                    <div className="flex items-center">
                      <ServerIcon className="h-8 w-8 text-amber-600 dark:text-amber-500" />
                      <div className="ml-4">
                        <p className="text-sm text-muted-foreground">System Status</p>
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
                <div className="bg-card rounded-lg shadow-sm border border-border p-6">
                  <h3 className="text-lg font-medium text-foreground mb-4">Today's Activity</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                    <div className="text-center">
                      <p className="text-2xl font-bold text-blue-600 dark:text-blue-500">{stats.daily_signups}</p>
                      <p className="text-sm text-muted-foreground">New Signups</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-green-600 dark:text-green-500">{stats.daily_urls_created}</p>
                      <p className="text-sm text-muted-foreground">URLs Created</p>
                    </div>
                    <div className="text-center">
                      <p className="text-2xl font-bold text-amber-600 dark:text-amber-500">{stats.active_users}</p>
                      <p className="text-sm text-muted-foreground">Active Users</p>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === 'users' && (
            <div className="bg-card rounded-lg shadow-sm border border-border">
              <div className="px-6 py-4 border-b border-border flex items-center justify-between">
                <h2 className="text-xl font-semibold text-foreground">User Management</h2>
                <button
                  onClick={() => setShowCreateUserModal(true)}
                  className="bg-primary hover:bg-primary/90 text-primary-foreground px-4 py-2 rounded-lg flex items-center space-x-2 transition-colors"
                >
                  <PlusIcon className="h-5 w-5" />
                  <span>Add User</span>
                </button>
              </div>

              {loadingUsers ? (
                <div className="p-6">
                  <UserTableSkeleton />
                </div>
              ) : (
                <div className="overflow-x-auto">
                  <table className="w-full">
                    <thead className="bg-muted">
                      <tr>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          User
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Account
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Roles
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          URLs/Clicks
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Last Login
                        </th>
                        <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">
                          Actions
                        </th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-gray-200">
                      {users.map((user) => (
                        <tr key={user.id} className="hover:bg-muted">
                          <td className="px-6 py-4">
                            <div>
                              <div className="text-sm font-medium text-foreground">{user.name}</div>
                              <div className="text-sm text-muted-foreground">{user.email}</div>
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex items-center space-x-2">
                              <span className={`inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium ${user.is_active ? 'bg-green-500/10 text-green-600 dark:text-green-500' : 'bg-destructive/10 text-destructive'
                                }`}>
                                {user.is_active ? 'Active' : 'Inactive'}
                              </span>
                              {user.is_email_verified && (
                                <CheckCircleIcon className="h-4 w-4 text-green-600 dark:text-green-500" />
                              )}
                              <span className="text-xs text-muted-foreground capitalize">{user.account_type}</span>
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex flex-wrap gap-1">
                              {user.is_admin && (
                                <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-red-500/10 text-red-600 dark:text-red-400">
                                  Admin
                                </span>
                              )}
                              {(user.roles ?? []).map((role) => (
                                <span
                                  key={role}
                                  className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary"
                                >
                                  {role}
                                </span>
                              ))}
                              {!user.is_admin && (user.roles ?? []).length === 0 && (
                                <span className="text-xs text-muted-foreground italic">No roles</span>
                              )}
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-foreground">
                              {user.urls_count} URLs, {user.total_clicks} clicks
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="text-sm text-muted-foreground">
                              {user.last_login_at ? formatDate(user.last_login_at) : 'Never'}
                            </div>
                          </td>
                          <td className="px-6 py-4">
                            <div className="flex items-center space-x-2">
                              <button
                                onClick={() => {
                                  setModalMode('manage');
                                  setSelectedUserId(user.id);
                                  setSelectedUserName(user.name);
                                  setShowRolesModal(true);
                                }}
                                aria-label="Manage user roles"
                                className="p-1 text-primary hover:text-primary/80 transition-colors"
                                title="Manage roles"
                              >
                                <ShieldCheckIcon className="h-4 w-4" />
                              </button>
                              <button
                                onClick={() => toggleUserStatus(user.id, user.is_active)}
                                aria-label={user.is_active ? 'Deactivate user' : 'Activate user'}
                                className={`p-1 rounded transition-colors ${user.is_active
                                  ? 'text-destructive hover:text-destructive/80'
                                  : 'text-green-600 dark:text-green-500 hover:text-green-700 dark:hover:text-green-400'
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
                                aria-label="Delete user"
                                className="p-1 text-destructive hover:text-destructive/80 transition-colors"
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
            <div className="bg-card rounded-lg shadow-sm border border-border">
              <div className="px-6 py-4 border-b border-border">
                <h2 className="text-xl font-semibold text-foreground">Recent Activity</h2>
              </div>

              {loadingActivity ? (
                <div className="p-6">
                  <ActivitySkeleton />
                </div>
              ) : (
                <div className="p-6">
                  <div className="space-y-4">
                    {recentActivity.map((activity) => (
                      <div key={activity.id} className="flex items-start space-x-3 p-3 bg-muted rounded-lg">
                        <div className="flex-shrink-0 p-1 bg-card rounded-full">
                          {getActivityIcon(activity.activity_type)}
                        </div>
                        <div className="flex-1 min-w-0">
                          <p className="text-sm text-foreground">
                            <span className="font-medium">{activity.user_name}</span>{' '}
                            {activity.description}
                          </p>
                          <div className="flex items-center space-x-2 mt-1">
                            <p className="text-xs text-muted-foreground">
                              {formatDate(activity.created_at)}
                            </p>
                            {activity.ip_address && (
                              <span className="text-xs text-muted-foreground/50">
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

          {activeTab === 'analytics' && stats && (
            <div className="space-y-6">
              {/* Platform KPIs */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                {[
                  { label: 'Total Users', value: stats.total_users, icon: UsersIcon, color: 'text-blue-600 dark:text-blue-400' },
                  { label: 'Active Users', value: stats.active_users, icon: UserGroupIcon, color: 'text-green-600 dark:text-green-400' },
                  { label: 'Total URLs', value: stats.total_urls, icon: LinkIcon, color: 'text-purple-600 dark:text-purple-400' },
                  { label: 'Total Clicks', value: stats.total_clicks, icon: ChartBarIcon, color: 'text-amber-600 dark:text-amber-400' },
                ].map(({ label, value, icon: Icon, color }) => (
                  <div key={label} className="bg-card rounded-lg border border-border p-5">
                    <div className="flex items-center gap-3">
                      <Icon className={`h-7 w-7 ${color}`} />
                      <div>
                        <p className="text-sm text-muted-foreground">{label}</p>
                        <p className="text-2xl font-bold text-foreground">{value?.toLocaleString() ?? '—'}</p>
                      </div>
                    </div>
                  </div>
                ))}
              </div>

              {/* Today's Activity */}
              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="bg-card rounded-lg border border-border p-5">
                  <h3 className="font-semibold text-foreground mb-3">Today's Activity</h3>
                  <div className="space-y-2 text-sm">
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">New Signups</span>
                      <span className="font-medium text-foreground">{stats.daily_signups}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-muted-foreground">URLs Created</span>
                      <span className="font-medium text-foreground">{stats.daily_urls_created}</span>
                    </div>
                  </div>
                </div>
                <div className="bg-card rounded-lg border border-border p-5">
                  <h3 className="font-semibold text-foreground mb-3">Plan Distribution</h3>
                  {users.length > 0 ? (
                    <div className="space-y-2 text-sm">
                      {(['free', 'premium', 'enterprise'] as const).map(plan => {
                        const count = users.filter(u => u.account_type === plan).length;
                        const pct = users.length > 0 ? Math.round((count / users.length) * 100) : 0;
                        const displayName = plan === 'free' ? PLAN_NAMES.free : plan === 'premium' ? PLAN_NAMES.premium : PLAN_NAMES.enterprise;
                        return (
                          <div key={plan} className="flex items-center gap-3">
                            <span className="w-10 text-muted-foreground">{displayName}</span>
                            <div className="flex-1 bg-muted rounded-full h-2">
                              <div
                                className="bg-primary h-2 rounded-full"
                                style={{ width: `${pct}%` }}
                              />
                            </div>
                            <span className="w-8 text-right font-medium text-foreground">{count}</span>
                          </div>
                        );
                      })}
                    </div>
                  ) : (
                    <p className="text-sm text-muted-foreground">Load the Users tab first to see plan data.</p>
                  )}
                </div>
              </div>

              {/* System Health */}
              <div className="bg-card rounded-lg border border-border p-5">
                <h3 className="font-semibold text-foreground mb-3">System Health</h3>
                <div className="flex items-center gap-3">
                  <span className={`inline-flex items-center gap-1.5 text-sm font-medium capitalize px-3 py-1 rounded-full ${
                    stats.system_status === 'healthy'
                      ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                      : stats.system_status === 'warning'
                      ? 'bg-amber-100 text-amber-700 dark:bg-amber-900/30 dark:text-amber-400'
                      : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                  }`}>
                    <span className={`h-2 w-2 rounded-full ${
                      stats.system_status === 'healthy' ? 'bg-green-500' :
                      stats.system_status === 'warning' ? 'bg-amber-500' : 'bg-red-500'
                    }`} />
                    {stats.system_status}
                  </span>
                  <span className="text-sm text-muted-foreground">Storage: {stats.storage_used}</span>
                </div>
              </div>
            </div>
          )}

          {activeTab === 'domains' && (
            <div className="space-y-6">
              {/* Domain Stats */}
              {loadingDomains ? (
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                  {[1, 2, 3, 4, 5, 6].map((i) => (
                    <div key={i} className="bg-card border border-border rounded-lg p-5 animate-pulse">
                      <div className="h-4 bg-muted rounded w-24 mb-2" />
                      <div className="h-8 bg-muted rounded w-16" />
                    </div>
                  ))}
                </div>
              ) : domainStats && (
                <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
                  {[
                    { label: 'Total Domains', value: domainStats.total_domains, color: 'text-blue-600 dark:text-blue-400' },
                    { label: 'Active', value: domainStats.active_domains, color: 'text-green-600 dark:text-green-400' },
                    { label: 'Verified', value: domainStats.verified_domains, color: 'text-primary' },
                    { label: 'SSL Enabled', value: domainStats.ssl_domains, color: 'text-purple-600 dark:text-purple-400' },
                    { label: 'Pending', value: domainStats.pending_domains, color: 'text-amber-600 dark:text-amber-400' },
                    { label: 'Failed', value: domainStats.failed_domains, color: 'text-destructive' },
                  ].map(({ label, value, color }) => (
                    <div key={label} className="bg-card border border-border rounded-lg p-5">
                      <p className="text-sm text-muted-foreground">{label}</p>
                      <p className={`text-2xl font-bold mt-1 ${color}`}>{value}</p>
                    </div>
                  ))}
                </div>
              )}

              {/* Domain List */}
              <div className="bg-card border border-border rounded-lg">
                <div className="px-6 py-4 border-b border-border flex items-center justify-between">
                  <h3 className="font-semibold text-foreground">All Custom Domains</h3>
                  <button
                    onClick={fetchDomainData}
                    className="flex items-center space-x-1.5 text-sm text-muted-foreground hover:text-foreground transition-colors"
                  >
                    <ArrowPathIcon className="h-4 w-4" />
                    <span>Refresh</span>
                  </button>
                </div>

                {loadingDomains ? (
                  <div className="p-6 space-y-3">
                    {[1, 2, 3].map((i) => (
                      <div key={i} className="h-12 bg-muted rounded animate-pulse" />
                    ))}
                  </div>
                ) : adminDomains.length === 0 ? (
                  <div className="p-12 text-center">
                    <GlobeAltIcon className="h-10 w-10 text-muted-foreground mx-auto mb-3" />
                    <p className="text-muted-foreground text-sm">No custom domains registered on the platform.</p>
                  </div>
                ) : (
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm">
                      <thead className="bg-muted">
                        <tr>
                          <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Domain</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Status</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">DNS</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Health</th>
                          <th className="px-6 py-3 text-left text-xs font-medium text-muted-foreground uppercase tracking-wider">Added</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border">
                        {adminDomains.map((d) => (
                          <tr key={d.id} className="hover:bg-muted/50 transition-colors">
                            <td className="px-6 py-4 font-mono text-foreground">{d.domain}</td>
                            <td className="px-6 py-4">
                              <div className="flex items-center space-x-1.5">
                                {d.is_verified ? (
                                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-500/10 text-green-600">Verified</span>
                                ) : (
                                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-yellow-500/10 text-yellow-600">Pending</span>
                                )}
                                {d.ssl_enabled && (
                                  <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary">SSL</span>
                                )}
                              </div>
                            </td>
                            <td className="px-6 py-4 text-muted-foreground capitalize">{d.dns_status}</td>
                            <td className="px-6 py-4 text-muted-foreground capitalize">{d.health_status}</td>
                            <td className="px-6 py-4 text-muted-foreground">
                              {new Date(d.created_at).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === 'cms' && (
            <CMSManager onEditInBuilder={(id) => { setBuilderPageId(id); setActiveTab('builder'); }} />
          )}

          {activeTab === 'builder' && (
            <VisualPageBuilder key={builderPageId ?? 'empty'} initialPageId={builderPageId} />
          )}

          {activeTab === 'seo' && (
            <SEOToolsManager />
          )}

          {activeTab === 'api-keys' && (
            <APIKeyManager />
          )}

          {activeTab === 'coupons' && (
            <div className="bg-card rounded-lg shadow-sm border border-border p-6 flex flex-col items-center justify-center min-h-[300px]">
              <TicketIcon className="w-16 h-16 text-muted-foreground/30 mb-4" />
              <h2 className="text-xl font-semibold mb-2 text-foreground">Discount Coupons</h2>
              <p className="text-muted-foreground text-center mb-6 max-w-md">Manage promotional discount codes like free-passes or automated billing integrations.</p>
              <button
                onClick={() => router.push('/admin/coupons')}
                className="bg-primary hover:bg-primary/90 text-primary-foreground px-6 py-2 rounded-lg font-medium transition-colors"
              >
                Go to Coupons Manager
              </button>
            </div>
          )}

          {activeTab === 'system' && (
            <div className="bg-card rounded-lg shadow-sm border border-border p-6">
              <h2 className="text-xl font-semibold text-foreground mb-4">System Information</h2>

              {stats && (
                <div className="space-y-6">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    <div>
                      <h3 className="text-md font-medium text-foreground mb-3">System Status</h3>
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">Overall Status</span>
                          <span className={`text-sm font-medium capitalize ${stats.system_status === 'healthy' ? 'text-green-600 dark:text-green-500' :
                            stats.system_status === 'warning' ? 'text-amber-600 dark:text-amber-500' : 'text-destructive'
                            }`}>
                            {stats.system_status}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-sm text-muted-foreground">Storage Used</span>
                          <span className="text-sm font-medium text-foreground">{stats.storage_used}</span>
                        </div>
                      </div>
                    </div>

                    <div>
                      <h3 className="text-md font-medium text-foreground mb-3">Quick Actions</h3>
                      <div className="space-y-2">
                        <button className="w-full text-left px-3 py-2 text-sm text-primary hover:bg-primary/5 rounded border border-transparent hover:border-primary/20 transition-colors">
                          Clear System Cache
                        </button>
                        <button className="w-full text-left px-3 py-2 text-sm text-green-600 dark:text-green-500 hover:bg-green-500/5 rounded border border-transparent hover:border-green-500/20 transition-colors">
                          Export System Logs
                        </button>
                        <button className="w-full text-left px-3 py-2 text-sm text-amber-600 dark:text-amber-500 hover:bg-amber-500/5 rounded border border-transparent hover:border-amber-500/20 transition-colors">
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

      {/* Create User Modal */}
      <CreateUserModal
        isOpen={showCreateUserModal}
        onClose={() => setShowCreateUserModal(false)}
        onUserCreated={() => {
          setShowCreateUserModal(false);
          fetchAdminData();
        }}
      />

      {/* User Roles Modal */}
      {selectedUserId && (
        <UserRolesModal
          isOpen={showRolesModal}
          onClose={() => {
            setShowRolesModal(false);
            setSelectedUserId(null);
            setSelectedUserName('');
          }}
          userId={selectedUserId}
          userName={selectedUserName}
          mode={modalMode}
        />
      )}
    </Layout>
  );
}