'use client';

import { useAdminAuth } from '@/contexts/AdminAuthContext';
import Link from 'next/link';
import {
  ShieldCheckIcon,
  ArrowRightOnRectangleIcon
} from '@heroicons/react/24/outline';

export default function AdminHeader() {
  const { adminUser, logout } = useAdminAuth();

  const handleLogout = async () => {
    if (confirm('Are you sure you want to logout?')) {
      await logout();
    }
  };

  return (
    <header className="bg-white border-b border-gray-200 shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo and Title */}
          <Link href="/admin" className="flex items-center space-x-3">
            <ShieldCheckIcon className="h-8 w-8 text-red-600" />
            <div>
              <h1 className="text-lg font-bold text-gray-900">Admin Portal</h1>
              <p className="text-xs text-gray-500">URL Shortener</p>
            </div>
          </Link>

          {/* User Info and Logout */}
          <div className="flex items-center space-x-4">
            {adminUser && (
              <>
                <div className="text-right">
                  <p className="text-sm font-medium text-gray-900">{adminUser.name}</p>
                  <p className="text-xs text-gray-500 capitalize">{adminUser.account_type.replace('_', ' ')}</p>
                </div>
                <button
                  onClick={handleLogout}
                  className="flex items-center space-x-2 px-4 py-2 text-sm font-medium text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg transition-colors"
                  title="Logout"
                >
                  <ArrowRightOnRectangleIcon className="h-5 w-5" />
                  <span>Logout</span>
                </button>
              </>
            )}
          </div>
        </div>
      </div>
    </header>
  );
}
