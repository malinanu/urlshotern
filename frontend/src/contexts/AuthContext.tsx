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
import { API_ENDPOINTS } from '@/lib/api-config';
import { createLogger } from '@/utils/logger';

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

// Create logger for auth operations
const logger = createLogger('Auth');

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
    logger.info('Logging out user');

    // Save token before removing for logout endpoint call
    const token = getStoredAccessToken();

    // Remove auth data using utility (Issue #35)
    clearAuthTokens();
    setUser(null);

    // Call logout endpoint if we had a token
    if (token) {
      fetch(API_ENDPOINTS.auth.logout, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }).catch((error) => {
        logger.error('Logout endpoint error:', error);
      });
    }
  }, []);

  const refreshToken = useCallback(async (retryCount = 0): Promise<void> => {
    // Check if refresh is already in progress (Issue #36 - Race condition fix)
    if (refreshPromiseRef.current) {
      logger.debug('Refresh already in progress, waiting for existing promise...');
      return refreshPromiseRef.current;
    }

    logger.info('Refreshing token, attempt:', retryCount + 1);

    // Create refresh promise and store reference
    const refreshPromise = (async () => {
      const refreshTokenValue = getStoredRefreshToken();

      if (!refreshTokenValue) {
        logger.error('No refresh token found');
        refreshPromiseRef.current = null;
        logout();
        return;
      }

      // Validate refresh token structure before attempting refresh (Issue #35)
      const refreshValidation = validateToken(refreshTokenValue);
      if (!refreshValidation.isValid && !refreshValidation.error?.includes('expired')) {
        logger.error('Invalid refresh token structure:', refreshValidation.error);
        refreshPromiseRef.current = null;
        logout();
        return;
      }

      setIsRefreshing(true);

      try {
        const response = await fetch(API_ENDPOINTS.auth.refresh, {
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

        logger.info('Token refresh successful');

        // Validate and store tokens using utility (Issues #35, #37)
        try {
          storeAuthTokens(data.access_token, data.refresh_token, data.expires_in);
          localStorage.setItem('user_data', JSON.stringify(data.user));
          setUser(data.user);
        } catch (validationError) {
          logger.error('Token validation failed:', validationError);
          refreshPromiseRef.current = null;
          logout();
          return;
        }
      } catch (error) {
        logger.error('Token refresh error:', error);

        // Handle network errors with exponential backoff (Issue #36)
        if (isNetworkError(error) && retryCount < 3) {
          const delaySeconds = Math.pow(2, retryCount);
          logger.warn(`Network error, retrying in ${delaySeconds} seconds...`);

          await new Promise(resolve => setTimeout(resolve, delaySeconds * 1000));

          setIsRefreshing(false);
          refreshPromiseRef.current = null;
          return refreshToken(retryCount + 1);
        }

        // For network errors that exceed retry limit, don't logout (Issue #36)
        if (isNetworkError(error)) {
          logger.warn('Backend unavailable after retries, continuing without token refresh');
          refreshPromiseRef.current = null;
          return;
        }

        // For auth errors, logout
        logger.error('Authentication error, logging out');
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
      logger.debug('Initializing auth state');

      const token = getStoredAccessToken();
      const userData = localStorage.getItem('user_data');

      if (token && userData && userData !== 'undefined') {
        try {
          // Validate token structure first (Issue #35)
          const tokenValidation = validateToken(token);

          if (!tokenValidation.isValid && !tokenValidation.error?.includes('expired')) {
            // Token has invalid structure (malformed)
            logger.error('Invalid token structure, logging out:', tokenValidation.error);
            logout();
            setIsLoading(false);
            return;
          }

          const parsedUser = JSON.parse(userData);

          // Always load user from cache first (for instant UI update)
          logger.debug('Loading user from cache');
          setUser(parsedUser);

          // Check if token should be refreshed using stored expiry (Issue #37)
          if (shouldRefreshToken()) {
            const remaining = getTokenRemainingTime();
            logger.info(`Token refresh needed (remaining: ${remaining}s), refreshing...`);
            try {
              await refreshToken();
            } catch (error) {
              logger.error('Failed to refresh token:', error);
              // Only logout if refresh fails due to auth error (not network)
              if (!isNetworkError(error)) {
                logout();
              }
            }
          } else {
            const remaining = getTokenRemainingTime();
            logger.debug(`Token valid (remaining: ${remaining}s)`);
          }
        } catch (error) {
          logger.error('Error parsing stored user data:', error);
          logout();
        }
      } else {
        logger.debug('No stored credentials found');
      }

      setIsLoading(false);
    };

    initAuth();
  }, [logout, refreshToken]);

  // Set up periodic token refresh check with adaptive polling
  useEffect(() => {
    if (!isAuthenticated) return;

    logger.debug('Setting up token refresh interval');

    // Calculate optimal check interval based on token lifetime
    const calculateCheckInterval = () => {
      const remaining = getTokenRemainingTime();
      if (remaining < 300) return 30000;  // 30s if < 5min left
      if (remaining < 1800) return 60000; // 1min if < 30min left
      return 300000; // 5min otherwise
    };

    const checkAndRefresh = () => {
      if (shouldRefreshToken()) {
        const remaining = getTokenRemainingTime();
        logger.info(`Token expiring soon (remaining: ${remaining}s), refreshing...`);
        refreshToken().catch((error) => {
          logger.error('Auto-refresh failed:', error);
        });
      }
    };

    // Initial check
    checkAndRefresh();

    // Set up interval with adaptive timing
    const tokenRefreshInterval = setInterval(checkAndRefresh, calculateCheckInterval());

    return () => {
      logger.debug('Clearing token refresh interval');
      clearInterval(tokenRefreshInterval);
    };
  }, [isAuthenticated, refreshToken]);

  // Sync token refresh across tabs
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'access_token' && e.newValue) {
        logger.debug('Token updated in another tab');
        // Update local state without triggering another refresh
        setIsAuthenticated(true);
      }
    };

    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, []);

  const login = async (email: string, password: string, rememberMe = false): Promise<AuthResponse> => {
    logger.info('Logging in user:', email);

    try {
      const response = await fetch(API_ENDPOINTS.auth.login, {
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

      logger.info('Login successful');

      // Validate and store auth data using utility (Issues #35, #37)
      try {
        storeAuthTokens(data.access_token, data.refresh_token, data.expires_in);
        localStorage.setItem('user_data', JSON.stringify(data.user));
        setUser(data.user);
      } catch (validationError) {
        logger.error('Token validation failed after login:', validationError);
        throw new Error('Received invalid tokens from server');
      }

      return data;
    } catch (error) {
      logger.error('Login error:', error);

      if (isNetworkError(error)) {
        throw new Error('Unable to connect to server. Please check your internet connection and try again.');
      }

      throw error;
    }
  };

  const register = async (data: RegisterData): Promise<AuthResponse> => {
    logger.info('Registering user:', data.email);

    try {
      const response = await fetch(API_ENDPOINTS.auth.register, {
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
        logger.error('Failed to parse registration response:', text);
        throw new Error('Invalid server response');
      }

      if (!response.ok) {
        throw new Error(result.error || 'Registration failed');
      }

      logger.info('Registration successful');

      if (result.access_token) {
        // Validate and store auth data using utility (Issues #35, #37)
        try {
          storeAuthTokens(result.access_token, result.refresh_token, result.expires_in);
          localStorage.setItem('user_data', JSON.stringify(result.user));
          setUser(result.user);
        } catch (validationError) {
          logger.error('Token validation failed after registration:', validationError);
          throw new Error('Received invalid tokens from server');
        }
      }

      return result;
    } catch (error) {
      logger.error('Registration error:', error);

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
