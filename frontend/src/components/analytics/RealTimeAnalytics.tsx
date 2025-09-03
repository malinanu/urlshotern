'use client';

import { useState, useEffect, useCallback, useRef } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { 
  ChartBarIcon, 
  GlobeAltIcon, 
  DevicePhoneMobileIcon,
  ComputerDesktopIcon,
  EyeIcon,
  ClockIcon,
  CalendarDaysIcon,
  ArrowTrendingUpIcon
} from '@heroicons/react/24/outline';

interface AnalyticsData {
  total_clicks: number;
  today_clicks: number;
  month_clicks: number;
  daily_clicks: DailyClick[];
  country_stats: CountryStat[];
  referrer_stats: ReferrerStat[];
  device_stats: DeviceStat[];
  browser_stats: BrowserStat[];
}

interface DailyClick {
  date: string;
  clicks: number;
}

interface CountryStat {
  country_code: string;
  country_name?: string;
  clicks: number;
}

interface ReferrerStat {
  referrer: string;
  clicks: number;
}

interface DeviceStat {
  device: string;
  clicks: number;
}

interface BrowserStat {
  browser: string;
  clicks: number;
}

interface RealTimeAnalyticsProps {
  shortCode: string;
  refreshInterval?: number; // in milliseconds
  className?: string;
}

export default function RealTimeAnalytics({ 
  shortCode, 
  refreshInterval = 30000, // 30 seconds default
  className = '' 
}: RealTimeAnalyticsProps) {
  const [analytics, setAnalytics] = useState<AnalyticsData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [lastUpdated, setLastUpdated] = useState<Date | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { getAccessToken, isAuthenticated } = useAuth();

  const fetchAnalytics = useCallback(async () => {
    if (!isAuthenticated || !shortCode) return;

    try {
      const token = getAccessToken();
      const response = await fetch(`http://localhost:8080/api/v1/analytics/${shortCode}?days=30`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setAnalytics(data);
        setLastUpdated(new Date());
        setError(null);
      } else if (response.status === 404) {
        setError('Analytics not found for this URL');
      } else {
        setError('Failed to fetch analytics data');
      }
    } catch (err) {
      console.error('Error fetching analytics:', err);
      setError('Error fetching analytics data');
    } finally {
      setIsLoading(false);
    }
  }, [shortCode, isAuthenticated, getAccessToken]);

  // Initial load
  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  // Set up real-time updates
  useEffect(() => {
    if (!refreshInterval || refreshInterval < 1000) return;

    const interval = setInterval(fetchAnalytics, refreshInterval);
    return () => clearInterval(interval);
  }, [fetchAnalytics, refreshInterval]);

  const formatNumber = (num: number | undefined | null): string => {
    // Handle null, undefined, or NaN values
    if (num == null || isNaN(num)) {
      return '0';
    }
    
    // Ensure we have a valid number
    const validNum = Number(num);
    if (isNaN(validNum)) {
      return '0';
    }
    
    if (validNum >= 1000000) {
      return (validNum / 1000000).toFixed(1) + 'M';
    }
    if (validNum >= 1000) {
      return (validNum / 1000).toFixed(1) + 'K';
    }
    return validNum.toString();
  };

  const getCountryName = (countryCode: string): string => {
    // Simple country code to name mapping (you could use a full lookup library)
    const countries: Record<string, string> = {
      'US': 'United States',
      'GB': 'United Kingdom',
      'DE': 'Germany',
      'FR': 'France',
      'CA': 'Canada',
      'AU': 'Australia',
      'JP': 'Japan',
      'IN': 'India',
      'BR': 'Brazil',
      'ES': 'Spain',
      'IT': 'Italy',
      'NL': 'Netherlands',
      'SE': 'Sweden',
      'NO': 'Norway',
      'DK': 'Denmark',
      'FI': 'Finland',
      'PL': 'Poland',
      'RU': 'Russia',
      'CN': 'China',
      'KR': 'South Korea',
    };
    return countries[countryCode] || countryCode;
  };

  if (isLoading) {
    return (
      <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 ${className}`}>
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-gray-200 rounded w-1/4"></div>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {[1, 2, 3].map(i => (
              <div key={i} className="h-20 bg-gray-200 rounded"></div>
            ))}
          </div>
          <div className="h-64 bg-gray-200 rounded"></div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 ${className}`}>
        <div className="text-center py-8">
          <ChartBarIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500">{error}</p>
          <button
            onClick={() => {
              setError(null);
              setIsLoading(true);
              fetchAnalytics();
            }}
            className="mt-4 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            Retry
          </button>
        </div>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 ${className}`}>
        <div className="text-center py-8">
          <ChartBarIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
          <p className="text-gray-500">No analytics data available</p>
        </div>
      </div>
    );
  }

  return (
    <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 ${className}`}>
      {/* Header */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <ArrowTrendingUpIcon className="h-6 w-6 text-primary-600" />
          <h2 className="text-lg font-semibold text-gray-900">Real-time Analytics</h2>
          <span className="inline-flex items-center px-2.5 py-0.5 rounded-full text-xs font-medium bg-green-100 text-green-800">
            Live
          </span>
        </div>
        {lastUpdated && (
          <div className="flex items-center gap-1 text-sm text-gray-500">
            <ClockIcon className="h-4 w-4" />
            Updated {lastUpdated.toLocaleTimeString()}
          </div>
        )}
      </div>

      {/* Key Metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-8">
        <div className="bg-gradient-to-r from-blue-50 to-blue-100 p-4 rounded-lg">
          <div className="flex items-center gap-3">
            <EyeIcon className="h-8 w-8 text-blue-600" />
            <div>
              <p className="text-sm font-medium text-blue-800">Total Clicks</p>
              <p className="text-2xl font-bold text-blue-900">{formatNumber(analytics?.total_clicks)}</p>
            </div>
          </div>
        </div>

        <div className="bg-gradient-to-r from-green-50 to-green-100 p-4 rounded-lg">
          <div className="flex items-center gap-3">
            <ClockIcon className="h-8 w-8 text-green-600" />
            <div>
              <p className="text-sm font-medium text-green-800">Today</p>
              <p className="text-2xl font-bold text-green-900">{formatNumber(analytics?.today_clicks)}</p>
            </div>
          </div>
        </div>

        <div className="bg-gradient-to-r from-purple-50 to-purple-100 p-4 rounded-lg">
          <div className="flex items-center gap-3">
            <CalendarDaysIcon className="h-8 w-8 text-purple-600" />
            <div>
              <p className="text-sm font-medium text-purple-800">This Month</p>
              <p className="text-2xl font-bold text-purple-900">{formatNumber(analytics?.month_clicks)}</p>
            </div>
          </div>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Daily Clicks Chart */}
        {analytics?.daily_clicks && analytics.daily_clicks.length > 0 && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Daily Clicks (Last 30 Days)</h3>
            <div className="space-y-2">
              {analytics.daily_clicks.slice(-7).map((day) => {
                const maxClicks = Math.max(...analytics.daily_clicks.map(d => d?.clicks || 0), 1);
                const dayClicks = day?.clicks || 0;
                return (
                  <div key={day?.date || Math.random()} className="flex items-center justify-between">
                    <span className="text-sm text-gray-600">
                      {day?.date ? new Date(day.date).toLocaleDateString('en-US', { 
                        month: 'short', 
                        day: 'numeric' 
                      }) : 'N/A'}
                    </span>
                    <div className="flex items-center gap-2">
                      <div className="w-24 bg-gray-200 rounded-full h-2">
                        <div 
                          className="bg-blue-500 h-2 rounded-full transition-all duration-300"
                          style={{ 
                            width: `${Math.max(5, (dayClicks / maxClicks) * 100)}%` 
                          }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium text-gray-900 w-8 text-right">
                        {dayClicks}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Geographic Stats */}
        {analytics?.country_stats && analytics.country_stats.length > 0 && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
              <GlobeAltIcon className="h-5 w-5" />
              Top Countries
            </h3>
            <div className="space-y-3">
              {analytics.country_stats.slice(0, 5).map((country) => {
                const topClicks = analytics.country_stats[0]?.clicks || 1;
                const countryClicks = country?.clicks || 0;
                return (
                  <div key={country?.country_code || Math.random()} className="flex items-center justify-between">
                    <span className="text-sm text-gray-700">
                      {getCountryName(country?.country_code || 'Unknown')}
                    </span>
                    <div className="flex items-center gap-2">
                      <div className="w-20 bg-gray-200 rounded-full h-1.5">
                        <div 
                          className="bg-green-500 h-1.5 rounded-full"
                          style={{ 
                            width: `${Math.max(10, (countryClicks / topClicks) * 100)}%` 
                          }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium text-gray-900 w-6 text-right">
                        {countryClicks}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Device Stats */}
        {analytics?.device_stats && analytics.device_stats.length > 0 && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
              <DevicePhoneMobileIcon className="h-5 w-5" />
              Device Types
            </h3>
            <div className="space-y-3">
              {analytics.device_stats.slice(0, 4).map((device) => {
                const topDeviceClicks = analytics.device_stats[0]?.clicks || 1;
                const deviceClicks = device?.clicks || 0;
                const deviceName = device?.device || 'Unknown';
                return (
                  <div key={deviceName} className="flex items-center justify-between">
                    <div className="flex items-center gap-2">
                      {deviceName.toLowerCase().includes('mobile') ? (
                        <DevicePhoneMobileIcon className="h-4 w-4 text-gray-500" />
                      ) : (
                        <ComputerDesktopIcon className="h-4 w-4 text-gray-500" />
                      )}
                      <span className="text-sm text-gray-700 capitalize">
                        {deviceName}
                      </span>
                    </div>
                    <div className="flex items-center gap-2">
                      <div className="w-16 bg-gray-200 rounded-full h-1.5">
                        <div 
                          className="bg-purple-500 h-1.5 rounded-full"
                          style={{ 
                            width: `${Math.max(15, (deviceClicks / topDeviceClicks) * 100)}%` 
                          }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium text-gray-900 w-6 text-right">
                        {deviceClicks}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {/* Browser Stats */}
        {analytics?.browser_stats && analytics.browser_stats.length > 0 && (
          <div className="bg-gray-50 p-4 rounded-lg">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Top Browsers</h3>
            <div className="space-y-3">
              {analytics.browser_stats.slice(0, 4).map((browser) => {
                const topBrowserClicks = analytics.browser_stats[0]?.clicks || 1;
                const browserClicks = browser?.clicks || 0;
                const browserName = browser?.browser || 'Unknown';
                return (
                  <div key={browserName} className="flex items-center justify-between">
                    <span className="text-sm text-gray-700">
                      {browserName}
                    </span>
                    <div className="flex items-center gap-2">
                      <div className="w-16 bg-gray-200 rounded-full h-1.5">
                        <div 
                          className="bg-orange-500 h-1.5 rounded-full"
                          style={{ 
                            width: `${Math.max(15, (browserClicks / topBrowserClicks) * 100)}%` 
                          }}
                        ></div>
                      </div>
                      <span className="text-sm font-medium text-gray-900 w-6 text-right">
                        {browserClicks}
                      </span>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>

      {/* Refresh Indicator */}
      <div className="mt-6 text-center">
        <p className="text-xs text-gray-500">
          Auto-refreshes every {refreshInterval / 1000} seconds
        </p>
      </div>
    </div>
  );
}