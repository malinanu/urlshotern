'use client';

import { ArrowDownIcon, ArrowUpIcon, PlusIcon, TrashIcon } from '@heroicons/react/24/outline';
import { useEffect, useState } from 'react';
import Link from 'next/link';
import { SOCIAL_PLATFORMS, newId, newProduct, type BlockData, type MicrositeBlock } from '@/lib/microsites';
import { ImageField, SelectField, TextAreaField, TextField, TimeZoneField, ToggleField } from './fields';

type Data = BlockData;

interface FormProps {
  block: MicrositeBlock;
  micrositeId: number;
  onChange: (data: Data) => void;
  /** Whether the owner's plan includes forms (Pro and Ultra). */
  canUseForms?: boolean;
}

function ProductFields({ d, set, micrositeId }: { d: Data; set: (patch: Data) => void; micrositeId: number }) {
  return (
    <>
      <ImageField label="Photo" value={d.image_url} onChange={(v) => set({ image_url: v })} micrositeId={micrositeId} />
      <TextField label="Name" value={d.name} onChange={(v) => set({ name: v })} maxLength={80} />
      <div className="grid grid-cols-2 gap-3">
        <TextField label="Price" value={d.price} onChange={(v) => set({ price: v })} maxLength={24} placeholder="$24.99" />
        <TextField label="Was (optional)" value={d.compare_price} onChange={(v) => set({ compare_price: v })} maxLength={24} placeholder="$34.99" />
      </div>
      <TextField label="Badge (optional)" value={d.badge} onChange={(v) => set({ badge: v })} maxLength={20} placeholder="-30%" />
      <TextAreaField label="Description" value={d.description} onChange={(v) => set({ description: v })} maxLength={300} rows={2} />
      <div className="grid grid-cols-2 gap-3">
        <TextField label="Buy link" value={d.buy_url} onChange={(v) => set({ buy_url: v })} placeholder="https:// or WhatsApp link" />
        <TextField label="Button text" value={d.button_text} onChange={(v) => set({ button_text: v })} maxLength={30} />
      </div>
    </>
  );
}

/**
 * A number input that keeps what the owner is typing ("-", "6.") locally and
 * only reports parseable values (or undefined when cleared).
 */
function NumberField({ label, value, onChange, placeholder }: { label: string; value?: number | null; onChange: (v: number | undefined) => void; placeholder?: string }) {
  const [text, setText] = useState(value === undefined || value === null ? '' : String(value));
  useEffect(() => {
    setText((t) => (Number(t) === value || (t.trim() === '' && (value === undefined || value === null)) ? t : value === undefined || value === null ? '' : String(value)));
  }, [value]);
  return (
    <TextField
      label={label}
      value={text}
      placeholder={placeholder}
      onChange={(v) => {
        setText(v);
        if (v.trim() === '') onChange(undefined);
        else if (Number.isFinite(Number(v))) onChange(Number(v));
      }}
    />
  );
}

/** Renders the configuration form for one block. */
export default function BlockForm({ block, micrositeId, onChange, canUseForms }: FormProps) {
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

    case 'product_card':
      return (
        <div className="grid gap-3">
          <ProductFields d={d} set={set} micrositeId={micrositeId} />
        </div>
      );

    case 'product_carousel':
      return (
        <div className="grid gap-3">
          <TextField label="Heading (optional)" value={d.title} onChange={(v) => set({ title: v })} maxLength={60} />
          <ItemList
            items={d.items ?? []}
            onChange={(items) => set({ items })}
            addLabel="Add product"
            max={12}
            create={newProduct}
            render={(it, update) => <ProductFields d={it} set={update} micrositeId={micrositeId} />}
          />
        </div>
      );

    case 'event_details':
      return (
        <div className="grid gap-3">
          <TextField label="Event name" value={d.title} onChange={(v) => set({ title: v })} maxLength={100} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Starts" type="datetime-local" value={d.start} onChange={(v) => set({ start: v })} />
            <TextField label="Ends (optional)" type="datetime-local" value={d.end} onChange={(v) => set({ end: v })} />
          </div>
          <TimeZoneField label="Time zone" value={d.timezone} onChange={(v) => set({ timezone: v })} />
          <TextField label="Venue" value={d.venue} onChange={(v) => set({ venue: v })} maxLength={100} />
          <TextField label="Address" value={d.address} onChange={(v) => set({ address: v })} maxLength={200} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Tickets link" value={d.ticket_url} onChange={(v) => set({ ticket_url: v })} placeholder="https://" />
            <TextField label="Tickets button" value={d.ticket_text} onChange={(v) => set({ ticket_text: v })} maxLength={30} />
          </div>
          <ToggleField
            label="Show “Add to calendar”"
            checked={d.show_add_to_calendar}
            onChange={(v) => set({ show_add_to_calendar: v })}
            hint="Visitors download an .ics file that works with Apple, Google and Outlook calendars."
          />
        </div>
      );

    case 'countdown':
      return (
        <div className="grid gap-3">
          <TextField label="Count down to" type="datetime-local" value={d.target} onChange={(v) => set({ target: v })} />
          <TimeZoneField label="Time zone" value={d.timezone} onChange={(v) => set({ timezone: v })} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Label" value={d.label} onChange={(v) => set({ label: v })} maxLength={40} placeholder="Starts in" />
            <TextField label="When it's over" value={d.done_text} onChange={(v) => set({ done_text: v })} maxLength={60} placeholder="Happening now" />
          </div>
        </div>
      );

    case 'agenda':
      return (
        <div className="grid gap-3">
          <TextField label="Heading (optional)" value={d.title} onChange={(v) => set({ title: v })} maxLength={60} />
          <ItemList
            items={d.items ?? []}
            onChange={(items) => set({ items })}
            addLabel="Add session"
            max={40}
            create={() => ({ id: newId('itm'), time: '', title: 'New session', speaker: '', track: '', detail: '' })}
            render={(it, update) => (
              <>
                <div className="grid grid-cols-[6rem_1fr] gap-3">
                  <TextField label="Time" value={it.time} onChange={(v) => update({ time: v })} maxLength={20} placeholder="09:30" />
                  <TextField label="Session" value={it.title} onChange={(v) => update({ title: v })} maxLength={100} />
                </div>
                <div className="grid grid-cols-2 gap-3">
                  <TextField label="Speaker (optional)" value={it.speaker} onChange={(v) => update({ speaker: v })} maxLength={80} />
                  <TextField label="Track (optional)" value={it.track} onChange={(v) => update({ track: v })} maxLength={30} placeholder="Main hall" />
                </div>
                <TextAreaField label="Details (shown when opened)" value={it.detail} onChange={(v) => update({ detail: v })} maxLength={500} rows={2} />
              </>
            )}
          />
        </div>
      );

    case 'map':
      return (
        <div className="grid gap-3">
          <TextField label="Label" value={d.label} onChange={(v) => set({ label: v })} maxLength={80} placeholder="Venue" />
          <TextField label="Address" value={d.address} onChange={(v) => set({ address: v })} maxLength={200} />
          <p className="text-xs text-muted-foreground">
            Visitors always get a “Get directions” link. Add coordinates to also show a map they can tap to load.
          </p>
          <div className="grid grid-cols-3 gap-3">
            <NumberField label="Latitude" value={d.lat} onChange={(v) => set({ lat: v })} placeholder="6.9271" />
            <NumberField label="Longitude" value={d.lng} onChange={(v) => set({ lng: v })} placeholder="79.8612" />
            <NumberField label="Zoom (3–19)" value={d.zoom || undefined} onChange={(v) => set({ zoom: v })} placeholder="15" />
          </div>
        </div>
      );

    case 'video':
      return (
        <div className="grid gap-3">
          <TextField
            label="YouTube or Vimeo link"
            value={d.url}
            onChange={(v) => set({ url: v })}
            placeholder="https://www.youtube.com/watch?v=…"
            hint="The player loads only when a visitor taps play."
          />
          <TextField label="Title" value={d.title} onChange={(v) => set({ title: v })} maxLength={100} />
          <ImageField label="Cover image (optional)" value={d.poster_url} onChange={(v) => set({ poster_url: v })} micrositeId={micrositeId} />
        </div>
      );

    case 'audio':
      return (
        <div className="grid gap-3">
          <TextField
            label="Spotify link"
            value={d.url}
            onChange={(v) => set({ url: v })}
            placeholder="https://open.spotify.com/track/…"
            hint="Tracks, albums, playlists, episodes and shows."
          />
          <TextField label="Title" value={d.title} onChange={(v) => set({ title: v })} maxLength={100} />
        </div>
      );

    case 'form':
      return (
        <div className="grid gap-3">
          {!canUseForms && (
            <p className="rounded-lg border border-amber-300 bg-amber-50 dark:bg-amber-950/30 px-3 py-2 text-xs text-amber-800 dark:text-amber-200">
              Forms are part of the Pro and Ultra plans. This form stays hidden on your live page until you{' '}
              <Link href="/billing" className="underline font-medium">
                upgrade
              </Link>
              .
            </p>
          )}
          <SelectField
            label="Form type"
            value={d.kind || 'newsletter'}
            onChange={(v) => set({ kind: v, button_text: v === 'contact' ? 'Send message' : 'Subscribe' })}
            options={[
              { value: 'newsletter', label: 'Newsletter: email only' },
              { value: 'contact', label: 'Contact: name, email and message' },
            ]}
          />
          <TextField label="Heading" value={d.title} onChange={(v) => set({ title: v })} maxLength={80} />
          <TextAreaField label="Intro (optional)" value={d.description} onChange={(v) => set({ description: v })} maxLength={300} rows={2} />
          <div className="grid grid-cols-2 gap-3">
            <TextField label="Button text" value={d.button_text} onChange={(v) => set({ button_text: v })} maxLength={30} />
            <TextField label="Thank-you message" value={d.success_text} onChange={(v) => set({ success_text: v })} maxLength={200} placeholder="Thanks! We got it." />
          </div>
          <TextField
            label="Consent note (optional)"
            value={d.consent_text}
            onChange={(v) => set({ consent_text: v })}
            maxLength={300}
            placeholder="We'll only email you about new offers. Unsubscribe any time."
          />
          <ToggleField
            label="Email me when someone submits"
            checked={d.notify}
            onChange={(v) => set({ notify: v })}
            hint="At most one alert a minute. Every submission is kept in the Inbox tab."
          />
        </div>
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
