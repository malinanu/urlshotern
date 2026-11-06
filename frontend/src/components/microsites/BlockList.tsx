'use client';

import {
  DndContext,
  KeyboardSensor,
  PointerSensor,
  TouchSensor,
  closestCenter,
  useSensor,
  useSensors,
  type DragEndEvent,
} from '@dnd-kit/core';
import {
  SortableContext,
  arrayMove,
  sortableKeyboardCoordinates,
  useSortable,
  verticalListSortingStrategy,
} from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import { Bars2Icon, ChevronDownIcon, EyeIcon, EyeSlashIcon, TrashIcon } from '@heroicons/react/24/outline';
import { BLOCK_NAMES, type BlockData, type MicrositeBlock } from '@/lib/microsites';
import BlockForm from './BlockForms';

interface Props {
  blocks: MicrositeBlock[];
  selectedId: string | null;
  micrositeId: number;
  onSelect: (id: string | null) => void;
  onReorder: (blocks: MicrositeBlock[]) => void;
  onUpdate: (id: string, patch: Partial<MicrositeBlock>) => void;
  onRemove: (id: string) => void;
  canUseForms?: boolean;
}

/** Draggable, accordion-style list of the page's blocks. */
export default function BlockList({ blocks, selectedId, micrositeId, onSelect, onReorder, onUpdate, onRemove, canUseForms }: Props) {
  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 6 } }),
    useSensor(TouchSensor, { activationConstraint: { delay: 150, tolerance: 6 } }),
    useSensor(KeyboardSensor, { coordinateGetter: sortableKeyboardCoordinates })
  );

  const handleDragEnd = (e: DragEndEvent) => {
    const { active, over } = e;
    if (!over || active.id === over.id) return;
    const from = blocks.findIndex((b) => b.id === active.id);
    const to = blocks.findIndex((b) => b.id === over.id);
    if (from >= 0 && to >= 0) onReorder(arrayMove(blocks, from, to));
  };

  if (blocks.length === 0) {
    return (
      <p className="rounded-lg border border-dashed border-border p-6 text-center text-sm text-muted-foreground">
        No blocks yet. Add one from the catalog.
      </p>
    );
  }

  return (
    <DndContext sensors={sensors} collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
      <SortableContext items={blocks.map((b) => b.id)} strategy={verticalListSortingStrategy}>
        <ul className="grid grid-cols-1 gap-2" aria-label="Page blocks">
          {blocks.map((b) => (
            <SortableBlock
              key={b.id}
              block={b}
              open={selectedId === b.id}
              micrositeId={micrositeId}
              canUseForms={canUseForms}
              onToggle={() => onSelect(selectedId === b.id ? null : b.id)}
              onChange={(data) => onUpdate(b.id, { data })}
              onToggleHidden={() => onUpdate(b.id, { hidden: !b.hidden })}
              onRemove={() => onRemove(b.id)}
            />
          ))}
        </ul>
      </SortableContext>
    </DndContext>
  );
}

function blockSummary(b: MicrositeBlock): string {
  const d = b.data ?? {};
  switch (b.type) {
    case 'profile_header':
      return d.display_name || '';
    case 'vcard_action':
      return [d.first_name, d.last_name].filter(Boolean).join(' ') || d.organization || '';
    case 'link_list':
    case 'contact_buttons':
    case 'social_icons':
      return `${(d.items ?? []).length} item${(d.items ?? []).length === 1 ? '' : 's'}`;
    case 'text':
      return (d.text || '').slice(0, 40);
    case 'product_card':
      return [d.name, d.price].filter(Boolean).join(' · ');
    case 'product_carousel':
    case 'agenda':
      return `${(d.items ?? []).length} item${(d.items ?? []).length === 1 ? '' : 's'}`;
    case 'event_details':
      return [d.title, (d.start || '').replace('T', ' ')].filter(Boolean).join(' · ');
    case 'countdown':
      return (d.target || '').replace('T', ' ');
    case 'map':
      return d.address || d.label || '';
    case 'video':
    case 'audio':
      return d.title || (d.url || '').replace(/^https?:\/\/(www\.)?/, '').slice(0, 40);
    case 'form':
      return d.kind === 'contact' ? 'Contact form' : 'Newsletter';
    default:
      return '';
  }
}

function SortableBlock({
  block,
  open,
  micrositeId,
  canUseForms,
  onToggle,
  onChange,
  onToggleHidden,
  onRemove,
}: {
  block: MicrositeBlock;
  open: boolean;
  micrositeId: number;
  canUseForms?: boolean;
  onToggle: () => void;
  onChange: (data: BlockData) => void;
  onToggleHidden: () => void;
  onRemove: () => void;
}) {
  const { attributes, listeners, setNodeRef, setActivatorNodeRef, transform, transition, isDragging } = useSortable({
    id: block.id,
  });
  const style = { transform: CSS.Transform.toString(transform), transition };
  const name = BLOCK_NAMES[block.type] ?? block.type;

  return (
    <li
      ref={setNodeRef}
      style={style}
      className={`rounded-lg border bg-card ${open ? 'border-primary' : 'border-border'} ${isDragging ? 'shadow-lg z-10 relative' : ''}`}
    >
      <div className="flex items-center gap-1 px-2 py-2">
        <button
          ref={setActivatorNodeRef}
          type="button"
          {...attributes}
          {...listeners}
          className="p-1 text-muted-foreground hover:text-foreground cursor-grab active:cursor-grabbing touch-none"
          aria-label={`Reorder ${name}`}
        >
          <Bars2Icon className="h-5 w-5" />
        </button>
        <button type="button" onClick={onToggle} className="flex-1 min-w-0 flex items-center gap-2 text-left" aria-expanded={open}>
          <span className={`text-sm font-medium ${block.hidden ? 'text-muted-foreground line-through' : 'text-foreground'}`}>{name}</span>
          <span className="text-xs text-muted-foreground truncate">{blockSummary(block)}</span>
          <ChevronDownIcon className={`ml-auto h-4 w-4 text-muted-foreground transition-transform ${open ? 'rotate-180' : ''}`} />
        </button>
        <button type="button" onClick={onToggleHidden} className="p-1 text-muted-foreground hover:text-foreground" aria-label={block.hidden ? `Show ${name}` : `Hide ${name}`}>
          {block.hidden ? <EyeSlashIcon className="h-4 w-4" /> : <EyeIcon className="h-4 w-4" />}
        </button>
        <button
          type="button"
          onClick={() => confirm(`Remove the ${name} block?`) && onRemove()}
          className="p-1 text-muted-foreground hover:text-red-600"
          aria-label={`Remove ${name}`}
        >
          <TrashIcon className="h-4 w-4" />
        </button>
      </div>
      {open && (
        <div className="border-t border-border px-3 py-3">
          <BlockForm block={block} micrositeId={micrositeId} onChange={onChange} canUseForms={canUseForms} />
        </div>
      )}
    </li>
  );
}
