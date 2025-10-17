'use client';

import React, { createContext, useContext, useState, useEffect, useCallback } from 'react';

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

// Helper function to decode JWT and check expiration (moved outside component)
const checkTokenExpiry = (token: string): boolean => {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Math.floor(Date.now() / 1000);
    // Check if token is actually expired (no buffer)
    return payload.exp < currentTime;
  } catch (error) {
    console.error('Error decoding token:', error);
    return true; // Treat invalid tokens as expired
  }
};

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const isAuthenticated = !!user;

  // Helper function to check expiration with optional buffer
  const isTokenExpired = useCallback((token?: string, bufferSeconds: number = 0): boolean => {
    const tokenToCheck = token || localStorage.getItem('access_token');
    if (!tokenToCheck) return true;

    try {
      const payload = JSON.parse(atob(tokenToCheck.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      // Check if token expires within the buffer time
      return payload.exp < (currentTime + bufferSeconds);
    } catch (error) {
      console.error('Error checking token expiry:', error);
      return true;
    }
  }, []);

  const getAccessToken = useCallback((): string | null => {
    return localStorage.getItem('access_token');
  }, []);

  const logout = useCallback(() => {
    console.log('[Auth] Logging out user');

    // Save token before removing for logout endpoint call
    const token = localStorage.getItem('access_token');

    // Remove auth data
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_data');
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

  const refreshToken = useCallback(async (retryCount = 0) => {
    console.log('[Auth] Refreshing token, attempt:', retryCount + 1);

    // Prevent multiple simultaneous refresh attempts
    if (isRefreshing) {
      console.log('[Auth] Refresh already in progress, waiting...');
      let attempts = 0;
      while (isRefreshing && attempts < 50) {
        await new Promise(resolve => setTimeout(resolve, 100));
        attempts++;
      }
      return;
    }

    const refreshTokenValue = localStorage.getItem('refresh_token');

    if (!refreshTokenValue) {
      console.error('[Auth] No refresh token found');
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

      // Update stored tokens
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_data', JSON.stringify(data.user));

      setUser(data.user);
    } catch (error) {
      console.error('[Auth] Token refresh error:', error);

      // Handle network errors with retry logic
      if (isNetworkError(error) && retryCount < 3) {
        console.log(`[Auth] Network error, retrying in ${Math.pow(2, retryCount)} seconds...`);

        await new Promise(resolve => setTimeout(resolve, Math.pow(2, retryCount) * 1000));

        setIsRefreshing(false);
        return refreshToken(retryCount + 1);
      }

      // For network errors that exceed retry limit, don't logout
      if (isNetworkError(error)) {
        console.warn('[Auth] Backend unavailable, continuing without token refresh');
        return;
      }

      // For auth errors, logout
      console.error('[Auth] Authentication error, logging out');
      logout();
      throw error;
    } finally {
      setIsRefreshing(false);
    }
  }, [isRefreshing, logout]);

  // Initialize auth state from localStorage
  useEffect(() => {
    const initAuth = async () => {
      console.log('[Auth] Initializing auth state');

      const token = localStorage.getItem('access_token');
      const userData = localStorage.getItem('user_data');

      if (token && userData && userData !== 'undefined') {
        try {
          const parsedUser = JSON.parse(userData);

          // Always load user from cache first (for instant UI update)
          console.log('[Auth] Loading user from cache');
          setUser(parsedUser);

          // Then check if token is expired and refresh if needed
          if (checkTokenExpiry(token)) {
            console.log('[Auth] Token expired, refreshing in background...');
            try {
              await refreshToken();
            } catch (error) {
              console.error('[Auth] Failed to refresh expired token:', error);
              // Only logout if refresh fails due to auth error (not network)
              if (!isNetworkError(error)) {
                logout();
              }
            }
          } else {
            console.log('[Auth] Token valid');
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
      const token = getAccessToken();

      if (token) {
        // Check if token will expire in the next 5 minutes
        if (isTokenExpired(token, 300)) {
          console.log('[Auth] Token expiring soon, refreshing...');
          refreshToken().catch((error) => {
            console.error('[Auth] Auto-refresh failed:', error);
          });
        }
      }
    }, 60000); // Check every minute

    return () => {
      console.log('[Auth] Clearing token refresh interval');
      clearInterval(tokenRefreshInterval);
    };
  }, [isAuthenticated, isTokenExpired, getAccessToken, refreshToken]);

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

      // Store auth data
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_data', JSON.stringify(data.user));

      setUser(data.user);

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
        localStorage.setItem('access_token', result.access_token);
        localStorage.setItem('refresh_token', result.refresh_token);
        localStorage.setItem('user_data', JSON.stringify(result.user));
        setUser(result.user);
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
