'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
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
} from '@heroicons/react/24/outline';
import clsx from 'clsx';

const profileSchema = z.object({
  name: z.string().min(2, 'Name must be at least 2 characters').max(100, 'Name must be less than 100 characters'),
  email: z.string().email('Please enter a valid email address'),
  phone: z.string().optional(),
});

const passwordSchema = z.object({
  currentPassword: z.string().min(1, 'Current password is required'),
  newPassword: z.string()
    .min(8, 'New password must be at least 8 characters')
    .regex(/[A-Z]/, 'Must contain at least one uppercase letter')
    .regex(/[a-z]/, 'Must contain at least one lowercase letter')
    .regex(/[0-9]/, 'Must contain at least one number'),
  confirmPassword: z.string(),
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
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const [activeTab, setActiveTab] = useState('profile');
  const [showCurrentPassword, setShowCurrentPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);
  const [isUpdating, setIsUpdating] = useState(false);
  const [updateSuccess, setUpdateSuccess] = useState<string | null>(null);
  const [updateError, setUpdateError] = useState<string | null>(null);
  const router = useRouter();

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
  }, [user, profileForm]);

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
      const token = getAccessToken();
      const response = await fetch('http://localhost:8080/api/v1/auth/profile', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          first_name: data.name.split(' ')[0] || data.name,
          last_name: data.name.split(' ').slice(1).join(' ') || '',
          email: data.email,
          phone: data.phone,
        }),
      });

      if (response.ok) {
        const updatedUser = await response.json();
        showSuccessMessage('Profile updated successfully!');
        // Update local user data if needed
      } else {
        const errorData = await response.json();
        showErrorMessage(errorData.message || 'Failed to update profile. Please try again.');
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
      const response = await fetch('http://localhost:8080/api/v1/auth/change-password', {
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
        showErrorMessage(errorData.message || 'Failed to change password. Please check your current password.');
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
      const response = await fetch('http://localhost:8080/api/v1/auth/preferences', {
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
      const response = await fetch('http://localhost:8080/api/v1/auth/delete-account', {
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

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  const tabs = [
    { id: 'profile', name: 'Profile', icon: UserCircleIcon },
    { id: 'password', name: 'Password', icon: KeyIcon },
    { id: 'preferences', name: 'Preferences', icon: BellIcon },
    { id: 'security', name: 'Security', icon: ShieldCheckIcon },
    { id: 'danger', name: 'Danger Zone', icon: ExclamationTriangleIcon },
  ];

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-black">Settings</h1>
            <p className="mt-2 text-gray-600">
              Manage your account settings and preferences.
            </p>
          </div>

          {/* Success/Error Messages */}
          {updateSuccess && (
            <div className="mb-6 p-4 bg-green-50 border border-green-200 rounded-lg flex items-center space-x-2">
              <CheckCircleIcon className="h-5 w-5 text-green-600" />
              <span className="text-green-700">{updateSuccess}</span>
            </div>
          )}

          {updateError && (
            <div className="mb-6 p-4 bg-red-50 border border-red-200 rounded-lg flex items-center space-x-2">
              <ExclamationTriangleIcon className="h-5 w-5 text-red-600" />
              <span className="text-red-700">{updateError}</span>
            </div>
          )}

          <div className="bg-white shadow rounded-lg">
            {/* Tab Navigation */}
            <div className="border-b border-gray-200">
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
                  <h2 className="text-lg font-medium text-black mb-4">Profile Information</h2>
                  <form onSubmit={profileForm.handleSubmit(onProfileSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label htmlFor="name" className="block text-sm font-medium text-black">
                          Full Name
                        </label>
                        <input
                          {...profileForm.register('name')}
                          type="text"
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        />
                        {profileForm.formState.errors.name && (
                          <p className="mt-1 text-sm text-red-600">{profileForm.formState.errors.name.message}</p>
                        )}
                      </div>

                      <div>
                        <label htmlFor="email" className="block text-sm font-medium text-black">
                          Email Address
                        </label>
                        <input
                          {...profileForm.register('email')}
                          type="email"
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        />
                        {profileForm.formState.errors.email && (
                          <p className="mt-1 text-sm text-red-600">{profileForm.formState.errors.email.message}</p>
                        )}
                      </div>
                    </div>

                    <div>
                      <label htmlFor="phone" className="block text-sm font-medium text-black">
                        Phone Number (Optional)
                      </label>
                      <input
                        {...profileForm.register('phone')}
                        type="tel"
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        placeholder="+1 (555) 123-4567"
                      />
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="bg-primary-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isUpdating ? 'Updating...' : 'Update Profile'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'password' && (
                <div>
                  <h2 className="text-lg font-medium text-black mb-4">Change Password</h2>
                  <form onSubmit={passwordForm.handleSubmit(onPasswordSubmit)} className="space-y-6">
                    <div>
                      <label htmlFor="currentPassword" className="block text-sm font-medium text-black">
                        Current Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('currentPassword')}
                          type={showCurrentPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                        >
                          {showCurrentPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-gray-400" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-gray-400" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.currentPassword && (
                        <p className="mt-1 text-sm text-red-600">{passwordForm.formState.errors.currentPassword.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="newPassword" className="block text-sm font-medium text-black">
                        New Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('newPassword')}
                          type={showNewPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowNewPassword(!showNewPassword)}
                        >
                          {showNewPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-gray-400" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-gray-400" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.newPassword && (
                        <p className="mt-1 text-sm text-red-600">{passwordForm.formState.errors.newPassword.message}</p>
                      )}
                    </div>

                    <div>
                      <label htmlFor="confirmPassword" className="block text-sm font-medium text-black">
                        Confirm New Password
                      </label>
                      <div className="mt-1 relative">
                        <input
                          {...passwordForm.register('confirmPassword')}
                          type={showConfirmPassword ? 'text' : 'password'}
                          className="block w-full px-3 py-2 pr-10 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                        />
                        <button
                          type="button"
                          className="absolute inset-y-0 right-0 pr-3 flex items-center"
                          onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                        >
                          {showConfirmPassword ? (
                            <EyeSlashIcon className="h-5 w-5 text-gray-400" />
                          ) : (
                            <EyeIcon className="h-5 w-5 text-gray-400" />
                          )}
                        </button>
                      </div>
                      {passwordForm.formState.errors.confirmPassword && (
                        <p className="mt-1 text-sm text-red-600">{passwordForm.formState.errors.confirmPassword.message}</p>
                      )}
                    </div>

                    <div className="flex justify-end">
                      <button
                        type="submit"
                        disabled={isUpdating}
                        className="bg-primary-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary-700 focus:outline-none focus:ring-2 focus:ring-primary-500 disabled:opacity-50 disabled:cursor-not-allowed"
                      >
                        {isUpdating ? 'Changing...' : 'Change Password'}
                      </button>
                    </div>
                  </form>
                </div>
              )}

              {activeTab === 'preferences' && (
                <div>
                  <h2 className="text-lg font-medium text-black mb-4">Preferences</h2>
                  <form onSubmit={preferencesForm.handleSubmit(onPreferencesSubmit)} className="space-y-6">
                    <div className="grid grid-cols-1 gap-6 sm:grid-cols-2">
                      <div>
                        <label htmlFor="defaultExpiration" className="block text-sm font-medium text-black">
                          Default URL Expiration
                        </label>
                        <select
                          {...preferencesForm.register('defaultExpiration')}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
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
                        <label htmlFor="timezone" className="block text-sm font-medium text-black">
                          Timezone
                        </label>
                        <select
                          {...preferencesForm.register('timezone')}
                          className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
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
                      <label htmlFor="theme" className="block text-sm font-medium text-black">
                        Theme
                      </label>
                      <select
                        {...preferencesForm.register('theme')}
                        className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm focus:outline-none focus:ring-primary-500 focus:border-primary-500 text-black"
                      >
                        <option value="light">Light</option>
                        <option value="dark">Dark</option>
                        <option value="system">System</option>
                      </select>
                    </div>

                    <div className="space-y-4">
                      <h3 className="text-md font-medium text-black">Notifications</h3>
                      
                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="emailNotifications" className="text-sm font-medium text-black">
                            Email Notifications
                          </label>
                          <p className="text-sm text-gray-600">Receive important updates about your account</p>
                        </div>
                        <input
                          {...preferencesForm.register('emailNotifications')}
                          type="checkbox"
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="marketingEmails" className="text-sm font-medium text-black">
                            Marketing Emails
                          </label>
                          <p className="text-sm text-gray-600">Receive product updates and promotional content</p>
                        </div>
                        <input
                          {...preferencesForm.register('marketingEmails')}
                          type="checkbox"
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
                        />
                      </div>

                      <div className="flex items-center justify-between">
                        <div>
                          <label htmlFor="analyticsPublic" className="text-sm font-medium text-black">
                            Public Analytics
                          </label>
                          <p className="text-sm text-gray-600">Allow others to view click statistics for your URLs</p>
                        </div>
                        <input
                          {...preferencesForm.register('analyticsPublic')}
                          type="checkbox"
                          className="h-4 w-4 text-primary-600 focus:ring-primary-500 border-gray-300 rounded"
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
                <div>
                  <h2 className="text-lg font-medium text-black mb-4">Security & Privacy</h2>
                  <div className="space-y-6">
                    <div className="bg-gray-50 p-4 rounded-lg">
                      <h3 className="text-md font-medium text-black mb-2">Account Status</h3>
                      <div className="space-y-2">
                        <div className="flex items-center justify-between">
                          <span className="text-sm text-gray-600">Email Verified</span>
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
                          <span className="text-sm text-gray-600">Phone Verified</span>
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
                          <span className="text-sm text-gray-600">Account Type</span>
                          <span className="text-sm text-black font-medium capitalize">{user?.account_type}</span>
                        </div>
                      </div>
                    </div>

                    <div className="space-y-3">
                      <h3 className="text-md font-medium text-black">Security Actions</h3>
                      
                      {!user?.email_verified && (
                        <button className="w-full text-left p-4 border border-yellow-200 bg-yellow-50 rounded-lg hover:bg-yellow-100 transition-colors">
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-sm font-medium text-yellow-800">Verify Email Address</h4>
                              <p className="text-sm text-yellow-600">Complete email verification to secure your account</p>
                            </div>
                            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600" />
                          </div>
                        </button>
                      )}

                      {user?.phone && !user?.phone_verified && (
                        <button className="w-full text-left p-4 border border-yellow-200 bg-yellow-50 rounded-lg hover:bg-yellow-100 transition-colors">
                          <div className="flex items-center justify-between">
                            <div>
                              <h4 className="text-sm font-medium text-yellow-800">Verify Phone Number</h4>
                              <p className="text-sm text-yellow-600">Add extra security with phone verification</p>
                            </div>
                            <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600" />
                          </div>
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              )}

              {activeTab === 'danger' && (
                <div>
                  <h2 className="text-lg font-medium text-red-600 mb-4">Danger Zone</h2>
                  <div className="space-y-4">
                    <div className="border border-red-200 bg-red-50 p-4 rounded-lg">
                      <h3 className="text-md font-medium text-red-800 mb-2">Delete Account</h3>
                      <p className="text-sm text-red-600 mb-4">
                        Permanently delete your account and all associated data. This action cannot be undone.
                      </p>
                      <div className="space-y-2 text-sm text-red-600">
                        <p>• All your shortened URLs will be deleted</p>
                        <p>• All analytics data will be permanently removed</p>
                        <p>• Your account cannot be recovered</p>
                      </div>
                      <button
                        onClick={handleDeleteAccount}
                        disabled={isUpdating}
                        className="mt-4 bg-red-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-red-500 disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
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
}