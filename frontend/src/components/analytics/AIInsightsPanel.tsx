'use client';

import React from 'react';
import { AIInsightCard, type InsightType } from './AIInsightCard';
import { AnalyticsTeaser } from './AnalyticsTeaser';
import { useAuth } from '@/contexts/AuthContext';

interface AIInsightsPanelProps {
  shortCode: string;
}

const INSIGHT_CARDS: Array<{
  insightType: InsightType;
  title: string;
  description: string;
  icon: string;
}> = [
  {
    insightType: 'summary',
    title: 'Performance Summary',
    description: 'AI-generated overview of link performance',
    icon: '📊',
  },
  {
    insightType: 'anomalies',
    title: 'Anomaly Detection',
    description: 'Unusual traffic patterns and spikes',
    icon: '🔍',
  },
  {
    insightType: 'recommendations',
    title: 'Recommendations',
    description: 'Actionable tips to boost performance',
    icon: '💡',
  },
  {
    insightType: 'forecast',
    title: 'Traffic Forecast',
    description: 'Projected clicks based on trend data',
    icon: '📈',
  },
  {
    insightType: 'audience',
    title: 'Audience Profile',
    description: 'Visitor segments and traffic sources',
    icon: '👥',
  },
];

// Mock layout shown inside the teaser blur
function MockLayout() {
  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 pointer-events-none">
      {INSIGHT_CARDS.map((card) => (
        <div
          key={card.insightType}
          className="bg-card rounded-xl border border-border p-5 min-h-[180px] flex flex-col gap-3"
        >
          <div className="flex items-center gap-2">
            <span className="text-xl">{card.icon}</span>
            <div>
              <p className="font-semibold text-sm">{card.title}</p>
              <p className="text-xs text-muted-foreground">{card.description}</p>
            </div>
          </div>
          <div className="flex flex-col gap-1.5 mt-2">
            <div className="h-3 w-full bg-muted rounded animate-pulse" />
            <div className="h-3 w-4/5 bg-muted rounded animate-pulse" />
            <div className="h-3 w-3/5 bg-muted rounded animate-pulse" />
          </div>
        </div>
      ))}
    </div>
  );
}

export function AIInsightsPanel({ shortCode }: AIInsightsPanelProps) {
  const { user, isAuthenticated } = useAuth();

  if (!isAuthenticated) {
    return (
      <AnalyticsTeaser variant="unauthenticated">
        <MockLayout />
      </AnalyticsTeaser>
    );
  }

  const isEnterprise = (user as any)?.account_type === 'enterprise';

  if (!isEnterprise) {
    return (
      <AnalyticsTeaser variant="upgrade-ultra">
        <MockLayout />
      </AnalyticsTeaser>
    );
  }

  return (
    <div className="flex flex-col gap-4">
      <p className="text-sm text-muted-foreground">
        AI insights are generated on-demand and cached for 1 hour. Click ↻ to force a refresh.
      </p>
      <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
        {INSIGHT_CARDS.map((card) => (
          <AIInsightCard
            key={card.insightType}
            shortCode={shortCode}
            insightType={card.insightType}
            title={card.title}
            description={card.description}
            icon={card.icon}
          />
        ))}
      </div>
    </div>
  );
}

export default AIInsightsPanel;
