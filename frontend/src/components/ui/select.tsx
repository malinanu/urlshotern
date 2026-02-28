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
        'flex h-10 w-full rounded-lg border border-input bg-background px-3 py-2 text-sm ring-offset-background focus:outline-none focus:ring-2 focus:ring-ring focus:border-transparent disabled:cursor-not-allowed disabled:opacity-50 text-foreground',
        error && 'border-destructive focus:ring-destructive',
        className
      )}
      {...props}
    >
      {children}
    </select>
  );
}