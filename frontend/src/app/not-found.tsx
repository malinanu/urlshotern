'use client';
import Link from 'next/link';
import { useTheme } from '@/contexts/ThemeContext';

export default function NotFound() {
  const { isDark } = useTheme();
  return (
    <div className={`min-h-screen flex flex-col items-center justify-center px-4 ${isDark ? 'bg-gray-950 text-white' : 'bg-gray-50 text-gray-900'}`}>
      <div className="text-center max-w-md">
        <h1 className="text-9xl font-black text-indigo-500 select-none">404</h1>
        <h2 className="mt-4 text-2xl font-bold">Page not found</h2>
        <p className={`mt-3 text-base ${isDark ? 'text-gray-400' : 'text-gray-500'}`}>
          This short link doesn&apos;t exist or has been removed.
        </p>
        <div className="mt-8 flex flex-col sm:flex-row gap-3 justify-center">
          <Link href="/" className="px-5 py-2.5 rounded-lg bg-indigo-600 hover:bg-indigo-700 text-white font-medium transition-colors">
            Go home
          </Link>
          <Link href="/dashboard" className={`px-5 py-2.5 rounded-lg border font-medium transition-colors ${isDark ? 'border-gray-700 hover:bg-gray-800' : 'border-gray-300 hover:bg-gray-100'}`}>
            My dashboard
          </Link>
        </div>
      </div>
    </div>
  );
}
