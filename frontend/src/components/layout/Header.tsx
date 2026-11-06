'use client';

import Link from 'next/link';
import { useState, useEffect, useRef } from 'react';
import { useRouter } from 'next/navigation';
import { Bars3Icon, XMarkIcon, LinkIcon, ChevronDownIcon, UserCircleIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import ThemeToggle from '@/components/ThemeToggle';
import {
  prefetchDashboard,
  prefetchAnalytics,
  prefetchBilling,
  prefetchSettings,
  prefetchApiKeys,
  prefetchTeams,
  prefetchCustomDomains,
  prefetchBulk,
} from '@/lib/page-prefetchers';

export default function Header() {
  const [isMenuOpen, setIsMenuOpen] = useState(false);
  const [isProfileMenuOpen, setIsProfileMenuOpen] = useState(false);
  const { user, isAuthenticated, logout, isLoading, getAccessToken } = useAuth();
  const router = useRouter();

  const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

  const handlePrefetch = (href: string) => {
    router.prefetch(href);
    const token = getAccessToken();
    if (!token) return;
    const prefetchers: Record<string, () => void> = {
      '/dashboard':      () => prefetchDashboard(token, API_BASE_URL),
      '/analytics':      () => prefetchAnalytics(token, API_BASE_URL),
      '/billing':        () => prefetchBilling(token, API_BASE_URL),
      '/settings':       () => prefetchSettings(token, API_BASE_URL),
      '/api-keys':       () => prefetchApiKeys(token, API_BASE_URL),
      '/teams':          () => prefetchTeams(token, API_BASE_URL),
      '/custom-domains': () => prefetchCustomDomains(token, API_BASE_URL),
      '/bulk':           () => prefetchBulk(token, API_BASE_URL),
    };
    prefetchers[href]?.();
  };
  const profileMenuRef = useRef<HTMLDivElement>(null);

  const navigation = [
    { name: 'Home', href: '/' },
    { name: 'Features', href: '/features' },
    { name: 'Pricing', href: '/pricing' },
    { name: 'Blog', href: '/blog' },
  ];

  const handleLogout = () => {
    logout();
    setIsProfileMenuOpen(false);
    router.push('/');
  };

  // Close profile menu when clicking outside or pressing Escape
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (profileMenuRef.current && !profileMenuRef.current.contains(event.target as Node)) {
        setIsProfileMenuOpen(false);
      }
    };
    const handleKeyDown = (event: KeyboardEvent) => {
      if (event.key === 'Escape') setIsProfileMenuOpen(false);
    };

    document.addEventListener('mousedown', handleClickOutside);
    document.addEventListener('keydown', handleKeyDown);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
      document.removeEventListener('keydown', handleKeyDown);
    };
  }, []);

  return (
    <header className="bg-card border-b border-border shadow-sm">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div className="flex justify-between items-center h-16">
          {/* Logo */}
          <div className="flex items-center">
            <Link href="/" className="flex items-center space-x-2">
              <LinkIcon className="h-8 w-8 text-primary" />
              <span className="text-xl font-bold text-foreground">
                Trunc
              </span>
            </Link>
          </div>

          {/* Desktop Navigation */}
          <nav className="hidden md:flex space-x-8">
            {navigation.map((item) => (
              <Link
                key={item.name}
                href={item.href}
                className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium transition-colors"
              >
                {item.name}
              </Link>
            ))}
          </nav>

          {/* Auth Buttons / User Menu */}
          <div className="hidden md:flex items-center space-x-4">
            <ThemeToggle />
            {isLoading ? (
              <div className="flex items-center space-x-3">
                <div className="h-8 w-14 bg-muted rounded animate-pulse"></div>
                <div className="h-9 w-28 bg-muted rounded-lg animate-pulse"></div>
              </div>
            ) : isAuthenticated && user ? (
              <div className="relative" ref={profileMenuRef}>
                <button
                  onClick={() => setIsProfileMenuOpen(!isProfileMenuOpen)}
                  aria-haspopup="true"
                  aria-expanded={isProfileMenuOpen}
                  aria-controls="profile-menu"
                  className="flex items-center space-x-2 text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium rounded-lg hover:bg-muted transition-colors"
                >
                  <UserCircleIcon className="h-5 w-5" />
                  <span>{user.name}</span>
                  <ChevronDownIcon className="h-4 w-4" />
                </button>

                {isProfileMenuOpen && (
                  <div id="profile-menu" className="absolute right-0 mt-2 w-48 bg-card rounded-md shadow-lg py-1 z-50 border border-border">
                    <div className="px-4 py-2 text-xs text-muted-foreground border-b border-border">
                      {user.email}
                    </div>
                    <Link
                      href="/dashboard"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/dashboard')}
                    >
                      Dashboard
                    </Link>
                    <Link
                      href="/micropages"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted font-medium text-blue-500"
                      onClick={() => setIsProfileMenuOpen(false)}
                    >
                      Micro-Pages & vCards
                    </Link>
                    <Link
                      href="/analytics"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/analytics')}
                    >
                      Analytics
                    </Link>
                    <Link
                      href="/bulk"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/bulk')}
                    >
                      Bulk Operations
                    </Link>
                    <Link
                      href="/api-keys"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/api-keys')}
                    >
                      API Keys
                    </Link>
                    <Link
                      href="/custom-domains"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/custom-domains')}
                    >
                      Custom Domains
                    </Link>
                    <Link
                      href="/teams"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/teams')}
                    >
                      Teams
                    </Link>
                    <Link
                      href="/settings"
                      className="block px-4 py-2 text-sm text-foreground hover:bg-muted"
                      onClick={() => setIsProfileMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/settings')}
                    >
                      Settings
                    </Link>
                    <button
                      onClick={handleLogout}
                      className="block w-full text-left px-4 py-2 text-sm text-foreground hover:bg-muted"
                    >
                      Sign out
                    </button>
                  </div>
                )}
              </div>
            ) : (
              <>
                <Link
                  href="/login"
                  className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium"
                >
                  Sign in
                </Link>
                <Link
                  href="/register"
                  className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors"
                >
                  Get Started
                </Link>
              </>
            )}
          </div>

          {/* Mobile menu button */}
          <div className="md:hidden">
            <button
              onClick={() => setIsMenuOpen(!isMenuOpen)}
              aria-label={isMenuOpen ? "Close menu" : "Open menu"}
              aria-expanded={isMenuOpen}
              aria-controls="mobile-menu"
              className="text-muted-foreground hover:text-foreground"
            >
              {isMenuOpen ? (
                <XMarkIcon className="h-6 w-6" />
              ) : (
                <Bars3Icon className="h-6 w-6" />
              )}
            </button>
          </div>
        </div>

        {/* Mobile Navigation */}
        {isMenuOpen && (
          <div id="mobile-menu" className="md:hidden py-4 border-t border-border">
            <div className="flex flex-col space-y-2 bg-card">
              {navigation.map((item) => (
                <Link
                  key={item.name}
                  href={item.href}
                  className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium"
                  onClick={() => setIsMenuOpen(false)}
                >
                  {item.name}
                </Link>
              ))}
              <div className="border-t border-border pt-2 mt-2">
                <div className="px-3 py-2 flex items-center justify-between">
                  <span className="text-sm font-medium text-foreground">Theme</span>
                  <ThemeToggle />
                </div>
                {isAuthenticated && user ? (
                  <>
                    <div className="px-3 py-2 text-xs text-muted-foreground">
                      {user.name} ({user.email})
                    </div>
                    <Link
                      href="/dashboard"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/dashboard')}
                    >
                      Dashboard
                    </Link>
                    <Link
                      href="/analytics"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/analytics')}
                    >
                      Analytics
                    </Link>
                    <Link
                      href="/bulk"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/bulk')}
                    >
                      Bulk Operations
                    </Link>
                    <Link
                      href="/api-keys"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/api-keys')}
                    >
                      API Keys
                    </Link>
                    <Link
                      href="/custom-domains"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/custom-domains')}
                    >
                      Custom Domains
                    </Link>
                    <Link
                      href="/teams"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/teams')}
                    >
                      Teams
                    </Link>
                    <Link
                      href="/settings"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                      onMouseEnter={() => handlePrefetch('/settings')}
                    >
                      Settings
                    </Link>
                    <button
                      onClick={() => {
                        setIsMenuOpen(false);
                        handleLogout();
                      }}
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block w-full text-left"
                    >
                      Sign out
                    </button>
                  </>
                ) : (
                  <>
                    <Link
                      href="/login"
                      className="text-muted-foreground hover:text-foreground px-3 py-2 text-sm font-medium block"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      Sign in
                    </Link>
                    <Link
                      href="/register"
                      className="bg-primary text-primary-foreground px-4 py-2 rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors block text-center mx-3 mt-2"
                      onClick={() => setIsMenuOpen(false)}
                    >
                      Get Started
                    </Link>
                  </>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </header>
  );
}