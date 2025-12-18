'use client';

import React, { createContext, useContext, useState, useEffect, useCallback, useRef } from 'react';
import {
  validateToken,
  storeAuthTokens,
  shouldRefreshToken,
  getTokenRemainingTime,
  clearAuthTokens,
  getAccessToken as getStoredAccessToken,
  getRefreshToken as getStoredRefreshToken,
} from '@/utils/tokenValidation';

interface User {
  id: number;
  name: string;
  email: string;
  phone?: string;
  phone_verified: boolean;
  email_verified: boolean;
  provider: string;
  avatar_url?: string;
  account_type: string;
  is_active: boolean;
  last_login_at?: string;
  created_at: string;
}

interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (email: string, password: string, rememberMe?: boolean) => Promise<AuthResponse>;
  register: (data: RegisterData) => Promise<AuthResponse>;
  logout: () => void;
  refreshToken: () => Promise<void>;
  getAccessToken: () => string | null;
  isTokenExpired: (token?: string) => boolean;
}

interface RegisterData {
  name: string;
  email: string;
  password: string;
  phone: string;
  terms_accepted: boolean;
  marketing_consent?: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Helper function to check if error is a network error
const isNetworkError = (error: any): boolean => {
  return (
    error instanceof TypeError &&
    (error.message === 'Failed to fetch' ||
     error.message === 'Network request failed' ||
     error.message.includes('fetch'))
  );
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  // Promise reference to prevent concurrent refresh attempts (Issue #36)
  const refreshPromiseRef = useRef<Promise<void> | null>(null);

  const isAuthenticated = !!user;

  // Wrapper for getAccessToken to maintain backward compatibility
  const getAccessToken = useCallback((): string | null => {
    return getStoredAccessToken();
  }, []);

  // Wrapper for isTokenExpired to maintain backward compatibility (uses utility internally)
  const isTokenExpired = useCallback((token?: string, bufferSeconds: number = 0): boolean => {
    const tokenToCheck = token || getStoredAccessToken();
    if (!tokenToCheck) return true;

    // Use utility function with proper validation (Issue #35)
    const validation = validateToken(tokenToCheck);
    if (!validation.isValid) return true;

    // Check if token expires within the buffer time
    const currentTime = Math.floor(Date.now() / 1000);
    return validation.payload!.exp < (currentTime + bufferSeconds);
  }, []);

  const logout = useCallback(() => {
    console.log('[Auth] Logging out user');

    // Save token before removing for logout endpoint call
    const token = getStoredAccessToken();

    // Remove auth data using utility (Issue #35)
    clearAuthTokens();
    setUser(null);

    // Call logout endpoint if we had a token
    if (token) {
      fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }).catch((error) => {
        console.error('[Auth] Logout endpoint error:', error);
      });
    }
  }, []);

  const refreshToken = useCallback(async (retryCount = 0): Promise<void> => {
    // Check if refresh is already in progress (Issue #36 - Race condition fix)
    if (refreshPromiseRef.current) {
      console.log('[Auth] Refresh already in progress, waiting for existing promise...');
      return refreshPromiseRef.current;
    }

    console.log('[Auth] Refreshing token, attempt:', retryCount + 1);

    // Create refresh promise and store reference
    const refreshPromise = (async () => {
      const refreshTokenValue = getStoredRefreshToken();

      if (!refreshTokenValue) {
        console.error('[Auth] No refresh token found');
        refreshPromiseRef.current = null;
        logout();
        return;
      }

      // Validate refresh token structure before attempting refresh (Issue #35)
      const refreshValidation = validateToken(refreshTokenValue);
      if (!refreshValidation.isValid && !refreshValidation.error?.includes('expired')) {
        console.error('[Auth] Invalid refresh token structure:', refreshValidation.error);
        refreshPromiseRef.current = null;
        logout();
        return;
      }

      setIsRefreshing(true);

      try {
        const response = await fetch(`${API_BASE_URL}/api/v1/auth/refresh`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify({
            refresh_token: refreshTokenValue,
          }),
          signal: AbortSignal.timeout(10000),
        });

        const data = await response.json();

        if (!response.ok) {
          throw new Error(data.error || 'Token refresh failed');
        }

        console.log('[Auth] Token refresh successful');

        // Validate and store tokens using utility (Issues #35, #37)
        try {
          storeAuthTokens(data.access_token, data.refresh_token, data.expires_in);
          localStorage.setItem('user_data', JSON.stringify(data.user));
          setUser(data.user);
        } catch (validationError) {
          console.error('[Auth] Token validation failed:', validationError);
          refreshPromiseRef.current = null;
          logout();
          return;
        }
      } catch (error) {
        console.error('[Auth] Token refresh error:', error);

        // Handle network errors with exponential backoff (Issue #36)
        if (isNetworkError(error) && retryCount < 3) {
          const delaySeconds = Math.pow(2, retryCount);
          console.log(`[Auth] Network error, retrying in ${delaySeconds} seconds...`);

          await new Promise(resolve => setTimeout(resolve, delaySeconds * 1000));

          setIsRefreshing(false);
          refreshPromiseRef.current = null;
          return refreshToken(retryCount + 1);
        }

        // For network errors that exceed retry limit, don't logout (Issue #36)
        if (isNetworkError(error)) {
          console.warn('[Auth] Backend unavailable after retries, continuing without token refresh');
          refreshPromiseRef.current = null;
          return;
        }

        // For auth errors, logout
        console.error('[Auth] Authentication error, logging out');
        refreshPromiseRef.current = null;
        logout();
        throw error;
      } finally {
        setIsRefreshing(false);
        refreshPromiseRef.current = null;
      }
    })();

    refreshPromiseRef.current = refreshPromise;
    return refreshPromise;
  }, [logout]);

  // Initialize auth state from localStorage
  useEffect(() => {
    const initAuth = async () => {
      console.log('[Auth] Initializing auth state');

      const token = getStoredAccessToken();
      const userData = localStorage.getItem('user_data');

      if (token && userData && userData !== 'undefined') {
        try {
          // Validate token structure first (Issue #35)
          const tokenValidation = validateToken(token);

          if (!tokenValidation.isValid && !tokenValidation.error?.includes('expired')) {
            // Token has invalid structure (malformed)
            console.error('[Auth] Invalid token structure, logging out:', tokenValidation.error);
            logout();
            setIsLoading(false);
            return;
          }

          const parsedUser = JSON.parse(userData);

          // Always load user from cache first (for instant UI update)
          console.log('[Auth] Loading user from cache');
          setUser(parsedUser);

          // Check if token should be refreshed using stored expiry (Issue #37)
          if (shouldRefreshToken()) {
            const remaining = getTokenRemainingTime();
            console.log(`[Auth] Token refresh needed (remaining: ${remaining}s), refreshing...`);
            try {
              await refreshToken();
            } catch (error) {
              console.error('[Auth] Failed to refresh token:', error);
              // Only logout if refresh fails due to auth error (not network)
              if (!isNetworkError(error)) {
                logout();
              }
            }
          } else {
            const remaining = getTokenRemainingTime();
            console.log(`[Auth] Token valid (remaining: ${remaining}s)`);
          }
        } catch (error) {
          console.error('[Auth] Error parsing stored user data:', error);
          logout();
        }
      } else {
        console.log('[Auth] No stored credentials found');
      }

      setIsLoading(false);
    };

    initAuth();
  }, [logout, refreshToken]);

  // Set up periodic token refresh check
  useEffect(() => {
    if (!isAuthenticated) return;

    console.log('[Auth] Setting up token refresh interval');

    const tokenRefreshInterval = setInterval(() => {
      // Use stored timestamp for efficient expiry check (Issue #37)
      if (shouldRefreshToken()) {
        const remaining = getTokenRemainingTime();
        console.log(`[Auth] Token expiring soon (remaining: ${remaining}s), refreshing...`);
        refreshToken().catch((error) => {
          console.error('[Auth] Auto-refresh failed:', error);
        });
      }
    }, 60000); // Check every minute

    return () => {
      console.log('[Auth] Clearing token refresh interval');
      clearInterval(tokenRefreshInterval);
    };
  }, [isAuthenticated, refreshToken]);

  const login = async (email: string, password: string, rememberMe = false): Promise<AuthResponse> => {
    console.log('[Auth] Logging in user:', email);

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/login`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          email,
          password,
          remember_me: rememberMe,
        }),
        signal: AbortSignal.timeout(15000),
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Login failed');
      }

      console.log('[Auth] Login successful');

      // Validate and store auth data using utility (Issues #35, #37)
      try {
        storeAuthTokens(data.access_token, data.refresh_token, data.expires_in);
        localStorage.setItem('user_data', JSON.stringify(data.user));
        setUser(data.user);
      } catch (validationError) {
        console.error('[Auth] Token validation failed after login:', validationError);
        throw new Error('Received invalid tokens from server');
      }

      return data;
    } catch (error) {
      console.error('[Auth] Login error:', error);

      if (isNetworkError(error)) {
        throw new Error('Unable to connect to server. Please check your internet connection and try again.');
      }

      throw error;
    }
  };

  const register = async (data: RegisterData): Promise<AuthResponse> => {
    console.log('[Auth] Registering user:', data.email);

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
        signal: AbortSignal.timeout(15000),
      });

      let result;
      const text = await response.text();
      try {
        result = text ? JSON.parse(text) : {};
      } catch (jsonError) {
        console.error('[Auth] Failed to parse registration response:', text);
        throw new Error('Invalid server response');
      }

      if (!response.ok) {
        throw new Error(result.error || 'Registration failed');
      }

      console.log('[Auth] Registration successful');

      if (result.access_token) {
        // Validate and store auth data using utility (Issues #35, #37)
        try {
          storeAuthTokens(result.access_token, result.refresh_token, result.expires_in);
          localStorage.setItem('user_data', JSON.stringify(result.user));
          setUser(result.user);
        } catch (validationError) {
          console.error('[Auth] Token validation failed after registration:', validationError);
          throw new Error('Received invalid tokens from server');
        }
      }

      return result;
    } catch (error) {
      console.error('[Auth] Registration error:', error);

      if (isNetworkError(error)) {
        throw new Error('Unable to connect to server. Please check your internet connection and try again.');
      }

      throw error;
    }
  };

  const value = {
    user,
    isAuthenticated,
    isLoading,
    login,
    register,
    logout,
    refreshToken,
    getAccessToken,
    isTokenExpired,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
