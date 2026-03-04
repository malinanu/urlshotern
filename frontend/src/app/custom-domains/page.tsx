'use client';

import { useState, useEffect, useCallback } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  GlobeAltIcon,
  PlusIcon,
  TrashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  ShieldCheckIcon,
} from '@heroicons/react/24/outline';

interface CustomDomain {
  id: number;
  domain: string;
  is_active: boolean;
  is_verified: boolean;
  verification_key?: string;
  ssl_enabled: boolean;
  ssl_status: string;
  dns_status: string;
  health_status: string;
  created_at: string;
}

interface DomainUsage {
  current: number;
  limit: number;
  account_type: string;
}

type ModalType = 'add' | 'verify' | null;

export default function CustomDomainsPage() {
  const { user, token } = useAuth();
  const [domains, setDomains] = useState<CustomDomain[]>([]);
  const [usage, setUsage] = useState<DomainUsage | null>(null);
  const [loading, setLoading] = useState(true);
  const [modal, setModal] = useState<ModalType>(null);
  const [selectedDomain, setSelectedDomain] = useState<CustomDomain | null>(null);
  const [newDomain, setNewDomain] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const authHeaders = useCallback(
    () => ({ Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' }),
    [token]
  );

  const fetchDomains = useCallback(async () => {
    try {
      const [domainsRes, usageRes] = await Promise.all([
        fetch(API_ENDPOINTS.domains.list, { headers: authHeaders() }),
        fetch(API_ENDPOINTS.domains.usage, { headers: authHeaders() }),
      ]);
      if (domainsRes.ok) {
        const data = await domainsRes.json();
        setDomains(data.domains ?? []);
      }
      if (usageRes.ok) {
        setUsage(await usageRes.json());
      }
    } finally {
      setLoading(false);
    }
  }, [authHeaders]);

  useEffect(() => {
    if (token) fetchDomains();
  }, [token, fetchDomains]);

  const handleAdd = async () => {
    setError('');
    if (!newDomain.trim()) return;
    setSubmitting(true);
    try {
      const res = await fetch(API_ENDPOINTS.domains.create, {
        method: 'POST',
        headers: authHeaders(),
        body: JSON.stringify({ domain: newDomain.trim() }),
      });
      const data = await res.json();
      if (!res.ok) {
        setError(data.message ?? 'Failed to add domain');
        return;
      }
      setNewDomain('');
      setModal(null);
      await fetchDomains();
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this domain?')) return;
    const res = await fetch(API_ENDPOINTS.domains.delete(id), { method: 'DELETE', headers: authHeaders() });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      alert(data.message ?? 'Failed to delete domain. Please try again.');
      return;
    }
    await fetchDomains();
  };

  const handleVerify = async (id: number) => {
    const res = await fetch(API_ENDPOINTS.domains.verify(id), {
      method: 'POST',
      headers: authHeaders(),
    });
    const data = await res.json();
    if (!res.ok) {
      alert(data.message ?? 'Verification failed');
    } else {
      alert('Domain verified!');
      setModal(null);
      await fetchDomains();
    }
  };

  const openVerifyModal = async (domain: CustomDomain) => {
    // Fetch full domain details (including verification_key) from the dedicated endpoint
    try {
      const res = await fetch(API_ENDPOINTS.domains.verification(domain.id), { headers: authHeaders() });
      if (res.ok) {
        const data = await res.json();
        // Merge the verification key into the selected domain object
        setSelectedDomain({ ...domain, verification_key: data.verification_key ?? data.key });
      } else {
        setSelectedDomain(domain);
      }
    } catch {
      setSelectedDomain(domain);
    }
    setModal('verify');
  };

  if (!user) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <p className="text-muted-foreground">Please log in to manage custom domains.</p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">

          {/* Page Header */}
          <div className="flex items-center justify-between mb-8">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <GlobeAltIcon className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-foreground">Custom Domains</h1>
                {usage && (
                  <p className="text-sm text-muted-foreground mt-0.5">
                    {usage.current} of {usage.limit} domains used
                  </p>
                )}
              </div>
            </div>
            <button
              onClick={() => setModal('add')}
              className="flex items-center space-x-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
            >
              <PlusIcon className="h-4 w-4" />
              <span>Add Domain</span>
            </button>
          </div>

          {/* Info Banner */}
          <div className="bg-card border border-border rounded-lg p-4 mb-6 flex items-start space-x-3">
            <ShieldCheckIcon className="h-5 w-5 text-primary shrink-0 mt-0.5" />
            <p className="text-sm text-muted-foreground">
              Custom domains let you use your own branded URL (e.g. <span className="font-mono text-foreground">links.yourcompany.com</span>) for all shortened links. DNS verification is required to prove ownership.
            </p>
          </div>

          {/* Domain List */}
          {loading ? (
            <div className="space-y-3">
              {[1, 2, 3].map((i) => (
                <div key={i} className="bg-card border border-border rounded-lg h-20 animate-pulse" />
              ))}
            </div>
          ) : domains.length === 0 ? (
            <div className="bg-card border border-border rounded-lg p-12 text-center">
              <GlobeAltIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
              <h3 className="text-lg font-medium text-foreground mb-1">No custom domains yet</h3>
              <p className="text-sm text-muted-foreground mb-4">
                Add a domain to start using your own branded short links.
              </p>
              <button
                onClick={() => setModal('add')}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
              >
                Add Your First Domain
              </button>
            </div>
          ) : (
            <div className="space-y-3">
              {domains.map((d) => (
                <div
                  key={d.id}
                  className="bg-card border border-border rounded-lg p-5 flex items-center justify-between"
                >
                  <div className="flex items-center space-x-4">
                    <div className={`p-2 rounded-lg ${d.is_verified ? 'bg-green-500/10' : 'bg-yellow-500/10'}`}>
                      {d.is_verified ? (
                        <CheckCircleIcon className="h-5 w-5 text-green-500" />
                      ) : (
                        <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />
                      )}
                    </div>
                    <div>
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-foreground">{d.domain}</span>
                        <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium ${
                          d.is_verified
                            ? 'bg-green-500/10 text-green-600'
                            : 'bg-yellow-500/10 text-yellow-600'
                        }`}>
                          {d.is_verified ? 'Verified' : 'Pending'}
                        </span>
                        {d.ssl_enabled && (
                          <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-primary/10 text-primary">
                            SSL
                          </span>
                        )}
                      </div>
                      <p className="text-xs text-muted-foreground mt-1">
                        DNS: {d.dns_status} · Health: {d.health_status}
                      </p>
                    </div>
                  </div>
                  <div className="flex items-center space-x-2">
                    {!d.is_verified && (
                      <button
                        onClick={() => openVerifyModal(d)}
                        className="flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium border border-border text-foreground rounded-lg hover:bg-muted transition-colors"
                      >
                        <ArrowPathIcon className="h-3.5 w-3.5" />
                        <span>Verify</span>
                      </button>
                    )}
                    <button
                      onClick={() => handleDelete(d.id)}
                      className="flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium text-destructive border border-destructive/30 rounded-lg hover:bg-destructive/10 transition-colors"
                    >
                      <TrashIcon className="h-3.5 w-3.5" />
                      <span>Delete</span>
                    </button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Add Domain Modal */}
      {modal === 'add' && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-xl p-6 w-full max-w-md mx-4">
            <h2 className="text-lg font-semibold text-foreground mb-4">Add Custom Domain</h2>
            <label className="block text-sm font-medium text-foreground mb-1">Domain</label>
            <input
              type="text"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
              placeholder="links.yourcompany.com"
              className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
            />
            {error && (
              <p className="text-destructive text-sm mt-2 flex items-center space-x-1">
                <ExclamationTriangleIcon className="h-4 w-4" />
                <span>{error}</span>
              </p>
            )}
            <p className="text-xs text-muted-foreground mt-3">
              After adding, you will need to add a DNS TXT record to verify domain ownership.
            </p>
            <div className="flex gap-3 mt-5">
              <button
                onClick={() => { setModal(null); setError(''); setNewDomain(''); }}
                className="flex-1 px-4 py-2 text-sm text-foreground border border-border rounded-lg hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleAdd}
                disabled={submitting}
                className="flex-1 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors"
              >
                {submitting ? 'Adding...' : 'Add Domain'}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Verify Domain Modal */}
      {modal === 'verify' && selectedDomain && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-lg shadow-xl p-6 w-full max-w-lg mx-4">
            <h2 className="text-lg font-semibold text-foreground mb-2">Verify Domain Ownership</h2>
            <p className="text-sm text-muted-foreground mb-4">
              Add this TXT record to your DNS for{' '}
              <span className="font-mono font-semibold text-foreground">{selectedDomain.domain}</span>:
            </p>
            <div className="bg-muted rounded-lg p-4 space-y-3 mb-4 text-sm">
              <div>
                <p className="text-xs text-muted-foreground mb-0.5">Record Type</p>
                <p className="font-mono text-foreground">TXT</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-0.5">Name / Host</p>
                <p className="font-mono text-foreground">_urlshortener-verify</p>
              </div>
              <div>
                <p className="text-xs text-muted-foreground mb-0.5">Value</p>
                <p className="font-mono text-foreground break-all">
                  {selectedDomain.verification_key ?? '(key shown when domain is loaded)'}
                </p>
              </div>
            </div>
            <p className="text-xs text-muted-foreground mb-5">
              DNS changes can take up to 48 hours to propagate. Click &quot;Check Now&quot; once the record is set.
            </p>
            <div className="flex gap-3">
              <button
                onClick={() => setModal(null)}
                className="flex-1 px-4 py-2 text-sm text-foreground border border-border rounded-lg hover:bg-muted transition-colors"
              >
                Close
              </button>
              <button
                onClick={() => handleVerify(selectedDomain.id)}
                className="flex-1 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 transition-colors"
              >
                Check Now
              </button>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
