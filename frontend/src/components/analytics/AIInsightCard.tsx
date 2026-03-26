'use client';

import React, { useCallback, useEffect, useState } from 'react';
import { API_ENDPOINTS } from '@/lib/api-config';

export type InsightType = 'summary' | 'anomalies' | 'recommendations' | 'forecast' | 'audience';

interface AIInsightCardProps {
  shortCode: string;
  insightType: InsightType;
  title: string;
  description: string;
  icon: string;
}

interface AIInsight {
  insight_type: string;
  content: string;
  generated_at: string;
}

// ─── Markdown renderer ────────────────────────────────────────────────────────
function parseMarkdown(md: string): string {
  return md
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/__(.*?)__/g, '<strong>$1</strong>')
    .replace(/\*([^*\n]+)\*/g, '<em>$1</em>')
    .replace(/^[\-\*] (.+)$/gm, '<li>$1</li>')
    .replace(/^\d+\. (.+)$/gm, '<li>$1</li>')
    .replace(/(<li>[\s\S]*?<\/li>\n?)+/g, (m) => `<ul class="list-disc list-inside space-y-1 my-1 pl-1">${m}</ul>`)
    .replace(/^###\s(.+)$/gm, '<p class="font-semibold text-xs mt-2 mb-0.5">$1</p>')
    .replace(/^##\s(.+)$/gm, '<p class="font-semibold text-sm mt-3 mb-1">$1</p>')
    .replace(/\n\n/g, '</p><p class="mt-2">')
    .replace(/\n/g, '<br/>')
    .replace(/^([\s\S]+)$/, '<p>$1</p>');
}

// ─── Anomaly renderer ─────────────────────────────────────────────────────────
interface AnomalyData {
  detected: boolean;
  severity: 'low' | 'medium' | 'high';
  explanation: string;
  possible_cause: string;
}

function AnomalyView({ content }: { content: string }) {
  let data: AnomalyData | null = null;
  try {
    data = JSON.parse(content);
  } catch {
    // fallback to markdown if parse fails
    return <div className="text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: parseMarkdown(content) }} />;
  }
  if (!data) return null;

  const severityColor: Record<string, string> = {
    low: 'bg-yellow-100 text-yellow-800 dark:bg-yellow-900/30 dark:text-yellow-300',
    medium: 'bg-orange-100 text-orange-800 dark:bg-orange-900/30 dark:text-orange-300',
    high: 'bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-300',
  };

  return (
    <div className="flex flex-col gap-2.5">
      {/* Status + severity badges */}
      <div className="flex items-center gap-2 flex-wrap">
        <span className={`inline-flex items-center gap-1 text-xs font-medium px-2.5 py-1 rounded-full ${data.detected ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-300' : 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-300'}`}>
          {data.detected ? '⚠ Anomaly detected' : '✓ No anomaly'}
        </span>
        {data.detected && (
          <span className={`inline-flex items-center text-xs font-medium px-2.5 py-1 rounded-full capitalize ${severityColor[data.severity] ?? severityColor.low}`}>
            {data.severity} severity
          </span>
        )}
      </div>

      {/* Explanation */}
      <div>
        <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-0.5">Explanation</p>
        <p className="text-sm leading-relaxed">{data.explanation}</p>
      </div>

      {/* Possible cause */}
      {data.possible_cause && (
        <div>
          <p className="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-0.5">Possible cause</p>
          <p className="text-sm leading-relaxed text-muted-foreground">{data.possible_cause}</p>
        </div>
      )}
    </div>
  );
}

// ─── Audience renderer ────────────────────────────────────────────────────────
interface AudienceSegment {
  category: string;
  percentage: number;
  top_domains: string[];
}

function AudienceView({ content }: { content: string }) {
  let segments: AudienceSegment[] = [];
  try {
    const parsed = JSON.parse(content);
    segments = Array.isArray(parsed) ? parsed : [];
  } catch {
    return <div className="text-sm leading-relaxed" dangerouslySetInnerHTML={{ __html: parseMarkdown(content) }} />;
  }

  const colors = [
    'bg-blue-500', 'bg-violet-500', 'bg-emerald-500', 'bg-amber-500', 'bg-rose-500',
  ];

  return (
    <div className="flex flex-col gap-2.5">
      {segments.map((seg, i) => (
        <div key={seg.category} className="flex flex-col gap-1">
          <div className="flex justify-between items-baseline">
            <span className="text-xs font-medium text-foreground">{seg.category}</span>
            <span className="text-xs font-semibold text-foreground">{seg.percentage}%</span>
          </div>
          {/* Progress bar */}
          <div className="h-1.5 w-full bg-muted rounded-full overflow-hidden">
            <div
              className={`h-full rounded-full ${colors[i % colors.length]}`}
              style={{ width: `${Math.min(seg.percentage, 100)}%` }}
            />
          </div>
          {seg.top_domains && seg.top_domains.length > 0 && (
            <p className="text-xs text-muted-foreground">{seg.top_domains.join(', ')}</p>
          )}
        </div>
      ))}
    </div>
  );
}

// ─── Card ─────────────────────────────────────────────────────────────────────
export function AIInsightCard({ shortCode, insightType, title, description, icon }: AIInsightCardProps) {
  const [state, setState] = useState<'idle' | 'loading' | 'done' | 'error'>('idle');
  const [insight, setInsight] = useState<AIInsight | null>(null);
  const [errorMsg, setErrorMsg] = useState('');

  const fetchInsight = useCallback(async (refresh = false) => {
    setState('loading');
    setErrorMsg('');
    try {
      const token = localStorage.getItem('access_token');
      const url = API_ENDPOINTS.insights[insightType](shortCode) + (refresh ? '?refresh=1' : '');
      const res = await fetch(url, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ message: 'Request failed' }));
        throw new Error(err.message || `HTTP ${res.status}`);
      }
      const data: AIInsight = await res.json();
      setInsight(data);
      setState('done');
    } catch (e: any) {
      setErrorMsg(e.message || 'Unknown error');
      setState('error');
    }
  }, [shortCode, insightType]);

  useEffect(() => {
    if (shortCode) fetchInsight();
  }, [shortCode, fetchInsight]);

  function renderContent(insight: AIInsight) {
    switch (insightType) {
      case 'anomalies':
        return <AnomalyView content={insight.content} />;
      case 'audience':
        return <AudienceView content={insight.content} />;
      default:
        return (
          <div
            className="text-sm text-foreground leading-relaxed [&_ul]:mt-1 [&_ul]:mb-1 [&_li]:ml-3 [&_strong]:font-semibold [&_em]:italic [&_p]:mb-1"
            dangerouslySetInnerHTML={{ __html: parseMarkdown(insight.content) }}
          />
        );
    }
  }

  return (
    <div className="bg-card rounded-xl border border-border p-5 flex flex-col gap-3 min-h-[180px]">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-2">
          <span className="text-xl">{icon}</span>
          <div>
            <p className="font-semibold text-sm text-foreground">{title}</p>
            <p className="text-xs text-muted-foreground">{description}</p>
          </div>
        </div>
        {state === 'done' && (
          <button
            onClick={() => fetchInsight(true)}
            title="Refresh insight"
            className="text-xs text-muted-foreground hover:text-primary transition-colors px-2 py-1 rounded"
          >
            ↻ Refresh
          </button>
        )}
      </div>

      {/* Body */}
      {state === 'idle' && (
        <p className="text-xs text-muted-foreground italic">Select a URL to load insight.</p>
      )}
      {state === 'loading' && (
        <div className="flex items-center gap-2 flex-1">
          <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-primary" />
          <span className="text-xs text-muted-foreground">Generating insight…</span>
        </div>
      )}
      {state === 'error' && (
        <div className="flex flex-col gap-2 flex-1">
          <p className="text-xs text-red-500">{errorMsg}</p>
          <button
            onClick={() => fetchInsight()}
            className="self-start text-xs underline text-primary hover:opacity-80"
          >
            Retry
          </button>
        </div>
      )}
      {state === 'done' && insight && (
        <div className="flex flex-col gap-1 flex-1">
          {renderContent(insight)}
          <p className="text-xs text-muted-foreground mt-auto pt-2">
            Generated {new Date(insight.generated_at).toLocaleString()}
          </p>
        </div>
      )}
    </div>
  );
}

export default AIInsightCard;
