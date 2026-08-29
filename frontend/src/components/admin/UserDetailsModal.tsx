'use client';

import { useState } from 'react';
import { Modal } from '@/components/ui/modal';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PLAN_NAMES } from '@/lib/plan-utils';
import {
  UserIcon,
  EnvelopeIcon,
  ShieldCheckIcon,
  ChartBarIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

interface User {
  id: number;
  name: string;
  email: string;
  phone?: string;
  phone_verified?: boolean;
  provider?: string;
  avatar_url?: string;
  is_active: boolean;
  is_admin: boolean;
  is_email_verified?: boolean;
  account_type: string;
  created_at?: string;
  updated_at?: string;
  last_login_at?: string;
  urls_count?: number;
  total_clicks?: number;
  roles?: string[];
}

interface Props {
  isOpen: boolean;
  onClose: () => void;
  user: User | null;
  onUserUpdated: () => void;
  onViewUrls: (userId: number) => void;
}

export function UserDetailsModal({ isOpen, onClose, user, onUserUpdated, onViewUrls }: Props) {
  const { getAccessToken } = useAuth();
  const [loading, setLoading] = useState(false);
  const [actionMessage, setActionMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null);

  if (!user) return null;

  const handleAction = async (actionFn: () => Promise<void>) => {
    setLoading(true);
    setActionMessage(null);
    try {
      await actionFn();
      onUserUpdated();
    } catch (err: any) {
      setActionMessage({ type: 'error', text: err.message || 'An error occurred' });
    } finally {
      setLoading(false);
    }
  };

  const toggleStatus = async () => {
    const token = getAccessToken();
    const res = await fetch(API_ENDPOINTS.admin.userStatus(user.id.toString()), {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ is_active: !user.is_active }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.message || 'Failed to update user status');
    }
    setActionMessage({ type: 'success', text: `User ${!user.is_active ? 'activated' : 'deactivated'} successfully` });
  };

  const sendPasswordReset = async () => {
    const token = getAccessToken();
    const res = await fetch(API_ENDPOINTS.admin.userResetPasswordEmail(user.id.toString()), {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
      },
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.message || 'Failed to send password reset email');
    }
    setActionMessage({ type: 'success', text: 'Password reset email sent successfully' });
  };

  const updateAccountType = async (newType: string) => {
    if (newType === user.account_type) return;
    
    const token = getAccessToken();
    const res = await fetch(API_ENDPOINTS.admin.updateAccountType(user.id.toString()), {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ account_type: newType }),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.message || 'Failed to update account type');
    }
    setActionMessage({ type: 'success', text: 'Account type updated successfully' });
  };

  return (
    <Modal isOpen={isOpen} onClose={onClose} title="User Details" size="lg">
      <div className="space-y-6">
        {/* Profile Info */}
        <div className="bg-muted p-4 rounded-lg flex items-start space-x-4">
          {user.avatar_url ? (
            <img src={user.avatar_url} alt={user.name} className="h-12 w-12 rounded-full border-2 border-primary/20 object-cover" />
          ) : (
            <div className="bg-primary/10 p-3 rounded-full flex-shrink-0">
              <UserIcon className="h-6 w-6 text-primary" />
            </div>
          )}
          <div className="flex-1">
            <h3 className="text-lg font-medium text-foreground">{user.name}</h3>
            <div className="flex flex-wrap items-center text-sm text-muted-foreground mt-1 gap-4">
              <span className="flex items-center">
                <EnvelopeIcon className="h-4 w-4 mr-1" />
                {user.email}
              </span>
              {user.is_admin && (
                <span className="flex items-center text-red-500 font-medium">
                  <ShieldCheckIcon className="h-4 w-4 mr-1" />
                  Admin Access
                </span>
              )}
            </div>
          </div>
          <div>
            <span
              className={`px-3 py-1 rounded-full text-xs font-medium ${
                user.is_active ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400' : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
              }`}
            >
              {user.is_active ? 'Active' : 'Inactive'}
            </span>
          </div>
        </div>

        {/* Account Additional Details */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
          <div className="bg-card border border-border p-4 rounded-lg">
            <h4 className="text-sm font-medium text-muted-foreground mb-3 uppercase tracking-wider">Account Info</h4>
            <ul className="space-y-2 text-sm">
              <li className="flex justify-between">
                <span className="text-muted-foreground">Status:</span>
                <span className="font-medium text-foreground">{user.is_active ? 'Active' : 'Disabled'}</span>
              </li>
              <li className="flex justify-between">
                <span className="text-muted-foreground">Email Verified:</span>
                <span className="font-medium text-foreground">{user.is_email_verified ? 'Yes' : 'No'}</span>
              </li>
              <li className="flex justify-between">
                <span className="text-muted-foreground">Joined:</span>
                <span className="font-medium text-foreground">
                  {user.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}
                </span>
              </li>
              <li className="flex justify-between">
                <span className="text-muted-foreground">Last Updated:</span>
                <span className="font-medium text-foreground">
                  {user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'N/A'}
                </span>
              </li>
              <li className="flex justify-between">
                <span className="text-muted-foreground">Last Login:</span>
                <span className="font-medium text-foreground">
                  {user.last_login_at ? new Date(user.last_login_at).toLocaleDateString() : 'Never'}
                </span>
              </li>
            </ul>
          </div>
          
          <div className="bg-card border border-border p-4 rounded-lg">
            <h4 className="text-sm font-medium text-muted-foreground mb-3 uppercase tracking-wider">Identity & Access</h4>
            <ul className="space-y-2 text-sm">
              <li className="flex justify-between">
                <span className="text-muted-foreground">Login Provider:</span>
                <span className="font-medium text-foreground capitalize">{user.provider || 'Email'}</span>
              </li>
              {user.phone && (
                <li className="flex justify-between">
                  <span className="text-muted-foreground">Phone Number:</span>
                  <span className="font-medium text-foreground">{user.phone} {user.phone_verified && '✓'}</span>
                </li>
              )}
              <li className="flex justify-between">
                <span className="text-muted-foreground">Account Tier:</span>
                <span className="font-medium text-foreground capitalize">{PLAN_NAMES[user.account_type as keyof typeof PLAN_NAMES] || user.account_type}</span>
              </li>
              <li className="flex justify-between">
                <span className="text-muted-foreground">Admin Access:</span>
                <span className="font-medium text-foreground">{user.is_admin ? 'Yes' : 'No'}</span>
              </li>
              <li className="flex flex-col mt-2">
                <span className="text-muted-foreground mb-1">Assigned Roles:</span>
                <div className="flex flex-wrap gap-1">
                  {user.roles && user.roles.length > 0 ? (
                    user.roles.map(role => (
                      <span key={role} className="bg-secondary text-secondary-foreground text-xs px-2 py-0.5 rounded">
                        {role}
                      </span>
                    ))
                  ) : (
                    <span className="text-xs italic text-muted-foreground">None</span>
                  )}
                </div>
              </li>
            </ul>
          </div>
        </div>

        {/* Stats Summary */}
        {(user.urls_count !== undefined || user.total_clicks !== undefined) && (
          <div className="grid grid-cols-2 gap-4">
            <div className="border border-border rounded-lg p-4 flex items-center space-x-3">
              <div className="bg-blue-500/10 p-2 rounded text-blue-600 dark:text-blue-400">
                <ChartBarIcon className="h-5 w-5" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Total URLs</p>
                <p className="text-lg font-semibold text-foreground">{user.urls_count || 0}</p>
              </div>
            </div>
            <div className="border border-border rounded-lg p-4 flex items-center space-x-3">
              <div className="bg-purple-500/10 p-2 rounded text-purple-600 dark:text-purple-400">
                <ChartBarIcon className="h-5 w-5" />
              </div>
              <div>
                <p className="text-xs text-muted-foreground uppercase">Total Clicks</p>
                <p className="text-lg font-semibold text-foreground">{user.total_clicks || 0}</p>
              </div>
            </div>
          </div>
        )}

        {/* Messages */}
        {actionMessage && (
          <div
            className={`p-3 rounded-lg flex items-center space-x-2 text-sm ${
              actionMessage.type === 'success'
                ? 'bg-green-500/10 text-green-600 dark:text-green-500'
                : 'bg-destructive/10 text-destructive'
            }`}
          >
            {actionMessage.type === 'error' && <ExclamationTriangleIcon className="h-5 w-5" />}
            <span>{actionMessage.text}</span>
          </div>
        )}

        {/* Admin Actions */}
        <div className="space-y-4">
          <h4 className="font-medium text-foreground border-b border-border pb-2">Admin Actions</h4>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
            {/* Password Reset */}
            <button
              onClick={() => handleAction(sendPasswordReset)}
              disabled={loading}
              className="flex items-center justify-center space-x-2 w-full px-4 py-2 bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-lg transition-colors disabled:opacity-50"
            >
              <EnvelopeIcon className="h-5 w-5" />
              <span>Send Password Reset</span>
            </button>

            {/* View URLs */}
            <button
              onClick={() => {
                onViewUrls(user.id);
                onClose();
              }}
              className="flex items-center justify-center space-x-2 w-full px-4 py-2 bg-primary/10 text-primary hover:bg-primary/20 rounded-lg transition-colors"
            >
              <ChartBarIcon className="h-5 w-5" />
              <span>View User's URLs</span>
            </button>

            {/* Toggle Status */}
            <button
              onClick={() => handleAction(toggleStatus)}
              disabled={loading || user.is_admin}
              className={`flex items-center justify-center space-x-2 w-full px-4 py-2 rounded-lg transition-colors disabled:opacity-50 ${
                user.is_active
                  ? 'bg-destructive/10 text-destructive hover:bg-destructive/20'
                  : 'bg-green-500/10 text-green-600 dark:text-green-500 hover:bg-green-500/20'
              }`}
            >
              <ShieldCheckIcon className="h-5 w-5" />
              <span>{user.is_active ? 'Deactivate User' : 'Activate User'}</span>
            </button>
          </div>

          {/* Account Tier */}
          <div className="pt-2">
            <label className="block text-sm font-medium text-foreground mb-1">Account Tier</label>
            <select
              value={user.account_type}
              onChange={(e) => handleAction(() => updateAccountType(e.target.value))}
              disabled={loading || user.is_admin}
              className="w-full px-3 py-2 bg-background border border-border rounded-lg text-foreground focus:ring-2 focus:ring-primary focus:border-primary disabled:opacity-50"
            >
              <option value="free">{PLAN_NAMES.free}</option>
              <option value="premium">{PLAN_NAMES.premium}</option>
              <option value="enterprise">{PLAN_NAMES.enterprise}</option>
            </select>
            {user.is_admin && (
              <p className="text-xs text-muted-foreground mt-1">Admin accounts do not use tiers.</p>
            )}
          </div>
        </div>

      </div>
    </Modal>
  );
}
