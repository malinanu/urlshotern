'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { 
  APIProvider, 
  Map, 
  Marker,
  InfoWindow,
  AdvancedMarker,
  Pin
} from '@vis.gl/react-google-maps';

interface MapPoint {
  latitude: number;
  longitude: number;
  clicks: number;
  location: string;
  country_code: string;
}

interface CountryDetail {
  country_code: string;
  country_name: string;
  clicks: number;
  percentage: number;
  unique_ips: number;
  last_click?: string;
}

interface GeographicAnalytics {
  short_code: string;
  total_clicks: number;
  countries: CountryDetail[];
  map_data: MapPoint[];
}

interface GeographicMapProps {
  shortCode: string;
  days?: number;
  height?: string;
  googleMapsApiKey?: string;
}

const FALLBACK_MAP_CENTER = { lat: 40.7128, lng: -74.0060 }; // New York City

export default function GeographicMap({ 
  shortCode, 
  days = 30, 
  height = '400px',
  googleMapsApiKey = process.env.NEXT_PUBLIC_GOOGLE_MAPS_API_KEY || ''
}: GeographicMapProps) {
  const [geoData, setGeoData] = useState<GeographicAnalytics | null>(null);
  const [selectedMarker, setSelectedMarker] = useState<MapPoint | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Fetch geographic data
  const fetchGeographicData = useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/analytics/geographic/${shortCode}?days=${days}`);
      
      if (!response.ok) {
        throw new Error('Failed to fetch geographic data');
      }
      
      const data = await response.json();
      setGeoData(data);
      setError(null);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load geographic data');
      console.error('Error fetching geographic data:', err);
    } finally {
      setLoading(false);
    }
  }, [shortCode, days]);

  useEffect(() => {
    fetchGeographicData();
  }, [fetchGeographicData]);

  // Calculate map center based on data
  const mapCenter = useMemo(() => {
    if (!geoData?.map_data || geoData.map_data.length === 0) {
      return FALLBACK_MAP_CENTER;
    }

    // Find the location with most clicks for center
    const topLocation = geoData.map_data.reduce((max, location) => 
      location.clicks > max.clicks ? location : max
    );

    return {
      lat: topLocation.latitude,
      lng: topLocation.longitude
    };
  }, [geoData]);

  // Calculate marker sizes based on click counts
  const getMarkerSize = useCallback((clicks: number, maxClicks: number) => {
    const minSize = 20;
    const maxSize = 60;
    const ratio = clicks / maxClicks;
    return Math.max(minSize, Math.min(maxSize, minSize + (maxSize - minSize) * ratio));
  }, []);

  const getMarkerColor = useCallback((clicks: number, maxClicks: number) => {
    const ratio = clicks / maxClicks;
    
    if (ratio > 0.7) return '#ef4444'; // red-500
    if (ratio > 0.4) return '#f97316'; // orange-500
    if (ratio > 0.2) return '#eab308'; // yellow-500
    return '#22c55e'; // green-500
  }, []);

  // Calculate max clicks for scaling
  const maxClicks = useMemo(() => {
    if (!geoData?.map_data) return 0;
    return Math.max(...geoData.map_data.map(point => point.clicks));
  }, [geoData]);

  if (!googleMapsApiKey) {
    return (
      <div className="bg-gray-100 rounded-lg p-8 text-center" style={{ height }}>
        <div className="text-gray-600">
          <p className="font-medium">Geographic Map Unavailable</p>
          <p className="text-sm mt-2">Google Maps API key is required to display the interactive map</p>
          <p className="text-xs mt-2 text-gray-500">Set NEXT_PUBLIC_GOOGLE_MAPS_API_KEY in your environment</p>
        </div>
        
        {/* Fallback: Show country list */}
        {geoData?.countries && (
          <div className="mt-6">
            <h4 className="font-medium text-gray-800 mb-3">Top Countries</h4>
            <div className="space-y-2 max-h-48 overflow-y-auto">
              {geoData.countries.slice(0, 10).map((country, index) => (
                <div key={country.country_code} className="flex justify-between items-center text-sm">
                  <span className="font-medium">{country.country_name}</span>
                  <span className="text-gray-600">{country.clicks.toLocaleString()} clicks</span>
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  }

  if (loading) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading geographic data...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-red-800 font-medium">Error loading map</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
            <button 
              onClick={fetchGeographicData}
              className="mt-3 px-4 py-2 bg-red-100 hover:bg-red-200 text-red-800 rounded text-sm"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!geoData?.map_data || geoData.map_data.length === 0) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center text-gray-600">
          <p className="font-medium">No geographic data available</p>
          <p className="text-sm mt-1">No clicks with location data found for the selected period</p>
        </div>
      </div>
    );
  }

  return (
    <div className="bg-white rounded-lg shadow-sm border" style={{ height }}>
      <div className="p-4 border-b">
        <div className="flex justify-between items-center">
          <div>
            <h3 className="font-medium text-gray-900">Geographic Distribution</h3>
            <p className="text-sm text-gray-600">
              {geoData.total_clicks.toLocaleString()} total clicks from {geoData.countries.length} countries
            </p>
          </div>
          <div className="flex items-center space-x-4 text-xs">
            <div className="flex items-center">
              <div className="w-3 h-3 bg-green-500 rounded-full mr-1"></div>
              <span>Low</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 bg-yellow-500 rounded-full mr-1"></div>
              <span>Medium</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 bg-orange-500 rounded-full mr-1"></div>
              <span>High</span>
            </div>
            <div className="flex items-center">
              <div className="w-3 h-3 bg-red-500 rounded-full mr-1"></div>
              <span>Highest</span>
            </div>
          </div>
        </div>
      </div>

      <div style={{ height: `calc(${height} - 80px)` }}>
        <APIProvider apiKey={googleMapsApiKey}>
          <Map
            defaultCenter={mapCenter}
            defaultZoom={3}
            mapId="geographic-analytics-map"
            style={{ width: '100%', height: '100%' }}
            gestureHandling="cooperative"
            disableDefaultUI={false}
            zoomControl={true}
            mapTypeControl={false}
            streetViewControl={false}
            fullscreenControl={true}
          >
            {geoData.map_data.map((point, index) => (
              <AdvancedMarker
                key={`${point.latitude}-${point.longitude}-${index}`}
                position={{ lat: point.latitude, lng: point.longitude }}
                onClick={() => setSelectedMarker(point)}
              >
                <Pin
                  background={getMarkerColor(point.clicks, maxClicks)}
                  borderColor="#ffffff"
                  glyphColor="#ffffff"
                  scale={getMarkerSize(point.clicks, maxClicks) / 30}
                />
              </AdvancedMarker>
            ))}

            {selectedMarker && (
              <InfoWindow
                position={{ lat: selectedMarker.latitude, lng: selectedMarker.longitude }}
                onCloseClick={() => setSelectedMarker(null)}
              >
                <div className="p-2 min-w-[200px]">
                  <h4 className="font-medium text-gray-900 mb-2">
                    {selectedMarker.location}
                  </h4>
                  <div className="space-y-1 text-sm">
                    <div className="flex justify-between">
                      <span className="text-gray-600">Clicks:</span>
                      <span className="font-medium">{selectedMarker.clicks.toLocaleString()}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Country:</span>
                      <span className="font-medium">{selectedMarker.country_code}</span>
                    </div>
                    <div className="flex justify-between">
                      <span className="text-gray-600">Percentage:</span>
                      <span className="font-medium">
                        {((selectedMarker.clicks / geoData.total_clicks) * 100).toFixed(1)}%
                      </span>
                    </div>
                  </div>
                </div>
              </InfoWindow>
            )}
          </Map>
        </APIProvider>
      </div>
    </div>
  );
}

// Summary stats component for the map
export function GeographicSummary({ geoData }: { geoData: GeographicAnalytics | null }) {
  if (!geoData) return null;

  const topCountries = geoData.countries.slice(0, 5);

  return (
    <div className="bg-white rounded-lg shadow-sm border p-4">
      <h4 className="font-medium text-gray-900 mb-3">Top Countries</h4>
      <div className="space-y-2">
        {topCountries.map((country, index) => (
          <div key={country.country_code} className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="w-6 h-4 bg-gray-200 rounded-sm mr-2 flex items-center justify-center text-xs font-mono">
                {country.country_code}
              </div>
              <span className="text-sm font-medium">{country.country_name}</span>
            </div>
            <div className="text-right">
              <div className="text-sm font-medium">{country.clicks.toLocaleString()}</div>
              <div className="text-xs text-gray-500">{country.percentage.toFixed(1)}%</div>
            </div>
          </div>
        ))}
      </div>
      
      {geoData.countries.length > 5 && (
        <div className="mt-3 pt-2 border-t text-center">
          <span className="text-xs text-gray-500">
            +{geoData.countries.length - 5} more countries
          </span>
        </div>
      )}
    </div>
  );
}