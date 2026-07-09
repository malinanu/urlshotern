'use client';

import { motion, type Variants } from 'framer-motion';
import type { ReactNode } from 'react';

interface StaggerProps {
  children: ReactNode;
  className?: string;
  /** Gap between each child's reveal, in seconds. */
  gap?: number;
  as?: 'div' | 'ul' | 'section';
}

// Reduced motion is honored globally via <MotionConfig reducedMotion="user">
// (see MotionProvider). We never branch the rendered tree on a reduced-motion
// hook, which would cause a server/client hydration mismatch.

/**
 * Container that reveals its <StaggerItem> children one after another as it
 * scrolls into view. Pair with StaggerItem.
 */
export function Stagger({ children, className, gap = 0.08, as = 'div' }: StaggerProps) {
  const MotionTag = motion[as];

  const container: Variants = {
    hidden: {},
    visible: { transition: { staggerChildren: gap } },
  };

  return (
    <MotionTag
      className={className}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: '0px 0px -80px 0px' }}
      variants={container}
    >
      {children}
    </MotionTag>
  );
}

interface StaggerItemProps {
  children: ReactNode;
  className?: string;
  distance?: number;
  as?: 'div' | 'li' | 'span';
}

export function StaggerItem({ children, className, distance = 20, as = 'div' }: StaggerItemProps) {
  const MotionTag = motion[as];

  const item: Variants = {
    hidden: { opacity: 0, y: distance },
    visible: {
      opacity: 1,
      y: 0,
      transition: { type: 'spring', stiffness: 100, damping: 20 },
    },
  };

  return (
    <MotionTag className={className} variants={item}>
      {children}
    </MotionTag>
  );
}

export default Stagger;
