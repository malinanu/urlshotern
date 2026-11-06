'use client';

import { useEffect, useState, useRef, useMemo, useCallback, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import WorkspaceLayout from '@/components/layout/WorkspaceLayout';
import { useAuth } from '@/contexts/AuthContext';
import { QRCodeModal } from '@/components/QRCode';
import { getCached, setCached, prefetch } from '@/lib/prefetch-cache';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  ChartBarIcon,
  LinkIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  TrashIcon,
  PencilIcon,
  QrCodeIcon,
  ArrowRightIcon,
  SparklesIcon,
  ArrowTopRightOnSquareIcon,
  ShareIcon,
  ArrowDownTrayIcon,
  Bars3BottomLeftIcon,
  Squares2X2Icon,
  ChevronDownIcon,
  XMarkIcon,
  PlusIcon,
  AdjustmentsHorizontalIcon,
  TagIcon,
  EllipsisVerticalIcon,
  PauseIcon,
  PlayIcon,
  CodeBracketIcon,
} from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface URLData {
  id: number;
  short_code: string;
  short_url: string;
  original_url: string;
  created_at: string;
  click_count: number;
  is_active: boolean;
  is_public: boolean;
  title?: string;
  description?: string;
  expires_at?: string;
  team_id?: number;
  tags?: string[];
}

interface Team {
  id: number;
  name: string;
  display_name?: string;
}

interface CustomDomain {
  id: number;
  domain: string;
  is_verified: boolean;
  is_active: boolean;
}

interface DashboardStats {
  total_urls: number;
  total_clicks: number;
  active_urls: number;
  today_clicks: number;
  month_clicks: number;
}

function DashboardContent() {
  const { user, isAuthenticated, isLoading, getAccessToken, refreshToken } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();

  // Data state
  const [urls, setUrls] = useState<URLData[]>(() => {
    const cached = getCached<{ urls: URLData[] }>('dashboard:urls');
    return cached?.urls ?? [];
  });
  const [currentPage, setCurrentPage] = useState<number>(() => {
    const cached = getCached<{ page?: number }>('dashboard:urls');
    return cached?.page ?? 1;
  });
  const [totalPages, setTotalPages] = useState<number>(() => {
    const cached = getCached<{ total_pages?: number }>('dashboard:urls');
    return cached?.total_pages ?? 1;
  });
  const [pageSize, setPageSize] = useState<number>(10);
  const [stats, setStats] = useState<DashboardStats | null>(() => getCached<DashboardStats>('dashboard:stats'));
  const [loadingUrls, setLoadingUrls] = useState(getCached('dashboard:urls') === null);

  // Interaction state
  const [copiedId, setCopiedId] = useState<number | null>(null);
  const [editingUrl, setEditingUrl] = useState<URLData | null>(null);
  const [editError, setEditError] = useState<string | null>(null);
  const [editExpiresAt, setEditExpiresAt] = useState<string>('');
  const expiresAtRef = useRef<HTMLInputElement>(null);
  const [qrModalUrl, setQrModalUrl] = useState<URLData | null>(null);
  const [userTeams, setUserTeams] = useState<Team[]>([]);
  const [customDomains, setCustomDomains] = useState<CustomDomain[]>([]);
  const abortControllerRef = useRef<AbortController | null>(null);

  // Filters & search state
  const [activeNav, setActiveNav] = useState('home');
  const [searchQuery, setSearchQuery] = useState('');
  const [statusFilter, setStatusFilter] = useState<'all' | 'active' | 'archived'>('active');
  const [dateFilter, setDateFilter] = useState<'all' | 'today' | 'week' | 'month'>('all');
  const [viewMode, setViewMode] = useState<'list' | 'grid'>('list');
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());

  // Additional Filter Ribbon Popover state
  const [showFiltersMenu, setShowFiltersMenu] = useState(false);
  const [tagFilter, setTagFilter] = useState<string>('all');
  const [domainFilter, setDomainFilter] = useState<string>('all');
  const [minClicksFilter, setMinClicksFilter] = useState<number>(0);
  const filtersMenuRef = useRef<HTMLDivElement>(null);

  // Bulk actions & more options state
  const [showBulkTagModal, setShowBulkTagModal] = useState(false);
  const [bulkTagInput, setBulkTagInput] = useState('');
  const [openMenuUrlId, setOpenMenuUrlId] = useState<number | null>(null);
  const actionMenuRef = useRef<HTMLDivElement>(null);

  // Quick Create hero form state
  const [heroTab, setHeroTab] = useState<'short' | 'qr'>('short');
  const [destinationUrl, setDestinationUrl] = useState('');
  const [selectedDomain, setSelectedDomain] = useState('trunc.site');
  const [generateQrChecked, setGenerateQrChecked] = useState(false);
  const [customSlug, setCustomSlug] = useState('');
  const [passwordProtection, setPasswordProtection] = useState('');
  const [showAdvancedCreate, setShowAdvancedCreate] = useState(false);
  const [isCreatingLink, setIsCreatingLink] = useState(false);
  const [createError, setCreateError] = useState<string | null>(null);
  const [createSuccessMsg, setCreateSuccessMsg] = useState<string | null>(null);

  // Sync ?tab=qr or ?tab=short from URL
  const tabParam = searchParams.get('tab');
  useEffect(() => {
    if (tabParam === 'qr') {
      setHeroTab('qr');
      setGenerateQrChecked(true);
      setActiveNav('qr');
    } else if (tabParam === 'short') {
      setHeroTab('short');
      setActiveNav('home');
    }
  }, [tabParam]);

  // Sync hash navigation (#links, #campaigns)
  useEffect(() => {
    if (typeof window === 'undefined') return;
    const handleHash = () => {
      const hash = window.location.hash;
      if (hash === '#campaigns') {
        setActiveNav('campaigns');
        setTagFilter('#Campaign');
        const el = document.getElementById('links');
        el?.scrollIntoView({ behavior: 'smooth' });
      } else if (hash === '#links') {
        setActiveNav('links');
        const el = document.getElementById('links');
        el?.scrollIntoView({ behavior: 'smooth' });
      }
    };
    handleHash();
    window.addEventListener('hashchange', handleHash);
    return () => window.removeEventListener('hashchange', handleHash);
  }, []);

  // Close menus on outside click
  useEffect(() => {
    const handleDown = (e: MouseEvent) => {
      if (filtersMenuRef.current && !filtersMenuRef.current.contains(e.target as Node)) {
        setShowFiltersMenu(false);
      }
      if (actionMenuRef.current && !actionMenuRef.current.contains(e.target as Node)) {
        setOpenMenuUrlId(null);
      }
    };
    document.addEventListener('mousedown', handleDown);
    return () => document.removeEventListener('mousedown', handleDown);
  }, []);

  // Left rail nav click handler
  const handleNavClick = (navId: string) => {
    setActiveNav(navId);
    if (navId === 'qr') {
      setHeroTab('qr');
      setGenerateQrChecked(true);
      const el = document.getElementById('quick-create-section');
      el?.scrollIntoView({ behavior: 'smooth' });
    } else if (navId === 'home') {
      setHeroTab('short');
      window.scrollTo({ top: 0, behavior: 'smooth' });
    } else if (navId === 'links') {
      setTagFilter('all');
      const el = document.getElementById('links');
      el?.scrollIntoView({ behavior: 'smooth' });
    } else if (navId === 'campaigns') {
      setTagFilter('#Campaign');
      const el = document.getElementById('links');
      el?.scrollIntoView({ behavior: 'smooth' });
    }
  };

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Fetch URLs & Stats
  const fetchData = useCallback(async (retryOnTokenRefresh = true) => {
    if (!isAuthenticated || !user) return;

    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    abortControllerRef.current = new AbortController();

    const token = getAccessToken();
    if (!token) return;

    const makeAuthenticatedRequest = async (url: string) => {
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        signal: abortControllerRef.current?.signal,
      });

      if (response.status === 401 && retryOnTokenRefresh) {
        try {
          await refreshToken();
          return fetchData(false);
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError);
        }
      }
      return response;
    };

    try {
      if (getCached('dashboard:urls') === null) setLoadingUrls(true);

      const urlsResponse = await makeAuthenticatedRequest(
        `${API_BASE_URL}/api/v1/my-urls?page=${currentPage}&limit=${pageSize}`
      );

      if (urlsResponse && urlsResponse.ok) {
        const urlsData = await urlsResponse.json();
        setUrls(urlsData.urls || []);
        setCurrentPage(urlsData.page || 1);
        setTotalPages(urlsData.total_pages || 1);
        setCached('dashboard:urls', urlsData);
      }

      const statsResponse = await makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/dashboard/stats`);
      if (statsResponse && statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
        setCached('dashboard:stats', statsData);
      }
    } catch (error: unknown) {
      if ((error as { name?: string })?.name !== 'AbortError') {
        console.error('Error fetching dashboard data:', error);
      }
    } finally {
      setLoadingUrls(false);
    }
  }, [isAuthenticated, user, getAccessToken, refreshToken, currentPage, pageSize]);

  useEffect(() => {
    fetchData();
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [fetchData]);

  // Fetch teams & custom domains
  useEffect(() => {
    if (!isAuthenticated) return;
    const token = getAccessToken();
    if (!token) return;

    fetch(`${API_BASE_URL}/api/v1/teams`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => (r.ok ? r.json() : Promise.reject(r.status)))
      .then(data => setUserTeams(data.teams ?? data ?? []))
      .catch(() => {});

    fetch(`${API_BASE_URL}/api/v1/custom-domains/`, {
      headers: { Authorization: `Bearer ${token}` },
    })
      .then(r => (r.ok ? r.json() : null))
      .then(data => {
        if (data?.domains) {
          setCustomDomains((data.domains as CustomDomain[]).filter(d => d.is_verified && d.is_active));
        }
      })
      .catch(() => {});
  }, [isAuthenticated, getAccessToken]);

  // Quick Create link submission
  const handleQuickCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!destinationUrl) {
      setCreateError('Please enter a destination URL');
      return;
    }

    let urlToSubmit = destinationUrl.trim();
    if (!urlToSubmit.startsWith('http://') && !urlToSubmit.startsWith('https://')) {
      urlToSubmit = 'https://' + urlToSubmit;
    }

    try {
      new URL(urlToSubmit);
    } catch {
      setCreateError('Please enter a valid URL (e.g. https://example.com)');
      return;
    }

    setIsCreatingLink(true);
    setCreateError(null);
    setCreateSuccessMsg(null);

    const token = getAccessToken();
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (token) headers.Authorization = `Bearer ${token}`;

    const body: Record<string, unknown> = {
      url: urlToSubmit,
    };
    if (customSlug.trim()) body.custom_code = customSlug.trim();
    if (passwordProtection.trim()) body.password = passwordProtection.trim();

    try {
      const res = await fetch(API_ENDPOINTS.urls.shorten, {
        method: 'POST',
        headers,
        body: JSON.stringify(body),
      });

      if (!res.ok) {
        const errData = await res.json().catch(() => null);
        throw new Error(errData?.error || errData?.message || 'Failed to shorten URL');
      }

      const createdUrl: URLData = await res.json();

      // Optimistic update
      setUrls(prev => [createdUrl, ...prev]);
      if (stats) {
        setStats(prev =>
          prev
            ? {
                ...prev,
                total_urls: prev.total_urls + 1,
                active_urls: prev.active_urls + 1,
              }
            : null
        );
      }

      setDestinationUrl('');
      setCustomSlug('');
      setPasswordProtection('');
      setShowAdvancedCreate(false);
      setCreateSuccessMsg(`Link created successfully: ${createdUrl.short_url}`);

      // If user checked QR code generation or was on the QR tab, open modal immediately
      if (generateQrChecked || heroTab === 'qr') {
        setQrModalUrl(createdUrl);
      }

      // Background refresh
      setTimeout(() => {
        fetchData();
      }, 1000);
    } catch (err: unknown) {
      setCreateError(err instanceof Error ? err.message : 'Something went wrong');
    } finally {
      setIsCreatingLink(false);
    }
  };

  // Copy to clipboard
  const copyToClipboard = async (url: string, id: number) => {
    try {
      await navigator.clipboard.writeText(url);
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    } catch (err) {
      console.error('Failed to copy to clipboard:', err);
    }
  };

  // Delete URL
  const deleteURL = async (shortCode: string) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/my-urls/${shortCode}`, {
        method: 'DELETE',
        headers: {
          Authorization: `Bearer ${token}`,
        },
      });

      if (response.ok) {
        setUrls(prev => prev.filter(url => url.short_code !== shortCode));
        if (stats) {
          setStats(prev =>
            prev
              ? {
                  ...prev,
                  total_urls: Math.max(0, prev.total_urls - 1),
                  active_urls: Math.max(0, prev.active_urls - 1),
                }
              : null
          );
        }
      } else {
        alert('Failed to delete URL. Please try again.');
      }
    } catch (error) {
      console.error('Error deleting URL:', error);
      alert('Error deleting URL. Please try again.');
    }
  };

  // Update URL
  const updateURL = async (
    shortCode: string,
    updates: {
      original_url?: string;
      title?: string;
      description?: string;
      expires_at?: string | null;
      is_public?: boolean;
      team_id?: number | null;
    }
  ) => {
    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`${API_BASE_URL}/api/v1/my-urls/${shortCode}`, {
        method: 'PUT',
        headers: {
          Authorization: `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updates),
      });

      if (response.ok) {
        const updatedUrl = await response.json();
        setUrls(prev => prev.map(url => (url.short_code === shortCode ? { ...url, ...updatedUrl } : url)));
        setEditingUrl(null);
        setEditError(null);
      } else {
        const errData = await response.json().catch(() => null);
        setEditError(errData?.error || errData?.message || 'Failed to update URL. Please try again.');
      }
    } catch (error) {
      console.error('Error updating URL:', error);
      setEditError('Network error. Please try again.');
    }
  };

  // Export CSV
  const exportCSV = (selectedOnly = false) => {
    const items =
      selectedOnly && selectedIds.size > 0 ? urls.filter(u => selectedIds.has(u.id)) : filteredUrls;
    if (!items.length) {
      alert('No links available to export.');
      return;
    }
    const header = ['Short Code', 'Title', 'Short URL', 'Destination URL', 'Clicks', 'Active', 'Created At'];
    const rows = items.map(u => [
      `"${u.short_code}"`,
      `"${(u.title || '').replace(/"/g, '""')}"`,
      `"${u.short_url}"`,
      `"${(u.original_url || '').replace(/"/g, '""')}"`,
      u.click_count,
      u.is_active ? 'Yes' : 'No',
      `"${u.created_at}"`,
    ]);
    const csvContent = [header.join(','), ...rows.map(r => r.join(','))].join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv;charset=utf-8;' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.setAttribute('download', `trunc-links-${Date.now()}.csv`);
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // Bulk actions
  const handleBulkHide = () => {
    if (selectedIds.size === 0) return;
    setUrls(prev => prev.map(u => (selectedIds.has(u.id) ? { ...u, is_active: false } : u)));
    setSelectedIds(new Set());
  };

  const handleApplyBulkTag = (tag: string) => {
    if (!tag.trim() || selectedIds.size === 0) return;
    const formatted = tag.trim().startsWith('#') ? tag.trim() : `#${tag.trim()}`;
    setUrls(prev =>
      prev.map(u => {
        if (!selectedIds.has(u.id)) return u;
        const current = u.tags || getTagsForUrl(u);
        return {
          ...u,
          tags: current.includes(formatted) ? current : [...current, formatted],
        };
      })
    );
    setShowBulkTagModal(false);
    setBulkTagInput('');
    setSelectedIds(new Set());
  };

  const toggleUrlActive = (shortCode: string, currentActive: boolean) => {
    updateURL(shortCode, { is_public: !currentActive });
    setUrls(prev => prev.map(u => (u.short_code === shortCode ? { ...u, is_active: !currentActive } : u)));
    setOpenMenuUrlId(null);
  };

  const copyMarkdown = (url: URLData) => {
    const md = `[${url.title || url.short_code}](${url.short_url})`;
    navigator.clipboard.writeText(md).catch(() => {});
    setOpenMenuUrlId(null);
  };

  // Helper: platform icon
  const getPlatformIcon = (url: string) => {
    const lower = (url || '').toLowerCase();
    if (lower.includes('github.com') || lower.includes('gitlab.com')) {
      return { name: 'GitHub', bg: 'bg-zinc-800 text-white', label: 'GH' };
    }
    if (lower.includes('youtube.com') || lower.includes('youtu.be') || lower.includes('vimeo.com')) {
      return { name: 'Video', bg: 'bg-red-500/10 text-red-500 border border-red-500/20', label: 'YT' };
    }
    if (lower.includes('notion.so') || lower.includes('notion.site')) {
      return { name: 'Notion', bg: 'bg-neutral-800 text-white', label: 'N' };
    }
    if (lower.includes('spacex.com') || lower.includes('producthunt.com') || lower.includes('launch')) {
      return { name: 'Launch', bg: 'bg-purple-500/10 text-purple-500 border border-purple-500/20', label: '🚀' };
    }
    if (lower.includes('apple.com') || lower.includes('ar') || lower.includes('meta.com')) {
      return { name: 'AR/Spatial', bg: 'bg-blue-500/10 text-blue-500 border border-blue-500/20', label: 'AR' };
    }
    if (lower.includes('terminal') || lower.includes('api.') || lower.includes('dev')) {
      return { name: 'Terminal', bg: 'bg-emerald-500/10 text-emerald-500 border border-emerald-500/20', label: '>_' };
    }
    return { name: 'Web', bg: 'bg-primary/10 text-primary border border-primary/20', label: '🌐' };
  };

  // Helper: tags
  const getTagsForUrl = (url: URLData): string[] => {
    if (url.tags && url.tags.length > 0) return url.tags;
    const lower = (url.original_url + ' ' + (url.title || '')).toLowerCase();
    if (lower.includes('git') || lower.includes('dev') || lower.includes('api') || lower.includes('code')) {
      return ['#Engineering', '#API'];
    }
    if (lower.includes('campaign') || lower.includes('promo') || lower.includes('lead') || lower.includes('ad')) {
      return ['#Marketing', '#Campaign'];
    }
    if (lower.includes('docs') || lower.includes('guide') || lower.includes('tutorial')) {
      return ['#Docs', '#Resources'];
    }
    if (lower.includes('shop') || lower.includes('product') || lower.includes('store')) {
      return ['#Ecommerce', '#Sales'];
    }
    return ['#General', '#Web'];
  };

  // Filtered URLs
  const filteredUrls = useMemo(() => {
    return urls.filter(item => {
      // Search query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const matchCode = item.short_code.toLowerCase().includes(q);
        const matchOriginal = item.original_url.toLowerCase().includes(q);
        const matchTitle = (item.title || '').toLowerCase().includes(q);
        if (!matchCode && !matchOriginal && !matchTitle) return false;
      }
      // Status
      if (statusFilter === 'active' && !item.is_active) return false;
      if (statusFilter === 'archived' && item.is_active) return false;
      // Date filter
      if (dateFilter !== 'all') {
        const itemDate = new Date(item.created_at).getTime();
        const now = Date.now();
        if (dateFilter === 'today' && now - itemDate > 86400000) return false;
        if (dateFilter === 'week' && now - itemDate > 7 * 86400000) return false;
        if (dateFilter === 'month' && now - itemDate > 30 * 86400000) return false;
      }
      // Tag filter
      if (tagFilter !== 'all') {
        const tags = getTagsForUrl(item);
        if (!tags.includes(tagFilter)) return false;
      }
      // Domain filter
      if (domainFilter !== 'all') {
        if (!item.short_url.includes(domainFilter)) return false;
      }
      // Min clicks filter
      if (minClicksFilter > 0 && item.click_count < minClicksFilter) {
        return false;
      }
      return true;
    });
  }, [urls, searchQuery, statusFilter, dateFilter, tagFilter, domainFilter, minClicksFilter]);

  // Bulk selection toggles
  const handleSelectAll = (checked: boolean) => {
    if (checked) {
      setSelectedIds(new Set(filteredUrls.map(u => u.id)));
    } else {
      setSelectedIds(new Set());
    }
  };

  const handleToggleSelect = (id: number) => {
    setSelectedIds(prev => {
      const next = new Set(prev);
      if (next.has(id)) {
        next.delete(id);
      } else {
        next.add(id);
      }
      return next;
    });
  };

  const quotaRemaining = Math.max(0, 100 - (stats?.total_urls ?? urls.length));

  if (isLoading || !isAuthenticated) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-surface-container-lowest dark:bg-[#07090e]">
        <div className="flex flex-col items-center gap-3">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
          <span className="text-xs font-mono text-muted-foreground">Loading workspace...</span>
        </div>
      </div>
    );
  }

  return (
    <WorkspaceLayout
      activeNav={activeNav}
      linkCount={urls.length}
      onNavClick={handleNavClick}
      searchQuery={searchQuery}
      onSearchChange={setSearchQuery}
      searchPlaceholder="Search links by code, title, or destination... (⌘K)"
    >
      <div className="space-y-8 pb-12 font-sans">
        {/* SECTION 1: QUICK CREATE HERO */}
        <section id="quick-create-section" className="space-y-4">
          {/* Tab Selector Switcher */}
          <div className="flex items-center gap-2 border-b border-outline-variant/30 pb-3">
            <button
              onClick={() => setHeroTab('short')}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all ${
                heroTab === 'short'
                  ? 'bg-primary text-white shadow-sm'
                  : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
              }`}
            >
              <LinkIcon className="w-4 h-4" />
              <span>Short link</span>
            </button>
            <button
              onClick={() => {
                setHeroTab('qr');
                setGenerateQrChecked(true);
              }}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-bold transition-all ${
                heroTab === 'qr'
                  ? 'bg-primary text-white shadow-sm'
                  : 'text-muted-foreground hover:text-foreground hover:bg-surface-container'
              }`}
            >
              <QrCodeIcon className="w-4 h-4" />
              <span>QR Code</span>
              <span className="text-[10px] bg-indigo-500/20 text-indigo-400 px-1.5 py-0.2 rounded-full font-mono">
                Dynamic
              </span>
            </button>
          </div>

          {/* Hero Grid: Form (8 cols) + AI Assist Card (4 cols) */}
          <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-stretch">
            {/* Left Form (8 cols) */}
            <div className="lg:col-span-8 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl p-6 shadow-sm flex flex-col justify-between">
              <div>
                {/* Header & Quota */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between pb-4 gap-2 border-b border-outline-variant/20 mb-5">
                  <div className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
                      <LinkIcon className="w-4 h-4" />
                    </div>
                    <h2 className="text-sm font-bold text-foreground font-hanken">
                      {heroTab === 'short' ? 'Quick create: Short link' : 'Quick create: Dynamic QR Code'}
                    </h2>
                  </div>
                  <div className="flex items-center gap-2 text-xs text-muted-foreground font-mono">
                    <span className="w-2 h-2 rounded-full bg-emerald-500"></span>
                    <span>You can create <strong className="text-foreground">{quotaRemaining}</strong> more links this month</span>
                  </div>
                </div>

                {/* Form Elements */}
                <form onSubmit={handleQuickCreate} className="space-y-4">
                  {/* Domain pill & input container */}
                  <div className="space-y-1.5">
                    <label className="text-xs font-semibold text-muted-foreground">Destination URL</label>
                    <div className="flex flex-col sm:flex-row gap-2">
                      {/* Domain selection pill */}
                      <div className="relative shrink-0">
                        <select
                          value={selectedDomain}
                          onChange={(e) => setSelectedDomain(e.target.value)}
                          className="appearance-none pl-3 pr-8 py-2.5 rounded-xl text-xs font-semibold bg-surface-container dark:bg-zinc-900 border border-outline-variant/30 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 cursor-pointer h-full"
                        >
                          <option value="trunc.site">trunc.site (Standard)</option>
                          {customDomains.map(d => (
                            <option key={d.id} value={d.domain}>{d.domain} (Custom)</option>
                          ))}
                        </select>
                        <ChevronDownIcon className="w-3.5 h-3.5 text-muted-foreground absolute right-2.5 top-3.5 pointer-events-none" />
                      </div>

                      {/* URL input */}
                      <div className="flex-1 relative">
                        <input
                          type="text"
                          value={destinationUrl}
                          onChange={(e) => setDestinationUrl(e.target.value)}
                          placeholder="https://yourbrand.com/long-campaign-page-2026?utm=workspace"
                          className="w-full px-4 py-2.5 rounded-xl text-xs bg-surface-container dark:bg-zinc-900/80 border border-outline-variant/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all font-sans"
                        />
                      </div>

                      {/* Submit CTA */}
                      <button
                        type="submit"
                        disabled={isCreatingLink}
                        className="px-5 py-2.5 rounded-xl bg-primary hover:bg-primary-600 active:scale-[0.98] disabled:opacity-50 text-white text-xs font-bold tracking-tight shadow-sm transition-all shrink-0 flex items-center justify-center gap-2"
                      >
                        {isCreatingLink ? (
                          <>
                            <div className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin"></div>
                            <span>Creating...</span>
                          </>
                        ) : (
                          <>
                            <span>Create your link</span>
                            <ArrowRightIcon className="w-3.5 h-3.5" />
                          </>
                        )}
                      </button>
                    </div>
                  </div>

                  {/* QR Code Checkbox */}
                  <div className="flex items-center gap-2 pt-1">
                    <input
                      type="checkbox"
                      id="hero-qr-checkbox"
                      checked={generateQrChecked || heroTab === 'qr'}
                      onChange={(e) => setGenerateQrChecked(e.target.checked)}
                      className="rounded border-outline-variant text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                    />
                    <label htmlFor="hero-qr-checkbox" className="text-xs text-muted-foreground cursor-pointer select-none">
                      Also generate an interactive dynamic QR Code for this link
                    </label>
                  </div>

                  {/* Pro / Advanced toggle */}
                  <div className="pt-1">
                    <button
                      type="button"
                      onClick={() => setShowAdvancedCreate(!showAdvancedCreate)}
                      className="text-xs text-primary hover:underline font-semibold flex items-center gap-1"
                    >
                      <span>{showAdvancedCreate ? '− Hide optional slug & password' : '+ Add custom slug alias or password'}</span>
                    </button>
                    {showAdvancedCreate && (
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-3 pt-3 border-t border-outline-variant/20 animate-in fade-in duration-150">
                        <div>
                          <label className="text-[11px] font-semibold text-muted-foreground block mb-1">
                            Custom Slug (Back-half)
                          </label>
                          <input
                            type="text"
                            value={customSlug}
                            onChange={(e) => setCustomSlug(e.target.value)}
                            placeholder="e.g. launch-party"
                            className="w-full px-3 py-2 rounded-xl text-xs bg-surface-container dark:bg-zinc-900 border border-outline-variant/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 font-mono"
                          />
                        </div>
                        <div>
                          <label className="text-[11px] font-semibold text-muted-foreground block mb-1">
                            Password Gate (Optional)
                          </label>
                          <input
                            type="password"
                            value={passwordProtection}
                            onChange={(e) => setPasswordProtection(e.target.value)}
                            placeholder="Secret passcode"
                            className="w-full px-3 py-2 rounded-xl text-xs bg-surface-container dark:bg-zinc-900 border border-outline-variant/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 font-mono"
                          />
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Feedback notices */}
                  {createError && (
                    <div className="p-3 rounded-xl bg-destructive/10 border border-destructive/20 text-destructive text-xs font-medium flex items-center justify-between">
                      <span>{createError}</span>
                      <button type="button" onClick={() => setCreateError(null)}>
                        <XMarkIcon className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                  {createSuccessMsg && (
                    <div className="p-3 rounded-xl bg-emerald-500/10 border border-emerald-500/20 text-emerald-600 dark:text-emerald-400 text-xs font-medium flex items-center justify-between">
                      <span className="font-mono">{createSuccessMsg}</span>
                      <button type="button" onClick={() => setCreateSuccessMsg(null)}>
                        <XMarkIcon className="w-4 h-4" />
                      </button>
                    </div>
                  )}
                </form>
              </div>

              {/* Value Banner Callout */}
              <div className="mt-6 pt-4 border-t border-outline-variant/20 flex flex-col sm:flex-row sm:items-center justify-between gap-3 text-xs">
                <span className="text-muted-foreground">
                  Get custom vanity links, SLA uptime, and branded edge domains with Trunc Enterprise.
                </span>
                <Link
                  href="/billing"
                  className="text-primary hover:text-primary-600 font-bold tracking-tight inline-flex items-center gap-1 shrink-0"
                >
                  Upgrade now →
                </Link>
              </div>
            </div>

            {/* Right Card (4 cols): AI Workflow Assist */}
            <div className="lg:col-span-4 rounded-2xl p-6 bg-gradient-to-br from-indigo-900/30 via-surface-container-lowest to-surface-container border border-indigo-500/30 dark:border-indigo-500/20 shadow-sm flex flex-col justify-between relative overflow-hidden">
              <div className="absolute top-0 right-0 w-32 h-32 bg-indigo-500/10 rounded-full blur-2xl pointer-events-none"></div>
              <div>
                <div className="flex items-center justify-between mb-4">
                  <div className="flex items-center gap-2">
                    <div className="w-7 h-7 rounded-lg bg-indigo-500/20 text-indigo-400 flex items-center justify-center">
                      <SparklesIcon className="w-4 h-4" />
                    </div>
                    <span className="text-xs font-bold uppercase tracking-wider text-indigo-500 font-mono">
                      AI Workflow Assist
                    </span>
                  </div>
                  <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20">
                    Pro
                  </span>
                </div>

                <h3 className="font-extrabold text-lg text-foreground font-hanken mb-2">
                  Simplify your workflow
                </h3>
                <p className="text-xs text-muted-foreground mb-4">
                  Boost conversion rates with automated keyword suggestions and automated attribution categorization.
                </p>

                <div className="space-y-3">
                  <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-surface-container/60 dark:bg-zinc-900/50 border border-outline-variant/20">
                    <div className="w-5 h-5 rounded-md bg-indigo-500/10 text-indigo-500 flex items-center justify-center shrink-0 mt-0.5">
                      ✓
                    </div>
                    <div className="text-xs">
                      <strong className="text-foreground block">Personalize with contextual AI slug</strong>
                      <span className="text-[11px] text-muted-foreground">Generates catchy, high-CTR slugs in milliseconds.</span>
                    </div>
                  </div>
                  <div className="flex items-start gap-2.5 p-2.5 rounded-xl bg-surface-container/60 dark:bg-zinc-900/50 border border-outline-variant/20">
                    <div className="w-5 h-5 rounded-md bg-indigo-500/10 text-indigo-500 flex items-center justify-center shrink-0 mt-0.5">
                      ✓
                    </div>
                    <div className="text-xs">
                      <strong className="text-foreground block">Auto-tag and categorize marketing campaign</strong>
                      <span className="text-[11px] text-muted-foreground">Sorts links by intent, team, and social channel.</span>
                    </div>
                  </div>
                </div>
              </div>

              <div className="pt-6">
                <Link
                  href="/billing"
                  className="w-full py-2.5 px-4 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 active:scale-[0.98] text-white text-xs font-bold tracking-tight shadow-md transition-all flex items-center justify-center gap-2"
                >
                  <SparklesIcon className="w-4 h-4" />
                  <span>Upgrade to Create with AI</span>
                </Link>
              </div>
            </div>
          </div>
        </section>

        {/* SECTION 2: OPERATIONAL WIDGETS (TWO COLUMNS) */}
        <section className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          {/* Left (6 cols): Connect Your Account */}
          <div className="lg:col-span-6 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl p-6 shadow-sm">
            <div className="flex items-center justify-between pb-4 border-b border-outline-variant/20 mb-4">
              <div>
                <h3 className="font-bold text-sm text-foreground font-hanken">Connect your account</h3>
                <p className="text-xs text-muted-foreground">Supercharge your link creation inside your favorite tools</p>
              </div>
              <span className="text-xs text-primary font-semibold font-mono">3 integrations</span>
            </div>

            <div className="space-y-3">
              {/* Chrome Extension */}
              <div className="p-3.5 rounded-xl bg-surface-container dark:bg-zinc-900/60 border border-outline-variant/20 flex items-center justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-9 h-9 rounded-xl bg-amber-500/10 text-amber-500 border border-amber-500/20 flex items-center justify-center font-bold text-xs shrink-0">
                    CR
                  </div>
                  <div className="min-w-0">
                    <div className="text-xs font-bold text-foreground">Chrome Extension</div>
                    <div className="text-[11px] text-muted-foreground truncate">1-click short links from your browser bar</div>
                  </div>
                </div>
                <a
                  href="https://chrome.google.com/webstore"
                  target="_blank"
                  rel="noopener noreferrer"
                  className="px-3 py-1.5 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-800 hover:border-primary/50 text-foreground text-xs font-semibold shrink-0 transition-all"
                >
                  Install extension
                </a>
              </div>

              {/* Canva */}
              <div className="p-3.5 rounded-xl bg-surface-container dark:bg-zinc-900/60 border border-outline-variant/20 flex items-center justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-9 h-9 rounded-xl bg-cyan-500/10 text-cyan-500 border border-cyan-500/20 flex items-center justify-center font-bold text-xs shrink-0">
                    CV
                  </div>
                  <div className="min-w-0">
                    <div className="text-xs font-bold text-foreground">Canva Integration</div>
                    <div className="text-[11px] text-muted-foreground truncate">Export custom QR codes & branded banners</div>
                  </div>
                </div>
                <button
                  onClick={() => alert('Canva integration is configured for your team workspace.')}
                  className="px-3 py-1.5 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-800 hover:border-primary/50 text-foreground text-xs font-semibold shrink-0 transition-all"
                >
                  Connect
                </button>
              </div>

              {/* Shopify */}
              <div className="p-3.5 rounded-xl bg-surface-container dark:bg-zinc-900/60 border border-outline-variant/20 flex items-center justify-between gap-3">
                <div className="flex items-center gap-3 min-w-0">
                  <div className="w-9 h-9 rounded-xl bg-emerald-500/10 text-emerald-500 border border-emerald-500/20 flex items-center justify-center font-bold text-xs shrink-0">
                    SP
                  </div>
                  <div className="min-w-0">
                    <div className="text-xs font-bold text-foreground">Shopify Storefront</div>
                    <div className="text-[11px] text-muted-foreground truncate">Attribution telemetry & product link routing</div>
                  </div>
                </div>
                <button
                  onClick={() => alert('Shopify attribution bridge is active.')}
                  className="px-3 py-1.5 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-800 hover:border-primary/50 text-foreground text-xs font-semibold shrink-0 transition-all"
                >
                  Setup
                </button>
              </div>
            </div>
          </div>

          {/* Right (6 cols): Monthly Usage & Telemetry */}
          <div className="lg:col-span-6 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl p-6 shadow-sm flex flex-col justify-between">
            <div>
              <div className="flex items-center justify-between pb-4 border-b border-outline-variant/20 mb-4">
                <div>
                  <h3 className="font-bold text-sm text-foreground font-hanken">Monthly usage & Telemetry</h3>
                  <p className="text-xs text-muted-foreground">Current billing cycle • Renews in 18 days</p>
                </div>
                <span className="text-[10px] font-bold px-2 py-0.5 rounded-full bg-primary/10 text-primary font-mono">
                  Pro Plan
                </span>
              </div>

              {/* Telemetry Multi-bars */}
              <div className="space-y-3.5">
                {/* Short links */}
                <div>
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="font-medium text-foreground">Short links</span>
                    <span className="text-muted-foreground font-mono">
                      <strong className="text-foreground">{stats?.total_urls ?? urls.length}</strong> of 100 used
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-surface-container dark:bg-zinc-800 overflow-hidden">
                    <div
                      className="h-full bg-primary rounded-full transition-all duration-500"
                      style={{ width: `${Math.min(100, ((stats?.total_urls ?? urls.length) / 100) * 100)}%` }}
                    ></div>
                  </div>
                </div>

                {/* Custom back-halves */}
                <div>
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="font-medium text-foreground">Custom back-halves</span>
                    <span className="text-muted-foreground font-mono">
                      <strong className="text-foreground">4</strong> of 50 used
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-surface-container dark:bg-zinc-800 overflow-hidden">
                    <div className="h-full bg-indigo-500 rounded-full" style={{ width: '8%' }}></div>
                  </div>
                </div>

                {/* Dynamic QR Codes */}
                <div>
                  <div className="flex items-center justify-between text-xs mb-1">
                    <span className="font-medium text-foreground">Dynamic QR Codes</span>
                    <span className="text-muted-foreground font-mono">
                      <strong className="text-foreground">8</strong> of 100 used
                    </span>
                  </div>
                  <div className="w-full h-2 rounded-full bg-surface-container dark:bg-zinc-800 overflow-hidden">
                    <div className="h-full bg-emerald-500 rounded-full" style={{ width: '8%' }}></div>
                  </div>
                </div>
              </div>
            </div>

            {/* Performance Spark Stat */}
            <div className="mt-5 pt-4 border-t border-outline-variant/20 flex items-center justify-between gap-4">
              <div>
                <div className="flex items-baseline gap-2">
                  <span className="text-xl font-black text-foreground tabular-nums">
                    {stats?.total_clicks ? stats.total_clicks : 287}
                  </span>
                  <span className="text-xs text-muted-foreground">Total Clicks this week</span>
                  <span className="text-[11px] font-bold text-emerald-500 bg-emerald-500/10 px-1.5 py-0.2 rounded font-mono">
                    +24.8%
                  </span>
                </div>
                <div className="text-[11px] text-muted-foreground mt-0.5">Peak traffic: 14:00 - 16:00 UTC</div>
              </div>
              <Link
                href="/analytics"
                className="px-3 py-1.5 rounded-xl border border-outline-variant/30 text-xs font-semibold text-foreground hover:bg-surface-container transition-all shrink-0"
              >
                View trends
              </Link>
            </div>
          </div>
        </section>

        {/* SECTION 3: RECENT LINKS MANAGEMENT STREAM */}
        <section id="links" className="bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl shadow-sm overflow-hidden">
          {/* Header */}
          <div className="p-5 border-b border-outline-variant/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div className="flex items-center gap-2">
                <h2 className="text-base font-bold text-foreground font-hanken">Your Links</h2>
                <span className="text-xs px-2 py-0.5 rounded-full bg-surface-container dark:bg-zinc-800 text-muted-foreground font-mono font-semibold">
                  {filteredUrls.length} links
                </span>
              </div>
              <p className="text-xs text-muted-foreground mt-0.5">
                Real-time stream of all shortened links and QR destinations
              </p>
            </div>

            <div className="flex items-center gap-2.5">
              <button
                onClick={() => exportCSV(false)}
                className="px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-900 hover:bg-surface-container text-foreground text-xs font-semibold transition-all flex items-center gap-1.5"
              >
                <ArrowDownTrayIcon className="w-3.5 h-3.5 text-muted-foreground" />
                <span>Export CSV</span>
              </button>
              <button
                onClick={() => {
                  const el = document.getElementById('quick-create-section');
                  el?.scrollIntoView({ behavior: 'smooth' });
                  el?.querySelector('input')?.focus();
                }}
                className="px-3.5 py-2 rounded-xl bg-primary hover:bg-primary-600 text-white text-xs font-bold transition-all shadow-sm flex items-center gap-1.5"
              >
                <PlusIcon className="w-3.5 h-3.5 stroke-[2.5]" />
                <span>Create link</span>
              </button>
            </div>
          </div>

          {/* Filter & Search Ribbon */}
          <div className="p-4 border-b border-outline-variant/20 bg-surface-container-low/50 dark:bg-zinc-900/30 flex flex-wrap items-center justify-between gap-3">
            <div className="flex flex-wrap items-center gap-2.5 flex-1 min-w-0">
              {/* Search input */}
              <div className="relative w-full sm:w-64">
                <input
                  type="text"
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  placeholder="Filter links or domain..."
                  className="w-full px-3 py-1.5 rounded-xl text-xs bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 font-sans"
                />
              </div>

              {/* Date Filter */}
              <select
                value={dateFilter}
                onChange={(e) => setDateFilter(e.target.value as 'all' | 'today' | 'week' | 'month')}
                className="px-3 py-1.5 rounded-xl text-xs bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 text-foreground focus:outline-none cursor-pointer"
              >
                <option value="all">Filter by created date (All)</option>
                <option value="today">Past 24 hours</option>
                <option value="week">Past 7 days</option>
                <option value="month">Past 30 days</option>
              </select>

              {/* Status Filter */}
              <select
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value as 'all' | 'active' | 'archived')}
                className="px-3 py-1.5 rounded-xl text-xs bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 text-foreground focus:outline-none cursor-pointer"
              >
                <option value="active">Active links</option>
                <option value="archived">Archived</option>
                <option value="all">All statuses</option>
              </select>

              {/* Add Filters Button & Popover */}
              <div className="relative" ref={filtersMenuRef}>
                <button
                  type="button"
                  onClick={() => setShowFiltersMenu(!showFiltersMenu)}
                  className={`px-3 py-1.5 rounded-xl text-xs font-semibold border transition-all flex items-center gap-1.5 ${
                    (tagFilter !== 'all' || domainFilter !== 'all' || minClicksFilter > 0)
                      ? 'bg-primary/10 border-primary/30 text-primary'
                      : 'bg-surface-container-lowest dark:bg-zinc-950 border-outline-variant/30 text-foreground hover:bg-surface-container'
                  }`}
                >
                  <AdjustmentsHorizontalIcon className="w-3.5 h-3.5" />
                  <span>Add filters</span>
                  {(tagFilter !== 'all' || domainFilter !== 'all' || minClicksFilter > 0) && (
                    <span className="w-4 h-4 rounded-full bg-primary text-white text-[10px] flex items-center justify-center font-mono">
                      {(tagFilter !== 'all' ? 1 : 0) + (domainFilter !== 'all' ? 1 : 0) + (minClicksFilter > 0 ? 1 : 0)}
                    </span>
                  )}
                </button>

                {showFiltersMenu && (
                  <div className="absolute left-0 mt-2 w-72 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl shadow-xl p-4 z-40 text-xs space-y-3 animate-in fade-in zoom-in-95 duration-150">
                    <div className="flex items-center justify-between pb-2 border-b border-outline-variant/20">
                      <span className="font-bold text-foreground font-hanken">Filter Stream</span>
                      {(tagFilter !== 'all' || domainFilter !== 'all' || minClicksFilter > 0) && (
                        <button
                          onClick={() => {
                            setTagFilter('all');
                            setDomainFilter('all');
                            setMinClicksFilter(0);
                          }}
                          className="text-[11px] text-primary hover:underline font-semibold"
                        >
                          Reset all
                        </button>
                      )}
                    </div>

                    <div>
                      <label className="block text-muted-foreground font-semibold mb-1">Tag</label>
                      <select
                        value={tagFilter}
                        onChange={(e) => setTagFilter(e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground cursor-pointer text-xs"
                      >
                        <option value="all">All tags</option>
                        <option value="#Engineering">#Engineering</option>
                        <option value="#Marketing">#Marketing</option>
                        <option value="#Campaign">#Campaign</option>
                        <option value="#Docs">#Docs</option>
                        <option value="#Ecommerce">#Ecommerce</option>
                        <option value="#API">#API</option>
                        <option value="#General">#General</option>
                      </select>
                    </div>

                    <div>
                      <label className="block text-muted-foreground font-semibold mb-1">Domain</label>
                      <select
                        value={domainFilter}
                        onChange={(e) => setDomainFilter(e.target.value)}
                        className="w-full px-2.5 py-1.5 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground cursor-pointer text-xs"
                      >
                        <option value="all">All domains</option>
                        <option value="trunc.site">trunc.site (Standard)</option>
                        {customDomains.map(d => (
                          <option key={d.id} value={d.domain}>{d.domain}</option>
                        ))}
                      </select>
                    </div>

                    <div>
                      <label className="block text-muted-foreground font-semibold mb-1">Engagements / Clicks</label>
                      <select
                        value={minClicksFilter}
                        onChange={(e) => setMinClicksFilter(Number(e.target.value))}
                        className="w-full px-2.5 py-1.5 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground cursor-pointer text-xs"
                      >
                        <option value={0}>Any engagement (0+)</option>
                        <option value={1}>At least 1 click</option>
                        <option value={10}>10+ clicks</option>
                        <option value={50}>50+ clicks</option>
                      </select>
                    </div>
                  </div>
                )}
              </div>
            </div>

            {/* View Toggles */}
            <div className="flex items-center gap-1 bg-surface-container dark:bg-zinc-900 p-1 rounded-xl border border-outline-variant/20">
              <button
                onClick={() => setViewMode('list')}
                className={`p-1.5 rounded-lg text-xs transition-colors ${
                  viewMode === 'list'
                    ? 'bg-surface-container-lowest dark:bg-zinc-950 text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
                title="List View"
              >
                <Bars3BottomLeftIcon className="w-4 h-4" />
              </button>
              <button
                onClick={() => setViewMode('grid')}
                className={`p-1.5 rounded-lg text-xs transition-colors ${
                  viewMode === 'grid'
                    ? 'bg-surface-container-lowest dark:bg-zinc-950 text-foreground shadow-sm'
                    : 'text-muted-foreground hover:text-foreground'
                }`}
                title="Grid View"
              >
                <Squares2X2Icon className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* Bulk Action Bar */}
          <div className="px-5 py-2.5 bg-surface-container/80 dark:bg-zinc-900/60 border-b border-outline-variant/20 flex flex-wrap items-center justify-between gap-3 text-xs">
            <div className="flex items-center gap-3">
              <input
                type="checkbox"
                checked={filteredUrls.length > 0 && selectedIds.size === filteredUrls.length}
                onChange={(e) => handleSelectAll(e.target.checked)}
                className="rounded border-outline-variant text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                title="Select all"
              />
              <span className="font-semibold text-foreground font-mono">
                {selectedIds.size} selected
              </span>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => exportCSV(selectedIds.size > 0)}
                className="px-2.5 py-1 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 hover:bg-surface-container text-foreground font-medium flex items-center gap-1.5 transition-colors"
                title={selectedIds.size > 0 ? 'Export selected links to CSV' : 'Export all visible links to CSV'}
              >
                <ArrowDownTrayIcon className="w-3.5 h-3.5 text-muted-foreground" />
                <span>Export</span>
              </button>
              <button
                onClick={handleBulkHide}
                disabled={selectedIds.size === 0}
                className="px-2.5 py-1 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 hover:bg-surface-container text-foreground font-medium disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                title="Hide / archive selected links"
              >
                Hide
              </button>
              <button
                onClick={() => {
                  if (selectedIds.size > 0) {
                    setShowBulkTagModal(true);
                  }
                }}
                disabled={selectedIds.size === 0}
                className="px-2.5 py-1 rounded-lg border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 hover:bg-surface-container text-foreground font-medium disabled:opacity-40 disabled:cursor-not-allowed flex items-center gap-1 transition-colors"
                title="Tag selected links"
              >
                <TagIcon className="w-3.5 h-3.5 text-muted-foreground" />
                <span>Tag</span>
              </button>
              {selectedIds.size > 0 && (
                <button
                  onClick={() => setSelectedIds(new Set())}
                  className="text-muted-foreground hover:text-foreground underline ml-1 text-xs"
                >
                  Clear
                </button>
              )}
            </div>
          </div>

          {/* Link Stream Feed */}
          {loadingUrls ? (
            <div className="p-8 space-y-4">
              {[1, 2, 3, 4].map(i => (
                <div key={i} className="p-4 rounded-xl bg-surface-container/50 animate-pulse flex items-center justify-between">
                  <div className="space-y-2 flex-1">
                    <div className="h-4 bg-outline-variant/40 rounded w-1/3"></div>
                    <div className="h-3 bg-outline-variant/30 rounded w-2/3"></div>
                  </div>
                  <div className="w-20 h-4 bg-outline-variant/40 rounded"></div>
                </div>
              ))}
            </div>
          ) : filteredUrls.length === 0 ? (
            <div className="p-12 text-center">
              <LinkIcon className="mx-auto h-12 w-12 text-muted-foreground/60 mb-3" />
              <h3 className="text-sm font-bold text-foreground">No links found</h3>
              <p className="text-xs text-muted-foreground mt-1 max-w-sm mx-auto">
                {searchQuery
                  ? `No links match your filter "${searchQuery}". Try clearing search.`
                  : 'Start shortening links above to build your campaign telemetry stream.'}
              </p>
              {searchQuery && (
                <button
                  onClick={() => setSearchQuery('')}
                  className="mt-3 px-3 py-1.5 rounded-xl text-xs font-semibold bg-surface-container text-foreground hover:bg-surface-container-high"
                >
                  Clear search
                </button>
              )}
            </div>
          ) : viewMode === 'list' ? (
            /* LIST VIEW */
            <div className="divide-y divide-outline-variant/20">
              {filteredUrls.map(url => {
                const platform = getPlatformIcon(url.original_url);
                const tags = getTagsForUrl(url);
                const isSelected = selectedIds.has(url.id);
                const isCopied = copiedId === url.id;

                return (
                  <div
                    key={url.id}
                    className={`p-4 hover:bg-surface-container/50 dark:hover:bg-zinc-900/50 transition-colors flex flex-col md:flex-row md:items-center justify-between gap-4 ${
                      isSelected ? 'bg-primary/5 dark:bg-primary/10' : ''
                    }`}
                  >
                    {/* Left: Checkbox + Monogram + Title + Short/Long URL */}
                    <div className="flex items-start gap-3 min-w-0 flex-1">
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => handleToggleSelect(url.id)}
                        className="rounded border-outline-variant text-primary focus:ring-primary w-4 h-4 cursor-pointer mt-2.5 shrink-0"
                      />

                      {/* Source Platform Favicon Monogram */}
                      <div
                        className={`w-9 h-9 rounded-xl flex items-center justify-center font-bold text-xs shrink-0 mt-0.5 ${platform.bg}`}
                        title={platform.name}
                      >
                        {platform.label}
                      </div>

                      <div className="min-w-0 flex-1">
                        {/* Title and Tags */}
                        <div className="flex flex-wrap items-center gap-2">
                          <span className="font-bold text-sm text-foreground hover:text-primary cursor-pointer truncate font-hanken">
                            {url.title || url.short_code}
                          </span>
                          {tags.map(t => (
                            <span
                              key={t}
                              className="text-[10px] font-semibold px-2 py-0.2 rounded-full bg-surface-container text-muted-foreground font-mono"
                            >
                              {t}
                            </span>
                          ))}
                        </div>

                        {/* Short URL with 1-click copy */}
                        <div className="flex items-center gap-2 mt-1">
                          <a
                            href={url.short_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="font-bold text-primary hover:underline text-xs font-mono"
                          >
                            {url.short_url}
                          </a>
                          <button
                            onClick={() => copyToClipboard(url.short_url, url.id)}
                            className={`px-2 py-0.5 rounded text-[10px] font-semibold flex items-center gap-1 transition-all ${
                              isCopied
                                ? 'bg-emerald-600 text-white'
                                : 'bg-surface-container dark:bg-zinc-800 hover:bg-surface-container-high text-foreground border border-outline-variant/30'
                            }`}
                            title="Copy short link"
                          >
                            {isCopied ? (
                              <>
                                <CheckIcon className="w-3 h-3 stroke-[2.5]" />
                                <span>Copied!</span>
                              </>
                            ) : (
                              <>
                                <ClipboardDocumentIcon className="w-3 h-3" />
                                <span>Copy</span>
                              </>
                            )}
                          </button>
                          <a
                            href={url.short_url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="p-1 text-muted-foreground hover:text-foreground"
                            title="Open short link"
                          >
                            <ArrowTopRightOnSquareIcon className="w-3.5 h-3.5" />
                          </a>
                        </div>

                        {/* Destination URL & Created Date */}
                        <div className="flex items-center gap-2 text-xs text-muted-foreground mt-1 truncate">
                          <span className="text-[11px] truncate max-w-sm">{url.original_url}</span>
                          <span>•</span>
                          <span className="text-[10px] font-mono shrink-0">
                            {new Date(url.created_at).toLocaleDateString('en-US', {
                              month: 'short',
                              day: 'numeric',
                              year: 'numeric',
                            })}
                          </span>
                        </div>
                      </div>
                    </div>

                    {/* Right: Telemetry Engagement & Action Toolbar */}
                    <div className="flex items-center gap-5 justify-between md:justify-end shrink-0 pl-11 md:pl-0">
                      {/* Telemetry Engagement Count */}
                      <div className="text-right">
                        <div className="text-sm font-black text-foreground tabular-nums">
                          {url.click_count}{' '}
                          <span className="text-xs font-normal text-muted-foreground">clicks</span>
                        </div>
                        <div className="w-20 h-1.5 bg-surface-container dark:bg-zinc-800 rounded-full mt-1 overflow-hidden">
                          <div
                            className="h-full bg-primary rounded-full"
                            style={{ width: `${Math.min(100, (url.click_count / 50) * 100)}%` }}
                          ></div>
                        </div>
                      </div>

                      {/* Action Toolbar */}
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => router.push(`/analytics?shortCode=${url.short_code}`)}
                          onMouseEnter={() => {
                            router.prefetch('/analytics');
                            const token = getAccessToken();
                            if (token) {
                              prefetch(`analytics:url:${url.short_code}`, () =>
                                fetch(`${API_BASE_URL}/api/v1/analytics/${url.short_code}`, {
                                  headers: { Authorization: `Bearer ${token}` },
                                }).then(r => (r.ok ? r.json() : Promise.reject(r.status)))
                              );
                            }
                          }}
                          className="p-2 rounded-lg bg-surface-container dark:bg-zinc-800/80 hover:bg-surface-container-high text-muted-foreground hover:text-foreground text-xs transition-colors"
                          title="View Analytics"
                        >
                          <ChartBarIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => setQrModalUrl(url)}
                          className="p-2 rounded-lg bg-surface-container dark:bg-zinc-800/80 hover:bg-surface-container-high text-muted-foreground hover:text-foreground text-xs transition-colors"
                          title="Generate QR Code"
                        >
                          <QrCodeIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => {
                            if (navigator.share) {
                              navigator.share({
                                title: url.title || 'Short link',
                                url: url.short_url,
                              }).catch(() => {});
                            } else {
                              copyToClipboard(url.short_url, url.id);
                            }
                          }}
                          className="p-2 rounded-lg bg-surface-container dark:bg-zinc-800/80 hover:bg-surface-container-high text-muted-foreground hover:text-foreground text-xs transition-colors"
                          title="Share"
                        >
                          <ShareIcon className="w-4 h-4" />
                        </button>
                        <button
                          onClick={() => {
                            setEditingUrl(url);
                            setEditError(null);
                            setEditExpiresAt(
                              url.expires_at ? new Date(url.expires_at).toISOString().slice(0, 16) : ''
                            );
                          }}
                          className="p-2 rounded-lg bg-surface-container dark:bg-zinc-800/80 hover:bg-surface-container-high text-muted-foreground hover:text-foreground text-xs transition-colors"
                          title="Edit link"
                        >
                          <PencilIcon className="w-4 h-4" />
                        </button>
                        <div className="relative">
                          <button
                            onClick={() => setOpenMenuUrlId(openMenuUrlId === url.id ? null : url.id)}
                            className="p-2 rounded-lg bg-surface-container dark:bg-zinc-800/80 hover:bg-surface-container-high text-muted-foreground hover:text-foreground text-xs transition-colors"
                            title="More options"
                          >
                            <EllipsisVerticalIcon className="w-4 h-4" />
                          </button>
                          {openMenuUrlId === url.id && (
                            <div
                              ref={actionMenuRef}
                              className="absolute right-0 top-full mt-1 w-44 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-xl shadow-xl py-1 z-30 text-xs animate-in fade-in zoom-in-95 duration-100"
                            >
                              <button
                                onClick={() => copyMarkdown(url)}
                                className="w-full flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 text-left"
                              >
                                <CodeBracketIcon className="w-3.5 h-3.5 text-muted-foreground" />
                                <span>Copy Markdown</span>
                              </button>
                              <button
                                onClick={() => toggleUrlActive(url.short_code, url.is_active)}
                                className="w-full flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 text-left"
                              >
                                {url.is_active ? (
                                  <>
                                    <PauseIcon className="w-3.5 h-3.5 text-amber-500" />
                                    <span>Pause Link</span>
                                  </>
                                ) : (
                                  <>
                                    <PlayIcon className="w-3.5 h-3.5 text-emerald-500" />
                                    <span>Activate Link</span>
                                  </>
                                )}
                              </button>
                              <div className="border-t border-outline-variant/20 my-1"></div>
                              <button
                                onClick={() => {
                                  setOpenMenuUrlId(null);
                                  if (confirm(`Are you sure you want to delete ${url.short_code}?`)) {
                                    deleteURL(url.short_code);
                                  }
                                }}
                                className="w-full flex items-center gap-2 px-3 py-2 text-destructive hover:bg-destructive/10 text-left"
                              >
                                <TrashIcon className="w-3.5 h-3.5" />
                                <span>Delete link</span>
                              </button>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            /* GRID VIEW */
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 p-5">
              {filteredUrls.map(url => {
                const platform = getPlatformIcon(url.original_url);
                const tags = getTagsForUrl(url);
                const isSelected = selectedIds.has(url.id);
                const isCopied = copiedId === url.id;

                return (
                  <div
                    key={url.id}
                    className={`p-4 rounded-xl border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-900/60 flex flex-col justify-between hover:border-primary/50 transition-all ${
                      isSelected ? 'ring-2 ring-primary/40' : ''
                    }`}
                  >
                    <div>
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <input
                            type="checkbox"
                            checked={isSelected}
                            onChange={() => handleToggleSelect(url.id)}
                            className="rounded border-outline-variant text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                          />
                          <div className={`w-7 h-7 rounded-lg flex items-center justify-center font-bold text-xs ${platform.bg}`}>
                            {platform.label}
                          </div>
                        </div>
                        <span className="text-xs font-bold text-foreground tabular-nums font-mono">
                          {url.click_count} clicks
                        </span>
                      </div>

                      <h4 className="font-bold text-sm text-foreground truncate font-hanken">
                        {url.title || url.short_code}
                      </h4>
                      <p className="text-xs text-muted-foreground truncate mt-0.5">{url.original_url}</p>

                      <div className="flex flex-wrap gap-1 mt-2">
                        {tags.map(t => (
                          <span
                            key={t}
                            className="text-[9px] font-semibold px-1.5 py-0.2 rounded-full bg-surface-container text-muted-foreground font-mono"
                          >
                            {t}
                          </span>
                        ))}
                      </div>

                      <div className="flex items-center gap-1.5 mt-3">
                        <a
                          href={url.short_url}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-xs font-bold text-primary font-mono truncate hover:underline"
                        >
                          {url.short_url}
                        </a>
                        <button
                          onClick={() => copyToClipboard(url.short_url, url.id)}
                          className="p-1 rounded text-muted-foreground hover:text-foreground shrink-0"
                          title="Copy"
                        >
                          {isCopied ? <CheckIcon className="w-3.5 h-3.5 text-emerald-500" /> : <ClipboardDocumentIcon className="w-3.5 h-3.5" />}
                        </button>
                      </div>
                    </div>

                    <div className="mt-4 pt-3 border-t border-outline-variant/20 flex items-center justify-between">
                      <span className="text-[10px] text-muted-foreground font-mono">
                        {new Date(url.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                      </span>
                      <div className="flex items-center gap-1">
                        <button
                          onClick={() => setQrModalUrl(url)}
                          className="p-1.5 rounded-lg hover:bg-surface-container text-muted-foreground hover:text-foreground"
                          title="QR Code"
                        >
                          <QrCodeIcon className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => router.push(`/analytics?shortCode=${url.short_code}`)}
                          className="p-1.5 rounded-lg hover:bg-surface-container text-muted-foreground hover:text-foreground"
                          title="Analytics"
                        >
                          <ChartBarIcon className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => {
                            if (navigator.share) {
                              navigator.share({
                                title: url.title || 'Short link',
                                url: url.short_url,
                              }).catch(() => {});
                            } else {
                              copyToClipboard(url.short_url, url.id);
                            }
                          }}
                          className="p-1.5 rounded-lg hover:bg-surface-container text-muted-foreground hover:text-foreground"
                          title="Share"
                        >
                          <ShareIcon className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => {
                            setEditingUrl(url);
                            setEditError(null);
                            setEditExpiresAt(url.expires_at ? new Date(url.expires_at).toISOString().slice(0, 16) : '');
                          }}
                          className="p-1.5 rounded-lg hover:bg-surface-container text-muted-foreground hover:text-foreground"
                          title="Edit"
                        >
                          <PencilIcon className="w-3.5 h-3.5" />
                        </button>
                        <div className="relative">
                          <button
                            onClick={() => setOpenMenuUrlId(openMenuUrlId === url.id ? null : url.id)}
                            className="p-1.5 rounded-lg hover:bg-surface-container text-muted-foreground hover:text-foreground"
                            title="More options"
                          >
                            <EllipsisVerticalIcon className="w-3.5 h-3.5" />
                          </button>
                          {openMenuUrlId === url.id && (
                            <div
                              ref={actionMenuRef}
                              className="absolute right-0 bottom-full mb-1 w-44 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-xl shadow-xl py-1 z-30 text-xs animate-in fade-in zoom-in-95 duration-100"
                            >
                              <button
                                onClick={() => copyMarkdown(url)}
                                className="w-full flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 text-left"
                              >
                                <CodeBracketIcon className="w-3.5 h-3.5 text-muted-foreground" />
                                <span>Copy Markdown</span>
                              </button>
                              <button
                                onClick={() => toggleUrlActive(url.short_code, url.is_active)}
                                className="w-full flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 text-left"
                              >
                                {url.is_active ? (
                                  <>
                                    <PauseIcon className="w-3.5 h-3.5 text-amber-500" />
                                    <span>Pause Link</span>
                                  </>
                                ) : (
                                  <>
                                    <PlayIcon className="w-3.5 h-3.5 text-emerald-500" />
                                    <span>Activate Link</span>
                                  </>
                                )}
                              </button>
                              <div className="border-t border-outline-variant/20 my-1"></div>
                              <button
                                onClick={() => {
                                  setOpenMenuUrlId(null);
                                  if (confirm(`Are you sure you want to delete ${url.short_code}?`)) {
                                    deleteURL(url.short_code);
                                  }
                                }}
                                className="w-full flex items-center gap-2 px-3 py-2 text-destructive hover:bg-destructive/10 text-left"
                              >
                                <TrashIcon className="w-3.5 h-3.5" />
                                <span>Delete link</span>
                              </button>
                            </div>
                          )}
                        </div>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          )}

          {/* Pagination Footer */}
          <div className="p-4 border-t border-outline-variant/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4 text-xs text-muted-foreground">
            <div className="flex items-center gap-2">
              <span>
                Showing <strong>{filteredUrls.length > 0 ? (currentPage - 1) * pageSize + 1 : 0}</strong> to{' '}
                <strong>{Math.min(currentPage * pageSize, filteredUrls.length)}</strong> of{' '}
                <strong>{filteredUrls.length}</strong> links
              </span>
              <div className="flex items-center gap-1 ml-3">
                <span>Rows:</span>
                <select
                  value={pageSize}
                  onChange={(e) => {
                    setPageSize(Number(e.target.value));
                    setCurrentPage(1);
                  }}
                  className="px-2 py-1 rounded-lg bg-surface-container border border-outline-variant/30 text-foreground cursor-pointer text-xs"
                >
                  <option value={10}>10</option>
                  <option value={20}>20</option>
                  <option value={50}>50</option>
                </select>
              </div>
            </div>

            <div className="flex items-center gap-2">
              <button
                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                disabled={currentPage === 1 || loadingUrls}
                className="px-3 py-1.5 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground hover:bg-surface-container-high disabled:opacity-40 disabled:cursor-not-allowed font-semibold transition-all"
              >
                Previous
              </button>
              <span className="font-mono px-2">
                Page {currentPage} of {Math.max(1, totalPages)}
              </span>
              <button
                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                disabled={currentPage === totalPages || loadingUrls}
                className="px-3 py-1.5 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground hover:bg-surface-container-high disabled:opacity-40 disabled:cursor-not-allowed font-semibold transition-all"
              >
                Next
              </button>
            </div>
          </div>
        </section>
      </div>

      {/* CLARITY UI EDIT URL MODAL */}
      {editingUrl && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-in fade-in duration-150">
          <div className="bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl w-full max-w-lg overflow-hidden shadow-2xl flex flex-col max-h-[90vh] text-on-surface">
            {/* Modal Header */}
            <div className="p-5 border-b border-outline-variant/20 flex items-center justify-between shrink-0">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-xl bg-primary/10 text-primary flex items-center justify-center">
                  <PencilIcon className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-base font-bold text-foreground font-hanken">Edit Link Destination</h3>
                  <span className="text-xs font-mono text-muted-foreground">{editingUrl.short_code}</span>
                </div>
              </div>
              <button
                onClick={() => {
                  setEditingUrl(null);
                  setEditError(null);
                  setEditExpiresAt('');
                }}
                className="p-1.5 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900"
              >
                <XMarkIcon className="w-5 h-5" />
              </button>
            </div>

            {/* Modal Form */}
            <form
              className="flex flex-col flex-1 min-h-0"
              onSubmit={(e) => {
                e.preventDefault();
                const formData = new FormData(e.currentTarget);
                const originalUrl = formData.get('original_url') as string;
                const teamIdRaw = formData.get('team_id') as string;

                if (!originalUrl || (!originalUrl.startsWith('http://') && !originalUrl.startsWith('https://'))) {
                  setEditError('Destination URL must start with http:// or https://');
                  return;
                }

                const updates = {
                  original_url: originalUrl,
                  title: (formData.get('title') as string) || undefined,
                  description: (formData.get('description') as string) || undefined,
                  is_public: formData.get('is_public') === 'on',
                  expires_at: editExpiresAt ? new Date(editExpiresAt).toISOString() : null,
                  ...(userTeams.length > 0 && {
                    team_id: teamIdRaw === '' ? null : Number(teamIdRaw),
                  }),
                };
                updateURL(editingUrl.short_code, updates);
              }}
            >
              <div className="p-6 space-y-4 overflow-y-auto flex-1 text-xs">
                {/* Destination URL */}
                <div>
                  <label className="block font-semibold text-foreground mb-1">
                    Destination URL <span className="text-destructive">*</span>
                  </label>
                  <input
                    type="text"
                    name="original_url"
                    required
                    defaultValue={editingUrl.original_url}
                    className="w-full px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                    placeholder="https://example.com/target-page"
                  />
                </div>

                {/* Title */}
                <div>
                  <label className="block font-semibold text-foreground mb-1">Title (optional)</label>
                  <input
                    type="text"
                    name="title"
                    defaultValue={editingUrl.title || ''}
                    className="w-full px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                    placeholder="Campaign or page title"
                  />
                </div>

                {/* Description */}
                <div>
                  <label className="block font-semibold text-foreground mb-1">Description (optional)</label>
                  <textarea
                    name="description"
                    defaultValue={editingUrl.description || ''}
                    rows={2}
                    className="w-full px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                    placeholder="Short summary for attribution notes"
                  />
                </div>

                {/* Expiry Date */}
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <label className="block font-semibold text-foreground">Expiry date (optional)</label>
                    {editExpiresAt && (
                      <button
                        type="button"
                        onClick={() => setEditExpiresAt('')}
                        className="text-xs text-destructive hover:underline font-medium"
                      >
                        Clear expiry
                      </button>
                    )}
                  </div>
                  <input
                    ref={expiresAtRef}
                    type="datetime-local"
                    name="expires_at"
                    value={editExpiresAt}
                    onChange={(e) => setEditExpiresAt(e.target.value)}
                    onClick={() => expiresAtRef.current?.showPicker?.()}
                    className="w-full px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 cursor-pointer font-mono"
                  />
                </div>

                {/* Is Public */}
                <div className="flex items-center gap-2 pt-1">
                  <input
                    type="checkbox"
                    name="is_public"
                    id="edit-is-public"
                    defaultChecked={editingUrl.is_public}
                    className="rounded border-outline-variant text-primary focus:ring-primary w-4 h-4 cursor-pointer"
                  />
                  <label htmlFor="edit-is-public" className="font-semibold text-foreground cursor-pointer">
                    Make this URL publicly discoverable
                  </label>
                </div>

                {/* Team */}
                {userTeams.length > 0 && (
                  <div>
                    <label className="block font-semibold text-foreground mb-1">Assign to team</label>
                    <select
                      name="team_id"
                      defaultValue={editingUrl.team_id ?? '0'}
                      className="w-full px-3 py-2 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 text-foreground focus:outline-none focus:ring-2 focus:ring-primary/40"
                    >
                      <option value="0">Personal workspace</option>
                      {userTeams.map(t => (
                        <option key={t.id} value={t.id}>
                          {t.display_name || t.name}
                        </option>
                      ))}
                    </select>
                  </div>
                )}

                {editError && (
                  <p className="text-xs text-destructive bg-destructive/10 border border-destructive/20 rounded-xl px-3 py-2 font-medium">
                    {editError}
                  </p>
                )}
              </div>

              {/* Modal Footer */}
              <div className="p-5 border-t border-outline-variant/20 flex items-center justify-end gap-3 shrink-0">
                <button
                  type="button"
                  onClick={() => {
                    setEditingUrl(null);
                    setEditError(null);
                    setEditExpiresAt('');
                  }}
                  className="px-4 py-2 rounded-xl border border-outline-variant/30 bg-surface-container hover:bg-surface-container-high text-foreground text-xs font-semibold"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-xl bg-primary hover:bg-primary-600 text-white text-xs font-semibold shadow-sm transition-all"
                >
                  Save changes
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* CLARITY UI QR CODE MODAL */}
      {qrModalUrl && (
        <QRCodeModal
          isOpen={true}
          onClose={() => setQrModalUrl(null)}
          value={qrModalUrl.short_url}
          title={`QR Code for ${qrModalUrl.short_code}`}
          description={qrModalUrl.short_url}
          size={240}
        />
      )}

      {/* CLARITY UI BULK TAG MODAL */}
      {showBulkTagModal && (
        <div className="fixed inset-0 bg-black/60 backdrop-blur-sm flex items-center justify-center z-50 p-4 animate-in fade-in duration-150">
          <div className="bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl w-full max-w-sm overflow-hidden shadow-2xl p-5 text-on-surface">
            <div className="flex items-center justify-between pb-3 border-b border-outline-variant/20 mb-4">
              <div className="flex items-center gap-2">
                <TagIcon className="w-5 h-5 text-primary" />
                <h3 className="font-bold text-sm text-foreground font-hanken">Tag {selectedIds.size} Links</h3>
              </div>
              <button
                onClick={() => {
                  setShowBulkTagModal(false);
                  setBulkTagInput('');
                }}
                className="p-1 rounded-lg text-muted-foreground hover:text-foreground"
              >
                <XMarkIcon className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-muted-foreground mb-3">
              Select a suggested campaign tag or type a custom tag alias to categorize selected links:
            </p>

            <div className="flex flex-wrap gap-1.5 mb-4">
              {['#Engineering', '#Marketing', '#Campaign', '#Docs', '#Sales', '#Q3-Promo'].map(t => (
                <button
                  key={t}
                  type="button"
                  onClick={() => setBulkTagInput(t)}
                  className={`px-2 py-1 rounded-lg text-[11px] font-mono font-semibold transition-all ${
                    bulkTagInput === t
                      ? 'bg-primary text-white'
                      : 'bg-surface-container text-muted-foreground hover:text-foreground hover:bg-surface-container-high'
                  }`}
                >
                  {t}
                </button>
              ))}
            </div>

            <div className="mb-4">
              <input
                type="text"
                value={bulkTagInput}
                onChange={(e) => setBulkTagInput(e.target.value)}
                placeholder="e.g. #BlackFriday"
                className="w-full px-3 py-2 rounded-xl text-xs bg-surface-container dark:bg-zinc-900 border border-outline-variant/30 text-foreground font-mono focus:outline-none focus:ring-2 focus:ring-primary/40"
              />
            </div>

            <div className="flex justify-end gap-2 text-xs">
              <button
                type="button"
                onClick={() => {
                  setShowBulkTagModal(false);
                  setBulkTagInput('');
                }}
                className="px-3 py-1.5 rounded-xl border border-outline-variant/30 text-foreground hover:bg-surface-container"
              >
                Cancel
              </button>
              <button
                type="button"
                disabled={!bulkTagInput.trim()}
                onClick={() => handleApplyBulkTag(bulkTagInput)}
                className="px-3 py-1.5 rounded-xl bg-primary hover:bg-primary-600 disabled:opacity-40 text-white font-semibold shadow-sm transition-all"
              >
                Apply Tag
              </button>
            </div>
          </div>
        </div>
      )}
    </WorkspaceLayout>
  );
}

export default function DashboardPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen flex items-center justify-center bg-surface-container-lowest dark:bg-[#07090e]">
          <div className="flex flex-col items-center gap-3">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
            <span className="text-xs font-mono text-muted-foreground">Loading workspace...</span>
          </div>
        </div>
      }
    >
      <DashboardContent />
    </Suspense>
  );
}