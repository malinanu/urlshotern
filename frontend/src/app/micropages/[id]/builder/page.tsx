'use client';

import React, { useState, useEffect, use } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { getAccessToken } from '@/utils/tokenValidation';
import { MicroPage, MicroPageTheme, MicroLinkItem, MicroSocialLink, VCardContact } from '@/types/micropage';
import { TEMPLATES } from '@/components/micropages/templates';
import { MicroPageRenderer } from '@/components/micropages/MicroPageRenderer';
import { QRCodeModal } from '@/components/QRCode';
import {
  ArrowLeftIcon,
  CheckIcon,
  EyeIcon,
  SparklesIcon,
  DevicePhoneMobileIcon,
  ComputerDesktopIcon,
  PlusIcon,
  TrashIcon,
  QrCodeIcon,
  ArrowDownTrayIcon,
  ArrowTopRightOnSquareIcon,
  UserCircleIcon,
  IdentificationIcon,
  LinkIcon,
} from '@heroicons/react/24/outline';

const SAMPLE_AVATARS = [
  'https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=400&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=400&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1517841905240-472988babdf9?w=400&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=400&auto=format&fit=crop&q=80',
];

const SAMPLE_COVERS = [
  'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?w=1200&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1579546929518-9e396f3cc809?w=1200&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1550684848-fac1c5b4e853?w=1200&auto=format&fit=crop&q=80',
  'https://images.unsplash.com/photo-1507525428034-b723cf961d3e?w=1200&auto=format&fit=crop&q=80',
];

export default function MicroPageBuilderPage({ params }: { params: Promise<{ id: string }> }) {
  const resolvedParams = use(params);
  const pageId = resolvedParams.id;

  const { isAuthenticated, isLoading: authLoading } = useAuth();
  const router = useRouter();

  const [page, setPage] = useState<MicroPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Active builder tab: 1 = Profile, 2 = vCard, 3 = Links
  const [activeTier, setActiveTier] = useState<1 | 2 | 3>(1);
  const [deviceView, setDeviceView] = useState<'mobile' | 'desktop'>('mobile');
  const [qrOpen, setQrOpen] = useState(false);

  useEffect(() => {
    if (!authLoading && !isAuthenticated) {
      router.push(`/login?redirect=/micropages/${pageId}/builder`);
    }
  }, [authLoading, isAuthenticated, pageId, router]);

  useEffect(() => {
    const fetchPage = async () => {
      try {
        setLoading(true);
        const token = getAccessToken();
        const res = await fetch(API_ENDPOINTS.micropages.get(pageId), {
          headers: {
            Authorization: `Bearer ${token}`,
          },
        });

        if (!res.ok) {
          throw new Error('Micro-page not found or unauthorized');
        }

        const data: MicroPage = await res.json();
        // Ensure default structures
        if (!data.vcard) {
          data.vcard = {
            first_name: data.title || '',
            last_name: '',
            organization: '',
            title: '',
            email: '',
            phone: '',
          };
        }
        if (!data.social_links) data.social_links = [];
        if (!data.links) data.links = [];
        setPage(data);
      } catch (err: any) {
        setError(err.message || 'Error loading builder');
      } finally {
        setLoading(false);
      }
    };

    if (isAuthenticated) {
      fetchPage();
    }
  }, [pageId, isAuthenticated]);

  const handleSave = async () => {
    if (!page) return;
    try {
      setSaving(true);
      setSaveSuccess(false);
      setError(null);

      const token = getAccessToken();
      const res = await fetch(API_ENDPOINTS.micropages.update(page.id), {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          title: page.title,
          slug: page.slug,
          bio: page.bio,
          avatar_url: page.avatar_url,
          cover_url: page.cover_url,
          theme: page.theme,
          custom_css: page.custom_css,
          config: page.config,
          vcard: page.vcard,
          social_links: page.social_links,
          links: page.links,
          is_published: page.is_published,
        }),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => ({}));
        throw new Error(errData.error || 'Failed to save changes');
      }

      const updated = await res.json();
      setPage(updated);
      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      setError(err.message || 'Error saving changes');
    } finally {
      setSaving(false);
    }
  };

  // Add / Edit / Remove Link
  const handleAddLink = () => {
    if (!page) return;
    const newLink: MicroLinkItem = {
      id: `link-${Date.now()}`,
      title: 'New Link',
      url: 'https://',
      icon: '🔗',
      description: '',
      is_highlighted: false,
      is_active: true,
      sort_order: page.links.length,
      click_count: 0,
    };
    setPage({ ...page, links: [...page.links, newLink] });
  };

  const handleUpdateLink = (id: string, updates: Partial<MicroLinkItem>) => {
    if (!page) return;
    setPage({
      ...page,
      links: page.links.map((l) => (l.id === id ? { ...l, ...updates } : l)),
    });
  };

  const handleDeleteLink = (id: string) => {
    if (!page) return;
    setPage({
      ...page,
      links: page.links.filter((l) => l.id !== id),
    });
  };

  // Add / Remove Social Link
  const handleAddSocial = () => {
    if (!page) return;
    const newSocial: MicroSocialLink = { platform: 'twitter', url: 'https://x.com/' };
    setPage({ ...page, social_links: [...page.social_links, newSocial] });
  };

  const handleUpdateSocial = (index: number, updates: Partial<MicroSocialLink>) => {
    if (!page) return;
    const updated = [...page.social_links];
    updated[index] = { ...updated[index], ...updates };
    setPage({ ...page, social_links: updated });
  };

  const handleDeleteSocial = (index: number) => {
    if (!page) return;
    setPage({
      ...page,
      social_links: page.social_links.filter((_, i) => i !== index),
    });
  };

  const handleDownloadPreviewVCard = () => {
    if (!page) return;
    window.open(API_ENDPOINTS.micropages.vcard(page.slug), '_blank');
  };

  if (loading) {
    return (
      <Layout>
        <div className="max-w-7xl mx-auto px-4 py-16 text-center animate-pulse">
          <div className="w-12 h-12 border-4 border-blue-500 border-t-transparent rounded-full animate-spin mx-auto mb-4" />
          <p className="text-slate-400 text-sm">Loading Micro-Page Builder...</p>
        </div>
      </Layout>
    );
  }

  if (error && !page) {
    return (
      <Layout>
        <div className="max-w-xl mx-auto px-4 py-16 text-center">
          <div className="p-6 bg-rose-500/10 border border-rose-500/30 rounded-2xl text-rose-400">
            {error}
          </div>
          <Link
            href="/micropages"
            className="mt-6 inline-flex items-center gap-2 px-4 py-2 rounded-xl text-sm font-medium text-white bg-slate-800 hover:bg-slate-700"
          >
            <ArrowLeftIcon className="w-4 h-4" />
            <span>Back to Dashboard</span>
          </Link>
        </div>
      </Layout>
    );
  }

  if (!page) return null;

  const publicUrl = `/m/${page.slug}`;

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-6 space-y-6">
        {/* Top Sticky Header */}
        <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 pb-4 border-b border-slate-800">
          <div className="flex items-center gap-3">
            <Link
              href="/micropages"
              className="p-2 rounded-xl bg-slate-800 text-slate-400 hover:text-white hover:bg-slate-700 transition-colors"
            >
              <ArrowLeftIcon className="w-5 h-5" />
            </Link>
            <div>
              <div className="flex items-center gap-2">
                <h1 className="text-xl font-bold text-white tracking-tight">{page.title}</h1>
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-blue-500/10 text-blue-400 border border-blue-500/20 font-mono">
                  /m/{page.slug}
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">Micro-Landing Page & Digital vCard Studio</p>
            </div>
          </div>

          <div className="flex items-center gap-2 flex-wrap">
            <button
              type="button"
              onClick={() => setQrOpen(true)}
              className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition-colors"
              title="Show QR Code"
            >
              <QrCodeIcon className="w-5 h-5" />
            </button>

            <a
              href={publicUrl}
              target="_blank"
              rel="noopener noreferrer"
              className="inline-flex items-center gap-1.5 px-3 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-xs font-medium text-slate-300 hover:text-white transition-colors"
            >
              <ArrowTopRightOnSquareIcon className="w-4 h-4" />
              <span>Preview Live</span>
            </a>

            <button
              type="button"
              onClick={() => setPage({ ...page, is_published: !page.is_published })}
              className={`px-3 py-2 rounded-xl text-xs font-semibold transition-colors ${
                page.is_published
                  ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 hover:bg-emerald-500/20'
                  : 'bg-amber-500/10 text-amber-400 border border-amber-500/20 hover:bg-amber-500/20'
              }`}
            >
              {page.is_published ? '● Published' : '○ Draft'}
            </button>

            <button
              type="button"
              onClick={handleSave}
              disabled={saving}
              className={`inline-flex items-center gap-2 px-5 py-2 rounded-xl text-xs font-bold text-white transition-all ${
                saveSuccess
                  ? 'bg-emerald-600 hover:bg-emerald-500'
                  : 'bg-blue-600 hover:bg-blue-500 shadow-lg shadow-blue-500/25'
              } disabled:opacity-50`}
            >
              {saveSuccess ? (
                <>
                  <CheckIcon className="w-4 h-4" />
                  <span>Saved!</span>
                </>
              ) : (
                <span>{saving ? 'Saving...' : 'Save Changes'}</span>
              )}
            </button>
          </div>
        </div>

        {error && (
          <div className="p-4 bg-rose-500/10 border border-rose-500/30 rounded-xl text-xs text-rose-400">
            {error}
          </div>
        )}

        {/* Builder Work Area: Left = 3-Tier Editor, Right = Live Device Preview */}
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-start">
          {/* Left Editor Pane (7 cols) */}
          <div className="lg:col-span-7 space-y-6">
            {/* 3-Tier Tabs Switcher */}
            <div className="flex bg-slate-900 border border-slate-800 rounded-2xl p-1.5 gap-1">
              <button
                type="button"
                onClick={() => setActiveTier(1)}
                className={`flex-1 py-2.5 px-3 rounded-xl text-xs font-bold flex items-center justify-center gap-2 transition-all ${
                  activeTier === 1
                    ? 'bg-blue-600 text-white shadow-md'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
                }`}
              >
                <UserCircleIcon className="w-4 h-4" />
                <span>1. Profile & Theme</span>
              </button>

              <button
                type="button"
                onClick={() => setActiveTier(2)}
                className={`flex-1 py-2.5 px-3 rounded-xl text-xs font-bold flex items-center justify-center gap-2 transition-all ${
                  activeTier === 2
                    ? 'bg-blue-600 text-white shadow-md'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
                }`}
              >
                <IdentificationIcon className="w-4 h-4" />
                <span>2. Digital vCard</span>
              </button>

              <button
                type="button"
                onClick={() => setActiveTier(3)}
                className={`flex-1 py-2.5 px-3 rounded-xl text-xs font-bold flex items-center justify-center gap-2 transition-all ${
                  activeTier === 3
                    ? 'bg-blue-600 text-white shadow-md'
                    : 'text-slate-400 hover:text-white hover:bg-slate-800/60'
                }`}
              >
                <LinkIcon className="w-4 h-4" />
                <span>3. Links & Social</span>
              </button>
            </div>

            {/* Tier 1: Profile & Identity */}
            {activeTier === 1 && (
              <div className="bg-slate-900/70 border border-slate-800 rounded-2xl p-6 space-y-6">
                <div>
                  <h3 className="text-base font-bold text-white">Tier 1: Profile & Visual Identity</h3>
                  <p className="text-xs text-slate-400 mt-0.5">Customize your page headers, theme aesthetic, and URLs</p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Display Title / Full Name
                    </label>
                    <input
                      type="text"
                      value={page.title}
                      onChange={(e) => setPage({ ...page, title: e.target.value })}
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Custom URL Slug
                    </label>
                    <div className="flex items-center bg-slate-800 border border-slate-700 rounded-xl px-3 text-sm focus-within:border-blue-500">
                      <span className="text-slate-500 select-none">/m/</span>
                      <input
                        type="text"
                        value={page.slug}
                        onChange={(e) =>
                          setPage({ ...page, slug: e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '') })
                        }
                        className="w-full bg-transparent border-0 focus:outline-none py-2 text-white pl-1"
                      />
                    </div>
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    Bio / Headline
                  </label>
                  <textarea
                    rows={3}
                    value={page.bio}
                    onChange={(e) => setPage({ ...page, bio: e.target.value })}
                    placeholder="Short summary about your mission or services..."
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                  />
                </div>

                {/* Avatar URL + presets */}
                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    Profile Avatar Image URL
                  </label>
                  <input
                    type="url"
                    value={page.avatar_url || ''}
                    onChange={(e) => setPage({ ...page, avatar_url: e.target.value })}
                    placeholder="https://example.com/avatar.jpg"
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none mb-2"
                  />
                  <div className="flex items-center gap-2">
                    <span className="text-[11px] text-slate-400">Quick picks:</span>
                    {SAMPLE_AVATARS.map((av, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setPage({ ...page, avatar_url: av })}
                        className="w-7 h-7 rounded-full overflow-hidden border border-slate-600 hover:scale-110 transition-transform"
                      >
                        <img src={av} alt="Preset" className="w-full h-full object-cover" />
                      </button>
                    ))}
                  </div>
                </div>

                {/* Cover URL + presets */}
                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    Cover Banner Image URL
                  </label>
                  <input
                    type="url"
                    value={page.cover_url || ''}
                    onChange={(e) => setPage({ ...page, cover_url: e.target.value })}
                    placeholder="https://example.com/cover.jpg"
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none mb-2"
                  />
                  <div className="flex items-center gap-2">
                    <span className="text-[11px] text-slate-400">Presets:</span>
                    {SAMPLE_COVERS.map((cov, idx) => (
                      <button
                        key={idx}
                        type="button"
                        onClick={() => setPage({ ...page, cover_url: cov })}
                        className="w-12 h-6 rounded-md overflow-hidden border border-slate-600 hover:scale-105 transition-transform"
                      >
                        <img src={cov} alt="Cover preset" className="w-full h-full object-cover" />
                      </button>
                    ))}
                  </div>
                </div>

                {/* 10 Curated Templates Selector */}
                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-2">
                    Curated Template (10 Handcrafted Designs)
                  </label>
                  <div className="grid grid-cols-2 sm:grid-cols-2 gap-3 max-h-72 overflow-y-auto pr-1">
                    {TEMPLATES.map((t) => (
                      <button
                        key={t.id}
                        type="button"
                        onClick={() => setPage({ ...page, theme: t.id })}
                        className={`text-left p-3.5 rounded-2xl border transition-all flex flex-col justify-between ${
                          page.theme === t.id
                            ? 'border-blue-500 bg-blue-500/10 shadow-md ring-1 ring-blue-500'
                            : 'border-slate-800 bg-slate-800/40 hover:border-slate-700'
                        }`}
                      >
                        <div>
                          <div className="flex items-center justify-between mb-1.5">
                            <span className="text-xs font-bold text-white truncate">{t.name}</span>
                            <span
                              className="w-3.5 h-3.5 rounded-full flex-shrink-0"
                              style={{ backgroundColor: t.accentColor }}
                            />
                          </div>
                          <p className="text-[11px] text-slate-400 leading-snug line-clamp-2">
                            {t.description}
                          </p>
                        </div>
                        <span className="mt-2 text-[10px] uppercase font-semibold tracking-wider text-slate-500">
                          {t.tag}
                        </span>
                      </button>
                    ))}
                  </div>
                </div>
              </div>
            )}

            {/* Tier 2: Digital vCard Information */}
            {activeTier === 2 && (
              <div className="bg-slate-900/70 border border-slate-800 rounded-2xl p-6 space-y-6">
                <div className="flex items-center justify-between">
                  <div>
                    <h3 className="text-base font-bold text-white">Tier 2: Digital vCard Information</h3>
                    <p className="text-xs text-slate-400 mt-0.5">
                      Information packaged into the .vcf contact card for smartphones
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={handleDownloadPreviewVCard}
                    className="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-semibold bg-slate-800 hover:bg-slate-700 text-slate-200 transition-colors"
                  >
                    <ArrowDownTrayIcon className="w-4 h-4 text-blue-400" />
                    <span>Test .vcf Download</span>
                  </button>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      First Name
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.first_name || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, first_name: e.target.value },
                        })
                      }
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Last Name
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.last_name || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, last_name: e.target.value },
                        })
                      }
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Job Title
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.title || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, title: e.target.value },
                        })
                      }
                      placeholder="e.g. Chief Product Officer"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Organization / Company
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.organization || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, organization: e.target.value },
                        })
                      }
                      placeholder="e.g. Acme Ventures"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Email Address
                    </label>
                    <input
                      type="email"
                      value={page.vcard?.email || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, email: e.target.value },
                        })
                      }
                      placeholder="alex@acme.com"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Primary Phone
                    </label>
                    <input
                      type="tel"
                      value={page.vcard?.phone || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, phone: e.target.value },
                        })
                      }
                      placeholder="+1 (555) 123-4567"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Cell / Mobile Phone
                    </label>
                    <input
                      type="tel"
                      value={page.vcard?.cell_phone || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, cell_phone: e.target.value },
                        })
                      }
                      placeholder="+1 (555) 987-6543"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Website / Portfolio
                    </label>
                    <input
                      type="url"
                      value={page.vcard?.website || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, website: e.target.value },
                        })
                      }
                      placeholder="https://example.com"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      City, State
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.city || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, city: e.target.value },
                        })
                      }
                      placeholder="San Francisco, CA"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>

                  <div>
                    <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                      Country
                    </label>
                    <input
                      type="text"
                      value={page.vcard?.country || ''}
                      onChange={(e) =>
                        setPage({
                          ...page,
                          vcard: { ...page.vcard!, country: e.target.value },
                        })
                      }
                      placeholder="United States"
                      className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                    />
                  </div>
                </div>

                <div>
                  <label className="block text-xs font-semibold text-slate-300 uppercase tracking-wider mb-1">
                    vCard Contact Notes / Memo
                  </label>
                  <textarea
                    rows={2}
                    value={page.vcard?.note || ''}
                    onChange={(e) =>
                      setPage({
                        ...page,
                        vcard: { ...page.vcard!, note: e.target.value },
                      })
                    }
                    placeholder="Notes embedded into recipient's address book..."
                    className="w-full bg-slate-800 border border-slate-700 rounded-xl px-3 py-2 text-sm text-white focus:border-blue-500 focus:outline-none"
                  />
                </div>
              </div>
            )}

            {/* Tier 3: Interactive Links & Social Icons */}
            {activeTier === 3 && (
              <div className="bg-slate-900/70 border border-slate-800 rounded-2xl p-6 space-y-6">
                <div>
                  <h3 className="text-base font-bold text-white">Tier 3: Interactive Links & Social Icons</h3>
                  <p className="text-xs text-slate-400 mt-0.5">
                    Connect social profiles and create interactive call-to-action buttons
                  </p>
                </div>

                {/* Social Profiles */}
                <div className="space-y-3">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-slate-300 uppercase tracking-wider">
                      Social Profiles
                    </span>
                    <button
                      type="button"
                      onClick={handleAddSocial}
                      className="inline-flex items-center gap-1 text-xs font-semibold text-blue-400 hover:text-blue-300"
                    >
                      <PlusIcon className="w-4 h-4" />
                      <span>Add Social Link</span>
                    </button>
                  </div>

                  {page.social_links.map((s, idx) => (
                    <div
                      key={idx}
                      className="flex items-center gap-2 bg-slate-800/60 p-2.5 rounded-xl border border-slate-700/80"
                    >
                      <select
                        value={s.platform}
                        onChange={(e) => handleUpdateSocial(idx, { platform: e.target.value })}
                        className="bg-slate-900 text-white text-xs rounded-lg px-2.5 py-1.5 border border-slate-700 focus:outline-none"
                      >
                        <option value="twitter">X / Twitter</option>
                        <option value="linkedin">LinkedIn</option>
                        <option value="github">GitHub</option>
                        <option value="instagram">Instagram</option>
                        <option value="youtube">YouTube</option>
                        <option value="whatsapp">WhatsApp</option>
                        <option value="telegram">Telegram</option>
                        <option value="email">Email</option>
                        <option value="website">Website</option>
                      </select>

                      <input
                        type="url"
                        value={s.url}
                        onChange={(e) => handleUpdateSocial(idx, { url: e.target.value })}
                        placeholder="https://..."
                        className="flex-1 bg-slate-900 text-white text-xs rounded-lg px-3 py-1.5 border border-slate-700 focus:outline-none"
                      />

                      <button
                        type="button"
                        onClick={() => handleDeleteSocial(idx)}
                        className="p-1.5 text-slate-400 hover:text-rose-400 transition-colors"
                      >
                        <TrashIcon className="w-4 h-4" />
                      </button>
                    </div>
                  ))}
                </div>

                {/* Custom Action Links */}
                <div className="space-y-3 pt-4 border-t border-slate-800">
                  <div className="flex items-center justify-between">
                    <span className="text-xs font-bold text-slate-300 uppercase tracking-wider">
                      Action Buttons / Links ({page.links.length})
                    </span>
                    <button
                      type="button"
                      onClick={handleAddLink}
                      className="inline-flex items-center gap-1 text-xs font-semibold text-blue-400 hover:text-blue-300"
                    >
                      <PlusIcon className="w-4 h-4" />
                      <span>Add Button</span>
                    </button>
                  </div>

                  {page.links.map((link) => (
                    <div
                      key={link.id}
                      className="bg-slate-800/60 border border-slate-700/80 rounded-xl p-3.5 space-y-3"
                    >
                      <div className="grid grid-cols-1 sm:grid-cols-12 gap-2.5">
                        <div className="sm:col-span-2">
                          <label className="block text-[10px] text-slate-400 uppercase font-semibold mb-0.5">Icon</label>
                          <input
                            type="text"
                            value={link.icon || ''}
                            onChange={(e) => handleUpdateLink(link.id, { icon: e.target.value })}
                            placeholder="📅"
                            className="w-full bg-slate-900 text-white text-center text-sm rounded-lg py-1.5 border border-slate-700 focus:outline-none"
                          />
                        </div>
                        <div className="sm:col-span-5">
                          <label className="block text-[10px] text-slate-400 uppercase font-semibold mb-0.5">Title</label>
                          <input
                            type="text"
                            value={link.title}
                            onChange={(e) => handleUpdateLink(link.id, { title: e.target.value })}
                            placeholder="Book a meeting"
                            className="w-full bg-slate-900 text-white text-xs rounded-lg px-3 py-1.5 border border-slate-700 focus:outline-none"
                          />
                        </div>
                        <div className="sm:col-span-5">
                          <label className="block text-[10px] text-slate-400 uppercase font-semibold mb-0.5">URL</label>
                          <input
                            type="url"
                            value={link.url}
                            onChange={(e) => handleUpdateLink(link.id, { url: e.target.value })}
                            placeholder="https://..."
                            className="w-full bg-slate-900 text-white text-xs rounded-lg px-3 py-1.5 border border-slate-700 focus:outline-none"
                          />
                        </div>
                      </div>

                      <div className="flex items-center justify-between gap-2 pt-1 text-xs">
                        <input
                          type="text"
                          value={link.description || ''}
                          onChange={(e) => handleUpdateLink(link.id, { description: e.target.value })}
                          placeholder="Subtitle description (optional)..."
                          className="flex-1 bg-slate-900 text-slate-300 text-xs rounded-lg px-3 py-1.5 border border-slate-700 focus:outline-none"
                        />

                        <label className="flex items-center gap-1.5 cursor-pointer text-slate-300 text-xs select-none">
                          <input
                            type="checkbox"
                            checked={link.is_highlighted || false}
                            onChange={(e) => handleUpdateLink(link.id, { is_highlighted: e.target.checked })}
                            className="rounded bg-slate-900 border-slate-700 text-blue-600 focus:ring-0"
                          />
                          <span>Highlight</span>
                        </label>

                        <button
                          type="button"
                          onClick={() => handleDeleteLink(link.id)}
                          className="p-1.5 text-slate-400 hover:text-rose-400 transition-colors ml-1"
                        >
                          <TrashIcon className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>

          {/* Right Live Device Preview Pane (5 cols) */}
          <div className="lg:col-span-5 flex flex-col items-center">
            {/* Device Switcher */}
            <div className="flex items-center gap-2 bg-slate-900 border border-slate-800 rounded-xl p-1 mb-4">
              <button
                type="button"
                onClick={() => setDeviceView('mobile')}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                  deviceView === 'mobile'
                    ? 'bg-blue-600 text-white shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                <DevicePhoneMobileIcon className="w-4 h-4" />
                <span>Mobile Preview</span>
              </button>
              <button
                type="button"
                onClick={() => setDeviceView('desktop')}
                className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-semibold transition-all ${
                  deviceView === 'desktop'
                    ? 'bg-blue-600 text-white shadow-sm'
                    : 'text-slate-400 hover:text-white'
                }`}
              >
                <ComputerDesktopIcon className="w-4 h-4" />
                <span>Desktop Card</span>
              </button>
            </div>

            {/* Device Frame */}
            {deviceView === 'mobile' ? (
              <div className="w-[360px] h-[720px] rounded-[48px] bg-black p-3.5 shadow-2xl border-4 border-slate-700 relative flex flex-col overflow-hidden">
                {/* Speaker Notch */}
                <div className="absolute top-4 left-1/2 -translate-x-1/2 w-28 h-5 bg-black rounded-full z-30 flex items-center justify-center">
                  <div className="w-10 h-1.5 bg-slate-900 rounded-full" />
                </div>

                {/* Inner Screen */}
                <div className="w-full h-full rounded-[38px] overflow-y-auto overflow-x-hidden relative flex flex-col bg-slate-950 scrollbar-none">
                  {/* Status Bar */}
                  <div className="w-full h-8 pt-2 px-6 flex items-center justify-between text-[11px] font-semibold text-white/70 select-none z-20 flex-shrink-0">
                    <span>9:41</span>
                    <div className="flex items-center gap-1.5">
                      <span>5G</span>
                      <span>100%</span>
                    </div>
                  </div>

                  {/* Rendered Live MicroPage */}
                  <div className="flex-1">
                    <MicroPageRenderer
                      page={page}
                      isPreview={true}
                      onSaveContact={handleDownloadPreviewVCard}
                    />
                  </div>

                  {/* Home Indicator */}
                  <div className="w-full py-2 flex items-center justify-center select-none flex-shrink-0">
                    <div className="w-32 h-1 bg-white/40 rounded-full" />
                  </div>
                </div>
              </div>
            ) : (
              /* Desktop Card Preview */
              <div className="w-full max-w-md h-[680px] rounded-3xl overflow-y-auto border border-slate-800 shadow-2xl bg-slate-950">
                <MicroPageRenderer
                  page={page}
                  isPreview={true}
                  onSaveContact={handleDownloadPreviewVCard}
                />
              </div>
            )}
          </div>
        </div>

        {/* QR Code Modal */}
        {qrOpen && (
          <QRCodeModal
            isOpen={true}
            onClose={() => setQrOpen(false)}
            value={`${typeof window !== 'undefined' ? window.location.origin : ''}/m/${page.slug}`}
            title={`Scan /m/${page.slug}`}
          />
        )}
      </div>
    </Layout>
  );
}
