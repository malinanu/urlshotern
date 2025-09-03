'use client';

import React, { useState, useEffect, useMemo, useCallback } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { 
  LinkIcon, 
  TrendingUpIcon, 
  GlobeIcon, 
  TagIcon,
  ChartBarIcon,
  EyeIcon,
  ArrowTrendingUpIcon,
  ExternalLinkIcon
} from '@heroicons/react/24/outline';

interface UTMParameters {
  utm_source: string;
  utm_medium: string;
  utm_campaign: string;
  utm_term: string;
  utm_content: string;
}

interface ReferrerCategoryStat {
  category: string;
  clicks: number;
  percentage: number;
}

interface CampaignStat {
  campaign: string;
  source: string;
  medium: string;
  clicks: number;
  percentage: number;
}

interface DomainStat {
  domain: string;
  platform: string;
  clicks: number;
  percentage: number;
}

interface UTMParamStat {
  value: string;
  clicks: number;
  percentage: number;
}

interface UTMBreakdown {
  sources: UTMParamStat[];
  mediums: UTMParamStat[];
  campaigns: UTMParamStat[];
}

interface OrganicVsPaidStat {
  organic_clicks: number;
  paid_clicks: number;
  organic_percent: number;
  paid_percent: number;
}

interface ReferrerAnalytics {
  short_code: string;
  total_clicks: number;
  categories: ReferrerCategoryStat[];
  campaigns: CampaignStat[];
  top_domains: DomainStat[];
  utm_breakdown: UTMBreakdown;
  organic_vs_paid: OrganicVsPaidStat;
}

interface ReferrerInsights {
  dominant_category: string;
  dominant_category_percentage: number;
  total_categories: number;
  organic_percentage: number;
  paid_percentage: number;
  unique_referring_domains: number;
  total_referrer_records: number;
  utm_tracking_percentage: number;
}

interface ReferrerAnalyticsProps {
  shortCode: string;
  days?: number;
  height?: string;
}

const CATEGORY_COLORS: { [key: string]: string } = {
  social: 'bg-blue-500',
  search: 'bg-green-500',
  direct: 'bg-gray-500',
  email: 'bg-purple-500',
  news: 'bg-yellow-500',
  forum: 'bg-red-500',
  other: 'bg-orange-500'
};

const CATEGORY_ICONS: { [key: string]: React.ReactNode } = {
  social: <GlobeIcon className="w-4 h-4" />,
  search: <TrendingUpIcon className="w-4 h-4" />,
  direct: <LinkIcon className="w-4 h-4" />,
  email: <TagIcon className="w-4 h-4" />,
  news: <ChartBarIcon className="w-4 h-4" />,
  forum: <EyeIcon className="w-4 h-4" />,
  other: <ArrowTrendingUpIcon className="w-4 h-4" />
};

export default function ReferrerAnalytics({ 
  shortCode, 
  days = 30, 
  height = '600px' 
}: ReferrerAnalyticsProps) {
  const [analytics, setAnalytics] = useState<ReferrerAnalytics | null>(null);
  const [insights, setInsights] = useState<ReferrerInsights | null>(null);
  const [campaigns, setCampaigns] = useState<CampaignStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState<'overview' | 'campaigns' | 'domains' | 'utm'>('overview');
  const { getAccessToken, isAuthenticated } = useAuth();

  // Fetch referrer analytics data
  const fetchAnalytics = useCallback(async () => {
    if (!isAuthenticated || !shortCode) return;

    try {
      setLoading(true);
      const token = getAccessToken();
      
      // Fetch enhanced referrer analytics
      const analyticsResponse = await fetch(
        `http://localhost:8080/api/v1/analytics/${shortCode}/referrers/enhanced?days=${days}`, 
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );
      
      // Fetch insights
      const insightsResponse = await fetch(
        `http://localhost:8080/api/v1/analytics/${shortCode}/referrers/insights?days=${days}`, 
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );
      
      // Fetch UTM campaigns
      const campaignsResponse = await fetch(
        `http://localhost:8080/api/v1/analytics/${shortCode}/utm-campaigns?days=${days}`, 
        {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      if (analyticsResponse.ok) {
        const analyticsData = await analyticsResponse.json();
        setAnalytics(analyticsData);
      }
      
      if (insightsResponse.ok) {
        const insightsData = await insightsResponse.json();
        setInsights(insightsData);
      }
      
      if (campaignsResponse.ok) {
        const campaignData = await campaignsResponse.json();
        setCampaigns(campaignData);
      }
      
      setError(null);
    } catch (err) {
      console.error('Error fetching referrer analytics:', err);
      setError('Failed to fetch referrer analytics data');
    } finally {
      setLoading(false);
    }
  }, [shortCode, days, isAuthenticated, getAccessToken]);

  useEffect(() => {
    fetchAnalytics();
  }, [fetchAnalytics]);

  const getCategoryIcon = (category: string) => {
    return CATEGORY_ICONS[category] || CATEGORY_ICONS.other;
  };

  const getCategoryColor = (category: string) => {
    return CATEGORY_COLORS[category] || CATEGORY_COLORS.other;
  };

  if (loading) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600 mx-auto"></div>
          <p className="mt-2 text-sm text-gray-600">Loading referrer analytics...</p>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4" style={{ height }}>
        <div className="flex items-center justify-center h-full">
          <div className="text-center">
            <p className="text-red-800 font-medium">Error loading referrer analytics</p>
            <p className="text-red-600 text-sm mt-1">{error}</p>
            <button 
              onClick={fetchAnalytics}
              className="mt-3 px-4 py-2 bg-red-100 hover:bg-red-200 text-red-800 rounded text-sm"
            >
              Retry
            </button>
          </div>
        </div>
      </div>
    );
  }

  if (!analytics) {
    return (
      <div className="bg-gray-50 rounded-lg flex items-center justify-center" style={{ height }}>
        <div className="text-center text-gray-600">
          <p className="font-medium">No referrer data available</p>
          <p className="text-sm mt-1">No referrer analytics found for the selected period</p>
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
            <h3 className="font-medium text-gray-900 flex items-center">
              <ExternalLinkIcon className="w-5 h-5 mr-2 text-blue-600" />
              Referrer Analytics
            </h3>
            <p className="text-sm text-gray-600">
              {analytics.total_clicks.toLocaleString()} total clicks analyzed
            </p>
          </div>
          
          {/* Tab selector */}
          <div className="flex bg-gray-100 rounded-lg p-1">
            {[
              { key: 'overview', label: 'Overview' },
              { key: 'campaigns', label: 'Campaigns' },
              { key: 'domains', label: 'Domains' },
              { key: 'utm', label: 'UTM' }
            ].map((tab) => (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key as any)}
                className={`px-3 py-1 text-xs rounded transition-colors ${
                  activeTab === tab.key
                    ? 'bg-white text-gray-900 shadow-sm'
                    : 'text-gray-600 hover:text-gray-900'
                }`}
              >
                {tab.label}
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* Content */}
      <div className="p-4" style={{ height: `calc(${height} - 80px)`, overflow: 'auto' }}>
        {activeTab === 'overview' && (
          <div className="space-y-6">
            {/* Key Insights */}
            {insights && (
              <div className="bg-gradient-to-r from-blue-50 to-purple-50 rounded-lg p-4">
                <h4 className="font-medium text-gray-900 mb-3">Key Insights</h4>
                <div className="grid grid-cols-2 lg:grid-cols-4 gap-4 text-sm">
                  <div className="text-center">
                    <div className="text-lg font-bold text-blue-600">
                      {insights.dominant_category}
                    </div>
                    <div className="text-gray-600">Top Category</div>
                    <div className="text-xs text-gray-500">
                      {insights.dominant_category_percentage?.toFixed(1)}%
                    </div>
                  </div>
                  <div className="text-center">
                    <div className="text-lg font-bold text-green-600">
                      {insights.organic_percentage?.toFixed(1)}%
                    </div>
                    <div className="text-gray-600">Organic Traffic</div>
                  </div>
                  <div className="text-center">
                    <div className="text-lg font-bold text-purple-600">
                      {insights.unique_referring_domains}
                    </div>
                    <div className="text-gray-600">Unique Domains</div>
                  </div>
                  <div className="text-center">
                    <div className="text-lg font-bold text-orange-600">
                      {insights.utm_tracking_percentage?.toFixed(1)}%
                    </div>
                    <div className="text-gray-600">UTM Tracked</div>
                  </div>
                </div>
              </div>
            )}

            {/* Organic vs Paid */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <div className="bg-gray-50 rounded-lg p-4">
                <h4 className="font-medium text-gray-900 mb-4">Organic vs Paid Traffic</h4>
                <div className="space-y-3">
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-700 flex items-center">
                      <div className="w-3 h-3 bg-green-500 rounded-full mr-2"></div>
                      Organic
                    </span>
                    <div className="text-right">
                      <div className="text-sm font-medium">{analytics.organic_vs_paid.organic_clicks.toLocaleString()}</div>
                      <div className="text-xs text-gray-500">{analytics.organic_vs_paid.organic_percent.toFixed(1)}%</div>
                    </div>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-green-500 h-2 rounded-full"
                      style={{ width: `${analytics.organic_vs_paid.organic_percent}%` }}
                    ></div>
                  </div>
                  
                  <div className="flex justify-between items-center">
                    <span className="text-sm text-gray-700 flex items-center">
                      <div className="w-3 h-3 bg-blue-500 rounded-full mr-2"></div>
                      Paid
                    </span>
                    <div className="text-right">
                      <div className="text-sm font-medium">{analytics.organic_vs_paid.paid_clicks.toLocaleString()}</div>
                      <div className="text-xs text-gray-500">{analytics.organic_vs_paid.paid_percent.toFixed(1)}%</div>
                    </div>
                  </div>
                  <div className="w-full bg-gray-200 rounded-full h-2">
                    <div 
                      className="bg-blue-500 h-2 rounded-full"
                      style={{ width: `${analytics.organic_vs_paid.paid_percent}%` }}
                    ></div>
                  </div>
                </div>
              </div>

              {/* Categories */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h4 className="font-medium text-gray-900 mb-4">Traffic Categories</h4>
                <div className="space-y-2">
                  {analytics.categories.slice(0, 6).map((category) => (
                    <div key={category.category} className="flex items-center justify-between">
                      <div className="flex items-center">
                        <div className={`w-3 h-3 rounded-full mr-2 ${getCategoryColor(category.category)}`}></div>
                        <div className="flex items-center">
                          {getCategoryIcon(category.category)}
                          <span className="text-sm font-medium ml-2 capitalize">{category.category}</span>
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-sm font-medium">{category.clicks.toLocaleString()}</div>
                        <div className="text-xs text-gray-500">{category.percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'campaigns' && (
          <div className="space-y-4">
            <h4 className="font-medium text-gray-900">UTM Campaigns</h4>
            {campaigns.length === 0 ? (
              <div className="text-center py-8 text-gray-500">
                <TagIcon className="w-8 h-8 mx-auto mb-2 opacity-50" />
                <p>No UTM campaigns found</p>
                <p className="text-sm mt-1">Add UTM parameters to track campaign performance</p>
              </div>
            ) : (
              <div className="space-y-3">
                {campaigns.map((campaign, index) => (
                  <div key={index} className="bg-gray-50 rounded-lg p-4">
                    <div className="flex justify-between items-start">
                      <div>
                        <div className="font-medium text-gray-900">
                          {campaign.campaign || 'Unnamed Campaign'}
                        </div>
                        <div className="text-sm text-gray-600 mt-1">
                          <span className="inline-block bg-blue-100 text-blue-800 px-2 py-1 rounded text-xs mr-2">
                            {campaign.source}
                          </span>
                          {campaign.medium && (
                            <span className="inline-block bg-green-100 text-green-800 px-2 py-1 rounded text-xs">
                              {campaign.medium}
                            </span>
                          )}
                        </div>
                      </div>
                      <div className="text-right">
                        <div className="text-lg font-bold text-blue-600">{campaign.clicks.toLocaleString()}</div>
                        <div className="text-sm text-gray-500">{campaign.percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </div>
        )}

        {activeTab === 'domains' && (
          <div className="space-y-4">
            <h4 className="font-medium text-gray-900">Top Referring Domains</h4>
            <div className="space-y-2">
              {analytics.top_domains.slice(0, 10).map((domain, index) => (
                <div key={domain.domain} className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                  <div className="flex items-center">
                    <div className="w-6 h-6 bg-gray-200 rounded-full flex items-center justify-center mr-3 text-xs font-mono">
                      {index + 1}
                    </div>
                    <div>
                      <div className="font-medium text-gray-900">{domain.platform}</div>
                      <div className="text-sm text-gray-500">{domain.domain}</div>
                    </div>
                  </div>
                  <div className="text-right">
                    <div className="font-medium">{domain.clicks.toLocaleString()}</div>
                    <div className="text-xs text-gray-500">{domain.percentage.toFixed(1)}%</div>
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {activeTab === 'utm' && (
          <div className="space-y-6">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* UTM Sources */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h5 className="font-medium text-gray-900 mb-3">Sources</h5>
                <div className="space-y-2">
                  {analytics.utm_breakdown.sources.slice(0, 5).map((source) => (
                    <div key={source.value} className="flex justify-between items-center text-sm">
                      <span className="font-medium">{source.value}</span>
                      <div className="text-right">
                        <div className="font-medium">{source.clicks.toLocaleString()}</div>
                        <div className="text-xs text-gray-500">{source.percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* UTM Mediums */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h5 className="font-medium text-gray-900 mb-3">Mediums</h5>
                <div className="space-y-2">
                  {analytics.utm_breakdown.mediums.slice(0, 5).map((medium) => (
                    <div key={medium.value} className="flex justify-between items-center text-sm">
                      <span className="font-medium">{medium.value}</span>
                      <div className="text-right">
                        <div className="font-medium">{medium.clicks.toLocaleString()}</div>
                        <div className="text-xs text-gray-500">{medium.percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* UTM Campaigns */}
              <div className="bg-gray-50 rounded-lg p-4">
                <h5 className="font-medium text-gray-900 mb-3">Campaigns</h5>
                <div className="space-y-2">
                  {analytics.utm_breakdown.campaigns.slice(0, 5).map((campaign) => (
                    <div key={campaign.value} className="flex justify-between items-center text-sm">
                      <span className="font-medium">{campaign.value}</span>
                      <div className="text-right">
                        <div className="font-medium">{campaign.clicks.toLocaleString()}</div>
                        <div className="text-xs text-gray-500">{campaign.percentage.toFixed(1)}%</div>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}