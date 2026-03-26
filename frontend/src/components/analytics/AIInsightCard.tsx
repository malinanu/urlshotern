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

/** Lightweight markdown → HTML converter (no external deps). */
function parseMarkdown(md: string): string {
  return md
    // Bold **text** or __text__
    .replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
    .replace(/__(.*?)__/g, '<strong>$1</strong>')
    // Italic *text* or _text_  (only standalone, not inside words)
    .replace(/\*([^*]+)\*/g, '<em>$1</em>')
    // Unordered list items: lines starting with - or *
    .replace(/^[\-\*] (.+)$/gm, '<li>$1</li>')
    // Numbered list items: lines starting with "1. "
    .replace(/^\d+\. (.+)$/gm, '<li>$1</li>')
    // Wrap consecutive <li> groups in <ul>
    .replace(/(<li>[\s\S]*?<\/li>\n?)+/g, (match) => `<ul class="list-disc list-inside space-y-1 my-1">${match}</ul>`)
    // Section headers ### / ##
    .replace(/^###\s(.+)$/gm, '<p class="font-semibold text-xs text-foreground mt-2 mb-0.5">$1</p>')
    .replace(/^##\s(.+)$/gm, '<p class="font-semibold text-sm text-foreground mt-3 mb-1">$1</p>')
    // Double newline → paragraph break
    .replace(/\n\n/g, '</p><p class="mt-2">')
    // Single newline → <br>
    .replace(/\n/g, '<br/>')
    // Wrap entire output in a paragraph
    .replace(/^([\s\S]+)$/, '<p>$1</p>');
}

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
          <div
            className="text-sm text-foreground leading-relaxed [&_ul]:mt-1 [&_ul]:mb-1 [&_li]:ml-4 [&_strong]:font-semibold [&_em]:italic [&_p]:mb-1"
            dangerouslySetInnerHTML={{ __html: parseMarkdown(insight.content) }}
          />
          <p className="text-xs text-muted-foreground mt-auto pt-2">
            Generated {new Date(insight.generated_at).toLocaleString()}
          </p>
        </div>
      )}
    </div>
  );
}

export default AIInsightCard;
