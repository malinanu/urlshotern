'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter, useParams } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { AnalyticsTeaser } from '@/components/analytics/AnalyticsTeaser';
import {
  UserGroupIcon,
  GlobeAltIcon,
  ArrowLeftIcon,
  SparklesIcon,
} from '@heroicons/react/24/outline';
import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  BarChart,
  Bar,
  XAxis,
  YAxis,
  CartesianGrid,
} from 'recharts';

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#8b5cf6', '#ec4899', '#6b7280'];

interface ReferrerItem {
  referrer: string;
  clicks: number;
}

interface AudienceInsight {
  insight_type: string;
  content: string;
  generated_at: string;
}

interface AudienceCategory {
  category: string;
  percentage: number;
  top_domains: string[];
}

export default function AudiencePage() {
  const { isAuthenticated, isLoading, getAccessToken, user } = useAuth();
  const router = useRouter();
  const params = useParams();
  const shortCode = params?.shortCode as string;

  const [referrers, setReferrers] = useState<ReferrerItem[]>([]);
  const [aiInsight, setAiInsight] = useState<AudienceInsight | null>(null);
  const [aiCategories, setAiCategories] = useState<AudienceCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [aiLoading, setAiLoading] = useState(false);

  const isEnterprise = (user as any)?.account_type === 'enterprise';

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchReferrers = useCallback(async () => {
    if (!shortCode || !isAuthenticated) return;
    const token = getAccessToken();
    try {
      const res = await fetch(API_ENDPOINTS.analytics.referrers(shortCode), {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setReferrers(data.referrers || data || []);
      }
    } catch {
      // silent
    } finally {
      setLoading(false);
    }
  }, [shortCode, isAuthenticated, getAccessToken]);

  const fetchAiAudience = useCallback(async () => {
    if (!shortCode || !isEnterprise) return;
    setAiLoading(true);
    const token = getAccessToken();
    try {
      const res = await fetch(API_ENDPOINTS.insights.audience(shortCode), {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data: AudienceInsight = await res.json();
        setAiInsight(data);
        try {
          const parsed = JSON.parse(data.content);
          if (Array.isArray(parsed)) setAiCategories(parsed);
        } catch {
          // content is prose, not JSON
        }
      }
    } catch {
      // silent
    } finally {
      setAiLoading(false);
    }
  }, [shortCode, isEnterprise, getAccessToken]);

  useEffect(() => {
    fetchReferrers();
    fetchAiAudience();
  }, [fetchReferrers, fetchAiAudience]);

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary" />
        </div>
      </Layout>
    );
  }

  if (!isAuthenticated) return null;

  const top8 = [...referrers].sort((a, b) => b.clicks - a.clicks).slice(0, 8);
  const total = referrers.reduce((s, r) => s + r.clicks, 0);

  return (
    <Layout>
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="flex items-center gap-4 mb-8">
            <button
              onClick={() => router.back()}
              className="p-2 rounded-lg border border-border hover:bg-muted transition-colors"
            >
              <ArrowLeftIcon className="h-5 w-5 text-muted-foreground" />
            </button>
            <div>
              <div className="flex items-center gap-3 mb-1">
                <UserGroupIcon className="h-8 w-8 text-primary" />
                <h1 className="text-3xl font-bold text-foreground">Audience Profile</h1>
              </div>
              <p className="text-muted-foreground">
                Short code: <span className="font-mono font-medium text-foreground">{shortCode}</span>
              </p>
            </div>
          </div>

          {loading ? (
            <div className="space-y-6">
              {[1, 2].map(i => (
                <div key={i} className="h-64 bg-muted rounded-lg animate-pulse" />
              ))}
            </div>
          ) : (
            <div className="space-y-6">
              {/* Referrer Distribution */}
              <div className="bg-card rounded-lg border border-border p-6">
                <h2 className="text-xl font-semibold text-foreground mb-4 flex items-center gap-2">
                  <GlobeAltIcon className="h-5 w-5 text-primary" />
                  Traffic Sources
                </h2>
                {referrers.length === 0 ? (
                  <p className="text-muted-foreground text-center py-8">No referrer data yet.</p>
                ) : (
                  <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
                    <div className="h-64">
                      <ResponsiveContainer width="100%" height="100%">
                        <PieChart>
                          <Pie
                            data={top8}
                            cx="50%"
                            cy="50%"
                            innerRadius={50}
                            outerRadius={100}
                            paddingAngle={2}
                            dataKey="clicks"
                            nameKey="referrer"
                          >
                            {top8.map((_, i) => (
                              <Cell key={i} fill={COLORS[i % COLORS.length]} />
                            ))}
                          </Pie>
                          <Tooltip
                            formatter={(v: number) => [`${v} clicks`, '']}
                            labelFormatter={(_: any, p: any) => p?.[0]?.payload?.referrer ?? ''}
                          />
                        </PieChart>
                      </ResponsiveContainer>
                    </div>
                    <div className="h-64">
                      <ResponsiveContainer width="100%" height="100%">
                        <BarChart data={top8} layout="vertical">
                          <CartesianGrid strokeDasharray="3 3" />
                          <XAxis type="number" />
                          <YAxis
                            type="category"
                            dataKey="referrer"
                            width={110}
                            tick={{ fontSize: 11 }}
                            tickFormatter={(v: string) => (v.length > 16 ? v.slice(0, 16) + '…' : v || '(direct)')}
                          />
                          <Tooltip formatter={(v: number) => [`${v} clicks`, 'Clicks']} />
                          <Bar dataKey="clicks" fill={COLORS[0]} />
                        </BarChart>
                      </ResponsiveContainer>
                    </div>
                  </div>
                )}
              </div>

              {/* AI Audience Profile */}
              <div className="bg-card rounded-lg border border-border p-6">
                <h2 className="text-xl font-semibold text-foreground mb-4 flex items-center gap-2">
                  <SparklesIcon className="h-5 w-5 text-primary" />
                  AI Audience Categories
                </h2>
                {!isEnterprise ? (
                  <AnalyticsTeaser variant="upgrade-ultra">
                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 pointer-events-none">
                      {['Social Media Users', 'Tech Enthusiasts', 'News Readers'].map(cat => (
                        <div key={cat} className="bg-muted rounded-lg p-4">
                          <p className="font-medium text-sm mb-2">{cat}</p>
                          <div className="h-2 w-full bg-muted-foreground/20 rounded animate-pulse" />
                          <div className="h-2 w-2/3 bg-muted-foreground/20 rounded animate-pulse mt-1" />
                        </div>
                      ))}
                    </div>
                  </AnalyticsTeaser>
                ) : aiLoading ? (
                  <div className="flex items-center gap-3 py-6">
                    <div className="animate-spin rounded-full h-5 w-5 border-b-2 border-primary" />
                    <span className="text-sm text-muted-foreground">Generating AI audience analysis…</span>
                  </div>
                ) : aiCategories.length > 0 ? (
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
                    {aiCategories.map((cat, i) => (
                      <div key={i} className="bg-muted rounded-lg p-4">
                        <div className="flex items-center justify-between mb-2">
                          <p className="font-medium text-sm text-foreground">{cat.category}</p>
                          <span className="text-sm font-bold text-primary">{cat.percentage}%</span>
                        </div>
                        <div className="w-full bg-muted-foreground/20 rounded-full h-1.5 mb-2">
                          <div
                            className="bg-primary h-1.5 rounded-full"
                            style={{ width: `${cat.percentage}%` }}
                          />
                        </div>
                        {cat.top_domains?.length > 0 && (
                          <p className="text-xs text-muted-foreground">
                            via {cat.top_domains.slice(0, 2).join(', ')}
                          </p>
                        )}
                      </div>
                    ))}
                  </div>
                ) : aiInsight ? (
                  <p className="text-sm text-foreground whitespace-pre-wrap leading-relaxed">{aiInsight.content}</p>
                ) : (
                  <p className="text-muted-foreground text-sm">No AI audience data available. Click refresh to generate.</p>
                )}
                {isEnterprise && aiInsight && (
                  <div className="flex items-center justify-between mt-4 pt-4 border-t border-border">
                    <p className="text-xs text-muted-foreground">
                      Generated {new Date(aiInsight.generated_at).toLocaleString()}
                    </p>
                    <button
                      onClick={fetchAiAudience}
                      className="text-xs text-primary hover:opacity-80 underline"
                    >
                      ↻ Refresh
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}
