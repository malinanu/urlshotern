import React from 'react';
import { cn } from '@/lib/utils';

export interface SelectProps extends React.SelectHTMLAttributes<HTMLSelectElement> {
  error?: boolean;
}

export function Select({
  className,
  error,
  children,
  ...props
}: SelectProps) {
  return (
    <select
      className={cn(
        'flex h-10 w-full rounded-lg border border-gray-300 bg-white px-3 py-2 text-sm ring-offset-white focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50',
        error && 'border-red-300 focus:ring-red-500',
        className
      )}
      {...props}
    >
      {children}
    </select>
  );
}