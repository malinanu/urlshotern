'use client';

import { useState, useEffect, useRef } from 'react';
import Link from 'next/link';
import { usePathname, useRouter } from 'next/navigation';
import { useAuth } from '@/contexts/AuthContext';
import ThemeToggle from '@/components/ThemeToggle';
import {
  BoltIcon,
  HomeIcon,
  LinkIcon,
  QrCodeIcon,
  DocumentDuplicateIcon,
  ChartBarIcon,
  MegaphoneIcon,
  GlobeAltIcon,
  PuzzlePieceIcon,
  Cog6ToothIcon,
  MagnifyingGlassIcon,
  QuestionMarkCircleIcon,
  SparklesIcon,
  PlusIcon,
  Bars3Icon,
  XMarkIcon,
  ArrowRightOnRectangleIcon,
  ChevronDownIcon,
  CreditCardIcon,
} from '@heroicons/react/24/outline';

interface WorkspaceLayoutProps {
  children: React.ReactNode;
  activeNav?: string;
  linkCount?: number;
  onCreateClick?: () => void;
  onNavClick?: (navId: string) => void;
  searchQuery?: string;
  onSearchChange?: (val: string) => void;
  searchPlaceholder?: string;
}

export default function WorkspaceLayout({
  children,
  activeNav = 'home',
  linkCount = 14,
  onCreateClick,
  onNavClick,
  searchQuery,
  onSearchChange,
  searchPlaceholder = 'Search workspace, links, or campaigns... (⌘K)',
}: WorkspaceLayoutProps) {
  const { user, logout } = useAuth();
  const pathname = usePathname();
  const router = useRouter();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [userMenuOpen, setUserMenuOpen] = useState(false);
  const [aiModalOpen, setAiModalOpen] = useState(false);
  const userMenuRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // Close menus on outside click
  useEffect(() => {
    const handleDown = (e: MouseEvent) => {
      if (userMenuRef.current && !userMenuRef.current.contains(e.target as Node)) {
        setUserMenuOpen(false);
      }
    };
    document.addEventListener('mousedown', handleDown);
    return () => document.removeEventListener('mousedown', handleDown);
  }, []);

  // Keyboard shortcut: ⌘K or Ctrl+K to focus search omnibar
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
        e.preventDefault();
        searchInputRef.current?.focus();
      }
      if (e.key === 'Escape') {
        setUserMenuOpen(false);
        setMobileMenuOpen(false);
        setAiModalOpen(false);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  const displayName = user?.name || 'Malin J.';
  const displayOrg = (user as { company_name?: string } | null)?.company_name || 'trunc-team';
  const userInitials = (user?.name
    ? user.name.split(' ').map(n => n[0]).join('').slice(0, 2).toUpperCase()
    : 'MJ');

  const mainNavItems = [
    { id: 'home', label: 'Home', href: '/dashboard', icon: HomeIcon, badge: null },
    { id: 'links', label: 'Links', href: '/dashboard#links', icon: LinkIcon, badge: String(linkCount) },
    { id: 'qr', label: 'QR Codes', href: '/dashboard?tab=qr', icon: QrCodeIcon, badge: 'Try it' },
    { id: 'pages', label: 'Pages', href: '/microsites', icon: DocumentDuplicateIcon, badge: 'Try it' },
    { id: 'analytics', label: 'Analytics', href: '/analytics', icon: ChartBarIcon, badge: 'Try it' },
    { id: 'campaigns', label: 'Campaigns', href: '/dashboard#campaigns', icon: MegaphoneIcon, badge: null },
    { id: 'domains', label: 'Custom domains', href: '/custom-domains', icon: GlobeAltIcon, badge: null },
  ];

  const utilityNavItems = [
    { id: 'integrations', label: 'Integrations', href: '/settings?tab=integrations', icon: PuzzlePieceIcon },
    { id: 'settings', label: 'Settings', href: '/settings', icon: Cog6ToothIcon },
  ];

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  const navContent = (
    <div className="flex flex-col h-full bg-surface-container-lowest dark:bg-[#0c1017] border-r border-outline-variant/30 text-on-surface">
      {/* Brand Header */}
      <div className="h-16 flex items-center px-5 border-b border-outline-variant/20 gap-3">
        <Link href="/dashboard" className="flex items-center gap-2.5 group">
          <div className="w-9 h-9 rounded-xl bg-gradient-to-tr from-primary to-blue-500 flex items-center justify-center text-white shadow-sm group-hover:scale-105 transition-transform">
            <BoltIcon className="w-5 h-5" />
          </div>
          <div className="flex flex-col">
            <span className="font-extrabold text-base tracking-tight text-foreground font-hanken">Trunc</span>
            <span className="text-[10px] font-medium text-muted-foreground uppercase tracking-wider -mt-0.5">Workspace</span>
          </div>
        </Link>
      </div>

      {/* Create New Action */}
      <div className="p-4">
        <button
          onClick={() => {
            if (onCreateClick) {
              onCreateClick();
            } else {
              const el = document.getElementById('quick-create-section');
              if (el) {
                el.scrollIntoView({ behavior: 'smooth' });
                const input = el.querySelector('input');
                input?.focus();
              } else {
                router.push('/dashboard');
              }
            }
          }}
          className="w-full flex items-center justify-center gap-2 py-2.5 px-4 rounded-xl bg-primary hover:bg-primary-600 active:scale-[0.98] text-white text-sm font-semibold shadow-sm transition-all"
        >
          <PlusIcon className="w-4 h-4 stroke-[2.5]" />
          <span>Create new</span>
        </button>
      </div>

      {/* Navigation Cluster */}
      <div className="flex-1 overflow-y-auto px-3 py-2 space-y-1">
        <div className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2 font-mono">
          Overview
        </div>
        {mainNavItems.map((item) => {
          const isActive = activeNav ? activeNav === item.id : (item.href === pathname && item.id !== 'links');
          const Icon = item.icon;
          return (
            <Link
              key={item.id}
              href={item.href}
              onClick={() => {
                setMobileMenuOpen(false);
                onNavClick?.(item.id);
              }}
              className={`flex items-center justify-between px-3 py-2 rounded-xl text-sm font-medium transition-all ${
                isActive
                  ? 'bg-primary/10 text-primary font-semibold dark:bg-primary/20'
                  : 'text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900/60'
              }`}
            >
              <div className="flex items-center gap-3">
                <Icon className={`w-4 h-4 ${isActive ? 'text-primary' : 'text-muted-foreground'}`} />
                <span>{item.label}</span>
              </div>
              {item.badge && (
                <span
                  className={`text-[10px] font-bold px-2 py-0.5 rounded-full ${
                    item.badge === 'Try it'
                      ? 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20'
                      : 'bg-primary/10 text-primary font-mono'
                  }`}
                >
                  {item.badge}
                </span>
              )}
            </Link>
          );
        })}

        <div className="pt-5 pb-1">
          <div className="text-[11px] font-semibold text-muted-foreground uppercase tracking-wider px-3 mb-2 font-mono">
            System
          </div>
          {utilityNavItems.map((item) => {
            const isActive = pathname === item.href || (item.id === 'settings' && pathname.startsWith('/settings'));
            const Icon = item.icon;
            return (
              <Link
                key={item.id}
                href={item.href}
                onClick={() => setMobileMenuOpen(false)}
                className={`flex items-center gap-3 px-3 py-2 rounded-xl text-sm font-medium transition-all ${
                  isActive
                    ? 'bg-primary/10 text-primary font-semibold dark:bg-primary/20'
                    : 'text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900/60'
                }`}
              >
                <Icon className={`w-4 h-4 ${isActive ? 'text-primary' : 'text-muted-foreground'}`} />
                <span>{item.label}</span>
              </Link>
            );
          })}
        </div>
      </div>

      {/* User Footer Tab */}
      <div className="p-3 border-t border-outline-variant/20 relative" ref={userMenuRef}>
        <button
          onClick={() => setUserMenuOpen(!userMenuOpen)}
          className="w-full flex items-center justify-between p-2 rounded-xl hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors text-left"
        >
          <div className="flex items-center gap-2.5 min-w-0">
            <div className="w-8 h-8 rounded-lg bg-gradient-to-tr from-blue-600 to-indigo-600 text-white font-bold text-xs flex items-center justify-center shrink-0 shadow-sm">
              {userInitials}
            </div>
            <div className="min-w-0 flex-1">
              <div className="text-xs font-semibold text-foreground truncate">{displayName}</div>
              <div className="text-[11px] text-muted-foreground truncate font-mono">{displayOrg}</div>
            </div>
          </div>
          <ChevronDownIcon className="w-4 h-4 text-muted-foreground shrink-0 ml-1" />
        </button>

        {/* User Popup Menu */}
        {userMenuOpen && (
          <div className="absolute bottom-16 left-3 right-3 bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-xl shadow-xl py-2 z-50 text-xs">
            <div className="px-3 py-2 border-b border-outline-variant/20">
              <div className="font-semibold text-foreground">{displayName}</div>
              <div className="text-muted-foreground truncate">{user?.email || 'user@trunc.site'}</div>
            </div>
            <Link
              href="/settings"
              onClick={() => setUserMenuOpen(false)}
              className="flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors"
            >
              <Cog6ToothIcon className="w-4 h-4 text-muted-foreground" />
              <span>Workspace Settings</span>
            </Link>
            <Link
              href="/billing"
              onClick={() => setUserMenuOpen(false)}
              className="flex items-center gap-2 px-3 py-2 text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors"
            >
              <CreditCardIcon className="w-4 h-4 text-muted-foreground" />
              <span>Billing & Invoices</span>
            </Link>
            <div className="border-t border-outline-variant/20 my-1"></div>
            <button
              onClick={handleLogout}
              className="w-full flex items-center gap-2 px-3 py-2 text-destructive hover:bg-destructive/10 transition-colors text-left"
            >
              <ArrowRightOnRectangleIcon className="w-4 h-4" />
              <span>Sign out</span>
            </button>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div className="min-h-screen bg-surface-container-lowest dark:bg-[#07090e] text-on-surface flex font-sans">
      {/* Desktop Left Rail (w-60) */}
      <aside className="hidden lg:flex w-60 flex-col shrink-0 sticky top-0 h-screen z-30">
        {navContent}
      </aside>

      {/* Mobile Backdrop & Drawer */}
      {mobileMenuOpen && (
        <div className="fixed inset-0 z-50 lg:hidden">
          <div
            className="fixed inset-0 bg-black/50 backdrop-blur-sm"
            onClick={() => setMobileMenuOpen(false)}
          />
          <div className="fixed inset-y-0 left-0 w-64 shadow-2xl z-10 animate-in slide-in-from-left duration-200">
            {navContent}
          </div>
        </div>
      )}

      {/* Canvas Area */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top App Header (h-16) */}
        <header className="h-16 sticky top-0 z-20 bg-surface-container-lowest/90 dark:bg-[#07090e]/90 backdrop-blur-md border-b border-outline-variant/20 px-4 sm:px-6 lg:px-8 flex items-center justify-between gap-4">
          <div className="flex items-center gap-3 flex-1 min-w-0">
            {/* Hamburger on mobile */}
            <button
              onClick={() => setMobileMenuOpen(true)}
              className="lg:hidden p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900"
              aria-label="Open sidebar"
            >
              <Bars3Icon className="w-5 h-5" />
            </button>

            {/* Omnibar Search with ⌘K */}
            <div className="relative flex-1 max-w-xl">
              <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
                <MagnifyingGlassIcon className="h-4 w-4 text-muted-foreground" />
              </div>
              <input
                ref={searchInputRef}
                type="text"
                value={searchQuery ?? ''}
                onChange={(e) => onSearchChange?.(e.target.value)}
                placeholder={searchPlaceholder}
                className="w-full pl-9 pr-14 py-2 rounded-xl text-xs bg-surface-container dark:bg-zinc-900/80 border border-outline-variant/30 text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-all font-sans"
              />
              <div className="absolute inset-y-0 right-0 pr-2.5 flex items-center pointer-events-none">
                <kbd className="inline-flex items-center px-1.5 py-0.5 rounded border border-outline-variant/40 bg-surface-container-high dark:bg-zinc-800 text-[10px] font-mono text-muted-foreground">
                  ⌘K
                </kbd>
              </div>
            </div>
          </div>

          {/* Right Header Cluster */}
          <div className="flex items-center gap-2 sm:gap-3 shrink-0">
            {/* Theme Toggle */}
            <ThemeToggle />

            {/* Help / Docs */}
            <Link
              href="/faq"
              className="p-2 rounded-lg text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors"
              title="Help & Documentation"
            >
              <QuestionMarkCircleIcon className="w-5 h-5" />
            </Link>

            {/* AI Assistant Button */}
            <button
              onClick={() => setAiModalOpen(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl border border-indigo-500/30 bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-600 dark:text-indigo-400 text-xs font-semibold transition-all relative"
            >
              <span className="w-2 h-2 rounded-full bg-indigo-500 animate-pulse"></span>
              <SparklesIcon className="w-3.5 h-3.5" />
              <span className="hidden sm:inline">AI Assistant</span>
            </button>

            {/* Upgrade CTA */}
            {user?.account_type !== 'enterprise' && (
              <Link
                href="/billing"
                className="hidden md:inline-flex items-center px-3.5 py-1.5 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 hover:from-blue-500 hover:to-indigo-500 text-white text-xs font-bold tracking-tight shadow-sm hover:shadow transition-all"
              >
                Upgrade
              </Link>
            )}

            {/* Profile Avatar */}
            <Link
              href="/settings"
              className="w-8 h-8 rounded-full bg-gradient-to-tr from-blue-600 to-indigo-600 text-white font-bold text-xs flex items-center justify-center shadow-sm ring-2 ring-primary/20 hover:ring-primary/40 transition-all ml-1"
              title={`${displayName} (${user?.email || 'Account'})`}
            >
              {userInitials}
            </Link>
          </div>
        </header>

        {/* Scrollable Canvas Body */}
        <main className="flex-1 px-4 sm:px-6 lg:px-8 py-6 max-w-7xl w-full mx-auto">
          {children}
        </main>
      </div>

      {/* AI Assistant Popup Modal */}
      {aiModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 backdrop-blur-sm animate-in fade-in duration-150">
          <div className="bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl max-w-md w-full p-6 shadow-2xl relative">
            <button
              onClick={() => setAiModalOpen(false)}
              className="absolute top-4 right-4 text-muted-foreground hover:text-foreground"
            >
              <XMarkIcon className="w-5 h-5" />
            </button>
            <div className="flex items-center gap-3 mb-4">
              <div className="w-10 h-10 rounded-xl bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 flex items-center justify-center">
                <SparklesIcon className="w-5 h-5" />
              </div>
              <div>
                <h3 className="font-bold text-base text-foreground font-hanken">Trunc AI Assistant</h3>
                <p className="text-xs text-muted-foreground">Automated slug generation & attribution</p>
              </div>
            </div>
            <div className="space-y-3 text-xs text-muted-foreground mb-6">
              <div className="p-3 rounded-xl bg-surface-container dark:bg-zinc-900 border border-outline-variant/20">
                <div className="font-semibold text-foreground mb-1">⚡ Smart Context Slugs</div>
                AI analyzes your target page and recommends memorable vanity keywords that improve CTR by up to 34%.
              </div>
              <div className="p-3 rounded-xl bg-surface-container dark:bg-zinc-900 border border-outline-variant/20">
                <div className="font-semibold text-foreground mb-1">🏷️ Campaign Auto-Tagging</div>
                Automatically extracts UTM tags, channels, and social intent to categorize your link stream.
              </div>
            </div>
            <div className="flex gap-2.5">
              <button
                onClick={() => setAiModalOpen(false)}
                className="flex-1 py-2 px-3 rounded-xl border border-outline-variant/30 text-xs font-semibold hover:bg-surface-container text-foreground"
              >
                Close
              </button>
              <Link
                href="/billing"
                onClick={() => setAiModalOpen(false)}
                className="flex-1 py-2 px-3 rounded-xl bg-gradient-to-r from-blue-600 to-indigo-600 text-white text-xs font-semibold text-center hover:opacity-95"
              >
                Enable AI Features
              </Link>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
