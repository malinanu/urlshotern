'use client';

import { useState, useEffect, useCallback } from 'react';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import Link from 'next/link';
import {
  GlobeAltIcon,
  PlusIcon,
  TrashIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
  ShieldCheckIcon,
  XMarkIcon,
  LockClosedIcon,
  SparklesIcon,
  ClipboardDocumentIcon,
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

export default function CustomDomainsPage() {
  const { user, getAccessToken } = useAuth();
  const [domains, setDomains] = useState<CustomDomain[]>([]);
  const [usage, setUsage] = useState<DomainUsage | null>(null);
  const [loading, setLoading] = useState(true);
  const [showAddModal, setShowAddModal] = useState(false);
  const [newDomain, setNewDomain] = useState('');
  const [addError, setAddError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  // Per-domain state
  const [verifyKeys, setVerifyKeys] = useState<Record<number, string>>({});
  const [verifyErrors, setVerifyErrors] = useState<Record<number, string>>({});
  const [verifyingId, setVerifyingId] = useState<number | null>(null);
  const [copied, setCopied] = useState<number | null>(null);

  const authHeaders = useCallback(
    () => ({ Authorization: `Bearer ${getAccessToken()}`, 'Content-Type': 'application/json' }),
    [getAccessToken]
  );

  const fetchVerificationKey = useCallback(
    async (domainId: number) => {
      try {
        const res = await fetch(API_ENDPOINTS.domains.verification(domainId), {
          headers: authHeaders(),
        });
        if (!res.ok) return;
        const data = await res.json();
        const key =
          data.verification_key ??
          data.key ??
          (data.verification_methods as { type: string; value: string }[] | undefined)?.find(
            (m) => m.type === 'dns'
          )?.value;
        if (key) {
          setVerifyKeys((prev) => ({ ...prev, [domainId]: key }));
        }
      } catch {
        // silently ignore
      }
    },
    [authHeaders]
  );

  const fetchDomains = useCallback(async () => {
    try {
      const [domainsRes, usageRes] = await Promise.all([
        fetch(API_ENDPOINTS.domains.list, { headers: authHeaders() }),
        fetch(API_ENDPOINTS.domains.usage, { headers: authHeaders() }),
      ]);
      if (domainsRes.ok) {
        const data = await domainsRes.json();
        const list: CustomDomain[] = data.domains ?? [];
        setDomains(list);
        // Pre-fetch verification keys for unverified domains
        list.filter((d) => !d.is_verified).forEach((d) => fetchVerificationKey(d.id));
      }
      if (usageRes.ok) {
        setUsage(await usageRes.json());
      }
    } finally {
      setLoading(false);
    }
  }, [authHeaders, fetchVerificationKey]);

  useEffect(() => {
    if (user) fetchDomains();
  }, [user, fetchDomains]);

  const handleAdd = async () => {
    setAddError('');
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
        setAddError(
          res.status === 403
            ? 'Custom domains require a Pro or Ultra plan. Please upgrade your plan.'
            : (data.message ?? 'Failed to add domain')
        );
        return;
      }
      setNewDomain('');
      setShowAddModal(false);
      await fetchDomains();
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Delete this domain?')) return;
    const res = await fetch(API_ENDPOINTS.domains.delete(id), {
      method: 'DELETE',
      headers: authHeaders(),
    });
    if (!res.ok) {
      const data = await res.json().catch(() => ({}));
      setVerifyErrors((prev) => ({
        ...prev,
        [id]: data.message ?? 'Failed to delete domain. Please try again.',
      }));
      return;
    }
    await fetchDomains();
  };

  const handleVerify = async (id: number) => {
    setVerifyErrors((prev) => ({ ...prev, [id]: '' }));
    setVerifyingId(id);
    try {
      const res = await fetch(API_ENDPOINTS.domains.verify(id), {
        method: 'POST',
        headers: authHeaders(),
      });
      const data = await res.json();
      if (!res.ok) {
        setVerifyErrors((prev) => ({
          ...prev,
          [id]: data.message ?? 'Verification failed. Please try again.',
        }));
      } else {
        await fetchDomains();
      }
    } finally {
      setVerifyingId(null);
    }
  };

  const copyToClipboard = (id: number, text: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopied(id);
      setTimeout(() => setCopied(null), 2000);
    });
  };

  const hasDomainAccess =
    user?.account_type === 'premium' || user?.account_type === 'enterprise';

  if (!user) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <p className="text-muted-foreground">Please log in to manage custom domains.</p>
        </div>
      </Layout>
    );
  }

  if (!hasDomainAccess) {
    return (
      <Layout>
        <div className="min-h-screen bg-background py-8">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center space-x-3 mb-8">
              <div className="p-2 bg-primary/10 rounded-lg">
                <GlobeAltIcon className="h-6 w-6 text-primary" />
              </div>
              <h1 className="text-2xl font-bold text-foreground">Custom Domains</h1>
            </div>
            <div className="bg-card border border-border rounded-xl p-12 text-center">
              <div className="mx-auto w-14 h-14 bg-primary/10 rounded-full flex items-center justify-center mb-4">
                <LockClosedIcon className="h-7 w-7 text-primary" />
              </div>
              <h2 className="text-xl font-semibold text-foreground mb-2">Pro Feature</h2>
              <p className="text-muted-foreground mb-2 max-w-md mx-auto">
                Custom domains let you use your own branded URL (e.g.{' '}
                <span className="font-mono text-foreground">links.yourcompany.com</span>) for all
                shortened links.
              </p>
              <p className="text-sm text-muted-foreground mb-6">
                Upgrade to <strong>Pro</strong> or <strong>Ultra</strong> to unlock this feature.
              </p>
              <Link
                href="/billing"
                className="inline-flex items-center space-x-2 px-6 py-2.5 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 font-medium transition-colors"
              >
                <SparklesIcon className="h-4 w-4" />
                <span>Upgrade Plan</span>
              </Link>
            </div>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">

          {/* Header */}
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
              onClick={() => { setShowAddModal(true); setAddError(''); setNewDomain(''); }}
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
              Custom domains let you use your own branded URL (e.g.{' '}
              <span className="font-mono text-foreground">links.yourcompany.com</span>) for all
              shortened links. DNS verification is required to prove ownership.
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
                onClick={() => { setShowAddModal(true); setAddError(''); setNewDomain(''); }}
                className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
              >
                Add Your First Domain
              </button>
            </div>
          ) : (
            <div className="space-y-4">
              {domains.map((d) => (
                <div
                  key={d.id}
                  className={`bg-card border rounded-xl overflow-hidden transition-colors ${
                    d.is_verified ? 'border-border' : 'border-yellow-500/40'
                  }`}
                >
                  {/* Domain header row */}
                  <div className="flex items-center justify-between px-5 py-4">
                    <div className="flex items-center space-x-3 min-w-0">
                      <div
                        className={`p-2 rounded-lg shrink-0 ${
                          d.is_verified ? 'bg-green-500/10' : 'bg-yellow-500/10'
                        }`}
                      >
                        {d.is_verified ? (
                          <CheckCircleIcon className="h-5 w-5 text-green-500" />
                        ) : (
                          <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />
                        )}
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-semibold text-foreground truncate">{d.domain}</span>
                          <span
                            className={`inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium shrink-0 ${
                              d.is_verified
                                ? 'bg-green-500/10 text-green-600 dark:text-green-400'
                                : 'bg-yellow-500/10 text-yellow-600 dark:text-yellow-400'
                            }`}
                          >
                            {d.is_verified ? 'Verified' : 'Pending verification'}
                          </span>
                          {d.ssl_enabled && (
                            <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-primary/10 text-primary shrink-0">
                              SSL
                            </span>
                          )}
                        </div>
                        {d.is_verified && (
                          <p className="text-xs text-muted-foreground mt-0.5">
                            DNS: {d.dns_status} · Health: {d.health_status}
                          </p>
                        )}
                      </div>
                    </div>
                    <button
                      onClick={() => handleDelete(d.id)}
                      className="flex items-center space-x-1.5 px-3 py-1.5 text-xs font-medium text-destructive border border-destructive/30 rounded-lg hover:bg-destructive/10 transition-colors shrink-0 ml-4"
                    >
                      <TrashIcon className="h-3.5 w-3.5" />
                      <span>Delete</span>
                    </button>
                  </div>

                  {/* Verification panel — shown only for unverified domains */}
                  {!d.is_verified && (
                    <div className="border-t border-yellow-500/20 bg-yellow-500/5 px-5 py-4">
                      <p className="text-sm font-medium text-foreground mb-3">
                        Add this DNS TXT record to verify ownership
                      </p>

                      {/* DNS record table */}
                      <div className="bg-background border border-border rounded-lg overflow-hidden mb-4">
                        <div className="grid grid-cols-[120px_1fr] divide-y divide-border text-sm">
                          <div className="px-3 py-2 bg-muted/50 text-xs font-medium text-muted-foreground flex items-center">
                            Record Type
                          </div>
                          <div className="px-3 py-2 font-mono text-foreground flex items-center">
                            TXT
                          </div>
                          <div className="px-3 py-2 bg-muted/50 text-xs font-medium text-muted-foreground flex items-center">
                            Name / Host
                          </div>
                          <div className="px-3 py-2 font-mono text-foreground flex items-center justify-between gap-2">
                            <span>_urlshortener-verify</span>
                            <button
                              onClick={() => copyToClipboard(-d.id, '_urlshortener-verify')}
                              className="text-muted-foreground hover:text-foreground transition-colors shrink-0"
                              title="Copy"
                            >
                              <ClipboardDocumentIcon className="h-4 w-4" />
                            </button>
                          </div>
                          <div className="px-3 py-2 bg-muted/50 text-xs font-medium text-muted-foreground flex items-center">
                            Value
                          </div>
                          <div className="px-3 py-2 font-mono text-foreground flex items-start justify-between gap-2">
                            {verifyKeys[d.id] ? (
                              <>
                                <span className="break-all text-xs">{verifyKeys[d.id]}</span>
                                <button
                                  onClick={() => copyToClipboard(d.id, verifyKeys[d.id])}
                                  className="text-muted-foreground hover:text-foreground transition-colors shrink-0 mt-0.5"
                                  title="Copy value"
                                >
                                  {copied === d.id ? (
                                    <CheckCircleIcon className="h-4 w-4 text-green-500" />
                                  ) : (
                                    <ClipboardDocumentIcon className="h-4 w-4" />
                                  )}
                                </button>
                              </>
                            ) : (
                              <span className="text-muted-foreground text-xs italic animate-pulse">
                                Loading verification token…
                              </span>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* Error message */}
                      {verifyErrors[d.id] && (
                        <div className="flex items-start gap-2.5 bg-destructive/10 border border-destructive/25 rounded-lg px-4 py-3 mb-4">
                          <ExclamationTriangleIcon className="h-4 w-4 text-destructive shrink-0 mt-0.5" />
                          <div>
                            <p className="text-xs font-medium text-destructive mb-0.5">
                              Verification failed
                            </p>
                            <p className="text-xs text-destructive/80">{verifyErrors[d.id]}</p>
                          </div>
                        </div>
                      )}

                      {/* Helper note + Check button */}
                      <div className="flex items-center justify-between gap-4">
                        <p className="text-xs text-muted-foreground">
                          DNS changes can take up to 48 h to propagate. Click{' '}
                          <strong>Check Verification</strong> once the record is live.
                        </p>
                        <button
                          onClick={() => handleVerify(d.id)}
                          disabled={verifyingId === d.id}
                          className="flex items-center space-x-1.5 px-4 py-2 text-xs font-semibold text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors shrink-0"
                        >
                          <ArrowPathIcon
                            className={`h-3.5 w-3.5 ${verifyingId === d.id ? 'animate-spin' : ''}`}
                          />
                          <span>{verifyingId === d.id ? 'Checking…' : 'Check Verification'}</span>
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Add Domain Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-background/80 backdrop-blur-sm flex items-center justify-center z-50">
          <div className="bg-card border border-border rounded-xl shadow-xl p-6 w-full max-w-md mx-4">
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-semibold text-foreground">Add Custom Domain</h2>
              <button
                onClick={() => { setShowAddModal(false); setAddError(''); setNewDomain(''); }}
                className="p-1 text-muted-foreground hover:text-foreground transition-colors rounded"
              >
                <XMarkIcon className="h-5 w-5" />
              </button>
            </div>
            <label className="block text-sm font-medium text-foreground mb-1">Domain</label>
            <input
              type="text"
              value={newDomain}
              onChange={(e) => setNewDomain(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && handleAdd()}
              placeholder="links.yourcompany.com"
              className="w-full border border-input bg-background text-foreground rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-primary/50"
              autoFocus
            />
            {addError && (
              <div className="flex items-start gap-2 mt-2">
                <ExclamationTriangleIcon className="h-4 w-4 text-destructive shrink-0 mt-0.5" />
                <p className="text-destructive text-sm">{addError}</p>
              </div>
            )}
            <p className="text-xs text-muted-foreground mt-3">
              After adding, you will see the DNS TXT record to add for verification.
            </p>
            <div className="flex gap-3 mt-5">
              <button
                onClick={() => { setShowAddModal(false); setAddError(''); setNewDomain(''); }}
                className="flex-1 px-4 py-2 text-sm text-foreground border border-border rounded-lg hover:bg-muted transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleAdd}
                disabled={submitting}
                className="flex-1 px-4 py-2 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors"
              >
                {submitting ? 'Adding…' : 'Add Domain'}
              </button>
            </div>
          </div>
        </div>
      )}
    </Layout>
  );
}
