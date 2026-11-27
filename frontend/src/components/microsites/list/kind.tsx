import {
  CalendarDaysIcon,
  IdentificationIcon,
  LinkIcon,
  ShoppingBagIcon,
  Squares2X2Icon,
} from '@heroicons/react/24/outline';
import type { PageKind } from '@/lib/microsites';

const ICONS: Record<PageKind, typeof LinkIcon> = {
  card: IdentificationIcon,
  links: LinkIcon,
  shop: ShoppingBagIcon,
  event: CalendarDaysIcon,
  other: Squares2X2Icon,
};

export function KindIcon({ kind, className }: { kind: PageKind; className?: string }) {
  const Icon = ICONS[kind];
  return <Icon className={className} aria-hidden="true" />;
}
