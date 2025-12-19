/**
 * Mobile-Optimized Virtual Scrolling Component
 * Renders large lists efficiently on mobile devices
 */

'use client';

import React, { useRef, useCallback } from 'react';
import { useVirtualizer } from '@tanstack/react-virtual';
import { useIntersectionObserver } from '@/hooks/useIntersectionObserver';
import { useNetworkQuality } from '@/hooks/useNetworkStatus';

interface VirtualScrollListProps<T> {
  items: T[];
  renderItem: (item: T, index: number) => React.ReactNode;
  estimateSize?: number;
  overscan?: number;
  onLoadMore?: () => void;
  hasMore?: boolean;
  isLoading?: boolean;
  emptyMessage?: string;
  className?: string;
  itemClassName?: string;
  gap?: number;
}

export function VirtualScrollList<T>({
  items,
  renderItem,
  estimateSize = 100,
  overscan = 5,
  onLoadMore,
  hasMore = false,
  isLoading = false,
  emptyMessage = 'No items to display',
  className = '',
  itemClassName = '',
  gap = 8,
}: VirtualScrollListProps<T>) {
  const parentRef = useRef<HTMLDivElement>(null);
  const networkQuality = useNetworkQuality();

  // Adjust overscan based on network quality
  const adjustedOverscan = networkQuality === 'low' ? 3 : networkQuality === 'high' ? 10 : overscan;

  // Initialize virtualizer
  const virtualizer = useVirtualizer({
    count: items.length,
    getScrollElement: () => parentRef.current,
    estimateSize: () => estimateSize,
    overscan: adjustedOverscan,
    gap,
  });

  // Setup infinite scroll
  const loadMoreRef = useRef<HTMLDivElement>(null);

  useIntersectionObserver({
    target: loadMoreRef,
    onIntersect: useCallback(() => {
      if (hasMore && !isLoading && onLoadMore) {
        onLoadMore();
      }
    }, [hasMore, isLoading, onLoadMore]),
    enabled: hasMore && !isLoading,
  });

  if (items.length === 0 && !isLoading) {
    return (
      <div className={`flex items-center justify-center min-h-[200px] ${className}`}>
        <p className="text-gray-500 text-center">{emptyMessage}</p>
      </div>
    );
  }

  return (
    <div
      ref={parentRef}
      className={`overflow-auto ${className}`}
      style={{
        contain: 'strict',
        WebkitOverflowScrolling: 'touch',
        willChange: 'transform',
      }}
    >
      <div
        style={{
          height: `${virtualizer.getTotalSize()}px`,
          width: '100%',
          position: 'relative',
        }}
      >
        {virtualizer.getVirtualItems().map((virtualItem) => (
          <div
            key={virtualItem.key}
            data-index={virtualItem.index}
            className={itemClassName}
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              width: '100%',
              height: `${virtualItem.size}px`,
              transform: `translateY(${virtualItem.start}px)`,
              willChange: 'transform',
            }}
          >
            {renderItem(items[virtualItem.index], virtualItem.index)}
          </div>
        ))}
      </div>

      {/* Infinite scroll trigger */}
      {hasMore && (
        <div
          ref={loadMoreRef}
          className="flex items-center justify-center py-4"
          style={{
            position: 'relative',
            height: '60px',
          }}
        >
          {isLoading && (
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600" />
          )}
        </div>
      )}
    </div>
  );
}

/**
 * Mobile-Optimized URL List Item Component
 */
interface URLListItemProps {
  url: {
    short_code: string;
    short_url: string;
    original_url: string;
    click_count: number;
    created_at: string;
  };
  onEdit?: () => void;
  onDelete?: () => void;
  onClick?: () => void;
}

export function URLListItem({ url, onEdit, onDelete, onClick }: URLListItemProps) {
  return (
    <div
      className="bg-white rounded-lg border border-gray-200 p-4 mb-2 active:bg-gray-50 transition-colors touch-manipulation"
      onClick={onClick}
      style={{
        contain: 'layout style',
        contentVisibility: 'auto',
      }}
    >
      <div className="flex items-start justify-between gap-3">
        <div className="flex-1 min-w-0">
          <div className="flex items-center gap-2 mb-1">
            <span className="text-sm font-mono font-medium text-primary-600 truncate">
              {url.short_url}
            </span>
            <button
              onClick={(e) => {
                e.stopPropagation();
                navigator.clipboard.writeText(url.short_url);
              }}
              className="flex-shrink-0 p-1.5 text-gray-400 hover:text-gray-600 active:text-primary-600 touch-manipulation"
              aria-label="Copy URL"
            >
              <svg
                className="w-4 h-4"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                />
              </svg>
            </button>
          </div>

          <p className="text-sm text-gray-600 truncate mb-2">{url.original_url}</p>

          <div className="flex items-center gap-4 text-xs text-gray-500">
            <span className="flex items-center gap-1">
              <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
                />
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M2.458 12C3.732 7.943 7.523 5 12 5c4.478 0 8.268 2.943 9.542 7-1.274 4.057-5.064 7-9.542 7-4.477 0-8.268-2.943-9.542-7z"
                />
              </svg>
              {url.click_count} clicks
            </span>
            <span>
              {new Date(url.created_at).toLocaleDateString()}
            </span>
          </div>
        </div>

        {(onEdit || onDelete) && (
          <div className="flex flex-col gap-1">
            {onEdit && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onEdit();
                }}
                className="p-2 text-gray-400 hover:text-primary-600 active:text-primary-700 touch-manipulation"
                aria-label="Edit"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
            )}
            {onDelete && (
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete();
                }}
                className="p-2 text-gray-400 hover:text-red-600 active:text-red-700 touch-manipulation"
                aria-label="Delete"
              >
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    strokeWidth={2}
                    d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                  />
                </svg>
              </button>
            )}
          </div>
        )}
      </div>
    </div>
  );
}
