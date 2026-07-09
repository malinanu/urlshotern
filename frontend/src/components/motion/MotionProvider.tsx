'use client';

import { MotionConfig } from 'framer-motion';
import type { ReactNode } from 'react';

/**
 * Applies reduced-motion handling to all descendant motion components without
 * changing the rendered tree (so SSR and client hydration stay identical).
 * reducedMotion="user" disables transform/layout movement for users who prefer
 * reduced motion, while keeping opacity fades so content still becomes visible.
 */
export default function MotionProvider({ children }: { children: ReactNode }) {
  return <MotionConfig reducedMotion="user">{children}</MotionConfig>;
}
