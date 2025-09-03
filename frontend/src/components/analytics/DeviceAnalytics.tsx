'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { PieChart, Pie, Cell, BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { Smartphone, Monitor, Tablet, Globe, Cpu, Palette } from 'lucide-react';

interface DeviceTypeStat {
  device_type: string;
  clicks: number;
  percentage: number;
}

interface BrowserDetailStat {
  browser_name: string;
  browser_version: string;
  clicks: number;
  percentage: number;
}

interface OSDetailStat {
  os_name: string;
  os_version: string;
  clicks: number;
  percentage: number;
}

interface DeviceAnalytics {
  short_code: string;
  device_types: DeviceTypeStat[];
  browser_stats: BrowserDetailStat[];
  operating_systems: OSDetailStat[];
}

interface DeviceAnalyticsProps {
  shortCode: string;
  days?: number;
  height?: string;
}

// Color palette for charts
const DEVICE_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];
const BROWSER_COLORS = ['#dc2626', '#ea580c', '#ca8a04', '#16a34a', '#2563eb', '#7c3aed'];
const OS_COLORS = ['#0891b2', '#059669', '#db2777', '#7c2d12', '#4338ca', '#9333ea'];

export default function DeviceAnalytics({ 
  shortCode, 
  days = 30, 
  height = '500px' 
}: DeviceAnalyticsProps) {
  const [deviceData, setDeviceData] = useState<DeviceAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'devices' | 'browsers' | 'os'>('devices');

  // Fetch device analytics data
  const fetchDeviceData = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/analytics/device/${shortCode}?days=${days}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch device analytics data');
      }
      
      const data = await response.json();
      setDeviceData(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load device data');
      console.error('Error fetching device data:', err);
    } finally {
      setLoading(false);
    }
  }, [shortCode, days]);

  useEffect(() => {
    fetchDeviceData();
  }, [fetchDeviceData]);

  // Process data for charts
  const deviceChartData = useMemo(() => {
    if (!deviceData?.device_types) return [];
    return deviceData.device_types.map((device, index) => ({
      ...device,
      color: DEVICE_COLORS[index % DEVICE_COLORS.length],
      icon: device.device_type === 'mobile' ? 'smartphone' : 
            device.device_type === 'tablet' ? 'tablet' : 'monitor'
    }));
  }, [deviceData]);

  const browserChartData = useMemo(() => {
    if (!deviceData?.browser_stats) return [];
    
    // Group browsers by name and sum clicks
    const grouped = deviceData.browser_stats.reduce((acc, browser) => {
      const key = browser.browser_name;
      if (!acc[key]) {
        acc[key] = {
          browser_name: key,
          clicks: 0,
          percentage: 0,
          versions: []
        };
      }
      acc[key].clicks += browser.clicks;
      acc[key].percentage += browser.percentage;
      acc[key].versions.push({
        version: browser.browser_version,
        clicks: browser.clicks
      });
      return acc;
    }, {} as any);

    return Object.values(grouped)
      .sort((a: any, b: any) => b.clicks - a.clicks)
      .slice(0, 8)
      .map((browser: any, index) => ({
        ...browser,
        color: BROWSER_COLORS[index % BROWSER_COLORS.length]
      }));
  }, [deviceData]);

  const osChartData = useMemo(() => {
    if (!deviceData?.operating_systems) return [];
    
    // Group OS by name and sum clicks
    const grouped = deviceData.operating_systems.reduce((acc, os) => {
      const key = os.os_name;
      if (!acc[key]) {
        acc[key] = {
          os_name: key,
          clicks: 0,
          percentage: 0,
          versions: []
        };
      }
      acc[key].clicks += os.clicks;
      acc[key].percentage += os.percentage;
      acc[key].versions.push({
        version: os.os_version,
        clicks: os.clicks
      });
      return acc;
    }, {} as any);

    return Object.values(grouped)
      .sort((a: any, b: any) => b.clicks - a.clicks)
      .slice(0, 8)
      .map((os: any, index) => ({
        ...os,
        color: OS_COLORS[index % OS_COLORS.length]
      }));
  }, [deviceData]);

  const getDeviceIcon = (deviceType: string) => {
    switch (deviceType.toLowerCase()) {
      case 'mobile': return <Smartphone className="w-4 h-4" />;
      case 'tablet': return <Tablet className="w-4 h-4" />;
      case 'desktop': return <Monitor className="w-4 h-4" />;
      default: return <Globe className="w-4 h-4" />;
    }
  };

  // Custom tooltip for pie charts
  const CustomTooltip = ({ active, payload }: any) => {
    if (active && payload && payload.length) {
      const data = payload[0];
      return (
        <div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">
          <p className="font-medium">{data.name}</p>
          <p className="text-sm text-gray-600">
            {data.value?.toLocaleString()} clicks ({data.payload?.percentage?.toFixed(1)}%)
          </p>
        </div>
      );
    }
    return null;
  };

  if (loading) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading device analytics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-red-800 font-medium">Error loading device analytics</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
            <button 
              onClick={fetchDeviceData}
              className="mt-3 px-4 py-2 bg-red-100 hover:bg-red-200 text-red-800 rounded text-sm"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!deviceData) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center text-gray-600">
          <p className="font-medium">No device data available</p>
          <p className="text-sm mt-1">No device analytics found for the selected period</p>
        </div>
      </div>
    );
  }

  const totalClicks = deviceData.device_types.reduce((sum, device) => sum + device.clicks, 0);

  return (
    <div className="bg-white rounded-lg shadow-sm border" style={{ height }}>
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex justify-between items-center">
          <div>
            <h3 className="font-medium text-gray-900">Device & Browser Analytics</h3>
            <p className="text-sm text-gray-600">
              {totalClicks.toLocaleString()} total clicks analyzed
            </p>
          </div>
          
          {/* Tab selector */}
          <div className="flex bg-gray-100 rounded-lg p-1">
            {[
              { key: 'devices', label: 'Devices', icon: <Smartphone className="w-4 h-4" /> },
              { key: 'browsers', label: 'Browsers', icon: <Globe className="w-4 h-4" /> },
              { key: 'os', label: 'OS', icon: <Cpu className="w-4 h-4" /> }
            ].map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key as any)}
                className={`flex items-center space-x-2 px-3 py-1 text-xs rounded transition-colors ${
                  activeTab === tab.key
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                {tab.icon}
                <span>{tab.label}</span>
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4" style={{ height: `calc(${height} - 80px)` }}>
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 h-full">
          {/* Pie Chart */}
          <div className="bg-gray-50 rounded-lg p-4">
            <h4 className="font-medium text-gray-900 mb-4">
              {activeTab === 'devices' && 'Device Types'}
              {activeTab === 'browsers' && 'Browsers'}
              {activeTab === 'os' && 'Operating Systems'}
            </h4>
            
            <ResponsiveContainer width="100%" height={280}>
              <PieChart>
                <Pie
                  data={activeTab === 'devices' ? deviceChartData :
                        activeTab === 'browsers' ? browserChartData : osChartData}
                  cx="50%"
                  cy="50%"
                  labelLine={false}
                  label={({ percentage }) => `${percentage.toFixed(1)}%`}
                  outerRadius={80}
                  fill="#8884d8"
                  dataKey="clicks"
                  nameKey={activeTab === 'devices' ? 'device_type' :
                          activeTab === 'browsers' ? 'browser_name' : 'os_name'}
                >
                  {(activeTab === 'devices' ? deviceChartData :
                    activeTab === 'browsers' ? browserChartData : osChartData).map((entry, index) => (
                    <Cell key={`cell-${index}`} fill={entry.color} />
                  ))}
                </Pie>
                <Tooltip content={<CustomTooltip />} />
              </PieChart>
            </ResponsiveContainer>
          </div>

          {/* Bar Chart and Details */}
          <div className="space-y-4">
            {/* Bar Chart */}
            <div className="bg-gray-50 rounded-lg p-4">
              <ResponsiveContainer width="100%" height={200}>
                <BarChart
                  data={activeTab === 'devices' ? deviceChartData :
                        activeTab === 'browsers' ? browserChartData : osChartData}
                  margin={{ top: 5, right: 30, left: 20, bottom: 5 }}
                >
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis 
                    dataKey={activeTab === 'devices' ? 'device_type' :
                            activeTab === 'browsers' ? 'browser_name' : 'os_name'}
                    angle={-45}
                    textAnchor="end"
                    height={80}
                    fontSize={12}
                  />
                  <YAxis fontSize={12} />
                  <Tooltip content={<CustomTooltip />} />
                  <Bar 
                    dataKey="clicks" 
                    fill="#3b82f6"
                    radius={[4, 4, 0, 0]}
                  />
                </BarChart>
              </ResponsiveContainer>
            </div>

            {/* Details List */}
            <div className="bg-gray-50 rounded-lg p-4 max-h-48 overflow-y-auto">
              <h5 className="font-medium text-gray-900 mb-3">Detailed Breakdown</h5>
              <div className="space-y-2">
                {(activeTab === 'devices' ? deviceChartData :
                  activeTab === 'browsers' ? browserChartData : osChartData).map((item, index) => (
                  <div key={index} className="flex items-center justify-between text-sm">
                    <div className="flex items-center space-x-2">
                      <div 
                        className="w-3 h-3 rounded-full"
                        style={{ backgroundColor: item.color }}
                      />
                      <span className="font-medium">
                        {activeTab === 'devices' && getDeviceIcon(item.device_type)}
                        <span className="ml-2">
                          {item[activeTab === 'devices' ? 'device_type' :
                               activeTab === 'browsers' ? 'browser_name' : 'os_name']}
                        </span>
                      </span>
                    </div>
                    <div className="text-right">
                      <div className="font-medium">{item.clicks.toLocaleString()}</div>
                      <div className="text-xs text-gray-500">{item.percentage.toFixed(1)}%</div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}

// Device summary stats component
export function DeviceSummary({ deviceData }: { deviceData: DeviceAnalytics | null }) {
  if (!deviceData) return null;

  const totalClicks = deviceData.device_types.reduce((sum, device) => sum + device.clicks, 0);
  const mobilePercentage = deviceData.device_types.find(d => d.device_type === 'mobile')?.percentage || 0;
  const topBrowser = deviceData.browser_stats.reduce((top, browser) => 
    browser.clicks > (top?.clicks || 0) ? browser : top, deviceData.browser_stats[0]);
  const topOS = deviceData.operating_systems.reduce((top, os) => 
    os.clicks > (top?.clicks || 0) ? os : top, deviceData.operating_systems[0]);

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <h4 className="font-medium text-gray-900 mb-4">Device Insights</h4>
      
      <div className="grid grid-cols-2 gap-4">
        <div className="text-center">
          <div className="text-2xl font-bold text-blue-600">{mobilePercentage.toFixed(1)}%</div>
          <div className="text-sm text-gray-600">Mobile Traffic</div>
        </div>
        
        <div className="text-center">
          <div className="text-2xl font-bold text-green-600">{deviceData.device_types.length}</div>
          <div className="text-sm text-gray-600">Device Types</div>
        </div>
      </div>

      <div className="mt-4 space-y-3">
        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-600">Top Browser:</span>
          <span className="text-sm font-medium">{topBrowser?.browser_name || 'N/A'}</span>
        </div>
        
        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-600">Top OS:</span>
          <span className="text-sm font-medium">{topOS?.os_name || 'N/A'}</span>
        </div>

        <div className="flex justify-between items-center">
          <span className="text-sm text-gray-600">Total Devices:</span>
          <span className="text-sm font-medium">{totalClicks.toLocaleString()}</span>
        </div>
      </div>
    </div>
  );
}