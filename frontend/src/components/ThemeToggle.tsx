'use client';

import { useState, useEffect } from 'react';
import { useTheme } from '@/contexts/ThemeContext';
import {
  SunIcon,
  MoonIcon,
  ComputerDesktopIcon,
} from '@heroicons/react/24/outline';

export default function ThemeToggle() {
  const { theme, setTheme } = useTheme();
  const [mounted, setMounted] = useState(false);

  useEffect(() => {
    setMounted(true);
  }, []);

  if (!mounted) {
    return (
      <div className="w-9 h-9 rounded-lg border border-gray-300 dark:border-gray-600 flex items-center justify-center">
        <div className="w-4 h-4 bg-gray-300 dark:bg-gray-600 rounded animate-pulse"></div>
      </div>
    );
  }

  const themes = [
    { id: 'light', icon: SunIcon, label: 'Light' },
    { id: 'dark', icon: MoonIcon, label: 'Dark' },
    { id: 'system', icon: ComputerDesktopIcon, label: 'System' },
  ];

  return (
    <div className="relative">
      <button
        onClick={() => {
          const currentIndex = themes.findIndex(t => t.id === theme);
          const nextIndex = (currentIndex + 1) % themes.length;
          setTheme(themes[nextIndex].id as any);
        }}
        className="w-9 h-9 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 flex items-center justify-center hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
        title={`Current: ${theme}, Click to cycle`}
      >
        {(() => {
          const currentTheme = themes.find(t => t.id === theme);
          const Icon = currentTheme?.icon || SunIcon;
          return <Icon className="w-4 h-4 text-gray-700 dark:text-gray-300" />;
        })()}
      </button>
    </div>
  );
}