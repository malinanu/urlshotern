'use client';

import React, { useState, useEffect } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { getAccessToken } from '@/utils/tokenValidation';
import { MicroPage, MicroPageTheme } from '@/types/micropage';
import { TEMPLATES } from '@/components/micropages/templates';
import { QRCodeModal } from '@/components/QRCode';
import {
  PlusIcon,
  EyeIcon,
  PencilSquareIcon,
  TrashIcon,
  QrCodeIcon,
  ArrowTopRightOnSquareIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  SparklesIcon,
  ArrowDownTrayIcon,
  GlobeAltIcon,
} from '@heroicons/react/24/outline';

export default function MicroPagesDashboard() {
  const { user, isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [pages, setPages] = useState<MicroPage[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Creation modal state
  const [isCreateOpen, setIsCreateOpen] = useState(false);
  const [newTitle, setNewTitle] = useState('');
  const [newSlug, setNewSlug] = useState('');
  const [selectedTheme, setSelectedTheme] = useState<MicroPageTheme>('executive-obsidian');
  const [creating, setCreating] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);

  // QR Modal state
  const [qrUrl, setQrUrl] = useState<string | null>(null);
  const [copiedSlug, setCopiedSlug] = useState<string | null>(null);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push('/login?redirect=/micropages');
    }
  }, [authLoading, isAuthenticated, router]);

  const fetchPages = async () => {
    try {
      setLoading(true);
      const token = getAccessToken();
      const res = await fetch(API_ENDPOINTS.micropages.list, {
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (!res.ok) {
        throw new Error('Failed to load micro-pages');
      }

      const data = await res.json();
      setPages(data.pages || []);
    } catch (err: any) {
      setError(err.message || 'Error loading micro-pages');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (isAuthenticated) {
      fetchPages();
    }
  }, [isAuthenticated]);

  const handleTitleChange = (val: string) => {
    setNewTitle(val);
    const slugified = val
      .toLowerCase()
      .trim()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
    setNewSlug(slugified);
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSlug.trim() || !newTitle.trim()) {
      setCreateError('Please provide both a title and slug');
      return;
    }

    try {
      setCreating(true);
      setCreateError(null);
      const token = getAccessToken();

      const payload = {
        title: newTitle.trim(),
        slug: newSlug.trim().toLowerCase(),
        theme: selectedTheme,
        bio: `Hi, I'm ${newTitle.trim()}. Connect with me below!`,
        vcard: {
          first_name: newTitle.trim().split(' ')[0] || newTitle.trim(),
          last_name: newTitle.trim().split(' ').slice(1).join(' ') || '',
          organization: '',
          title: '',
          email: user?.email || '',
          phone: user?.phone || '',
        },
        social_links: [
          { platform: 'email', url: user?.email ? `mailto:${user.email}` : '' },
        ],
        links: [
          {
            id: 'link-1',
            title: 'Schedule a Meeting',
            url: 'https://calendly.com',
            icon: '📅',
            description: 'Book a 15-minute quick chat',
            is_highlighted: true,
            is_active: true,
            sort_order: 0,
          },
        ],
        is_published: true,
      };

      const res = await fetch(API_ENDPOINTS.micropages.create, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || 'Failed to create micro-page');
      }

      const created = await res.json();
      setIsCreateOpen(false);
      router.push(`/micropages/${created.id}/builder`);
    } catch (err: any) {
      setCreateError(err.message || 'Error creating micro-page');
    } finally {
      setCreating(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm('Are you sure you want to delete this micro-page? This action cannot be undone.')) {
      return;
    }

    try {
      const token = getAccessToken();
      const res = await fetch(API_ENDPOINTS.micropages.delete(id), {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (!res.ok) {
        throw new Error('Failed to delete micro-page');
      }

      setPages((prev) => prev.filter((p) => p.id !== id));
    } catch (err: any) {
      alert(err.message || 'Failed to delete micro-page');
    }
  };

  const copyToClipboard = (slug: string) => {
    const fullUrl = `${window.location.origin}/m/${slug}`;
    navigator.clipboard.writeText(fullUrl);
    setCopiedSlug(slug);
    setTimeout(() => setCopiedSlug(null), 2500);
  };

  const totalViews = pages.reduce((acc, p) => acc + (p.view_count || 0), 0);
  const totalLinks = pages.reduce((acc, p) => acc + (p.links ? p.links.length : 0), 0);

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8 space-y-8">
        {/* Top Header & CTA */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-6 border-b border-slate-800">
          <div>
            <div className="inline-flex items-center gap-2 px-3 py-1 rounded-full text-xs font-semibold bg-blue-500/10 text-blue-400 border border-blue-500/20 mb-2">
              <SparklesIcon className="w-4 h-4" />
              <span>Digital vCard & Bio Link Engine</span>
            </div>
            <h1 className="text-3xl font-extrabold text-white tracking-tight">Micro-Landing Pages</h1>
            <p className="mt-1 text-sm text-slate-400">
              Create contact cards, bio link trees, and digital landing pages with one-click phone contact downloads.
            </p>
          </div>
          <div>
            <button
              onClick={() => {
                setNewTitle('');
                setNewSlug('');
                setCreateError(null);
                setIsCreateOpen(true);
              }}
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-xl font-semibold text-white bg-blue-600 hover:bg-blue-500 shadow-lg shadow-blue-500/25 transition-all active:scale-95"
            >
              <PlusIcon className="w-5 h-5" />
              <span>Create Micro-Page</span>
            </button>
          </div>
        </div>

        {/* Stats Row */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-5">
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-5">
            <div className="text-xs font-medium text-slate-400 uppercase tracking-wider">Active Pages</div>
            <div className="mt-2 text-3xl font-extrabold text-white">{pages.length}</div>
            <div className="mt-1 text-xs text-slate-500">Curated digital vCards</div>
          </div>
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-5">
            <div className="text-xs font-medium text-slate-400 uppercase tracking-wider">Total Views</div>
            <div className="mt-2 text-3xl font-extrabold text-blue-400">{totalViews.toLocaleString()}</div>
            <div className="mt-1 text-xs text-slate-500">Impressions across all pages</div>
          </div>
          <div className="bg-slate-900/60 border border-slate-800 rounded-2xl p-5">
            <div className="text-xs font-medium text-slate-400 uppercase tracking-wider">Interactive Links</div>
            <div className="mt-2 text-3xl font-extrabold text-emerald-400">{totalLinks}</div>
            <div className="mt-1 text-xs text-slate-500">Action buttons & social endpoints</div>
          </div>
        </div>

        {/* Loading / Error States */}
        {loading ? (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 animate-pulse">
            {[1, 2, 3].map((i) => (
              <div key={i} className="h-64 bg-slate-900/60 border border-slate-800 rounded-2xl" />
            ))}
          </div>
        ) : error ? (
          <div className="bg-rose-500/10 border border-rose-500/30 rounded-2xl p-6 text-center text-rose-400">
            {error}
          </div>
        ) : pages.length === 0 ? (
          /* Empty State */
          <div className="bg-slate-900/40 border border-slate-800/80 rounded-3xl p-12 text-center max-w-xl mx-auto">
            <div className="w-16 h-16 bg-blue-500/10 border border-blue-500/20 text-blue-400 rounded-2xl flex items-center justify-center mx-auto mb-4">
              <SparklesIcon className="w-8 h-8" />
            </div>
            <h3 className="text-xl font-bold text-white">No Micro-Pages Yet</h3>
            <p className="mt-2 text-sm text-slate-400 leading-relaxed">
              Design your first digital business card with one of 10 handcrafted templates. Share it on social bios, NFC badges, or email signatures.
            </p>
            <button
              onClick={() => setIsCreateOpen(true)}
              className="mt-6 inline-flex items-center gap-2 px-6 py-3 rounded-xl font-semibold text-white bg-blue-600 hover:bg-blue-500 shadow-lg shadow-blue-500/25 transition-all"
            >
              <PlusIcon className="w-5 h-5" />
              <span>Build Your First vCard</span>
            </button>
          </div>
        ) : (
          /* Cards Grid */
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
            {pages.map((p) => {
              const templateInfo = TEMPLATES.find((t) => t.id === p.theme) || TEMPLATES[0];
              const publicUrl = `/m/${p.slug}`;

              return (
                <div
                  key={p.id}
                  className="bg-slate-900/80 border border-slate-800 hover:border-slate-700 rounded-2xl overflow-hidden flex flex-col justify-between transition-all group"
                >
                  <div>
                    {/* Top Preview Banner */}
                    <div className="h-28 w-full relative overflow-hidden bg-slate-950 flex items-center justify-center">
                      {p.cover_url ? (
                        <img src={p.cover_url} alt="Cover" className="w-full h-full object-cover" />
                      ) : (
                        <div className={`w-full h-full opacity-60 ${templateInfo.previewBg}`} />
                      )}
                      <span className="absolute top-3 right-3 text-xs font-semibold px-2.5 py-1 rounded-full bg-black/60 backdrop-blur-md text-white border border-white/10">
                        {templateInfo.name}
                      </span>
                    </div>

                    {/* Avatar & Title */}
                    <div className="px-5 pt-3 pb-4 relative">
                      <div className="-mt-12 mb-3">
                        {p.avatar_url ? (
                          <img
                            src={p.avatar_url}
                            alt={p.title}
                            className="w-16 h-16 rounded-full object-cover border-2 border-slate-900 shadow-md"
                          />
                        ) : (
                          <div className="w-16 h-16 rounded-full bg-blue-600 text-white font-bold text-xl flex items-center justify-center border-2 border-slate-900 shadow-md">
                            {(p.title || 'V')[0]}
                          </div>
                        )}
                      </div>

                      <div className="flex items-start justify-between gap-2">
                        <div>
                          <h3 className="font-bold text-lg text-white group-hover:text-blue-400 transition-colors">
                            {p.title}
                          </h3>
                          <div className="text-xs font-mono text-slate-400 mt-0.5">/m/{p.slug}</div>
                        </div>
                        <span
                          className={`text-xs px-2.5 py-0.5 rounded-full font-medium ${
                            p.is_published
                              ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20'
                              : 'bg-amber-500/10 text-amber-400 border border-amber-500/20'
                          }`}
                        >
                          {p.is_published ? 'Published' : 'Draft'}
                        </span>
                      </div>

                      {p.bio && (
                        <p className="mt-2 text-xs text-slate-400 line-clamp-2 leading-relaxed">
                          {p.bio}
                        </p>
                      )}

                      {/* Micro stats */}
                      <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs text-slate-400">
                        <span className="flex items-center gap-1.5">
                          <EyeIcon className="w-4 h-4 text-blue-400" />
                          <span>{p.view_count || 0} views</span>
                        </span>
                        <span>{p.links?.length || 0} links</span>
                      </div>
                    </div>
                  </div>

                  {/* Actions Footer */}
                  <div className="p-3 bg-slate-950/60 border-t border-slate-800/80 flex items-center justify-between gap-1">
                    <Link
                      href={`/micropages/${p.id}/builder`}
                      className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold text-white bg-blue-600 hover:bg-blue-500 transition-colors"
                    >
                      <PencilSquareIcon className="w-3.5 h-3.5" />
                      <span>Edit</span>
                    </Link>

                    <div className="flex items-center gap-1">
                      <a
                        href={publicUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        title="View Public Page"
                        className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                      >
                        <ArrowTopRightOnSquareIcon className="w-4 h-4" />
                      </a>

                      <button
                        type="button"
                        onClick={() => copyToClipboard(p.slug)}
                        title="Copy Public Link"
                        className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                      >
                        {copiedSlug === p.slug ? (
                          <CheckIcon className="w-4 h-4 text-emerald-400" />
                        ) : (
                          <ClipboardDocumentIcon className="w-4 h-4" />
                        )}
                      </button>

                      <button
                        type="button"
                        onClick={() => setQrUrl(`${window.location.origin}/m/${p.slug}`)}
                        title="QR Code"
                        className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                      >
                        <QrCodeIcon className="w-4 h-4" />
                      </button>

                      <a
                        href={API_ENDPOINTS.micropages.vcard(p.slug)}
                        download
                        title="Download .vcf"
                        className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition-colors"
                      >
                        <ArrowDownTrayIcon className="w-4 h-4" />
                      </a>

                      <button
                        type="button"
                        onClick={() => handleDelete(p.id)}
                        title="Delete Page"
                        className="p-1.5 rounded-lg text-slate-400 hover:text-rose-400 hover:bg-slate-800 transition-colors"
                      >
                        <TrashIcon className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        )}

        {/* Modal: Create Micro-Page */}
        {isCreateOpen && (
          <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/75 backdrop-blur-sm">
            <div className="bg-slate-900 border border-slate-800 rounded-3xl max-w-xl w-full p-6 sm:p-8 shadow-2xl relative">
              <div className="flex items-center justify-between pb-4 border-b border-slate-800">
                <div>
                  <h3 className="text-xl font-bold text-white">New Micro-Page & Digital vCard</h3>
                  <p className="text-xs text-slate-400 mt-0.5">Select a template and choose your custom URL slug</p>
                </div>
                <button
                  onClick={() => setIsCreateOpen(false)}
                  className="text-slate-400 hover:text-white text-xl p-1"
                >
                  ✕
                </button>
              </div>

              {createError && (
                <div className="mt-4 p-3 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs text-rose-400">
                  {createError}
                </div>
              )}

              <form onSubmit={handleCreate} className="mt-5 space-y-4">
                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    Your Name / Display Title
                  </label>
                  <input
                    type="text"
                    required
                    value={newTitle}
                    onChange={(e) => handleTitleChange(e.target.value)}
                    placeholder="e.g. Alex Carter"
                    className="w-full bg-slate-800/80 border border-slate-700 focus:border-blue-500 focus:outline-none rounded-xl px-4 py-2.5 text-white text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    Custom Slug URL
                  </label>
                  <div className="flex items-center bg-slate-800/80 border border-slate-700 rounded-xl px-3 text-sm focus-within:border-blue-500">
                    <span className="text-slate-500 select-none">trunc.site/m/</span>
                    <input
                      type="text"
                      required
                      value={newSlug}
                      onChange={(e) => setNewSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ''))}
                      placeholder="alex-carter"
                      className="w-full bg-transparent border-0 focus:outline-none py-2.5 text-white pl-1"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                    Curated Template (10 Options)
                  </label>
                  <div className="grid grid-cols-2 gap-2.5 max-h-52 overflow-y-auto pr-1">
                    {TEMPLATES.map((t) => (
                      <button
                        type="button"
                        key={t.id}
                        onClick={() => setSelectedTheme(t.id)}
                        className={`text-left p-3 rounded-xl border transition-all flex flex-col justify-between ${
                          selectedTheme === t.id
                            ? 'border-blue-500 bg-blue-500/10 shadow-sm'
                            : 'border-slate-800 bg-slate-800/40 hover:border-slate-700'
                        }`}
                      >
                        <div className="flex items-center justify-between mb-1">
                          <span className="text-xs font-bold text-white truncate">{t.name}</span>
                          <span
                            className="w-3 h-3 rounded-full flex-shrink-0"
                            style={{ backgroundColor: t.accentColor }}
                          />
                        </div>
                        <div className="text-[10px] text-slate-400 line-clamp-1">{t.tag}</div>
                      </button>
                    ))}
                  </div>
                </div>

                <div className="pt-4 border-t border-slate-800 flex items-center justify-end gap-3">
                  <button
                    type="button"
                    onClick={() => setIsCreateOpen(false)}
                    className="px-4 py-2.5 rounded-xl text-sm font-medium text-slate-400 hover:text-white transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    disabled={creating}
                    className="px-6 py-2.5 rounded-xl text-sm font-semibold text-white bg-blue-600 hover:bg-blue-500 shadow-lg shadow-blue-500/25 transition-all disabled:opacity-50"
                  >
                    {creating ? 'Creating...' : 'Create & Open Builder'}
                  </button>
                </div>
              </form>
            </div>
          </div>
        )}

        {/* QR Code Modal */}
        {qrUrl && (
          <QRCodeModal
            isOpen={true}
            onClose={() => setQrUrl(null)}
            value={qrUrl}
            title="Scan to Save Contact"
          />
        )}
      </div>
    </Layout>
  );
}
