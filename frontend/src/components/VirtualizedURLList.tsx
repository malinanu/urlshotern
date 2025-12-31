'use client';

import { memo } from 'react';
import {
  ClipboardDocumentIcon,
  CheckIcon,
  CalendarIcon,
  EyeIcon,
  QrCodeIcon,
  PencilIcon,
  TrashIcon,
} from '@heroicons/react/24/outline';

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
}

interface VirtualizedURLListProps {
  urls: URLData[];
  copiedId: number | null;
  onCopy: (url: string, id: number) => void;
  onAnalytics: (shortCode: string) => void;
  onQRCode: (url: URLData) => void;
  onEdit: (url: URLData) => void;
  onDelete: (shortCode: string) => void;
  formatDate: (date: string) => string;
}

// Simple non-virtualized version (react-window removed for build compatibility)
// This component is not currently used in the application
export const VirtualizedURLList = memo(({
  urls,
  copiedId,
  onCopy,
  onAnalytics,
  onQRCode,
  onEdit,
  onDelete,
  formatDate,
}: VirtualizedURLListProps) => {
  return (
    <div className="bg-white rounded-lg shadow-sm border">
      {/* Header */}
      <div className="flex items-center px-6 py-3 border-b bg-gray-50 text-xs font-medium text-gray-500 uppercase tracking-wider">
        <div className="flex-1 pr-4">Original URL</div>
        <div className="w-64 pr-4">Short URL</div>
        <div className="w-24 pr-4">Clicks</div>
        <div className="w-48 pr-4">Created</div>
        <div className="w-40">Actions</div>
      </div>

      {/* URL List */}
      <div className="max-h-[720px] overflow-y-auto">
        {urls.map((url, index) => {
          const isEven = index % 2 === 0;
          return (
            <div
              key={url.id}
              className={`flex items-center px-6 py-4 border-b border-gray-200 hover:bg-gray-50 transition-colors ${
                isEven ? 'bg-white' : 'bg-gray-50'
              }`}
            >
              {/* Original URL */}
              <div className="flex-1 min-w-0 pr-4">
                <a
                  href={url.original_url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="text-primary-600 hover:text-primary-800 text-sm truncate block"
                >
                  {url.original_url}
                </a>
              </div>

              {/* Short URL */}
              <div className="flex items-center space-x-2 w-64 pr-4">
                <span className="text-sm text-black font-mono truncate">
                  {url.short_url}
                </span>
                <button
                  onClick={() => onCopy(url.short_url, url.id)}
                  className="p-1 text-gray-400 hover:text-gray-600 transition-colors flex-shrink-0"
                  aria-label="Copy URL"
                >
                  {copiedId === url.id ? (
                    <CheckIcon className="h-4 w-4 text-green-600" />
                  ) : (
                    <ClipboardDocumentIcon className="h-4 w-4" />
                  )}
                </button>
              </div>

              {/* Clicks */}
              <div className="flex items-center space-x-1 w-24 pr-4">
                <EyeIcon className="h-4 w-4 text-gray-400 flex-shrink-0" />
                <span className="text-sm text-black">{url.click_count}</span>
              </div>

              {/* Created */}
              <div className="flex items-center space-x-1 w-48 pr-4">
                <CalendarIcon className="h-4 w-4 text-gray-400 flex-shrink-0" />
                <span className="text-sm text-gray-600 truncate">{formatDate(url.created_at)}</span>
              </div>

              {/* Actions */}
              <div className="flex items-center space-x-2 w-40 flex-shrink-0">
                <button
                  onClick={() => onAnalytics(url.short_code)}
                  className="text-primary-600 hover:text-primary-800 text-sm font-medium"
                >
                  Analytics
                </button>
                <button
                  onClick={() => onQRCode(url)}
                  className="p-1 text-purple-600 hover:text-purple-800 transition-colors"
                  title="Show QR Code"
                >
                  <QrCodeIcon className="h-4 w-4" />
                </button>
                <button
                  onClick={() => onEdit(url)}
                  className="p-1 text-blue-600 hover:text-blue-800 transition-colors"
                  title="Edit URL"
                >
                  <PencilIcon className="h-4 w-4" />
                </button>
                <button
                  onClick={() => {
                    if (confirm('Are you sure you want to delete this URL? This action cannot be undone.')) {
                      onDelete(url.short_code);
                    }
                  }}
                  className="p-1 text-red-600 hover:text-red-800 transition-colors"
                  title="Delete URL"
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
            </div>
          );
        })}
      </div>
    </div>
  );
});

VirtualizedURLList.displayName = 'VirtualizedURLList';
