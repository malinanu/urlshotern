'use client';

import { useState } from 'react';
import Header from './Header';
import Footer from './Footer';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';

interface LayoutProps {
  children: React.ReactNode;
}

export default function Layout({ children }: LayoutProps) {
  const { user, getAccessToken } = useAuth();
  const [emailBannerDismissed, setEmailBannerDismissed] = useState(false);
  const [phoneBannerDismissed, setPhoneBannerDismissed] = useState(false);
  const [resendState, setResendState] = useState<'idle' | 'sending' | 'sent'>('idle');

  const handleResend = async () => {
    if (resendState !== 'idle') return;
    setResendState('sending');
    try {
      await fetch(API_ENDPOINTS.auth.resendEmail, {
        method: 'POST',
        headers: { Authorization: `Bearer ${getAccessToken()}` },
      });
      setResendState('sent');
    } catch {
      setResendState('idle');
    }
  };

  const showEmailBanner = !!user && !user.email_verified && !emailBannerDismissed;
  const showPhoneBanner = !!user && user.email_verified && !user.phone_verified && !phoneBannerDismissed;

  return (
    <div className="min-h-screen flex flex-col">
      <Header />

      {showEmailBanner && (
        <div className="bg-yellow-500/10 border-b border-yellow-500/20 px-4 py-2 flex items-center justify-between text-sm">
          <span className="text-yellow-700 dark:text-yellow-400">
            Please verify your email address to fully activate your account.{' '}
            <button
              onClick={handleResend}
              className="underline font-medium hover:opacity-80"
            >
              {resendState === 'sent' ? 'Email sent!' : resendState === 'sending' ? 'Sending…' : 'Resend verification email'}
            </button>
          </span>
          <button
            onClick={() => setEmailBannerDismissed(true)}
            className="text-yellow-600 hover:text-yellow-800 ml-4 text-base leading-none"
            aria-label="Dismiss"
          >
            ✕
          </button>
        </div>
      )}

      {showPhoneBanner && (
        <div className="bg-blue-500/10 border-b border-blue-500/20 px-4 py-2 flex items-center justify-between text-sm">
          <span className="text-blue-700 dark:text-blue-400">
            Add a phone number for extra security.{' '}
            <a href="/settings#security" className="underline font-medium hover:opacity-80">
              Go to Security settings
            </a>
          </span>
          <button
            onClick={() => setPhoneBannerDismissed(true)}
            className="text-blue-600 hover:text-blue-800 ml-4 text-base leading-none"
            aria-label="Dismiss"
          >
            ✕
          </button>
        </div>
      )}

      <main className="flex-1">{children}</main>
      <Footer />
    </div>
  );
}
