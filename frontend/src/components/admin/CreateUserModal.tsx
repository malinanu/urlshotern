'use client';

import { useState, useEffect } from 'react';
import { Modal } from '@/components/ui/modal';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  Role,
  CreateUserRequest,
  CreateUserResponse,
} from '@/types/rbac';
import {
  ExclamationTriangleIcon,
  CheckCircleIcon,
  ClipboardDocumentIcon,
} from '@heroicons/react/24/outline';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  onUserCreated: () => void;
}

export function CreateUserModal({ isOpen, onClose, onUserCreated }: Props) {
  const { getAccessToken } = useAuth();
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [createdUserInfo, setCreatedUserInfo] = useState<CreateUserResponse | null>(null);
  const [copiedPassword, setCopiedPassword] = useState(false);

  const [formData, setFormData] = useState<CreateUserRequest>({
    name: '',
    email: '',
    password: '',
    phone: '',
    account_type: 'free',
    role_ids: [],
    send_invite: false,
    email_verified: false,
    is_active: true,
    force_password_reset: false,
  });

  // Fetch available roles
  useEffect(() => {
    if (isOpen) {
      fetchRoles();
      // Reset form when modal opens
      setFormData({
        name: '',
        email: '',
        password: '',
        phone: '',
        account_type: 'free',
        role_ids: [],
        send_invite: false,
        email_verified: false,
        is_active: true,
        force_password_reset: false,
      });
      setCreatedUserInfo(null);
      setError('');
    }
  }, [isOpen]);

  const fetchRoles = async () => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(API_ENDPOINTS.admin.rbac.roles, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setRoles(data.roles || []);
      }
    } catch (err) {
      console.error('Error fetching roles:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setLoading(true);

    // Validate form
    if (!formData.name.trim()) {
      setError('Name is required');
      setLoading(false);
      return;
    }

    if (!formData.email.trim()) {
      setError('Email is required');
      setLoading(false);
      return;
    }

    if (formData.role_ids.length === 0) {
      setError('Please select at least one role');
      setLoading(false);
      return;
    }

    const token = getAccessToken();
    if (!token) {
      setError('Not authenticated');
      setLoading(false);
      return;
    }

    try {
      // Prepare request data - convert empty strings to null for optional fields
      const requestData = {
        name: formData.name.trim(),
        email: formData.email.trim(),
        password: formData.password?.trim() || null,
        phone: formData.phone?.trim() || null,
        account_type: formData.account_type,
        role_ids: formData.role_ids,
        send_invite: formData.send_invite,
        email_verified: formData.email_verified,
        is_active: formData.is_active,
        force_password_reset: formData.force_password_reset,
      };

      // Remove null values for cleaner request
      const cleanedData = Object.fromEntries(
        Object.entries(requestData).filter(([_, v]) => v !== null && v !== '')
      );

      console.log('Creating user with data:', cleanedData);

      const response = await fetch(API_ENDPOINTS.admin.createUser, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(cleanedData),
      });

      if (response.ok) {
        const data: CreateUserResponse = await response.json();
        console.log('User created successfully:', data);
        setCreatedUserInfo(data);
        // Call the callback to refresh user list
        onUserCreated();
      } else {
        const errorData = await response.json().catch(() => ({ error: 'Failed to create user' }));
        console.error('Error response:', errorData);
        setError(errorData.error || errorData.message || 'Failed to create user');
      }
    } catch (err) {
      setError('Network error. Please try again.');
      console.error('Error creating user:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleRoleToggle = (roleId: number) => {
    setFormData(prev => ({
      ...prev,
      role_ids: prev.role_ids.includes(roleId)
        ? prev.role_ids.filter(id => id !== roleId)
        : [...prev.role_ids, roleId],
    }));
  };

  const copyPassword = () => {
    if (createdUserInfo?.temp_password) {
      navigator.clipboard.writeText(createdUserInfo.temp_password);
      setCopiedPassword(true);
      setTimeout(() => setCopiedPassword(false), 2000);
    }
  };

  const handleClose = () => {
    setCreatedUserInfo(null);
    setError('');
    onClose();
  };

  return (
    <Modal isOpen={isOpen} onClose={handleClose} title="Create New User" size="2xl">
      {!createdUserInfo ? (
        <form onSubmit={handleSubmit} className="space-y-6">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-start space-x-2">
              <ExclamationTriangleIcon className="h-5 w-5 flex-shrink-0 mt-0.5" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          {/* User Information Section */}
          <div>
            <h3 className="text-sm font-medium text-gray-900 mb-3">User Information</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Name <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="John Doe"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Email <span className="text-red-500">*</span>
                </label>
                <input
                  type="email"
                  value={formData.email}
                  onChange={(e) => setFormData({ ...formData, email: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="john@example.com"
                  required
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Phone (Optional)
                </label>
                <input
                  type="tel"
                  value={formData.phone}
                  onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="+1 (555) 123-4567"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Password (Optional)
                </label>
                <input
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  placeholder="Leave empty to auto-generate"
                />
                <p className="mt-1 text-xs text-gray-500">
                  If left empty, a temporary password will be generated
                </p>
              </div>
            </div>
          </div>

          {/* Account Settings Section */}
          <div className="border-t pt-6">
            <h3 className="text-sm font-medium text-gray-900 mb-3">Account Settings</h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Account Type
                </label>
                <select
                  value={formData.account_type}
                  onChange={(e) => setFormData({ ...formData, account_type: e.target.value as any })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="free">Free</option>
                  <option value="premium">Premium</option>
                  <option value="enterprise">Enterprise</option>
                </select>
              </div>

              <div className="space-y-3">
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={formData.email_verified}
                    onChange={(e) => setFormData({ ...formData, email_verified: e.target.checked })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Email Verified</span>
                </label>

                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Active</span>
                </label>
              </div>

              <div className="space-y-3">
                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={formData.force_password_reset}
                    onChange={(e) => setFormData({ ...formData, force_password_reset: e.target.checked })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Force Password Reset</span>
                </label>

                <label className="flex items-center space-x-2">
                  <input
                    type="checkbox"
                    checked={formData.send_invite}
                    onChange={(e) => setFormData({ ...formData, send_invite: e.target.checked })}
                    className="rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <span className="text-sm text-gray-700">Send Invitation Email</span>
                </label>
              </div>
            </div>
          </div>

          {/* Role Assignment Section */}
          <div className="border-t pt-6">
            <h3 className="text-sm font-medium text-gray-900 mb-3">
              Role Assignment <span className="text-red-500">*</span>
            </h3>
            <div className="space-y-2 max-h-60 overflow-y-auto">
              {roles.map((role) => (
                <label
                  key={role.id}
                  className="flex items-start space-x-3 p-3 border border-gray-200 rounded-lg hover:bg-gray-50 cursor-pointer"
                >
                  <input
                    type="checkbox"
                    checked={formData.role_ids.includes(role.id)}
                    onChange={() => handleRoleToggle(role.id)}
                    className="mt-0.5 rounded border-gray-300 text-blue-600 focus:ring-blue-500"
                  />
                  <div className="flex-1">
                    <div className="flex items-center space-x-2">
                      <span className="text-sm font-medium text-gray-900">{role.display_name}</span>
                      {role.is_system && (
                        <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                          System
                        </span>
                      )}
                    </div>
                    <p className="text-xs text-gray-500 mt-0.5">{role.description}</p>
                  </div>
                </label>
              ))}
            </div>
          </div>

          {/* Form Actions */}
          <div className="border-t pt-6 flex justify-end space-x-3">
            <button
              type="button"
              onClick={handleClose}
              className="px-4 py-2 border border-gray-300 rounded-lg text-gray-700 hover:bg-gray-50 transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={loading}
            >
              {loading ? 'Creating...' : 'Create User'}
            </button>
          </div>
        </form>
      ) : (
        <div className="space-y-6">
          {/* Success Message */}
          <div className="bg-green-50 border border-green-200 text-green-700 px-4 py-3 rounded-lg flex items-start space-x-2">
            <CheckCircleIcon className="h-5 w-5 flex-shrink-0 mt-0.5" />
            <div className="flex-1">
              <p className="text-sm font-medium">User created successfully!</p>
              <p className="text-xs mt-1">
                {createdUserInfo.user.name} ({createdUserInfo.user.email})
              </p>
            </div>
          </div>

          {/* Temporary Password Display */}
          {createdUserInfo.temp_password && (
            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-4">
              <div className="flex items-start space-x-2">
                <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600 flex-shrink-0 mt-0.5" />
                <div className="flex-1">
                  <p className="text-sm font-medium text-yellow-900 mb-2">
                    Temporary Password Generated
                  </p>
                  <p className="text-xs text-yellow-700 mb-3">
                    This password will only be shown once. Please copy it now.
                  </p>
                  <div className="bg-white border border-yellow-300 rounded px-3 py-2 flex items-center justify-between">
                    <code className="text-sm font-mono text-gray-900">
                      {createdUserInfo.temp_password}
                    </code>
                    <button
                      onClick={copyPassword}
                      className="ml-2 p-1 text-yellow-700 hover:text-yellow-900 transition-colors"
                      title="Copy password"
                    >
                      {copiedPassword ? (
                        <CheckCircleIcon className="h-5 w-5 text-green-600" />
                      ) : (
                        <ClipboardDocumentIcon className="h-5 w-5" />
                      )}
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Invitation URL */}
          {createdUserInfo.invite_url && (
            <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <p className="text-sm font-medium text-blue-900 mb-2">Invitation Sent</p>
              <p className="text-xs text-blue-700">
                An invitation email has been sent to {createdUserInfo.user.email}
              </p>
            </div>
          )}

          {/* Assigned Roles */}
          <div>
            <h3 className="text-sm font-medium text-gray-900 mb-2">Assigned Roles</h3>
            <div className="flex flex-wrap gap-2">
              {createdUserInfo.roles.map((role) => (
                <span
                  key={role.id}
                  className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800"
                >
                  {role.display_name}
                </span>
              ))}
            </div>
          </div>

          {/* Close Button */}
          <div className="border-t pt-6 flex justify-end">
            <button
              onClick={handleClose}
              className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors"
            >
              Done
            </button>
          </div>
        </div>
      )}
    </Modal>
  );
}
