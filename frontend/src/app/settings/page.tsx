'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import { API_ENDPOINTS } from '@/lib/api-config';
import { getPlanDisplayName, PLAN_NAMES } from '@/lib/plan-utils';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import {
  UserCircleIcon,
  KeyIcon,
  BellIcon,
  GlobeAltIcon,
  ShieldCheckIcon,
  TrashIcon,
  EyeIcon,
  EyeSlashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  SwatchIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';
import clsx from 'clsx';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

const profileSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100, 'Name must be less than 100 characters'),
  email: z.string().email('Please enter a valid email address'),
  phone: z.string().min(10, 'Phone number must be at least 10 digits').max(20, 'Phone number is too long'),
});

const passwordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: z.string()
    .min(8, 'New password must be at least 8 characters')
    .regex(/[A-Z]/, 'Must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Must contain at least one number'),
  confirmPassword: z.string(),
}).refine((data) => data.currentPassword !== data.newPassword, {
  message: "New password must be different from your current password",
  path: ["newPassword"],
}).refine((data) => data.newPassword === data.confirmPassword, {
  message: "Passwords don't match",
  path: ["confirmPassword"],
});

const preferencesSchema = z.object({
  defaultExpiration: z.string().optional(),
  analyticsPublic: z.boolean(),
  emailNotifications: z.boolean(),
  marketingEmails: z.boolean(),
  timezone: z.string(),
  theme: z.enum(['light', 'dark', 'system']),
});

type ProfileFormData = z.infer<typeof profileSchema>;
type PasswordFormData = z.infer<typeof passwordSchema>;
type PreferencesFormData = z.infer<typeof preferencesSchema>;

export default function SettingsPage() {
  const { user, updateUser, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const [activeTab, setActiveTab] = useState('profile');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateSuccess, setUpdateSuccess] = useState<string | null>(null);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const [isLoadingPreferences, setIsLoadingPreferences] = useState(true);
  const [brandingForm, setBrandingForm] = useState({
    company_name: '', logo_url: '', primary_color: '#3b82f6',
    accent_color: '#10b981', custom_css: '', favicon_url: '',
  });
  const [brandingSaving, setBrandingSaving] = useState(false);
  const [brandingMsg, setBrandingMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [seoForm, setSeoForm] = useState({
    site_title: '', site_description: '', site_keywords: '',
    og_image: '', twitter_site: '', google_analytics_id: '',
  });
  const [seoSaving, setSeoSaving] = useState(false);
  const [seoMsg, setSeoMsg] = useState<{ type: 'success' | 'error'; text: string } | null>(null);
  const [otpModalOpen, setOtpModalOpen] = useState(false);
  const [otpCode, setOtpCode] = useState('');
  const [otpSending, setOtpSending] = useState(false);
  const [otpVerifying, setOtpVerifying] = useState(false);
  const [otpError, setOtpError] = useState<string | null>(null);
  const [otpExpiresAt, setOtpExpiresAt] = useState<string | null>(null);
  const [otpSuccess, setOtpSuccess] = useState<string | null>(null);
  const [sendingVerification, setSendingVerification] = useState(false);
  const [verificationSent, setVerificationSent] = useState(false);
  const router = useRouter();

  const handleSendVerification = async () => {
    setSendingVerification(true);
    try {
      const res = await fetch(API_ENDPOINTS.auth.resendEmail, {
        method: 'POST',
        headers: { Authorization: `Bearer ${getAccessToken()}` },
      });
      if (res.ok) setVerificationSent(true);
    } finally {
      setSendingVerification(false);
    }
  };

  // Profile form
  const profileForm = useForm<ProfileFormData>({
    resolver: zodResolver(profileSchema),
    defaultValues: {
      name: user?.name || '',
      email: user?.email || '',
      phone: user?.phone || '',
    },
  });

  // Password form
  const passwordForm = useForm<PasswordFormData>({
    resolver: zodResolver(passwordSchema),
  });

  // Preferences form
  const preferencesForm = useForm<PreferencesFormData>({
    resolver: zodResolver(preferencesSchema),
    defaultValues: {
      defaultExpiration: '30',
      analyticsPublic: false,
      emailNotifications: true,
      marketingEmails: false,
      timezone: 'UTC',
      theme: 'light',
    },
  });

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Update form defaults when user data loads
  useEffect(() => {
    if (user) {
      profileForm.reset({
        name: user.name,
        email: user.email,
        phone: user.phone || '',
      });
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]); // Only depend on user, not profileForm (which is stable)

  // Fetch and populate preferences
  useEffect(() => {
    const fetchPreferences = async () => {
      if (!user) return;

      setIsLoadingPreferences(true);
      try {
        const token = getAccessToken();
        const response = await fetch(`${API_BASE_URL}/api/v1/auth/preferences`, {
          method: 'GET',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (response.ok) {
          const data = await response.json();
          const prefs = data.preferences;

          // Reset form with fetched preferences
          preferencesForm.reset({
            defaultExpiration: prefs.default_expiration?.toString() || '30',
            analyticsPublic: prefs.analytics_public || false,
            emailNotifications: prefs.email_notifications !== undefined ? prefs.email_notifications : true,
            marketingEmails: prefs.marketing_emails || false,
            timezone: prefs.timezone || 'UTC',
            theme: prefs.theme || 'light',
          });
        }
      } catch (error) {
        console.error('Failed to fetch preferences:', error);
      } finally {
        setIsLoadingPreferences(false);
      }
    };

    fetchPreferences();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]); // Only depend on user, not preferencesForm (which is stable)

  const showSuccessMessage = (message: string) => {
    setUpdateSuccess(message);
    setUpdateError(null);
    setTimeout(() => setUpdateSuccess(null), 5000);
  };

  const showErrorMessage = (message: string) => {
    setUpdateError(message);
    setUpdateSuccess(null);
    setTimeout(() => setUpdateError(null), 5000);
  };

  const onProfileSubmit = async (data: ProfileFormData) => {
    setIsUpdating(true);
    try {
      const payload = {
        name: data.name,
        phone: data.phone.trim(),
      };

      const token = getAccessToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/profile`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (response.ok) {
        const responseData = await response.json();
        showSuccessMessage('Profile updated successfully!');
        // Update local user data and localStorage
        if (responseData.user) {
          updateUser(responseData.user);
        }
      } else {
        const errorData = await response.json();
        const errorMessage = errorData.error || errorData.message || errorData.details || 'Failed to update profile. Please try again.';
        showErrorMessage(errorMessage);
      }
    } catch (error) {
      showErrorMessage('Failed to update profile. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  const onPasswordSubmit = async (data: PasswordFormData) => {
    setIsUpdating(true);
    try {
      const token = getAccessToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/change-password`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          current_password: data.currentPassword,
          new_password: data.newPassword,
        }),
      });

      if (response.ok) {
        passwordForm.reset();
        showSuccessMessage('Password changed successfully!');
      } else {
        const errorData = await response.json();
        showErrorMessage(errorData.error || 'Failed to change password. Please check your current password.');
      }
    } catch (error) {
      showErrorMessage('Failed to change password. Please check your current password.');
    } finally {
      setIsUpdating(false);
    }
  };

  const onPreferencesSubmit = async (data: PreferencesFormData) => {
    setIsUpdating(true);
    try {
      const token = getAccessToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/preferences`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          timezone: data.timezone,
          theme: data.theme,
          notifications_email: data.emailNotifications,
          marketing_consent: data.marketingEmails,
          public_profile: data.analyticsPublic,
          default_url_expiration: data.defaultExpiration ? parseInt(data.defaultExpiration) : null,
        }),
      });

      if (response.ok) {
        const responseData = await response.json();
        const updatedPrefs = responseData.preferences;

        // Update form with actual saved values from backend
        preferencesForm.reset({
          defaultExpiration: updatedPrefs.default_expiration?.toString() || '30',
          analyticsPublic: updatedPrefs.analytics_public || false,
          emailNotifications: updatedPrefs.email_notifications !== undefined ? updatedPrefs.email_notifications : true,
          marketingEmails: updatedPrefs.marketing_emails || false,
          timezone: updatedPrefs.timezone || 'UTC',
          theme: updatedPrefs.theme || 'light',
        });

        showSuccessMessage('Preferences updated successfully!');
      } else {
        const errorData = await response.json();
        showErrorMessage(errorData.message || 'Failed to update preferences. Please try again.');
      }
    } catch (error) {
      showErrorMessage('Failed to update preferences. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  const handleDeleteAccount = async () => {
    if (!confirm('Are you sure you want to delete your account? This action cannot be undone.')) {
      return;
    }

    if (!confirm('This will permanently delete all your URLs and data. Are you absolutely sure?')) {
      return;
    }

    setIsUpdating(true);
    try {
      const token = getAccessToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/delete-account`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        showSuccessMessage('Account deletion initiated. You will be logged out.');
        setTimeout(() => {
          // logout();
          router.push('/');
        }, 2000);
      } else {
        const errorData = await response.json();
        showErrorMessage(errorData.message || 'Failed to delete account. Please try again.');
      }
    } catch (error) {
      showErrorMessage('Failed to delete account. Please try again.');
    } finally {
      setIsUpdating(false);
    }
  };

  // Load SEO settings for admin
  useEffect(() => {
    if (!user) return;
    const token = getAccessToken();
    fetch(`${API_BASE_URL}/api/v1/admin/settings?category=seo`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then((r) => r.ok ? r.json() : null)
      .then((data) => {
        if (!data?.settings) return;
        const flat: Record<string, string> = {};
        const list = Array.isArray(data.settings) ? data.settings : Object.values(data.settings).flat();
        for (const s of list as Array<{ key: string; value: string }>) {
          flat[s.key] = s.value ?? '';
        }
        setSeoForm({
          site_title: flat.site_title ?? '',
          site_description: flat.site_description ?? '',
          site_keywords: flat.site_keywords ?? '',
          og_image: flat.og_image ?? '',
          twitter_site: flat.twitter_site ?? '',
          google_analytics_id: flat.google_analytics_id ?? '',
        });
      })
      .catch(() => {});
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [user]);

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  const isEnterprise = user?.account_type === 'enterprise';

  const tabs = [
    { id: 'profile', name: 'Profile', icon: UserCircleIcon },
    { id: 'password', name: 'Password', icon: KeyIcon },
    { id: 'preferences', name: 'Preferences', icon: BellIcon },
    { id: 'security', name: 'Security', icon: ShieldCheckIcon },
    ...(isEnterprise ? [{ id: 'branding', name: 'Branding', icon: SwatchIcon }] : []),
    ...(isEnterprise ? [{ id: 'seo', name: 'SEO', icon: MagnifyingGlassIcon }] : []),
    { id: 'danger', name: 'Danger Zone', icon: ExclamationTriangleIcon },
  ];

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground">Settings</h1>
            <p className="mt-2 text-muted-foreground">
              Manage your account settings and preferences.
            </p>
          </div>

          {/* Success/Error Messages */}
          {updateSuccess && (
            <div className="mb-6 p-4 bg-green-500/10 border border-green-500/20 rounded-lg flex items-center space-x-2">
              <CheckCircleIcon className="h-5 w-5 text-green-600 dark:text-green-500" />
              <span className="text-green-700 dark:text-green-400">{updateSuccess}</span>
            </div>
          )}

          {updateError && (
            <div className="mb-6 p-4 bg-destructive/10 border border-destructive/20 rounded-lg flex items-center space-x-2">
              <ExclamationTriangleIcon className="h-5 w-5 text-destructive" />
              <span className="text-destructive font-medium">{updateError}</span>
            </div>
          )}

          <div className="bg-card shadow-sm border border-border rounded-lg">
            {/* Tab Navigation */}
            <div className="border-b border-border">
              <nav className="flex space-x-8 px-6" aria-label="Tabs">
                {tabs.map((tab) => {
                  const Icon = tab.icon;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id)}
                      className={clsx(
                        'flex items-center space-x-2 py-4 px-1 border-b-2 font-medium text-sm transition-colors',
                        activeTab === tab.id
                          ? 'border-primary-500 text-primary-600'
                          : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                      )}
                    >
                      <Icon className="h-5 w-5" />
                      <span>{tab.name}</span>
                    </button>
                  );
                })}
              </nav>
            </div>

            {/* Tab Content */}
            <div className="p-6">
              {activeTab === 'profile' && (
                <div>
                  <h2 className="text-lg font-medium text-foreground mb-4">Profile Information</h2>
                  <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label htmlFor="name" className="block text-sm font-medium text-foreground">
                          Full Name
                        </label>
                        <input
                          {...profileForm.register('name')}
                          type="text"
                          className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        />
                        {profileForm.formState.errors.name && (
                          <p className="mt-1 text-sm text-destructive">{profileForm.formState.errors.name.message}</p>
                        )}
                      </div>

                      <div>
                        <label htmlFor="email" className="block text-sm font-medium text-foreground">
                          Email Address
                        </label>
                        <input
                          {...profileForm.register('email')}
                          type="email"
                          disabled
                          className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm bg-muted text-muted-foreground cursor-not-allowed"
                        />
                        <p className="mt-1 text-sm text-muted-foreground">Email address cannot be changed</p>
                      </div>
                    </div>

                    <div>
                      <label htmlFor="phone" className="block text-sm font-medium text-foreground">
                        Phone Number
                      </label>
                      <input
                        {...profileForm.register('phone')}
                        type="tel"
                        className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        placeholder="+1 (555) 123-4567"
                      />
                      {profileForm.formState.errors.phone && (
                        <p className="mt-1 text-sm text-destructive">{profileForm.formState.errors.phone.message}</p>
                      )}
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isUpdating ? 'Updating...' : 'Update Profile'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'password' && (
                <div>
                  <h2 className="text-lg font-medium text-foreground mb-4">Change Password</h2>
                  <form onSubmit={passwordForm.handleSubmit(onPasswordSubmit)} className="space-y-6">
                    <div>
                      <label htmlFor="currentPassword" className="block text-sm font-medium text-foreground">
                        Current Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('currentPassword')}
                          type={showCurrentPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                        >
                          {showCurrentPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-muted-foreground" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.currentPassword && (
                        <p className="mt-1 text-sm text-destructive">{passwordForm.formState.errors.currentPassword.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="newPassword" className="block text-sm font-medium text-foreground">
                        New Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('newPassword')}
                          type={showNewPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowNewPassword(!showNewPassword)}
                        >
                          {showNewPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-muted-foreground" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.newPassword && (
                        <p className="mt-1 text-sm text-destructive">{passwordForm.formState.errors.newPassword.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="confirmPassword" className="block text-sm font-medium text-foreground">
                        Confirm New Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('confirmPassword')}
                          type={showConfirmPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                        >
                          {showConfirmPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-muted-foreground" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-muted-foreground" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.confirmPassword && (
                        <p className="mt-1 text-sm text-destructive">{passwordForm.formState.errors.confirmPassword.message}</p>
                      )}
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 focus:outline-none focus:ring-2 focus:ring-primary disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isUpdating ? 'Changing...' : 'Change Password'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'preferences' && (
                <div>
                  <h2 className="text-lg font-medium text-foreground mb-4">Preferences</h2>
                  <form onSubmit={preferencesForm.handleSubmit(onPreferencesSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label htmlFor="defaultExpiration" className="block text-sm font-medium text-foreground">
                          Default URL Expiration
                        </label>
                        <select
                          {...preferencesForm.register('defaultExpiration')}
                          className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        >
                          <option value="">Never expire</option>
                          <option value="1">1 day</option>
                          <option value="7">7 days</option>
                          <option value="30">30 days</option>
                          <option value="90">90 days</option>
                          <option value="365">1 year</option>
                        </select>
                      </div>

                      <div>
                        <label htmlFor="timezone" className="block text-sm font-medium text-foreground">
                          Timezone
                        </label>
                        <select
                          {...preferencesForm.register('timezone')}
                          className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                        >
                          <option value="UTC">UTC</option>
                          <option value="America/New_York">Eastern Time</option>
                          <option value="America/Chicago">Central Time</option>
                          <option value="America/Denver">Mountain Time</option>
                          <option value="America/Los_Angeles">Pacific Time</option>
                          <option value="Europe/London">London</option>
                          <option value="Europe/Paris">Paris</option>
                          <option value="Asia/Tokyo">Tokyo</option>
                        </select>
                      </div>
                    </div>

                    <div>
                      <label htmlFor="theme" className="block text-sm font-medium text-foreground">
                        Theme
                      </label>
                      <select
                        {...preferencesForm.register('theme')}
                        className="mt-1 block w-full px-3 py-2 border border-input rounded-md shadow-sm focus:outline-none focus:ring-primary focus:border-primary bg-background text-foreground"
                      >
                        <option value="light">Light</option>
                        <option value="dark">Dark</option>
                        <option value="system">System</option>
                      </select>
                    </div>

                    <div className="space-y-4">
                      <h3 className="text-md font-medium text-foreground">Notifications</h3>

                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="emailNotifications" className="text-sm font-medium text-foreground">
                            Email Notifications
                          </label>
                          <p className="text-sm text-muted-foreground">Receive important updates about your account</p>
                        </div>
                        <input
                          id="emailNotifications"
                          type="checkbox"
                          checked={preferencesForm.watch('emailNotifications')}
                          onChange={(e) => preferencesForm.setValue('emailNotifications', e.target.checked, { shouldDirty: true })}
                          className="h-4 w-4 text-primary focus:ring-primary border-input rounded cursor-pointer"
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="marketingEmails" className="text-sm font-medium text-foreground">
                            Marketing Emails
                          </label>
                          <p className="text-sm text-muted-foreground">Receive product updates and promotional content</p>
                        </div>
                        <input
                          id="marketingEmails"
                          type="checkbox"
                          checked={preferencesForm.watch('marketingEmails')}
                          onChange={(e) => preferencesForm.setValue('marketingEmails', e.target.checked, { shouldDirty: true })}
                          className="h-4 w-4 text-primary focus:ring-primary border-input rounded cursor-pointer"
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="analyticsPublic" className="text-sm font-medium text-foreground">
                            Public Analytics
                          </label>
                          <p className="text-sm text-muted-foreground">Allow others to view click statistics for your URLs</p>
                        </div>
                        <input
                          id="analyticsPublic"
                          type="checkbox"
                          checked={preferencesForm.watch('analyticsPublic')}
                          onChange={(e) => preferencesForm.setValue('analyticsPublic', e.target.checked, { shouldDirty: true })}
                          className="h-4 w-4 text-primary focus:ring-primary border-input rounded cursor-pointer"
                        />
                      </div>
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="bg-primary-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isUpdating ? 'Saving...' : 'Save Preferences'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'security' && (
                <>
                  <h2 className="text-lg font-medium text-foreground mb-4">Security & Privacy</h2>
                  <div className="space-y-6">
                    <div className="bg-muted p-4 rounded-lg">
                      <h3 className="text-md font-medium text-foreground mb-2">Account Status</h3>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-muted-foreground">Email Verified</span>
                          <div className="flex items-center space-x-1">
                            {user?.email_verified ? (
                              <>
                                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                                <span className="text-sm text-green-600">Verified</span>
                              </>
                            ) : (
                              <>
                                <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />
                                <span className="text-sm text-yellow-600">Not Verified</span>
                              </>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-muted-foreground">Phone Verified</span>
                          <div className="flex items-center space-x-1">
                            {user?.phone_verified ? (
                              <>
                                <CheckCircleIcon className="h-4 w-4 text-green-600" />
                                <span className="text-sm text-green-600">Verified</span>
                              </>
                            ) : (
                              <>
                                <ExclamationTriangleIcon className="h-4 w-4 text-yellow-600" />
                                <span className="text-sm text-yellow-600">Not Verified</span>
                              </>
                            )}
                          </div>
                        </div>
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-muted-foreground">Account Type</span>
                          <span className="text-sm text-foreground font-medium">{getPlanDisplayName(user?.account_type ?? '')}</span>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-3">
                      <h3 className="text-md font-medium text-foreground">Security Actions</h3>

                      {!user?.email_verified && (
                        <button
                          onClick={handleSendVerification}
                          disabled={sendingVerification || verificationSent}
                          className="w-full text-left p-4 border border-yellow-500/20 bg-yellow-500/10 rounded-lg hover:bg-yellow-500/20 transition-colors disabled:opacity-60"
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-sm font-medium dark:text-yellow-500 text-yellow-600">
                                {verificationSent ? 'Verification email sent!' : sendingVerification ? 'Sending…' : 'Verify Email Address'}
                              </h4>
                              <p className="text-sm dark:text-yellow-500/80 text-yellow-600/80">Complete email verification to secure your account</p>
                            </div>
                            {verificationSent ? (
                              <CheckCircleIcon className="h-5 w-5 text-green-500" />
                            ) : (
                              <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />
                            )}
                          </div>
                        </button>
                      )}

                      {user?.phone && !user?.phone_verified && (
                        <button
                          disabled={otpSending}
                          onClick={async () => {
                            const token = getAccessToken();
                            setOtpSending(true);
                            setOtpError(null);
                            setOtpSuccess(null);
                            try {
                              const res = await fetch(API_ENDPOINTS.auth.sendOtp, {
                                method: 'POST',
                                headers: { Authorization: `Bearer ${token}` },
                              });
                              const data = await res.json();
                              if (!res.ok) {
                                setOtpError(data.error || 'Failed to send OTP');
                              } else {
                                setOtpExpiresAt(data.expires_at ?? null);
                                setOtpCode('');
                                setOtpModalOpen(true);
                              }
                            } catch {
                              setOtpError('Network error. Please try again.');
                            } finally {
                              setOtpSending(false);
                            }
                          }}
                          className="w-full text-left p-4 border border-yellow-500/20 bg-yellow-500/10 rounded-lg hover:bg-yellow-500/20 transition-colors disabled:opacity-60"
                        >
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-sm font-medium dark:text-yellow-500 text-yellow-600">
                                {otpSending ? 'Sending OTP…' : 'Verify Phone Number'}
                              </h4>
                              <p className="text-sm dark:text-yellow-500/80 text-yellow-600/80">Add extra security with phone verification</p>
                            </div>
                            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />
                          </div>
                        </button>
                      )}

                      {otpError && (
                        <p className="text-sm text-destructive mt-1">{otpError}</p>
                      )}
                      {otpSuccess && (
                        <p className="text-sm text-green-600 mt-1">{otpSuccess}</p>
                      )}
                    </div>
                  </div>

                {/* OTP Verification Modal */}
                {otpModalOpen && (
                  <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
                    <div className="bg-card border border-border rounded-xl shadow-xl p-6 w-full max-w-sm mx-4">
                      <h3 className="text-lg font-semibold text-foreground mb-1">Verify Phone Number</h3>
                      <p className="text-sm text-muted-foreground mb-4">
                        Enter the 6-digit code sent to <span className="font-medium text-foreground">{user?.phone}</span>.
                        {otpExpiresAt && (
                          <span className="block mt-1">Expires at {new Date(otpExpiresAt).toLocaleTimeString()}</span>
                        )}
                      </p>

                      <input
                        type="text"
                        inputMode="numeric"
                        maxLength={6}
                        value={otpCode}
                        onChange={(e) => {
                          setOtpCode(e.target.value.replace(/\D/g, ''));
                          setOtpError(null);
                        }}
                        placeholder="000000"
                        className="w-full border border-input rounded-lg px-4 py-3 text-center text-2xl font-mono tracking-widest bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary mb-3"
                      />

                      {otpError && (
                        <p className="text-sm text-destructive mb-3">{otpError}</p>
                      )}

                      <button
                        disabled={otpVerifying || otpCode.length !== 6}
                        onClick={async () => {
                          if (!user) return;
                          setOtpVerifying(true);
                          setOtpError(null);
                          try {
                            const res = await fetch(API_ENDPOINTS.auth.otp.verify, {
                              method: 'POST',
                              headers: { 'Content-Type': 'application/json' },
                              body: JSON.stringify({ user_id: user.id, otp: otpCode }),
                            });
                            const data = await res.json();
                            if (!res.ok) {
                              setOtpError(data.error || 'Invalid code. Please try again.');
                            } else {
                              updateUser({ ...user, phone_verified: true });
                              setOtpModalOpen(false);
                              setOtpSuccess('Phone number verified successfully.');
                            }
                          } catch {
                            setOtpError('Network error. Please try again.');
                          } finally {
                            setOtpVerifying(false);
                          }
                        }}
                        className="w-full py-2 px-4 bg-primary text-primary-foreground rounded-lg font-medium hover:bg-primary/90 disabled:opacity-50 mb-2"
                      >
                        {otpVerifying ? 'Verifying…' : 'Verify'}
                      </button>

                      <div className="flex items-center justify-between mt-1">
                        <button
                          disabled={otpSending}
                          onClick={async () => {
                            const token = getAccessToken();
                            setOtpSending(true);
                            setOtpError(null);
                            try {
                              const res = await fetch(API_ENDPOINTS.auth.sendOtp, {
                                method: 'POST',
                                headers: { Authorization: `Bearer ${token}` },
                              });
                              const data = await res.json();
                              if (!res.ok) {
                                setOtpError(data.error || 'Failed to resend OTP');
                              } else {
                                setOtpExpiresAt(data.expires_at ?? null);
                                setOtpCode('');
                              }
                            } catch {
                              setOtpError('Network error.');
                            } finally {
                              setOtpSending(false);
                            }
                          }}
                          className="text-sm text-primary hover:underline disabled:opacity-50"
                        >
                          {otpSending ? 'Sending…' : 'Resend code'}
                        </button>
                        <button
                          onClick={() => { setOtpModalOpen(false); setOtpError(null); }}
                          className="text-sm text-muted-foreground hover:text-foreground"
                        >
                          Cancel
                        </button>
                      </div>
                    </div>
                  </div>
                )}
                </>
              )}

              {activeTab === 'branding' && (
                <div>
                  <h2 className="text-lg font-medium text-foreground mb-1">White-Label Branding</h2>
                  <p className="text-sm text-gray-500 mb-6">Customize the appearance of your branded short links ({PLAN_NAMES.enterprise} only).</p>
                  {brandingMsg && (
                    <div className={`mb-4 px-4 py-3 rounded-lg text-sm ${brandingMsg.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                      {brandingMsg.text}
                    </div>
                  )}
                  <div className="space-y-4">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Company Name</label>
                        <input type="text" value={brandingForm.company_name}
                          onChange={e => setBrandingForm(f => ({ ...f, company_name: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="Acme Corp" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Logo URL</label>
                        <input type="url" value={brandingForm.logo_url}
                          onChange={e => setBrandingForm(f => ({ ...f, logo_url: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="https://..." />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Primary Color</label>
                        <div className="flex items-center gap-2">
                          <input type="color" value={brandingForm.primary_color}
                            onChange={e => setBrandingForm(f => ({ ...f, primary_color: e.target.value }))}
                            className="h-9 w-14 border border-gray-300 rounded cursor-pointer" />
                          <input type="text" value={brandingForm.primary_color}
                            onChange={e => setBrandingForm(f => ({ ...f, primary_color: e.target.value }))}
                            className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono" />
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Accent Color</label>
                        <div className="flex items-center gap-2">
                          <input type="color" value={brandingForm.accent_color}
                            onChange={e => setBrandingForm(f => ({ ...f, accent_color: e.target.value }))}
                            className="h-9 w-14 border border-gray-300 rounded cursor-pointer" />
                          <input type="text" value={brandingForm.accent_color}
                            onChange={e => setBrandingForm(f => ({ ...f, accent_color: e.target.value }))}
                            className="flex-1 border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono" />
                        </div>
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Favicon URL</label>
                        <input type="url" value={brandingForm.favicon_url}
                          onChange={e => setBrandingForm(f => ({ ...f, favicon_url: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm" placeholder="https://..." />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Custom CSS</label>
                      <textarea value={brandingForm.custom_css}
                        onChange={e => setBrandingForm(f => ({ ...f, custom_css: e.target.value }))}
                        rows={4}
                        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm font-mono"
                        placeholder="/* custom styles */" />
                    </div>
                    <button
                      disabled={brandingSaving}
                      onClick={async () => {
                        setBrandingSaving(true); setBrandingMsg(null);
                        try {
                          const token = await getAccessToken();
                          const res = await fetch(`${API_BASE_URL}/api/v1/branding`, {
                            method: 'PUT',
                            headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                              company_name: brandingForm.company_name,
                              logo_url: brandingForm.logo_url,
                              primary_color: brandingForm.primary_color,
                              accent_color: brandingForm.accent_color,
                              custom_css: brandingForm.custom_css,
                              favicon_url: brandingForm.favicon_url,
                            }),
                          });
                          const data = await res.json();
                          if (res.ok) setBrandingMsg({ type: 'success', text: 'Branding saved successfully.' });
                          else setBrandingMsg({ type: 'error', text: data.error ?? 'Failed to save branding.' });
                        } finally {
                          setBrandingSaving(false);
                        }
                      }}
                      className="px-5 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      {brandingSaving ? 'Saving...' : 'Save Branding'}
                    </button>
                  </div>
                </div>
              )}

              {activeTab === 'seo' && (
                <div>
                  <h2 className="text-lg font-medium text-foreground mb-1">Site-wide SEO Settings</h2>
                  <p className="text-sm text-gray-500 mb-6">Control metadata, Open Graph tags, and analytics used across the site.</p>
                  {seoMsg && (
                    <div className={`mb-4 px-4 py-3 rounded-lg text-sm ${seoMsg.type === 'success' ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                      {seoMsg.text}
                    </div>
                  )}
                  <div className="space-y-4">
                    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Site Title</label>
                        <input type="text" value={seoForm.site_title}
                          onChange={e => setSeoForm(f => ({ ...f, site_title: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                          placeholder="Trunc - Professional URL Shortener" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Twitter Handle</label>
                        <input type="text" value={seoForm.twitter_site}
                          onChange={e => setSeoForm(f => ({ ...f, twitter_site: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                          placeholder="@trunc" />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Default OG Image URL</label>
                        <input type="url" value={seoForm.og_image}
                          onChange={e => setSeoForm(f => ({ ...f, og_image: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                          placeholder="https://..." />
                      </div>
                      <div>
                        <label className="block text-sm font-medium text-gray-700 mb-1">Google Analytics ID</label>
                        <input type="text" value={seoForm.google_analytics_id}
                          onChange={e => setSeoForm(f => ({ ...f, google_analytics_id: e.target.value }))}
                          className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                          placeholder="G-XXXXXXXXXX" />
                      </div>
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Site Description</label>
                      <textarea value={seoForm.site_description}
                        onChange={e => setSeoForm(f => ({ ...f, site_description: e.target.value }))}
                        rows={3}
                        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                        placeholder="Professional URL shortener with analytics, custom domains..." />
                    </div>
                    <div>
                      <label className="block text-sm font-medium text-gray-700 mb-1">Keywords</label>
                      <input type="text" value={seoForm.site_keywords}
                        onChange={e => setSeoForm(f => ({ ...f, site_keywords: e.target.value }))}
                        className="w-full border border-gray-300 rounded-lg px-3 py-2 text-sm"
                        placeholder="url shortener, link shortener, custom links" />
                    </div>
                    <button
                      disabled={seoSaving}
                      onClick={async () => {
                        setSeoSaving(true); setSeoMsg(null);
                        try {
                          const token = await getAccessToken();
                          const res = await fetch(`${API_BASE_URL}/api/v1/admin/settings`, {
                            method: 'PUT',
                            headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                              site_title: seoForm.site_title,
                              site_description: seoForm.site_description,
                              site_keywords: seoForm.site_keywords,
                              og_image: seoForm.og_image,
                              twitter_site: seoForm.twitter_site,
                              google_analytics_id: seoForm.google_analytics_id,
                            }),
                          });
                          if (res.ok) setSeoMsg({ type: 'success', text: 'SEO settings saved.' });
                          else {
                            const d = await res.json();
                            setSeoMsg({ type: 'error', text: d.error ?? 'Failed to save SEO settings.' });
                          }
                        } finally {
                          setSeoSaving(false);
                        }
                      }}
                      className="px-5 py-2 bg-blue-600 text-white text-sm font-medium rounded-lg hover:bg-blue-700 disabled:opacity-50"
                    >
                      {seoSaving ? 'Saving...' : 'Save SEO Settings'}
                    </button>
                  </div>
                </div>
              )}

              {activeTab === 'danger' && (
                <div>
                  <h2 className="text-lg font-medium text-destructive mb-4">Danger Zone</h2>
                  <div className="space-y-4">
                    <div className="border border-destructive/20 bg-destructive/5 p-4 rounded-lg">
                      <h3 className="text-md font-medium text-destructive mb-2">Delete Account</h3>
                      <p className="text-sm text-destructive/80 mb-4">
                        Permanently delete your account and all associated data. This action cannot be undone.
                      </p>
                      <div className="space-y-2 text-sm text-destructive/80">
                        <p>• All your shortened URLs will be deleted</p>
                        <p>• All analytics data will be permanently removed</p>
                        <p>• Your account cannot be recovered</p>
                      </div>
                      <button
                        onClick={handleDeleteAccount}
                        disabled={isUpdating}
                        className="mt-4 bg-destructive text-destructive-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-destructive/90 focus:outline-none focus:ring-2 focus:ring-destructive disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                      >
                        <TrashIcon className="h-4 w-4" />
                        <span>{isUpdating ? 'Deleting...' : 'Delete Account'}</span>
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}// Force recompile trigger
