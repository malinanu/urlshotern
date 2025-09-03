'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import {
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  PieChart,
  Pie,
  Cell,
  LineChart,
  Line,
  AreaChart,
  Area,
} from 'recharts';
import { 
  GlobeAltIcon, 
  DevicePhoneMobileIcon,
  ClockIcon,
  ArrowTrendingUpIcon,
  ChartBarIcon,
  MapIcon,
  CalendarIcon,
} from '@heroicons/react/24/outline';

// Color palette for charts
const COLORS = {
  primary: '#3b82f6',
  secondary: '#10b981',
  accent: '#f59e0b',
  danger: '#ef4444',
  purple: '#8b5cf6',
  indigo: '#6366f1',
  pink: '#ec4899',
  gray: '#6b7280',
};

const CHART_COLORS = [
  COLORS.primary,
  COLORS.secondary,
  COLORS.accent,
  COLORS.purple,
  COLORS.indigo,
  COLORS.pink,
  COLORS.danger,
  COLORS.gray,
];

// Interfaces for the advanced analytics data
interface AdvancedAnalyticsData {
  short_code: string;
  original_url: string;
  total_clicks: number;
  created_at: string;
  geographic: {
    short_code: string;
    total_clicks: number;
    countries: Array<{
      country_code: string;
      country_name: string;
      clicks: number;
      percentage: number;
      unique_ips: number;
      last_click?: string;
    }>;
    cities: Array<{
      country_code: string;
      region?: string;
      city: string;
      latitude?: number;
      longitude?: number;
      clicks: number;
      percentage: number;
    }>;
    map_data: Array<{
      latitude: number;
      longitude: number;
      clicks: number;
      location: string;
      country_code: string;
    }>;
  };
  time_analytics: {
    short_code: string;
    hourly_pattern: Array<{
      hour: number;
      clicks: number;
    }>;
    weekly_pattern: Array<{
      weekday: number;
      day: string;
      clicks: number;
    }>;
    heatmap_data: Array<{
      date: string;
      hour: number;
      clicks: number;
    }>;
    peak_times: {
      peak_hour: number;
      peak_weekday: number;
      peak_day: string;
      max_clicks: number;
    };
  };
  device_analytics: {
    short_code: string;
    device_types: Array<{
      device_type: string;
      clicks: number;
      percentage: number;
    }>;
    browsers: Array<{
      browser_name: string;
      browser_version: string;
      clicks: number;
      percentage: number;
    }>;
    operating_systems: Array<{
      os_name: string;
      os_version: string;
      clicks: number;
      percentage: number;
    }>;
  };
  referrers: Array<{
    referrer: string;
    clicks: number;
  }>;
  last_updated: string;
}

interface AdvancedAnalyticsProps {
  shortCode: string;
  days?: number;
  className?: string;
}

export default function AdvancedAnalytics({ 
  shortCode, 
  days = 30,
  className = '' 
}: AdvancedAnalyticsProps) {
  const [analytics, setAnalytics] = useState<AdvancedAnalyticsData | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'geographic' | 'time' | 'device'>('geographic');
  const { getAccessToken, isAuthenticated } = useAuth();

  useEffect(() => {
    fetchAdvancedAnalytics();
  }, [shortCode, days]);

  const fetchAdvancedAnalytics = async () => {
    if (!isAuthenticated || !shortCode) return;

    setIsLoading(true);
    try {
      const token = getAccessToken();
      const response = await fetch(`/api/v1/analytics/${shortCode}/advanced?days=${days}`, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setAnalytics(data);
        setError(null);
      } else {
        setError('Failed to fetch advanced analytics data');
      }
    } catch (err) {
      console.error('Error fetching advanced analytics:', err);
      setError('Error fetching analytics data');
    } finally {
      setIsLoading(false);
    }
  };

  const formatNumber = (num: number): string => {
    if (num >= 1000000) {
      return (num / 1000000).toFixed(1) + 'M';
    }
    if (num >= 1000) {
      return (num / 1000).toFixed(1) + 'K';
    }
    return num.toString();
  };

  const formatPercentage = (percentage: number): string => {
    return `${percentage.toFixed(1)}%`;
  };

  if (isLoading) {
    return (
      <div className={`bg-white rounded-lg shadow-sm border border-gray-200 p-6 ${className}`}>
        <div className="animate-pulse space-y-4">
          <div className="h-4 bg-gray-200 rounded w-1/4"></div>
          <div className="space-y-3">
            {[1, 2, 3, 4].map(i => (
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
            onClick={fetchAdvancedAnalytics}
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
          <p className="text-gray-500">No advanced analytics data available</p>
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
          <h2 className="text-xl font-semibold text-gray-900">Advanced Analytics</h2>
        </div>
        <div className="text-sm text-gray-500">
          Last updated: {new Date(analytics.last_updated).toLocaleString()}
        </div>
      </div>

      {/* Tab Navigation */}
      <div className="border-b border-gray-200 mb-6">
        <nav className="flex space-x-8">
          <button
            onClick={() => setActiveTab('geographic')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'geographic'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <div className="flex items-center gap-2">
              <GlobeAltIcon className="h-4 w-4" />
              Geographic
            </div>
          </button>
          <button
            onClick={() => setActiveTab('time')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'time'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <div className="flex items-center gap-2">
              <ClockIcon className="h-4 w-4" />
              Time Patterns
            </div>
          </button>
          <button
            onClick={() => setActiveTab('device')}
            className={`py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'device'
                ? 'border-primary-500 text-primary-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <div className="flex items-center gap-2">
              <DevicePhoneMobileIcon className="h-4 w-4" />
              Devices & Browsers
            </div>
          </button>
        </nav>
      </div>

      {/* Tab Content */}
      {activeTab === 'geographic' && (
        <GeographicAnalytics data={analytics.geographic} />
      )}
      {activeTab === 'time' && (
        <TimeAnalytics data={analytics.time_analytics} />
      )}
      {activeTab === 'device' && (
        <DeviceAnalytics data={analytics.device_analytics} />
      )}
    </div>
  );
}

// Geographic Analytics Component
function GeographicAnalytics({ data }: { data: AdvancedAnalyticsData['geographic'] }) {
  return (
    <div className="space-y-6">
      {/* Countries Chart */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
            <MapIcon className="h-5 w-5" />
            Top Countries
          </h3>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data.countries.slice(0, 10)}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="country_code" />
                <YAxis />
                <Tooltip 
                  formatter={(value: number, name: string) => [formatNumber(value), 'Clicks']}
                  labelFormatter={(label: string) => {
                    const country = data.countries.find(c => c.country_code === label);
                    return country?.country_name || label;
                  }}
                />
                <Bar dataKey="clicks" fill={COLORS.primary} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Country Distribution</h3>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={data.countries.slice(0, 8)}
                  cx="50%"
                  cy="50%"
                  innerRadius={60}
                  outerRadius={120}
                  paddingAngle={2}
                  dataKey="clicks"
                >
                  {data.countries.slice(0, 8).map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={CHART_COLORS[index % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip 
                  formatter={(value: number) => [formatNumber(value), 'Clicks']}
                  labelFormatter={(label: string, payload: any) => {
                    if (payload && payload.length > 0) {
                      const data = payload[0].payload;
                      return data.country_name;
                    }
                    return label;
                  }}
                />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>

      {/* Countries Table */}
      <div>
        <h3 className="text-lg font-medium text-gray-900 mb-4">Detailed Country Statistics</h3>
        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Country
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Clicks
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Percentage
                </th>
                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Unique IPs
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {data.countries.map((country, index) => (
                <tr key={country.country_code} className={index % 2 === 0 ? 'bg-white' : 'bg-gray-50'}>
                  <td className="px-6 py-4 whitespace-nowrap">
                    <div className="flex items-center">
                      <span className="text-sm font-medium text-gray-900">
                        {country.country_name}
                      </span>
                      <span className="ml-2 text-xs text-gray-500">
                        ({country.country_code})
                      </span>
                    </div>
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatNumber(country.clicks)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatPercentage(country.percentage)}
                  </td>
                  <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                    {formatNumber(country.unique_ips)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  );
}

// Time Analytics Component
function TimeAnalytics({ data }: { data: AdvancedAnalyticsData['time_analytics'] }) {
  // Convert hourly data for 24-hour display
  const hourlyData = Array.from({ length: 24 }, (_, hour) => {
    const found = data.hourly_pattern.find(h => h.hour === hour);
    return {
      hour: `${hour.toString().padStart(2, '0')}:00`,
      clicks: found?.clicks || 0,
    };
  });

  return (
    <div className="space-y-6">
      {/* Peak Times Summary */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <div className="bg-blue-50 p-4 rounded-lg">
          <div className="text-sm font-medium text-blue-800">Peak Hour</div>
          <div className="text-2xl font-bold text-blue-900">
            {data.peak_times.peak_hour.toString().padStart(2, '0')}:00
          </div>
        </div>
        <div className="bg-green-50 p-4 rounded-lg">
          <div className="text-sm font-medium text-green-800">Peak Day</div>
          <div className="text-2xl font-bold text-green-900">
            {data.peak_times.peak_day}
          </div>
        </div>
        <div className="bg-purple-50 p-4 rounded-lg">
          <div className="text-sm font-medium text-purple-800">Max Clicks</div>
          <div className="text-2xl font-bold text-purple-900">
            {formatNumber(data.peak_times.max_clicks)}
          </div>
        </div>
      </div>

      {/* Charts */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
            <ClockIcon className="h-5 w-5" />
            Hourly Pattern
          </h3>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <AreaChart data={hourlyData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis 
                  dataKey="hour" 
                  tick={{ fontSize: 12 }}
                  interval={3}
                />
                <YAxis />
                <Tooltip formatter={(value: number) => [formatNumber(value), 'Clicks']} />
                <Area 
                  type="monotone" 
                  dataKey="clicks" 
                  stroke={COLORS.primary} 
                  fill={COLORS.primary}
                  fillOpacity={0.3}
                />
              </AreaChart>
            </ResponsiveContainer>
          </div>
        </div>

        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4 flex items-center gap-2">
            <CalendarIcon className="h-5 w-5" />
            Weekly Pattern
          </h3>
          <div className="h-80">
            <ResponsiveContainer width="100%" height="100%">
              <BarChart data={data.weekly_pattern}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="day" />
                <YAxis />
                <Tooltip formatter={(value: number) => [formatNumber(value), 'Clicks']} />
                <Bar dataKey="clicks" fill={COLORS.secondary} />
              </BarChart>
            </ResponsiveContainer>
          </div>
        </div>
      </div>
    </div>
  );
}

// Device Analytics Component
function DeviceAnalytics({ data }: { data: AdvancedAnalyticsData['device_analytics'] }) {
  return (
    <div className="space-y-6">
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Device Types */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Device Types</h3>
          <div className="h-60">
            <ResponsiveContainer width="100%" height="100%">
              <PieChart>
                <Pie
                  data={data.device_types}
                  cx="50%"
                  cy="50%"
                  innerRadius={40}
                  outerRadius={80}
                  paddingAngle={2}
                  dataKey="clicks"
                >
                  {data.device_types.map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={CHART_COLORS[index % CHART_COLORS.length]} />
                  ))}
                </Pie>
                <Tooltip formatter={(value: number) => [formatNumber(value), 'Clicks']} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>

        {/* Top Browsers */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Top Browsers</h3>
          <div className="space-y-3">
            {data.browsers.slice(0, 5).map((browser, index) => (
              <div key={`${browser.browser_name}-${browser.browser_version}`} className="flex items-center justify-between">
                <span className="text-sm text-gray-700">
                  {browser.browser_name} {browser.browser_version}
                </span>
                <div className="flex items-center gap-2">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-500 h-2 rounded-full"
                      style={{ width: `${Math.max(10, browser.percentage)}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-8 text-right">
                    {formatNumber(browser.clicks)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Operating Systems */}
        <div>
          <h3 className="text-lg font-medium text-gray-900 mb-4">Operating Systems</h3>
          <div className="space-y-3">
            {data.operating_systems.slice(0, 5).map((os, index) => (
              <div key={`${os.os_name}-${os.os_version}`} className="flex items-center justify-between">
                <span className="text-sm text-gray-700">
                  {os.os_name} {os.os_version}
                </span>
                <div className="flex items-center gap-2">
                  <div className="w-20 bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-500 h-2 rounded-full"
                      style={{ width: `${Math.max(10, os.percentage)}%` }}
                    ></div>
                  </div>
                  <span className="text-sm font-medium text-gray-900 w-8 text-right">
                    {formatNumber(os.clicks)}
                  </span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}