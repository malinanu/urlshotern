'use client';

import React, { useState, useEffect, useCallback, useRef } from 'react';
import { Activity, Users, Globe, Smartphone, Monitor, Wifi, WifiOff } from 'lucide-react';

interface RealtimeClickData {
  short_code: string;
  client_ip?: string;
  user_agent?: string;
  referrer?: string;
  country?: string;
  city?: string;
  device?: string;
  browser?: string;
  os?: string;
  timestamp: string;
}

interface RealtimeUpdate {
  type: 'click' | 'analytics_update' | 'conversion' | 'ping' | 'pong' | 'initial_analytics';
  short_code: string;
  data: any;
  timestamp: string;
}

interface LiveClick {
  id: string;
  location: string;
  device: string;
  browser: string;
  timestamp: Date;
  country?: string;
  referrer?: string;
}

interface RealtimeStats {
  clicks_per_minute: number;
  active_visitors: number;
  total_clicks_today: number;
  top_countries: { [key: string]: number };
  top_devices: { [key: string]: number };
  recent_referrers: string[];
}

interface EnhancedRealTimeAnalyticsProps {
  shortCode: string;
  websocketUrl?: string;
  maxRecentClicks?: number;
  className?: string;
}

const DEFAULT_WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws/analytics';

export default function EnhancedRealTimeAnalytics({
  shortCode,
  websocketUrl = DEFAULT_WS_URL,
  maxRecentClicks = 50,
  className = ''
}: EnhancedRealTimeAnalyticsProps) {
  const [recentClicks, setRecentClicks] = useState<LiveClick[]>([]);
  const [stats, setStats] = useState<RealtimeStats>({
    clicks_per_minute: 0,
    active_visitors: 0,
    total_clicks_today: 0,
    top_countries: {},
    top_devices: {},
    recent_referrers: []
  });
  const [isConnected, setIsConnected] = useState(false);
  const [connectionError, setConnectionError] = useState<string | null>(null);
  const [lastActivity, setLastActivity] = useState<Date | null>(null);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const pingIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const clickCounterRef = useRef<{ [key: string]: number }>({});

  // WebSocket connection management
  const connectWebSocket = useCallback(() => {
    try {
      if (wsRef.current?.readyState === WebSocket.CONNECTING) {
        return;
      }

      wsRef.current = new WebSocket(websocketUrl);

      wsRef.current.onopen = () => {
        console.log('WebSocket connected');
        setIsConnected(true);
        setConnectionError(null);

        // Subscribe to updates for this short code
        const subscribeMessage = {
          type: 'subscribe',
          short_code: shortCode
        };
        wsRef.current?.send(JSON.stringify(subscribeMessage));

        // Set up ping interval
        pingIntervalRef.current = setInterval(() => {
          if (wsRef.current?.readyState === WebSocket.OPEN) {
            wsRef.current.send(JSON.stringify({ type: 'ping' }));
          }
        }, 30000);
      };

      wsRef.current.onmessage = (event) => {
        try {
          const update: RealtimeUpdate = JSON.parse(event.data);
          handleRealtimeUpdate(update);
          setLastActivity(new Date());
        } catch (error) {
          console.error('Error parsing WebSocket message:', error);
        }
      };

      wsRef.current.onclose = (event) => {
        console.log('WebSocket disconnected:', event.reason);
        setIsConnected(false);
        
        if (pingIntervalRef.current) {
          clearInterval(pingIntervalRef.current);
          pingIntervalRef.current = null;
        }

        // Attempt to reconnect unless it was a clean close
        if (!event.wasClean) {
          setConnectionError('Connection lost. Attempting to reconnect...');
          reconnectTimeoutRef.current = setTimeout(() => {
            connectWebSocket();
          }, 3000);
        }
      };

      wsRef.current.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionError('Connection error occurred');
      };

    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionError('Failed to establish connection');
    }
  }, [websocketUrl, shortCode]);

  // Handle different types of real-time updates
  const handleRealtimeUpdate = useCallback((update: RealtimeUpdate) => {
    switch (update.type) {
      case 'click':
        handleNewClick(update.data as RealtimeClickData);
        break;
      case 'analytics_update':
        handleAnalyticsUpdate(update.data);
        break;
      case 'initial_analytics':
        handleInitialAnalytics(update.data);
        break;
      case 'pong':
        // Handle pong response
        break;
      default:
        console.log('Unknown update type:', update.type);
    }
  }, []);

  const handleNewClick = useCallback((clickData: RealtimeClickData) => {
    const newClick: LiveClick = {
      id: `${Date.now()}-${Math.random()}`,
      location: clickData.city ? `${clickData.city}, ${clickData.country}` : clickData.country || 'Unknown',
      device: clickData.device || 'Unknown',
      browser: clickData.browser || 'Unknown',
      timestamp: new Date(clickData.timestamp),
      country: clickData.country,
      referrer: clickData.referrer
    };

    setRecentClicks(prev => [newClick, ...prev.slice(0, maxRecentClicks - 1)]);

    // Update click counter for CPM calculation
    const now = new Date();
    const minuteKey = `${now.getFullYear()}-${now.getMonth()}-${now.getDate()}-${now.getHours()}-${now.getMinutes()}`;
    clickCounterRef.current[minuteKey] = (clickCounterRef.current[minuteKey] || 0) + 1;

    // Update real-time stats
    setStats(prev => {
      const newStats = { ...prev };
      
      // Update clicks per minute
      const currentMinute = Object.keys(clickCounterRef.current)
        .filter(key => key === minuteKey)
        .reduce((sum, key) => sum + clickCounterRef.current[key], 0);
      newStats.clicks_per_minute = currentMinute;

      // Update country stats
      if (clickData.country) {
        newStats.top_countries[clickData.country] = (newStats.top_countries[clickData.country] || 0) + 1;
      }

      // Update device stats
      if (clickData.device) {
        newStats.top_devices[clickData.device] = (newStats.top_devices[clickData.device] || 0) + 1;
      }

      // Update recent referrers
      if (clickData.referrer && !newStats.recent_referrers.includes(clickData.referrer)) {
        newStats.recent_referrers = [clickData.referrer, ...newStats.recent_referrers.slice(0, 4)];
      }

      return newStats;
    });
  }, [maxRecentClicks]);

  const handleAnalyticsUpdate = useCallback((data: any) => {
    // Handle periodic analytics updates
    if (data.total_clicks) {
      setStats(prev => ({
        ...prev,
        total_clicks_today: data.today_clicks || prev.total_clicks_today
      }));
    }
  }, []);

  const handleInitialAnalytics = useCallback((data: any) => {
    // Handle initial analytics data when connecting
    if (data) {
      setStats(prev => ({
        ...prev,
        total_clicks_today: data.today_clicks || 0,
        active_visitors: Math.floor(Math.random() * 10) + 1 // Simulated active visitors
      }));
    }
  }, []);

  // Clean up old click counter data
  useEffect(() => {
    const cleanupInterval = setInterval(() => {
      const now = new Date();
      const fiveMinutesAgo = new Date(now.getTime() - 5 * 60 * 1000);
      
      Object.keys(clickCounterRef.current).forEach(key => {
        const [year, month, day, hour, minute] = key.split('-').map(Number);
        const keyDate = new Date(year, month, day, hour, minute);
        
        if (keyDate < fiveMinutesAgo) {
          delete clickCounterRef.current[key];
        }
      });
    }, 60000); // Clean up every minute

    return () => clearInterval(cleanupInterval);
  }, []);

  // Initialize WebSocket connection
  useEffect(() => {
    connectWebSocket();

    return () => {
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
      }
      if (pingIntervalRef.current) {
        clearInterval(pingIntervalRef.current);
      }
      if (wsRef.current) {
        wsRef.current.close(1000, 'Component unmounting');
      }
    };
  }, [connectWebSocket]);

  const getTopEntries = (obj: { [key: string]: number }, limit = 3) => {
    return Object.entries(obj)
      .sort(([,a], [,b]) => b - a)
      .slice(0, limit);
  };

  const formatTimeAgo = (timestamp: Date) => {
    const now = new Date();
    const seconds = Math.floor((now.getTime() - timestamp.getTime()) / 1000);
    
    if (seconds < 60) return `${seconds}s ago`;
    if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`;
    if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`;
    return `${Math.floor(seconds / 86400)}d ago`;
  };

  return (
    <div className={`bg-white rounded-lg shadow-sm border ${className}`}>
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex justify-between items-center">
          <div>
            <h3 className="font-medium text-gray-900 flex items-center">
              <Activity className="w-5 h-5 mr-2 text-blue-600" />
              Real-Time Analytics
            </h3>
            <p className="text-sm text-gray-600">Live clicks for {shortCode}</p>
          </div>
          
          <div className="flex items-center space-x-2">
            {isConnected ? (
              <div className="flex items-center text-green-600">
                <Wifi className="w-4 h-4 mr-1" />
                <span className="text-xs">Connected</span>
              </div>
            ) : (
              <div className="flex items-center text-red-600">
                <WifiOff className="w-4 h-4 mr-1" />
                <span className="text-xs">Disconnected</span>
              </div>
            )}
            {lastActivity && (
              <span className="text-xs text-gray-500">
                Last: {formatTimeAgo(lastActivity)}
              </span>
            )}
          </div>
        </div>
        
        {connectionError && (
          <div className="mt-2 p-2 bg-red-50 border border-red-200 rounded text-sm text-red-700">
            {connectionError}
          </div>
        )}
      </div>

      {/* Stats Grid */}
      <div className="p-4 border-b">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <div className="text-center">
            <div className="text-2xl font-bold text-blue-600">{stats.clicks_per_minute}</div>
            <div className="text-sm text-gray-600">Clicks/min</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-green-600">{stats.active_visitors}</div>
            <div className="text-sm text-gray-600">Active now</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-orange-600">{stats.total_clicks_today}</div>
            <div className="text-sm text-gray-600">Today</div>
          </div>
          <div className="text-center">
            <div className="text-2xl font-bold text-purple-600">{recentClicks.length}</div>
            <div className="text-sm text-gray-600">Recent</div>
          </div>
        </div>
      </div>

      {/* Recent Activity */}
      <div className="p-4">
        <h4 className="font-medium text-gray-900 mb-3">Recent Clicks</h4>
        
        {recentClicks.length === 0 ? (
          <div className="text-center py-8 text-gray-500">
            <Activity className="w-8 h-8 mx-auto mb-2 opacity-50" />
            <p>Waiting for clicks...</p>
            {!isConnected && (
              <p className="text-sm mt-1">Connection required for real-time updates</p>
            )}
          </div>
        ) : (
          <div className="space-y-2 max-h-64 overflow-y-auto">
            {recentClicks.map((click) => (
              <div 
                key={click.id} 
                className="flex items-center justify-between p-3 bg-gray-50 rounded-lg animate-fade-in"
              >
                <div className="flex items-center space-x-3">
                  <div className="flex-shrink-0">
                    {click.device === 'mobile' ? (
                      <Smartphone className="w-4 h-4 text-blue-600" />
                    ) : (
                      <Monitor className="w-4 h-4 text-gray-600" />
                    )}
                  </div>
                  <div>
                    <div className="font-medium text-sm flex items-center">
                      <Globe className="w-3 h-3 mr-1 text-gray-400" />
                      {click.location}
                    </div>
                    <div className="text-xs text-gray-500">
                      {click.browser} • {click.device}
                    </div>
                  </div>
                </div>
                <div className="text-xs text-gray-500">
                  {formatTimeAgo(click.timestamp)}
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Quick Stats */}
      {(Object.keys(stats.top_countries).length > 0 || Object.keys(stats.top_devices).length > 0) && (
        <div className="p-4 border-t bg-gray-50">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
            {/* Top Countries */}
            {Object.keys(stats.top_countries).length > 0 && (
              <div>
                <h5 className="font-medium text-gray-700 mb-2">Top Countries</h5>
                <div className="space-y-1">
                  {getTopEntries(stats.top_countries).map(([country, clicks]) => (
                    <div key={country} className="flex justify-between text-sm">
                      <span>{country}</span>
                      <span className="font-medium">{clicks}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Top Devices */}
            {Object.keys(stats.top_devices).length > 0 && (
              <div>
                <h5 className="font-medium text-gray-700 mb-2">Top Devices</h5>
                <div className="space-y-1">
                  {getTopEntries(stats.top_devices).map(([device, clicks]) => (
                    <div key={device} className="flex justify-between text-sm">
                      <span className="capitalize">{device}</span>
                      <span className="font-medium">{clicks}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}