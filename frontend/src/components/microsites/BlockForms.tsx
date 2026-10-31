'use client';

import { ArrowDownIcon, ArrowUpIcon, PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import { SOCIAL_PLATFORMS, newId, type BlockData, type MicrositeBlock } from '@/lib/microsites';
import { ImageField, SelectField, TextAreaField, TextField, ToggleField } from './fields';

type Data = BlockData;

interface FormProps {
  block: MicrositeBlock;
  micrositeId: number;
  onChange: (data: Data) => void;
}

/** Renders the configuration form for one block. */
export default function BlockForm({ block, micrositeId, onChange }: FormProps) {
  const d = block.data;
  const set = (patch: Data) => onChange({ ...d, ...patch });

  switch (block.type) {
    case 'profile_header':
      return (
        <div className="grid gap-3">
          <ImageField label="Avatar" value={d.avatar_url} onChange={(v) => set({ avatar_url: v })} micrositeId={micrositeId} />
          <TextField label="Display name" value={d.display_name} onChange={(v) => set({ display_name: v })} maxLength={80} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Title" value={d.title} onChange={(v) => set({ title: v })} maxLength={100} />
            <TextField label="Company" value={d.company} onChange={(v) => set({ company: v })} maxLength={100} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Pronouns" value={d.pronouns} onChange={(v) => set({ pronouns: v })} maxLength={30} placeholder="she/her" />
            <TextField label="Location" value={d.location} onChange={(v) => set({ location: v })} maxLength={100} />
          </div>
          <TextAreaField label="Bio" value={d.bio} onChange={(v) => set({ bio: v })} maxLength={500} />
          <ImageField label="Company logo" value={d.company_logo_url} onChange={(v) => set({ company_logo_url: v })} micrositeId={micrositeId} />
          <ToggleField label="Show verified badge" checked={d.verified} onChange={(v) => set({ verified: v })} />
        </div>
      );

    case 'vcard_action':
      return (
        <div className="grid gap-3">
          <p className="text-xs text-muted-foreground">
            Visitors tap this button to save you to their phone&apos;s contacts (.vcf).
          </p>
          <TextField label="Button text" value={d.button_text} onChange={(v) => set({ button_text: v })} maxLength={40} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="First name" value={d.first_name} onChange={(v) => set({ first_name: v })} maxLength={60} />
            <TextField label="Last name" value={d.last_name} onChange={(v) => set({ last_name: v })} maxLength={60} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Organization" value={d.organization} onChange={(v) => set({ organization: v })} maxLength={100} />
            <TextField label="Job title" value={d.title} onChange={(v) => set({ title: v })} maxLength={100} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Mobile phone" type="tel" value={d.phone_mobile} onChange={(v) => set({ phone_mobile: v })} placeholder="+1 303 333 4444" />
            <TextField label="Work phone" type="tel" value={d.phone_work} onChange={(v) => set({ phone_work: v })} />
          </div>
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Work email" type="email" value={d.email_work} onChange={(v) => set({ email_work: v })} />
            <TextField label="Personal email" type="email" value={d.email_personal} onChange={(v) => set({ email_personal: v })} />
          </div>
          <TextField label="Website" type="url" value={d.website} onChange={(v) => set({ website: v })} placeholder="https://" />
          <details className="rounded-lg border border-border px-3 py-2">
            <summary className="cursor-pointer text-xs font-medium text-muted-foreground">Address, note and photo</summary>
            <div className="grid gap-3 mt-3">
              <TextField label="Street" value={d.street} onChange={(v) => set({ street: v })} maxLength={120} />
              <div className="grid grid-cols-2 gap-3">
                <TextField label="City" value={d.city} onChange={(v) => set({ city: v })} maxLength={60} />
                <TextField label="Region" value={d.region} onChange={(v) => set({ region: v })} maxLength={60} />
              </div>
              <div className="grid grid-cols-2 gap-3">
                <TextField label="Postal code" value={d.postal_code} onChange={(v) => set({ postal_code: v })} maxLength={20} />
                <TextField label="Country" value={d.country} onChange={(v) => set({ country: v })} maxLength={60} />
              </div>
              <TextAreaField label="Note" value={d.note} onChange={(v) => set({ note: v })} maxLength={500} rows={2} />
              <ImageField label="Contact photo (defaults to the avatar)" value={d.photo_url} onChange={(v) => set({ photo_url: v })} micrositeId={micrositeId} />
            </div>
          </details>
        </div>
      );

    case 'link_list':
      return (
        <ItemList
          items={d.items ?? []}
          onChange={(items) => set({ items })}
          addLabel="Add link"
          create={() => ({ id: newId('itm'), label: 'New link', url: '', highlight: false })}
          render={(it, update) => (
            <>
              <TextField label="Label" value={it.label} onChange={(v) => update({ label: v })} maxLength={80} />
              <TextField label="URL" value={it.url} onChange={(v) => update({ url: v })} placeholder="https://" />
              <ToggleField label="Highlight" checked={it.highlight} onChange={(v) => update({ highlight: v })} />
            </>
          )}
        />
      );

    case 'contact_buttons':
      return (
        <ItemList
          items={d.items ?? []}
          onChange={(items) => set({ items })}
          addLabel="Add button"
          max={6}
          create={() => ({ id: newId('itm'), kind: 'call', value: '', label: '' })}
          render={(it, update) => (
            <>
              <SelectField
                label="Action"
                value={it.kind}
                onChange={(v) => update({ kind: v })}
                options={[
                  { value: 'call', label: 'Call (tel:)' },
                  { value: 'sms', label: 'Text message (sms:)' },
                  { value: 'email', label: 'Email (mailto:)' },
                  { value: 'whatsapp', label: 'WhatsApp (wa.me)' },
                ]}
              />
              <TextField
                label={it.kind === 'email' ? 'Email address' : 'Phone number (with country code)'}
                value={it.value}
                onChange={(v) => update({ value: v })}
                type={it.kind === 'email' ? 'email' : 'tel'}
              />
              <TextField label="Label (optional)" value={it.label} onChange={(v) => update({ label: v })} maxLength={40} />
            </>
          )}
        />
      );

    case 'social_icons':
      return (
        <ItemList
          items={d.items ?? []}
          onChange={(items) => set({ items })}
          addLabel="Add profile"
          max={20}
          create={() => ({ id: newId('itm'), platform: 'instagram', url: '' })}
          render={(it, update) => (
            <>
              <SelectField
                label="Platform"
                value={it.platform}
                onChange={(v) => update({ platform: v })}
                options={SOCIAL_PLATFORMS.map((p) => ({ value: p.id, label: p.name }))}
              />
              <TextField label="Profile URL" value={it.url} onChange={(v) => update({ url: v })} placeholder="https://" />
            </>
          )}
        />
      );

    case 'text':
      return (
        <div className="grid gap-3">
          <TextAreaField label="Text" value={d.text} onChange={(v) => set({ text: v })} maxLength={2000} rows={5} />
          <SelectField
            label="Alignment"
            value={d.align || 'left'}
            onChange={(v) => set({ align: v })}
            options={[
              { value: 'left', label: 'Left' },
              { value: 'center', label: 'Center' },
            ]}
          />
        </div>
      );

    case 'image':
      return (
        <div className="grid gap-3">
          <ImageField label="Image" value={d.url} onChange={(v) => set({ url: v })} micrositeId={micrositeId} />
          <TextField label="Alt text (describe the image)" value={d.alt} onChange={(v) => set({ alt: v })} maxLength={150} />
          <TextField label="Link (optional)" value={d.link_url} onChange={(v) => set({ link_url: v })} placeholder="https://" />
        </div>
      );

    case 'divider':
      return (
        <SelectField
          label="Style"
          value={d.style || 'line'}
          onChange={(v) => set({ style: v })}
          options={[
            { value: 'line', label: 'Line' },
            { value: 'space', label: 'Space' },
          ]}
        />
      );

    default:
      return <p className="text-sm text-muted-foreground">This block has no settings.</p>;
  }
}

function ItemList({
  items,
  onChange,
  create,
  render,
  addLabel,
  max = 30,
}: {
  items: Data[];
  onChange: (items: Data[]) => void;
  create: () => Data;
  render: (item: Data, update: (patch: Data) => void) => React.ReactNode;
  addLabel: string;
  max?: number;
}) {
  const move = (i: number, dir: -1 | 1) => {
    const j = i + dir;
    if (j < 0 || j >= items.length) return;
    const next = [...items];
    [next[i], next[j]] = [next[j], next[i]];
    onChange(next);
  };
  return (
    <div className="grid gap-3">
      {items.map((it, i) => (
        <div key={it.id ?? i} className="rounded-lg border border-border p-3 grid gap-2">
          {render(it, (patch) => onChange(items.map((x, k) => (k === i ? { ...x, ...patch } : x))))}
          <div className="flex items-center gap-1 justify-end">
            <button type="button" onClick={() => move(i, -1)} disabled={i === 0} className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30" aria-label="Move up">
              <ArrowUpIcon className="h-4 w-4" />
            </button>
            <button type="button" onClick={() => move(i, 1)} disabled={i === items.length - 1} className="p-1 text-muted-foreground hover:text-foreground disabled:opacity-30" aria-label="Move down">
              <ArrowDownIcon className="h-4 w-4" />
            </button>
            <button type="button" onClick={() => onChange(items.filter((_, k) => k !== i))} className="p-1 text-muted-foreground hover:text-red-600" aria-label="Remove">
              <TrashIcon className="h-4 w-4" />
            </button>
          </div>
        </div>
      ))}
      {items.length < max && (
        <button
          type="button"
          onClick={() => onChange([...items, create()])}
          className="inline-flex items-center justify-center gap-1 rounded-lg border border-dashed border-border px-3 py-2 text-sm text-muted-foreground hover:text-foreground hover:bg-muted"
        >
          <PlusIcon className="h-4 w-4" />
          {addLabel}
        </button>
      )}
    </div>
  );
}
