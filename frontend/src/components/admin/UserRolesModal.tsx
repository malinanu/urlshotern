'use client';

import { useState, useEffect } from 'react';
import { Modal } from '@/components/ui/modal';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { Role, UserRoles, Permission } from '@/types/rbac';
import {
  ShieldCheckIcon,
  XMarkIcon,
  PlusIcon,
  ChevronDownIcon,
  ChevronUpIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

interface Props {
  isOpen: boolean;
  onClose: () => void;
  userId: number;
  userName: string;
}

export function UserRolesModal({ isOpen, onClose, userId, userName }: Props) {
  const { getAccessToken } = useAuth();
  const [userRoles, setUserRoles] = useState<UserRoles | null>(null);
  const [availableRoles, setAvailableRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedRoleId, setSelectedRoleId] = useState<number | null>(null);
  const [showPermissions, setShowPermissions] = useState(false);
  const [actionLoading, setActionLoading] = useState(false);

  useEffect(() => {
    if (isOpen && userId) {
      fetchData();
    }
  }, [isOpen, userId]);

  const fetchData = async () => {
    setLoading(true);
    setError('');
    await Promise.all([fetchUserRoles(), fetchAvailableRoles()]);
    setLoading(false);
  };

  const fetchUserRoles = async () => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(API_ENDPOINTS.admin.getUserRoles(userId.toString()), {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data: UserRoles = await response.json();
        // Ensure data has valid structure
        setUserRoles({
          roles: data.roles || [],
          permissions: data.permissions || [],
        });
      } else {
        const errorData = await response.json().catch(() => ({ error: 'Failed to fetch user roles' }));
        setError(errorData.error || 'Failed to fetch user roles');
      }
    } catch (err) {
      console.error('Error fetching user roles:', err);
      setError('Network error. Please try again.');
    }
  };

  const fetchAvailableRoles = async () => {
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
        setAvailableRoles(data.roles || []);
      }
    } catch (err) {
      console.error('Error fetching available roles:', err);
    }
  };

  const handleAssignRole = async () => {
    if (!selectedRoleId) return;

    setActionLoading(true);
    setError('');

    const token = getAccessToken();
    if (!token) {
      setError('Not authenticated');
      setActionLoading(false);
      return;
    }

    try {
      const response = await fetch(API_ENDPOINTS.admin.assignUserRole(userId.toString()), {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ role_id: selectedRoleId }),
      });

      if (response.ok) {
        await fetchUserRoles();
        setSelectedRoleId(null);
      } else {
        const errorData = await response.json().catch(() => ({ error: 'Failed to assign role' }));
        setError(errorData.error || 'Failed to assign role');
      }
    } catch (err) {
      console.error('Error assigning role:', err);
      setError('Network error. Please try again.');
    } finally {
      setActionLoading(false);
    }
  };

  const handleRemoveRole = async (roleId: number) => {
    if (!confirm('Are you sure you want to remove this role from the user?')) {
      return;
    }

    setActionLoading(true);
    setError('');

    const token = getAccessToken();
    if (!token) {
      setError('Not authenticated');
      setActionLoading(false);
      return;
    }

    try {
      const response = await fetch(
        API_ENDPOINTS.admin.removeUserRole(userId.toString(), roleId.toString()),
        {
          method: 'DELETE',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (response.ok) {
        await fetchUserRoles();
      } else {
        const errorData = await response.json().catch(() => ({ error: 'Failed to remove role' }));
        setError(errorData.error || 'Failed to remove role');
      }
    } catch (err) {
      console.error('Error removing role:', err);
      setError('Network error. Please try again.');
    } finally {
      setActionLoading(false);
    }
  };

  // Get roles that can be added (not already assigned)
  const assignableRoles = availableRoles.filter(
    (role) => !(userRoles?.roles || []).some((ur) => ur.id === role.id)
  );

  return (
    <Modal isOpen={isOpen} onClose={onClose} title={`Manage Roles: ${userName}`} size="2xl">
      {loading ? (
        <div className="flex items-center justify-center py-12">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
        </div>
      ) : (
        <div className="space-y-6">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 px-4 py-3 rounded-lg flex items-start space-x-2">
              <ExclamationTriangleIcon className="h-5 w-5 flex-shrink-0 mt-0.5" />
              <span className="text-sm">{error}</span>
            </div>
          )}

          {/* Current Roles Section */}
          <div>
            <h3 className="text-sm font-medium text-gray-900 mb-3">Current Roles</h3>
            {userRoles && userRoles.roles.length > 0 ? (
              <div className="space-y-2">
                {userRoles.roles.map((role) => (
                  <div
                    key={role.id}
                    className="flex items-center justify-between p-3 border border-gray-200 rounded-lg bg-gray-50"
                  >
                    <div className="flex items-center space-x-3">
                      <ShieldCheckIcon className="h-5 w-5 text-blue-600" />
                      <div>
                        <div className="flex items-center space-x-2">
                          <span className="text-sm font-medium text-gray-900">
                            {role.display_name}
                          </span>
                          {role.is_system && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                              System
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-gray-500 mt-0.5">{role.description}</p>
                      </div>
                    </div>
                    <button
                      onClick={() => handleRemoveRole(role.id)}
                      disabled={actionLoading}
                      className="p-1 text-red-600 hover:text-red-800 transition-colors disabled:opacity-50"
                      title="Remove role"
                    >
                      <XMarkIcon className="h-5 w-5" />
                    </button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 italic">No roles assigned</p>
            )}
          </div>

          {/* Add Role Section */}
          {assignableRoles.length > 0 && (
            <div className="border-t pt-6">
              <h3 className="text-sm font-medium text-gray-900 mb-3">Add Role</h3>
              <div className="flex items-end space-x-3">
                <div className="flex-1">
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Select Role
                  </label>
                  <select
                    value={selectedRoleId || ''}
                    onChange={(e) => setSelectedRoleId(Number(e.target.value))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    disabled={actionLoading}
                  >
                    <option value="">Choose a role...</option>
                    {assignableRoles.map((role) => (
                      <option key={role.id} value={role.id}>
                        {role.display_name} - {role.description}
                      </option>
                    ))}
                  </select>
                </div>
                <button
                  onClick={handleAssignRole}
                  disabled={!selectedRoleId || actionLoading}
                  className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center space-x-2"
                >
                  <PlusIcon className="h-4 w-4" />
                  <span>Add</span>
                </button>
              </div>
            </div>
          )}

          {/* Computed Permissions Section */}
          {userRoles && userRoles.permissions.length > 0 && (
            <div className="border-t pt-6">
              <button
                onClick={() => setShowPermissions(!showPermissions)}
                className="flex items-center justify-between w-full text-left"
              >
                <h3 className="text-sm font-medium text-gray-900">
                  Computed Permissions ({userRoles.permissions.length})
                </h3>
                {showPermissions ? (
                  <ChevronUpIcon className="h-5 w-5 text-gray-500" />
                ) : (
                  <ChevronDownIcon className="h-5 w-5 text-gray-500" />
                )}
              </button>

              {showPermissions && (
                <div className="mt-3 max-h-60 overflow-y-auto">
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-2">
                    {userRoles.permissions.map((permission) => (
                      <div
                        key={permission.id}
                        className="px-3 py-2 bg-gray-50 border border-gray-200 rounded-lg"
                      >
                        <div className="text-xs font-medium text-gray-900">
                          {permission.name}
                        </div>
                        <div className="text-xs text-gray-500 mt-0.5">
                          {permission.description}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          )}

          {/* Close Button */}
          <div className="border-t pt-6 flex justify-end">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-gray-200 text-gray-700 rounded-lg hover:bg-gray-300 transition-colors"
            >
              Close
            </button>
          </div>
        </div>
      )}
    </Modal>
  );
}
