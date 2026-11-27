'use client';

import type { ReactNode } from 'react';
import { Dialog, DialogBackdrop, DialogPanel, DialogTitle, Description } from '@headlessui/react';
import { ExclamationTriangleIcon } from '@heroicons/react/24/outline';
import { Button } from './ui';

/** Confirmation for a destructive action. Focus starts on Cancel. */
export default function ConfirmDialog({
  open,
  title,
  children,
  confirmLabel,
  busy,
  error,
  onConfirm,
  onClose,
}: {
  open: boolean;
  title: string;
  children: ReactNode;
  confirmLabel: string;
  busy?: boolean;
  error?: string;
  onConfirm: () => void;
  onClose: () => void;
}) {
  return (
    <Dialog open={open} onClose={() => !busy && onClose()} className="relative z-50">
      <DialogBackdrop transition className="fixed inset-0 bg-black/60 backdrop-blur-[2px] transition-opacity duration-150 data-[closed]:opacity-0" />
      <div className="fixed inset-0 flex items-center justify-center p-4">
        <DialogPanel
          transition
          className="w-full max-w-md rounded-2xl border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 p-5 shadow-2xl transition duration-150 data-[closed]:scale-95 data-[closed]:opacity-0"
        >
          <div className="flex gap-3">
            <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-destructive/10 text-destructive">
              <ExclamationTriangleIcon className="h-5 w-5" />
            </div>
            <div className="min-w-0">
              <DialogTitle className="font-hanken text-base font-bold text-foreground">{title}</DialogTitle>
              <Description as="div" className="mt-1 text-sm text-muted-foreground">
                {children}
              </Description>
            </div>
          </div>
          {error && (
            <p role="alert" className="mt-4 rounded-xl border border-destructive/20 bg-destructive/10 px-3 py-2 text-xs font-medium text-destructive">
              {error}
            </p>
          )}
          <div className="mt-5 flex justify-end gap-2">
            <Button autoFocus onClick={onClose} disabled={busy}>
              Cancel
            </Button>
            <Button variant="danger" onClick={onConfirm} disabled={busy}>
              {busy ? 'Working…' : confirmLabel}
            </Button>
          </div>
        </DialogPanel>
      </div>
    </Dialog>
  );
}
