'use client';

import React, { createContext, useContext, useState, useEffect } from 'react';

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

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);

  const isAuthenticated = !!user;

  // Initialize auth state from localStorage
  useEffect(() => {
    const initAuth = async () => {
      const token = localStorage.getItem('access_token');
      const userData = localStorage.getItem('user_data');
      
      if (token && userData) {
        try {
          const parsedUser = JSON.parse(userData);
          setUser(parsedUser);
          
          // If token is expired, try to refresh it
          if (isTokenExpired(token)) {
            try {
              await refreshToken();
            } catch (error) {
              console.error('Failed to refresh expired token on init:', error);
              logout();
            }
          } else {
            // Verify token is still valid by calling profile endpoint
            const response = await fetch(`${API_BASE_URL}/api/v1/auth/profile`, {
              headers: {
                'Authorization': `Bearer ${token}`,
                'Content-Type': 'application/json',
              },
            });
            
            if (!response.ok) {
              // Token is invalid, clear auth state
              logout();
            }
          }
        } catch (error) {
          console.error('Error verifying auth state:', error);
          logout();
        }
      }
      
      setIsLoading(false);
    };
    
    initAuth();
  }, []);

  // Set up periodic token refresh check
  useEffect(() => {
    if (!isAuthenticated) return;

    const tokenRefreshInterval = setInterval(async () => {
      const token = getAccessToken();
      if (token && isTokenExpired(token)) {
        try {
          console.log('Token expired, refreshing automatically...');
          await refreshToken();
        } catch (error) {
          console.error('Automatic token refresh failed:', error);
          logout();
        }
      }
    }, 60000); // Check every minute

    return () => clearInterval(tokenRefreshInterval);
  }, [isAuthenticated]);

  const login = async (email: string, password: string, rememberMe = false): Promise<AuthResponse> => {
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
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Login failed');
      }

      // Store auth data
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_data', JSON.stringify(data.user));
      
      setUser(data.user);
      
      return data;
    } catch (error) {
      console.error('Login error:', error);
      throw error;
    }
  };

  const register = async (data: RegisterData): Promise<AuthResponse> => {
    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/auth/register`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(data),
      });

      const result = await response.json();

      if (!response.ok) {
        throw new Error(result.error || 'Registration failed');
      }

      // For registration, we might not get tokens immediately if verification is required
      // Check if we got an auth response or just a registration confirmation
      if (result.access_token) {
        localStorage.setItem('access_token', result.access_token);
        localStorage.setItem('refresh_token', result.refresh_token);
        localStorage.setItem('user_data', JSON.stringify(result.user));
        setUser(result.user);
      }
      
      return result;
    } catch (error) {
      console.error('Registration error:', error);
      throw error;
    }
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user_data');
    setUser(null);
    
    // Call logout endpoint if we have a token
    const token = localStorage.getItem('access_token');
    if (token) {
      fetch(`${API_BASE_URL}/api/v1/auth/logout`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      }).catch(() => {
        // Ignore errors on logout
      });
    }
  };

  const refreshToken = async () => {
    // Prevent multiple simultaneous refresh attempts
    if (isRefreshing) {
      // Wait for ongoing refresh to complete
      let attempts = 0;
      while (isRefreshing && attempts < 50) { // Max 5 seconds wait
        await new Promise(resolve => setTimeout(resolve, 100));
        attempts++;
      }
      return;
    }

    const refreshTokenValue = localStorage.getItem('refresh_token');
    
    if (!refreshTokenValue) {
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
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Token refresh failed');
      }

      // Update stored tokens
      localStorage.setItem('access_token', data.access_token);
      localStorage.setItem('refresh_token', data.refresh_token);
      localStorage.setItem('user_data', JSON.stringify(data.user));
      
      setUser(data.user);
    } catch (error) {
      console.error('Token refresh error:', error);
      logout();
      throw error;
    } finally {
      setIsRefreshing(false);
    }
  };

  // Helper function to decode JWT and check expiration
  const isTokenExpired = (token?: string): boolean => {
    const tokenToCheck = token || localStorage.getItem('access_token');
    if (!tokenToCheck) return true;
    
    try {
      const payload = JSON.parse(atob(tokenToCheck.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      // Check if token expires within the next 5 minutes
      return payload.exp < (currentTime + 300);
    } catch (error) {
      return true; // Treat invalid tokens as expired
    }
  };

  const getAccessToken = (): string | null => {
    return localStorage.getItem('access_token');
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