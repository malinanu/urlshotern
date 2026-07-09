'use client';

import { motion, type Variants } from 'framer-motion';
import type { ReactNode } from 'react';

type Direction = 'up' | 'down' | 'left' | 'right' | 'none';

interface RevealProps {
  children: ReactNode;
  /** How far, in px, the element travels into place. */
  distance?: number;
  direction?: Direction;
  /** Delay before the reveal starts, in seconds. */
  delay?: number;
  className?: string;
  /** Render as a different element while keeping motion (defaults to div). */
  as?: 'div' | 'section' | 'li' | 'span';
}

function offset(direction: Direction, distance: number) {
  switch (direction) {
    case 'up':
      return { y: distance };
    case 'down':
      return { y: -distance };
    case 'left':
      return { x: distance };
    case 'right':
      return { x: -distance };
    default:
      return {};
  }
}

/**
 * Scroll-into-view reveal. Animates opacity + transform only (GPU friendly).
 *
 * Reduced motion is honored globally via <MotionConfig reducedMotion="user">
 * (see MotionProvider), which disables transform movement while keeping a safe
 * opacity fade. We deliberately do NOT branch the rendered tree on a
 * reduced-motion hook: doing so makes the server and client render different
 * elements and causes a hydration mismatch.
 */
export function Reveal({
  children,
  distance = 24,
  direction = 'up',
  delay = 0,
  className,
  as = 'div',
}: RevealProps) {
  const MotionTag = motion[as];

  const variants: Variants = {
    hidden: { opacity: 0, ...offset(direction, distance) },
    visible: {
      opacity: 1,
      x: 0,
      y: 0,
      transition: { type: 'spring', stiffness: 100, damping: 20, delay },
    },
  };

  return (
    <MotionTag
      className={className}
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, margin: '0px 0px -80px 0px' }}
      variants={variants}
    >
      {children}
    </MotionTag>
  );
}

export default Reveal;
