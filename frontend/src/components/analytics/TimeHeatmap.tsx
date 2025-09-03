'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { format, startOfDay, addDays, subDays, parseISO } from 'date-fns';

interface HeatmapPoint {
  date: string; // YYYY-MM-DD format
  hour: number; // 0-23
  clicks: number;
}

interface HourlyClick {
  hour: number;
  clicks: number;
}

interface WeekdayClick {
  weekday: number; // 0-6 (Sunday-Saturday)
  day: string;
  clicks: number;
}

interface PeakTimeInfo {
  peak_hour: number;
  peak_weekday: number;
  peak_day: string;
  max_clicks: number;
}

interface TimeAnalytics {
  short_code: string;
  hourly_pattern: HourlyClick[];
  weekly_pattern: WeekdayClick[];
  heatmap_data: HeatmapPoint[];
  peak_times: PeakTimeInfo;
}

interface TimeHeatmapProps {
  shortCode: string;
  days?: number;
  height?: string;
}

const HOURS = Array.from({ length: 24 }, (_, i) => i);
const WEEKDAYS = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];

export default function TimeHeatmap({ 
  shortCode, 
  days = 30, 
  height = '400px' 
}: TimeHeatmapProps) {
  const [timeData, setTimeData] = useState<TimeAnalytics | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<'heatmap' | 'hourly' | 'weekly'>('heatmap');

  // Fetch time analytics data
  const fetchTimeData = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/analytics/time/${shortCode}?days=${days}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch time analytics data');
      }
      
      const data = await response.json();
      setTimeData(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load time data');
      console.error('Error fetching time data:', err);
    } finally {
      setLoading(false);
    }
  }, [shortCode, days]);

  useEffect(() => {
    fetchTimeData();
  }, [fetchTimeData]);

  // Process heatmap data into a grid format
  const heatmapGrid = useMemo(() => {
    if (!timeData?.heatmap_data) return [];

    const endDate = new Date();
    const startDate = subDays(endDate, days - 1);
    const grid: { date: string; data: number[] }[] = [];

    // Create grid for each day
    for (let d = startDate; d <= endDate; d = addDays(d, 1)) {
      const dateStr = format(d, 'yyyy-MM-dd');
      const dayData = new Array(24).fill(0);

      // Fill in actual data
      timeData.heatmap_data
        .filter(point => point.date === dateStr)
        .forEach(point => {
          dayData[point.hour] = point.clicks;
        });

      grid.push({
        date: dateStr,
        data: dayData
      });
    }

    return grid;
  }, [timeData, days]);

  // Calculate max clicks for color scaling
  const maxClicks = useMemo(() => {
    if (!timeData?.heatmap_data) return 0;
    return Math.max(...timeData.heatmap_data.map(point => point.clicks));
  }, [timeData]);

  // Get color for heatmap cell
  const getHeatmapColor = useCallback((clicks: number) => {
    if (clicks === 0) return 'bg-gray-100';
    
    const intensity = clicks / maxClicks;
    if (intensity >= 0.8) return 'bg-red-500';
    if (intensity >= 0.6) return 'bg-red-400';
    if (intensity >= 0.4) return 'bg-orange-400';
    if (intensity >= 0.2) return 'bg-yellow-400';
    return 'bg-green-300';
  }, [maxClicks]);

  const formatHour = (hour: number) => {
    if (hour === 0) return '12 AM';
    if (hour < 12) return `${hour} AM`;
    if (hour === 12) return '12 PM';
    return `${hour - 12} PM`;
  };

  if (loading) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading time analytics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-red-800 font-medium">Error loading time analytics</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
            <button 
              onClick={fetchTimeData}
              className="mt-3 px-4 py-2 bg-red-100 hover:bg-red-200 text-red-800 rounded text-sm"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!timeData) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center text-gray-600">
          <p className="font-medium">No time data available</p>
          <p className="text-sm mt-1">No time-based analytics found for the selected period</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border" style={{ height }}>
      {/* Header */}
      <div className="p-4 border-b">
        <div className="flex justify-between items-center">
          <div>
            <h3 className="font-medium text-gray-900">Time-based Analytics</h3>
            <p className="text-sm text-gray-600">
              Peak time: {formatHour(timeData.peak_times.peak_hour)} on {timeData.peak_times.peak_day}
            </p>
          </div>
          
          {/* View mode selector */}
          <div className="flex bg-gray-100 rounded-lg p-1">
            {[
              { key: 'heatmap', label: 'Heatmap' },
              { key: 'hourly', label: 'Hourly' },
              { key: 'weekly', label: 'Weekly' }
            ].map((mode) => (
              <button
                key={mode.key}
                onClick={() => setViewMode(mode.key as any)}
                className={`px-3 py-1 text-xs rounded transition-colors ${
                  viewMode === mode.key
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                {mode.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4" style={{ height: `calc(${height} - 80px)`, overflow: 'auto' }}>
        {viewMode === 'heatmap' && (
          <div className="space-y-4">
            {/* Color Legend */}
            <div className="flex items-center justify-between text-xs">
              <span className="text-gray-600">Less activity</span>
              <div className="flex space-x-1">
                <div className="w-3 h-3 bg-gray-100 border"></div>
                <div className="w-3 h-3 bg-green-300"></div>
                <div className="w-3 h-3 bg-yellow-400"></div>
                <div className="w-3 h-3 bg-orange-400"></div>
                <div className="w-3 h-3 bg-red-400"></div>
                <div className="w-3 h-3 bg-red-500"></div>
              </div>
              <span className="text-gray-600">More activity</span>
            </div>

            {/* Heatmap Grid */}
            <div className="overflow-x-auto">
              <div className="inline-block min-w-full">
                {/* Hour labels */}
                <div className="flex">
                  <div className="w-20"></div> {/* Date column */}
                  {HOURS.map(hour => (
                    <div key={hour} className="w-6 text-center">
                      <span className="text-xs text-gray-500">
                        {hour % 4 === 0 ? hour : ''}
                      </span>
                    </div>
                  ))}
                </div>

                {/* Heatmap rows */}
                {heatmapGrid.map((day, index) => (
                  <div key={day.date} className="flex items-center">
                    <div className="w-20 text-xs text-gray-600 pr-2">
                      {index % 7 === 0 || index === heatmapGrid.length - 1
                        ? format(parseISO(day.date), 'MMM d')
                        : ''}
                    </div>
                    {day.data.map((clicks, hour) => (
                      <div
                        key={`${day.date}-${hour}`}
                        className={`w-6 h-4 m-px rounded-sm ${getHeatmapColor(clicks)} cursor-pointer`}
                        title={`${format(parseISO(day.date), 'MMM d')} at ${formatHour(hour)}: ${clicks} clicks`}
                      />
                    ))}
                  </div>
                ))}
              </div>
            </div>
          </div>
        )}

        {viewMode === 'hourly' && (
          <div className="space-y-4">
            <h4 className="font-medium text-gray-900">Hourly Pattern</h4>
            <div className="grid grid-cols-6 gap-2">
              {timeData.hourly_pattern.map((hourData, index) => {
                const maxHourlyClicks = Math.max(...timeData.hourly_pattern.map(h => h.clicks));
                const height = maxHourlyClicks > 0 ? (hourData.clicks / maxHourlyClicks) * 100 : 0;
                
                return (
                  <div key={hourData.hour} className="text-center">
                    <div className="h-24 flex items-end justify-center mb-2">
                      <div
                        className="w-8 bg-blue-500 rounded-t"
                        style={{ height: `${Math.max(height, 2)}%` }}
                        title={`${formatHour(hourData.hour)}: ${hourData.clicks} clicks`}
                      />
                    </div>
                    <div className="text-xs text-gray-600">
                      {hourData.hour % 4 === 0 ? formatHour(hourData.hour) : hourData.hour}
                    </div>
                    <div className="text-xs font-medium">{hourData.clicks}</div>
                  </div>
                );
              })}
            </div>
          </div>
        )}

        {viewMode === 'weekly' && (
          <div className="space-y-4">
            <h4 className="font-medium text-gray-900">Weekly Pattern</h4>
            <div className="grid grid-cols-7 gap-2">
              {timeData.weekly_pattern.map((dayData) => {
                const maxWeeklyClicks = Math.max(...timeData.weekly_pattern.map(d => d.clicks));
                const height = maxWeeklyClicks > 0 ? (dayData.clicks / maxWeeklyClicks) * 100 : 0;
                
                return (
                  <div key={dayData.weekday} className="text-center">
                    <div className="h-32 flex items-end justify-center mb-2">
                      <div
                        className="w-12 bg-green-500 rounded-t"
                        style={{ height: `${Math.max(height, 2)}%` }}
                        title={`${dayData.day}: ${dayData.clicks} clicks`}
                      />
                    </div>
                    <div className="text-sm font-medium text-gray-900">{dayData.day}</div>
                    <div className="text-xs text-gray-600">{dayData.clicks} clicks</div>
                  </div>
                );
              })}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}

// Peak times summary component
export function PeakTimesSummary({ timeData }: { timeData: TimeAnalytics | null }) {
  if (!timeData) return null;

  const formatHour = (hour: number) => {
    if (hour === 0) return '12 AM';
    if (hour < 12) return `${hour} AM`;
    if (hour === 12) return '12 PM';
    return `${hour - 12} PM`;
  };

  const topHours = timeData.hourly_pattern
    .sort((a, b) => b.clicks - a.clicks)
    .slice(0, 3);

  const topDays = timeData.weekly_pattern
    .sort((a, b) => b.clicks - a.clicks)
    .slice(0, 3);

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <h4 className="font-medium text-gray-900 mb-4">Peak Activity Times</h4>
      
      <div className="space-y-4">
        {/* Peak overall */}
        <div className="bg-blue-50 rounded-lg p-3">
          <div className="text-sm font-medium text-blue-900">Peak Time</div>
          <div className="text-lg font-bold text-blue-600">
            {formatHour(timeData.peak_times.peak_hour)} on {timeData.peak_times.peak_day}
          </div>
          <div className="text-sm text-blue-700">
            {timeData.peak_times.max_clicks} clicks
          </div>
        </div>

        {/* Top hours */}
        <div>
          <div className="text-sm font-medium text-gray-700 mb-2">Top Hours</div>
          <div className="space-y-1">
            {topHours.map((hour, index) => (
              <div key={hour.hour} className="flex justify-between items-center text-sm">
                <span className="flex items-center">
                  <span className="w-4 h-4 bg-blue-100 text-blue-800 rounded text-xs flex items-center justify-center mr-2">
                    {index + 1}
                  </span>
                  {formatHour(hour.hour)}
                </span>
                <span className="font-medium">{hour.clicks} clicks</span>
              </div>
            ))}
          </div>
        </div>

        {/* Top days */}
        <div>
          <div className="text-sm font-medium text-gray-700 mb-2">Top Days</div>
          <div className="space-y-1">
            {topDays.map((day, index) => (
              <div key={day.weekday} className="flex justify-between items-center text-sm">
                <span className="flex items-center">
                  <span className="w-4 h-4 bg-green-100 text-green-800 rounded text-xs flex items-center justify-center mr-2">
                    {index + 1}
                  </span>
                  {day.day}
                </span>
                <span className="font-medium">{day.clicks} clicks</span>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
}