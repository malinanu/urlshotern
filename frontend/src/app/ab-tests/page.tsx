'use client';

import { useState, useEffect, useCallback } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  BeakerIcon,
  PlusIcon,
  PlayIcon,
  StopIcon,
  ChartBarIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';

interface ABTest {
  id: string;
  name: string;
  description: string;
  short_code: string;
  status: 'draft' | 'running' | 'completed' | 'stopped';
  variant_a_url: string;
  variant_b_url: string;
  variant_a_clicks: number;
  variant_b_clicks: number;
  variant_a_conversions: number;
  variant_b_conversions: number;
  confidence_level: number;
  p_value: number;
  winner?: 'a' | 'b' | null;
  created_at: string;
  started_at?: string;
  stopped_at?: string;
}

interface CreateTestForm {
  name: string;
  description: string;
  short_code: string;
  variant_a_url: string;
  variant_b_url: string;
}

const emptyForm: CreateTestForm = {
  name: '',
  description: '',
  short_code: '',
  variant_a_url: '',
  variant_b_url: '',
};

function statusBadge(status: ABTest['status']) {
  const map: Record<ABTest['status'], string> = {
    draft: 'bg-muted text-muted-foreground',
    running: 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400',
    completed: 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400',
    stopped: 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400',
  };
  return map[status] ?? 'bg-muted text-muted-foreground';
}

export default function ABTestsPage() {
  const { isAuthenticated, isLoading, getAccessToken } = useAuth();
  const router = useRouter();
  const [tests, setTests] = useState<ABTest[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [form, setForm] = useState<CreateTestForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');
  const [selectedTest, setSelectedTest] = useState<ABTest | null>(null);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchTests = useCallback(async () => {
    const token = getAccessToken();
    try {
      const res = await fetch(API_ENDPOINTS.abTesting.tests, {
        headers: { Authorization: `Bearer ${token}` },
      });
      if (res.ok) {
        const data = await res.json();
        setTests(data.tests || data || []);
      }
    } catch {
      // silent — empty state shown
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    if (isAuthenticated) fetchTests();
  }, [isAuthenticated, fetchTests]);

  const createTest = async () => {
    setSaving(true);
    setError('');
    try {
      const token = getAccessToken();
      const res = await fetch(API_ENDPOINTS.abTesting.tests, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(form),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ message: 'Failed to create test' }));
        throw new Error(err.message || 'Failed to create test');
      }
      setShowCreateModal(false);
      setForm(emptyForm);
      fetchTests();
    } catch (e: any) {
      setError(e.message || 'Unknown error');
    } finally {
      setSaving(false);
    }
  };

  const startTest = async (id: string) => {
    const token = getAccessToken();
    await fetch(API_ENDPOINTS.abTesting.start(id), {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    fetchTests();
  };

  const stopTest = async (id: string) => {
    const token = getAccessToken();
    await fetch(API_ENDPOINTS.abTesting.stop(id), {
      method: 'POST',
      headers: { Authorization: `Bearer ${token}` },
    });
    fetchTests();
  };

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

  return (
    <Layout>
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Header */}
          <div className="flex items-center justify-between mb-8">
            <div>
              <div className="flex items-center gap-3 mb-2">
                <BeakerIcon className="h-8 w-8 text-primary" />
                <h1 className="text-3xl font-bold text-foreground">A/B Tests</h1>
              </div>
              <p className="text-muted-foreground">
                Split-test your short links to find the highest-converting destination URL.
              </p>
            </div>
            <button
              onClick={() => setShowCreateModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
            >
              <PlusIcon className="h-4 w-4" />
              New Test
            </button>
          </div>

          {/* Tests List */}
          {loading ? (
            <div className="space-y-4">
              {[1, 2, 3].map(i => (
                <div key={i} className="h-28 bg-muted rounded-lg animate-pulse" />
              ))}
            </div>
          ) : tests.length === 0 ? (
            <div className="bg-card rounded-lg border border-border p-12 text-center">
              <BeakerIcon className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-xl font-semibold text-foreground mb-2">No A/B Tests Yet</h3>
              <p className="text-muted-foreground mb-6">
                Create your first test to compare two destination URLs and find the winner.
              </p>
              <button
                onClick={() => setShowCreateModal(true)}
                className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
              >
                Create First Test
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {tests.map(test => (
                <div key={test.id} className="bg-card rounded-lg border border-border p-6">
                  <div className="flex items-start justify-between gap-4">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-3 mb-1">
                        <h3 className="font-semibold text-foreground truncate">{test.name}</h3>
                        <span className={`text-xs font-medium px-2 py-0.5 rounded-full capitalize ${statusBadge(test.status)}`}>
                          {test.status}
                        </span>
                      </div>
                      {test.description && (
                        <p className="text-sm text-muted-foreground mb-3">{test.description}</p>
                      )}
                      <div className="grid grid-cols-2 gap-4 text-sm">
                        <div className="space-y-1">
                          <p className="font-medium text-foreground">Variant A</p>
                          <p className="text-muted-foreground truncate">{test.variant_a_url}</p>
                          <p className="text-foreground">{test.variant_a_clicks} clicks · {test.variant_a_conversions} conversions</p>
                        </div>
                        <div className="space-y-1">
                          <p className="font-medium text-foreground">Variant B</p>
                          <p className="text-muted-foreground truncate">{test.variant_b_url}</p>
                          <p className="text-foreground">{test.variant_b_clicks} clicks · {test.variant_b_conversions} conversions</p>
                        </div>
                      </div>
                      {test.confidence_level > 0 && (
                        <div className="mt-3 flex items-center gap-4 text-sm">
                          <span className="text-muted-foreground">
                            Confidence: <span className="font-medium text-foreground">{test.confidence_level.toFixed(1)}%</span>
                          </span>
                          <span className="text-muted-foreground">
                            p-value: <span className="font-medium text-foreground">{test.p_value.toFixed(4)}</span>
                          </span>
                          {test.winner && (
                            <span className="font-medium text-green-600 dark:text-green-400">
                              Winner: Variant {test.winner.toUpperCase()}
                            </span>
                          )}
                        </div>
                      )}
                    </div>
                    <div className="flex items-center gap-2 shrink-0">
                      {test.status === 'draft' && (
                        <button
                          onClick={() => startTest(test.id)}
                          className="flex items-center gap-1 px-3 py-1.5 bg-green-600 text-white rounded-lg hover:bg-green-700 text-sm transition-colors"
                        >
                          <PlayIcon className="h-3.5 w-3.5" />
                          Start
                        </button>
                      )}
                      {test.status === 'running' && (
                        <button
                          onClick={() => stopTest(test.id)}
                          className="flex items-center gap-1 px-3 py-1.5 bg-red-600 text-white rounded-lg hover:bg-red-700 text-sm transition-colors"
                        >
                          <StopIcon className="h-3.5 w-3.5" />
                          Stop
                        </button>
                      )}
                      <button
                        onClick={() => setSelectedTest(test)}
                        className="flex items-center gap-1 px-3 py-1.5 border border-border rounded-lg hover:bg-muted text-sm transition-colors"
                      >
                        <ChartBarIcon className="h-3.5 w-3.5" />
                        Results
                      </button>
                    </div>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Create Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-card rounded-xl border border-border p-6 w-full max-w-lg mx-4 shadow-xl">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-foreground">Create A/B Test</h2>
              <button onClick={() => { setShowCreateModal(false); setForm(emptyForm); setError(''); }}>
                <XMarkIcon className="h-5 w-5 text-muted-foreground hover:text-foreground" />
              </button>
            </div>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Test Name</label>
                <input
                  type="text"
                  value={form.name}
                  onChange={e => setForm(f => ({ ...f, name: e.target.value }))}
                  placeholder="e.g. Homepage CTA Test"
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:ring-2 focus:ring-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Short Code</label>
                <input
                  type="text"
                  value={form.short_code}
                  onChange={e => setForm(f => ({ ...f, short_code: e.target.value }))}
                  placeholder="e.g. abc123"
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:ring-2 focus:ring-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Variant A URL</label>
                <input
                  type="url"
                  value={form.variant_a_url}
                  onChange={e => setForm(f => ({ ...f, variant_a_url: e.target.value }))}
                  placeholder="https://example.com/page-a"
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:ring-2 focus:ring-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Variant B URL</label>
                <input
                  type="url"
                  value={form.variant_b_url}
                  onChange={e => setForm(f => ({ ...f, variant_b_url: e.target.value }))}
                  placeholder="https://example.com/page-b"
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:ring-2 focus:ring-primary"
                />
              </div>
              <div>
                <label className="block text-sm font-medium text-foreground mb-1">Description (optional)</label>
                <textarea
                  value={form.description}
                  onChange={e => setForm(f => ({ ...f, description: e.target.value }))}
                  rows={2}
                  className="w-full px-3 py-2 border border-input rounded-md bg-background text-foreground focus:ring-2 focus:ring-primary"
                />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
              <div className="flex gap-3 pt-2">
                <button
                  onClick={createTest}
                  disabled={saving || !form.name || !form.short_code || !form.variant_a_url || !form.variant_b_url}
                  className="flex-1 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors"
                >
                  {saving ? 'Creating…' : 'Create Test'}
                </button>
                <button
                  onClick={() => { setShowCreateModal(false); setForm(emptyForm); setError(''); }}
                  className="flex-1 py-2 border border-border rounded-lg hover:bg-muted transition-colors"
                >
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Results Modal */}
      {selectedTest && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="bg-card rounded-xl border border-border p-6 w-full max-w-lg mx-4 shadow-xl">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-semibold text-foreground">Test Results: {selectedTest.name}</h2>
              <button onClick={() => setSelectedTest(null)}>
                <XMarkIcon className="h-5 w-5 text-muted-foreground hover:text-foreground" />
              </button>
            </div>
            <div className="space-y-4">
              <div className="grid grid-cols-2 gap-4">
                {(['a', 'b'] as const).map(variant => {
                  const clicks = variant === 'a' ? selectedTest.variant_a_clicks : selectedTest.variant_b_clicks;
                  const conversions = variant === 'a' ? selectedTest.variant_a_conversions : selectedTest.variant_b_conversions;
                  const url = variant === 'a' ? selectedTest.variant_a_url : selectedTest.variant_b_url;
                  const cvr = clicks > 0 ? ((conversions / clicks) * 100).toFixed(2) : '0.00';
                  const isWinner = selectedTest.winner === variant;
                  return (
                    <div key={variant} className={`p-4 rounded-lg border ${isWinner ? 'border-green-500 bg-green-50 dark:bg-green-900/20' : 'border-border'}`}>
                      <div className="flex items-center justify-between mb-2">
                        <p className="font-semibold text-foreground">Variant {variant.toUpperCase()}</p>
                        {isWinner && <span className="text-xs font-medium text-green-600 dark:text-green-400">Winner</span>}
                      </div>
                      <p className="text-xs text-muted-foreground mb-3 truncate">{url}</p>
                      <div className="space-y-1 text-sm">
                        <div className="flex justify-between"><span className="text-muted-foreground">Clicks</span><span className="font-medium">{clicks}</span></div>
                        <div className="flex justify-between"><span className="text-muted-foreground">Conversions</span><span className="font-medium">{conversions}</span></div>
                        <div className="flex justify-between"><span className="text-muted-foreground">CVR</span><span className="font-medium">{cvr}%</span></div>
                      </div>
                    </div>
                  );
                })}
              </div>
              <div className="bg-muted rounded-lg p-4 text-sm space-y-1">
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Statistical Confidence</span>
                  <span className={`font-medium ${selectedTest.confidence_level >= 95 ? 'text-green-600 dark:text-green-400' : 'text-foreground'}`}>
                    {selectedTest.confidence_level.toFixed(1)}%
                  </span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">p-value</span>
                  <span className="font-medium">{selectedTest.p_value.toFixed(4)}</span>
                </div>
                <div className="flex justify-between">
                  <span className="text-muted-foreground">Status</span>
                  <span className={`font-medium capitalize ${statusBadge(selectedTest.status)} px-2 rounded-full`}>{selectedTest.status}</span>
                </div>
              </div>
              {selectedTest.confidence_level < 95 && selectedTest.status === 'running' && (
                <p className="text-xs text-muted-foreground italic">
                  ≥ 95% confidence needed to declare a statistically significant winner. Keep the test running.
                </p>
              )}
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
