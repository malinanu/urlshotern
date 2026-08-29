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
  ExclamationTriangleIcon,
  PhoneIcon,
  CalendarDaysIcon,
  ClockIcon,
  KeyIcon,
  IdentificationIcon,
  LinkIcon,
  CursorArrowRaysIcon,
  Cog6ToothIcon
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
    <Modal isOpen={isOpen} onClose={onClose} title="User Profile" size="lg">
      <div className="space-y-6 mt-2">
        {/* Profile Header */}
        <div className="flex items-start gap-5 pb-6 border-b border-border/50">
          <div className="relative shrink-0">
            {user.avatar_url ? (
              <img src={user.avatar_url} alt={user.name} className="h-16 w-16 rounded-full ring-4 ring-background shadow-sm object-cover" />
            ) : (
              <div className="h-16 w-16 rounded-full bg-primary/10 flex items-center justify-center ring-4 ring-background shadow-sm">
                <UserIcon className="h-8 w-8 text-primary" />
              </div>
            )}
            <div 
              className={`absolute -bottom-1 -right-1 h-5 w-5 rounded-full border-2 border-background flex items-center justify-center ${user.is_active ? 'bg-green-500' : 'bg-red-500'}`} 
              title={user.is_active ? 'Active' : 'Inactive'}
            />
          </div>
          
          <div className="flex-1 min-w-0">
            <div className="flex items-center gap-3 mb-1.5 flex-wrap">
              <h3 className="text-xl font-semibold text-foreground truncate">{user.name}</h3>
              {user.is_admin && (
                <span className="inline-flex items-center gap-1 px-2.5 py-0.5 rounded-full bg-purple-500/10 text-purple-600 dark:text-purple-400 text-xs font-medium border border-purple-500/20">
                  <ShieldCheckIcon className="h-3.5 w-3.5" />
                  Admin
                </span>
              )}
            </div>
            <div className="flex flex-col gap-1.5 text-sm text-muted-foreground">
              <span className="flex items-center gap-2">
                <EnvelopeIcon className="h-4 w-4 shrink-0" />
                <span className="truncate">{user.email}</span>
                {user.is_email_verified && <span className="text-green-600 dark:text-green-400 text-[10px] uppercase font-bold tracking-wider bg-green-500/10 px-1.5 py-0.5 rounded ml-1">Verified</span>}
              </span>
              {user.phone && (
                <span className="flex items-center gap-2">
                  <PhoneIcon className="h-4 w-4 shrink-0" />
                  {user.phone}
                  {user.phone_verified && <span className="text-green-600 dark:text-green-400 text-[10px] uppercase font-bold tracking-wider bg-green-500/10 px-1.5 py-0.5 rounded ml-1">Verified</span>}
                </span>
              )}
            </div>
          </div>
        </div>

        {/* Stats Row */}
        {(user.urls_count !== undefined || user.total_clicks !== undefined) && (
          <div className="grid grid-cols-2 gap-4">
            <div className="bg-blue-500/5 border border-blue-500/10 rounded-xl p-4 flex items-center gap-4 transition-colors hover:bg-blue-500/10">
              <div className="bg-blue-500/20 p-2.5 rounded-lg text-blue-600 dark:text-blue-400">
                <LinkIcon className="h-6 w-6" />
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Total URLs</p>
                <p className="text-2xl font-bold text-foreground tracking-tight">{user.urls_count || 0}</p>
              </div>
            </div>
            <div className="bg-purple-500/5 border border-purple-500/10 rounded-xl p-4 flex items-center gap-4 transition-colors hover:bg-purple-500/10">
              <div className="bg-purple-500/20 p-2.5 rounded-lg text-purple-600 dark:text-purple-400">
                <CursorArrowRaysIcon className="h-6 w-6" />
              </div>
              <div>
                <p className="text-sm font-medium text-muted-foreground">Total Clicks</p>
                <p className="text-2xl font-bold text-foreground tracking-tight">{user.total_clicks || 0}</p>
              </div>
            </div>
          </div>
        )}

        {/* Details Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
          {/* Account Info */}
          <div className="space-y-3">
            <h4 className="flex items-center gap-2 text-sm font-semibold text-foreground/80 uppercase tracking-wider">
              <IdentificationIcon className="h-4 w-4" />
              Account Info
            </h4>
            <div className="bg-secondary/30 rounded-xl p-1 border border-border/50 shadow-sm">
              <div className="flex items-center justify-between p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground flex items-center gap-2"><CalendarDaysIcon className="h-4 w-4" /> Joined</span>
                <span className="text-sm font-medium text-foreground">{user.created_at ? new Date(user.created_at).toLocaleDateString() : 'N/A'}</span>
              </div>
              <div className="flex items-center justify-between p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground flex items-center gap-2"><ClockIcon className="h-4 w-4" /> Last Login</span>
                <span className="text-sm font-medium text-foreground">{user.last_login_at ? new Date(user.last_login_at).toLocaleDateString() : 'Never'}</span>
              </div>
              <div className="flex items-center justify-between p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground flex items-center gap-2"><Cog6ToothIcon className="h-4 w-4" /> Updated</span>
                <span className="text-sm font-medium text-foreground">{user.updated_at ? new Date(user.updated_at).toLocaleDateString() : 'N/A'}</span>
              </div>
            </div>
          </div>

          {/* Identity & Access */}
          <div className="space-y-3">
            <h4 className="flex items-center gap-2 text-sm font-semibold text-foreground/80 uppercase tracking-wider">
              <KeyIcon className="h-4 w-4" />
              Identity & Access
            </h4>
            <div className="bg-secondary/30 rounded-xl p-1 border border-border/50 shadow-sm">
              <div className="flex items-center justify-between p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground">Login Provider</span>
                <span className="text-sm font-medium text-foreground capitalize bg-background px-2.5 py-1 rounded-md shadow-sm border border-border/50">{user.provider || 'Email'}</span>
              </div>
              <div className="flex items-center justify-between p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground">Account Tier</span>
                <span className="text-sm font-medium text-foreground capitalize bg-background px-2.5 py-1 rounded-md shadow-sm border border-border/50">{PLAN_NAMES[user.account_type as keyof typeof PLAN_NAMES] || user.account_type}</span>
              </div>
              <div className="flex flex-col gap-2 p-3 border-b border-border/50 last:border-0">
                <span className="text-sm text-muted-foreground">Assigned Roles</span>
                <div className="flex flex-wrap gap-1.5">
                  {user.roles && user.roles.length > 0 ? (
                    user.roles.map(role => (
                      <span key={role} className="bg-background border border-border/50 text-foreground text-xs px-2.5 py-1 rounded-md shadow-sm">
                        {role}
                      </span>
                    ))
                  ) : (
                    <span className="text-sm italic text-muted-foreground">None assigned</span>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Messages */}
        {actionMessage && (
          <div
            className={`p-4 rounded-xl flex items-center gap-3 text-sm font-medium shadow-sm ${
              actionMessage.type === 'success'
                ? 'bg-green-500/10 text-green-700 dark:text-green-400 border border-green-500/20'
                : 'bg-destructive/10 text-destructive border border-destructive/20'
            }`}
          >
            {actionMessage.type === 'error' && <ExclamationTriangleIcon className="h-5 w-5 shrink-0" />}
            <span>{actionMessage.text}</span>
          </div>
        )}

        {/* Admin Actions */}
        <div className="space-y-4 pt-6 border-t border-border/50">
          <h4 className="text-base font-semibold text-foreground">Management & Actions</h4>
          
          <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
            <div className="space-y-2 sm:col-span-2 bg-secondary/20 p-4 rounded-xl border border-border/50">
              <label className="text-sm font-medium text-foreground">Change Account Tier</label>
              <div className="flex gap-3">
                <select
                  value={user.account_type}
                  onChange={(e) => handleAction(() => updateAccountType(e.target.value))}
                  disabled={loading || user.is_admin}
                  className="flex-1 px-3 py-2 bg-background border border-input rounded-lg text-foreground text-sm focus:ring-2 focus:ring-primary focus:border-primary disabled:opacity-60 transition-shadow shadow-sm outline-none"
                >
                  <option value="free">{PLAN_NAMES.free}</option>
                  <option value="premium">{PLAN_NAMES.premium}</option>
                  <option value="enterprise">{PLAN_NAMES.enterprise}</option>
                </select>
                {user.is_admin && (
                  <div className="flex items-center text-xs font-medium text-amber-600 dark:text-amber-400 bg-amber-500/10 px-3 rounded-lg border border-amber-500/20">
                    Admin Fixed
                  </div>
                )}
              </div>
            </div>

            <button
              onClick={() => {
                onViewUrls(user.id);
                onClose();
              }}
              className="flex items-center justify-center gap-2 w-full px-4 py-2.5 bg-primary text-primary-foreground hover:bg-primary/90 rounded-xl transition-all shadow-sm font-medium text-sm"
            >
              <LinkIcon className="h-4 w-4" />
              <span>View User's URLs</span>
            </button>

            <button
              onClick={() => handleAction(sendPasswordReset)}
              disabled={loading}
              className="flex items-center justify-center gap-2 w-full px-4 py-2.5 bg-secondary text-secondary-foreground hover:bg-secondary/80 rounded-xl transition-all shadow-sm font-medium text-sm disabled:opacity-50"
            >
              <KeyIcon className="h-4 w-4" />
              <span>Reset Password</span>
            </button>

            <button
              onClick={() => handleAction(toggleStatus)}
              disabled={loading || user.is_admin}
              className={`sm:col-span-2 flex items-center justify-center gap-2 w-full px-4 py-3 rounded-xl transition-all shadow-sm font-medium text-sm disabled:opacity-50 disabled:cursor-not-allowed ${
                user.is_active
                  ? 'bg-destructive/10 text-destructive hover:bg-destructive/20 border border-destructive/20'
                  : 'bg-green-500/10 text-green-700 dark:text-green-400 hover:bg-green-500/20 border border-green-500/20'
              }`}
            >
              <ShieldCheckIcon className="h-5 w-5" />
              <span>{user.is_active ? 'Deactivate User Account' : 'Activate User Account'}</span>
            </button>
          </div>
        </div>

      </div>
    </Modal>
  );
}
