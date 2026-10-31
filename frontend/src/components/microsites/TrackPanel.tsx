'use client';

import { useEffect, useState } from 'react';
import { CartesianGrid, Legend, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { apiFetch, isApiError, type MicrositeAnalytics, type MicrositeCount } from '@/lib/microsites';

const RANGES = [7, 30, 90];

const SOURCE_LABELS: Record<string, string> = {
  qr: 'QR code',
  social: 'Social',
  search: 'Search',
  email: 'Email',
  direct: 'Direct / link',
  other: 'Other sites',
};

const EVENT_LABELS: Record<string, string> = {
  click: 'Link',
  call: 'Call',
  sms: 'Text',
  email: 'Email',
  whatsapp: 'WhatsApp',
  vcard_download: 'Save contact',
};

function formatDay(d: string) {
  const [, m, day] = d.split('-');
  return `${Number(m)}/${Number(day)}`;
}

function StatTile({ label, value, hint }: { label: string; value: string; hint?: string }) {
  return (
    <div className="rounded-xl border border-border bg-card px-4 py-3">
      <p className="text-xs font-medium text-muted-foreground">{label}</p>
      <p className="mt-1 text-2xl font-semibold text-foreground tabular-nums">{value}</p>
      {hint && <p className="text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

function BarList({ title, rows, labelFor }: { title: string; rows: MicrositeCount[]; labelFor?: (l: string) => string }) {
  const max = Math.max(1, ...rows.map((r) => r.count));
  return (
    <section className="rounded-xl border border-border bg-card p-4">
      <h3 className="text-sm font-semibold text-foreground mb-3">{title}</h3>
      {rows.length === 0 ? (
        <p className="text-sm text-muted-foreground">No data yet.</p>
      ) : (
        <ul className="grid gap-2">
          {rows.map((r) => (
            <li key={r.label} className="grid gap-1" title={`${labelFor ? labelFor(r.label) : r.label}: ${r.count.toLocaleString()}`}>
              <div className="flex justify-between gap-2 text-sm">
                <span className="text-foreground truncate">{labelFor ? labelFor(r.label) : r.label}</span>
                <span className="text-muted-foreground tabular-nums">{r.count.toLocaleString()}</span>
              </div>
              <div className="h-1.5 rounded-full bg-muted overflow-hidden">
                <div className="h-full rounded-full viz-bar" style={{ width: `${(r.count / max) * 100}%` }} />
              </div>
            </li>
          ))}
        </ul>
      )}
    </section>
  );
}

/** Track tab: page views, unique visitors, taps and contact saves. */
export default function TrackPanel({ micrositeId, maxDays }: { micrositeId: number; maxDays?: number }) {
  const { getAccessToken } = useAuth();
  const [days, setDays] = useState(30);
  const [data, setData] = useState<MicrositeAnalytics | null>(null);
  const [error, setError] = useState('');
  const [showTable, setShowTable] = useState(false);

  useEffect(() => {
    let cancelled = false;
    setError('');
    apiFetch<MicrositeAnalytics>(`${API_ENDPOINTS.microsites.analytics(micrositeId)}?days=${days}`, getAccessToken())
      .then((d) => !cancelled && setData(d))
      .catch((e) => !cancelled && setError(isApiError(e) ? e.message : 'Could not load analytics.'));
    return () => {
      cancelled = true;
    };
  }, [micrositeId, days, getAccessToken]);

  const ranges = RANGES.filter((r) => !maxDays || maxDays < 0 || r <= maxDays || r === RANGES[0]);
  const contactTaps = data ? Object.values(data.contact_actions ?? {}).reduce((a, b) => a + b, 0) : 0;

  return (
    <div className="viz-root grid gap-6">
      <style>{`
        .viz-root { --series-1: #2a78d6; --series-2: #eb6834; --grid: #e5e7eb; --axis: #6b7280; }
        .dark .viz-root { --series-1: #3987e5; --series-2: #d95926; --grid: #2a2f3a; --axis: #9ca3af; }
        .viz-root .viz-bar { background: var(--series-1); }
      `}</style>

      <div className="flex flex-wrap items-center justify-between gap-3">
        <h2 className="text-lg font-semibold text-foreground">Performance</h2>
        <div className="inline-flex rounded-lg border border-border p-0.5 bg-muted/50" role="group" aria-label="Date range">
          {ranges.map((r) => (
            <button
              key={r}
              type="button"
              onClick={() => setDays(r)}
              aria-pressed={days === r}
              className={`px-3 py-1 text-xs font-medium rounded-md ${days === r ? 'bg-card text-foreground shadow-sm' : 'text-muted-foreground hover:text-foreground'}`}
            >
              {r} days
            </button>
          ))}
        </div>
      </div>

      {error && <p role="alert" className="text-sm text-red-600">{error}</p>}

      {data && (
        <>
          <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-6 gap-3">
            <StatTile label="Views" value={data.views.toLocaleString()} />
            <StatTile label="Unique visitors" value={data.unique_visitors.toLocaleString()} hint="Daily unique" />
            <StatTile label="Link taps" value={data.link_clicks.toLocaleString()} />
            <StatTile label="Contact taps" value={contactTaps.toLocaleString()} hint="Call, text, email, WhatsApp" />
            <StatTile label="Contacts saved" value={data.vcard_downloads.toLocaleString()} hint={`${data.contact_save_rate}% save rate`} />
            <StatTile label="Click-through rate" value={`${data.ctr}%`} hint="Taps ÷ views" />
          </div>

          <section className="rounded-xl border border-border bg-card p-4">
            <div className="flex items-center justify-between mb-3">
              <h3 className="text-sm font-semibold text-foreground">Views and unique visitors</h3>
              <button type="button" onClick={() => setShowTable((v) => !v)} className="text-xs text-primary hover:underline">
                {showTable ? 'Show chart' : 'Show table'}
              </button>
            </div>
            {showTable ? (
              <div className="max-h-72 overflow-auto">
                <table className="w-full text-sm">
                  <thead className="text-xs text-muted-foreground text-left">
                    <tr>
                      <th className="py-1 font-medium">Date</th>
                      <th className="py-1 font-medium text-right">Views</th>
                      <th className="py-1 font-medium text-right">Unique</th>
                      <th className="py-1 font-medium text-right">Taps</th>
                      <th className="py-1 font-medium text-right">Saves</th>
                    </tr>
                  </thead>
                  <tbody className="tabular-nums text-foreground">
                    {[...(data.daily ?? [])].reverse().map((d) => (
                      <tr key={d.date} className="border-t border-border">
                        <td className="py-1">{d.date}</td>
                        <td className="py-1 text-right">{d.views}</td>
                        <td className="py-1 text-right">{d.unique_visitors}</td>
                        <td className="py-1 text-right">{d.clicks}</td>
                        <td className="py-1 text-right">{d.vcard_downloads}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="h-64">
                <ResponsiveContainer width="100%" height="100%">
                  <LineChart data={data.daily ?? []} margin={{ top: 8, right: 12, left: -16, bottom: 0 }}>
                    <CartesianGrid stroke="var(--grid)" vertical={false} />
                    <XAxis dataKey="date" tickFormatter={formatDay} tick={{ fontSize: 11, fill: 'var(--axis)' }} axisLine={false} tickLine={false} minTickGap={16} />
                    <YAxis allowDecimals={false} tick={{ fontSize: 11, fill: 'var(--axis)' }} axisLine={false} tickLine={false} />
                    <Tooltip
                      cursor={{ stroke: 'var(--axis)', strokeWidth: 1 }}
                      contentStyle={{ borderRadius: 8, fontSize: 12 }}
                      labelFormatter={(l) => String(l)}
                    />
                    <Legend iconType="plainline" wrapperStyle={{ fontSize: 12 }} />
                    <Line type="monotone" dataKey="views" name="Views" stroke="var(--series-1)" strokeWidth={2} dot={false} activeDot={{ r: 4 }} />
                    <Line type="monotone" dataKey="unique_visitors" name="Unique visitors" stroke="var(--series-2)" strokeWidth={2} dot={false} activeDot={{ r: 4 }} />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            )}
          </section>

          <div className="grid md:grid-cols-2 gap-4">
            <BarList
              title="Taps by link and button"
              rows={(data.blocks ?? []).map((b) => ({ label: `${b.label} · ${EVENT_LABELS[b.event_type] ?? b.event_type}`, count: b.count }))}
            />
            <BarList title="Traffic sources" rows={data.referrer_sources ?? []} labelFor={(l) => SOURCE_LABELS[l] ?? l} />
            <BarList title="Top referrers" rows={data.top_referrers ?? []} />
            <BarList title="Devices" rows={data.devices ?? []} labelFor={(l) => l.charAt(0).toUpperCase() + l.slice(1)} />
            <BarList title="Operating systems" rows={data.operating_systems ?? []} />
            <BarList title="Countries" rows={data.countries ?? []} labelFor={(l) => (l === '--' ? 'Unknown' : l)} />
          </div>
        </>
      )}
      {!data && !error && <div className="h-64 rounded-xl border border-border bg-card animate-pulse" />}
    </div>
  );
}
