'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import { adminAuthService } from '@/lib/admin-auth-service';
import { API_ENDPOINTS } from '@/lib/api-config';

interface AdminUser {
  id: number;
  name: string;
  email: string;
  account_type: string;
  is_admin: boolean;
  is_active: boolean;
  created_at: string;
}

interface AdminAuthContextType {
  adminUser: AdminUser | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshToken: () => Promise<void>;
}

const AdminAuthContext = createContext<AdminAuthContextType | undefined>(undefined);

export function AdminAuthProvider({ children }: { children: React.ReactNode }) {
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const router = useRouter();

  const isAuthenticated = !!adminUser;

  // Validate current session
  const validateSession = useCallback(async (): Promise<boolean> => {
    const token = adminAuthService.getAdminToken();

    if (!token) {
      return false;
    }

    try {
      // Validate token with backend by calling dashboard endpoint
      const response = await fetch(API_ENDPOINTS.admin.dashboard, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!response.ok) {
        throw new Error('Invalid token');
      }

      // Token is valid, get user data from localStorage
      const storedUser = adminAuthService.getAdminUser();
      if (storedUser) {
        setAdminUser(storedUser);
        return true;
      }

      return false;
    } catch (error) {
      console.error('Session validation failed:', error);
      // Clear invalid tokens
      adminAuthService.logout();
      return false;
    }
  }, []);

  // Initialize auth state
  useEffect(() => {
    const initAuth = async () => {
      console.log('[AdminAuth] Initializing admin authentication');

      const isValid = await validateSession();

      if (!isValid) {
        console.log('[AdminAuth] No valid session found');
        setAdminUser(null);
      } else {
        console.log('[AdminAuth] Valid session restored');
      }

      setIsLoading(false);
    };

    initAuth();
  }, [validateSession]);

  const login = useCallback(async (email: string, password: string): Promise<void> => {
    console.log('[AdminAuth] Attempting admin login');

    try {
      const response = await fetch(API_ENDPOINTS.adminAuth.login, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ email, password }),
        signal: AbortSignal.timeout(15000),
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({ error: 'Invalid credentials' }));
        throw new Error(errorData.error || 'Admin login failed');
      }

      const data = await response.json();

      // Validate response structure
      if (!data || !data.access_token || !data.refresh_token || !data.user) {
        throw new Error('Invalid response from server');
      }

      // Check if user has admin privileges
      if (!data.user.is_admin && !isAdminAccountType(data.user.account_type)) {
        throw new Error('Admin access required');
      }

      // Store admin auth data
      adminAuthService.setAuthData(data.access_token, data.refresh_token, data.user);
      setAdminUser(data.user);

      console.log('[AdminAuth] Admin login successful');
    } catch (error) {
      console.error('[AdminAuth] Login error:', error);
      throw error;
    }
  }, []);

  const logout = useCallback(async (): Promise<void> => {
    console.log('[AdminAuth] Logging out admin user');

    const token = adminAuthService.getAdminToken();

    // Clear local state first
    setAdminUser(null);
    adminAuthService.logout();

    // Call backend logout endpoint
    if (token) {
      try {
        await fetch(API_ENDPOINTS.adminAuth.logout, {
          method: 'POST',
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });
      } catch (error) {
        console.error('[AdminAuth] Logout endpoint error:', error);
      }
    }

    // Redirect to admin login
    router.push('/admin/login');
  }, [router]);

  const refreshToken = useCallback(async (): Promise<void> => {
    console.log('[AdminAuth] Refreshing admin token');

    try {
      const newToken = await adminAuthService.refreshToken();
      const updatedUser = adminAuthService.getAdminUser();

      if (updatedUser) {
        setAdminUser(updatedUser);
      }

      console.log('[AdminAuth] Token refresh successful');
    } catch (error) {
      console.error('[AdminAuth] Token refresh failed:', error);
      // Clear invalid session
      setAdminUser(null);
      adminAuthService.logout();
      throw error;
    }
  }, []);

  const value = {
    adminUser,
    isAuthenticated,
    isLoading,
    login,
    logout,
    refreshToken,
  };

  return (
    <AdminAuthContext.Provider value={value}>
      {children}
    </AdminAuthContext.Provider>
  );
}

export function useAdminAuth() {
  const context = useContext(AdminAuthContext);
  if (context === undefined) {
    throw new Error('useAdminAuth must be used within an AdminAuthProvider');
  }
  return context;
}

// Helper function to check admin account types
function isAdminAccountType(accountType: string): boolean {
  const adminTypes = [
    'super_admin',
    'system_admin',
    'support_admin',
    'billing_admin',
    'content_admin',
  ];
  return adminTypes.includes(accountType);
}
