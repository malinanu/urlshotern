'use client';

import { useState, useCallback } from 'react';
import { Collapsible, CollapsibleTrigger, CollapsibleContent } from '@/components/ui/collapsible';
import { PlusIcon, TrashIcon, ChevronDownIcon } from '@heroicons/react/24/outline';

interface StructuredContentEditorProps {
  value: string;
  onChange: (v: string) => void;
}

function formatLabel(key: string): string {
  return key.replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase());
}

function isPlainObjectOfStrings(val: unknown): val is Record<string, string> {
  return (
    typeof val === 'object' &&
    val !== null &&
    !Array.isArray(val) &&
    Object.values(val).every((v) => typeof v === 'string')
  );
}

function isArrayOfObjects(val: unknown): val is Record<string, unknown>[] {
  return Array.isArray(val) && val.length > 0 && typeof val[0] === 'object' && val[0] !== null;
}

function isStringArray(val: unknown): val is string[] {
  return Array.isArray(val) && (val.length === 0 || typeof val[0] === 'string');
}

function emptyShapeOf(item: Record<string, unknown>): Record<string, unknown> {
  const empty: Record<string, unknown> = {};
  for (const key of Object.keys(item)) {
    const v = item[key];
    if (typeof v === 'string') empty[key] = '';
    else if (typeof v === 'number') empty[key] = 0;
    else if (typeof v === 'boolean') empty[key] = false;
    else if (Array.isArray(v)) empty[key] = [];
    else empty[key] = '';
  }
  return empty;
}

// Renders a single field value (not array-of-objects, not plain-object-of-strings)
function FieldInput({
  fieldKey,
  value,
  onChange,
}: {
  fieldKey: string;
  value: unknown;
  onChange: (v: unknown) => void;
}) {
  const isTextarea = /tagline|description|content/i.test(fieldKey);
  const isUrl = fieldKey === 'url';

  if (typeof value === 'string') {
    if (isTextarea) {
      return (
        <textarea
          rows={2}
          value={value}
          onChange={(e) => onChange(e.target.value)}
          className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      );
    }
    return (
      <input
        type={isUrl ? 'url' : 'text'}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    );
  }

  if (typeof value === 'number') {
    return (
      <input
        type="number"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
      />
    );
  }

  // fallback for booleans or other primitives
  return (
    <input
      type="text"
      value={String(value)}
      onChange={(e) => onChange(e.target.value)}
      className="w-full px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
    />
  );
}

function StringArrayField({
  value,
  onChange,
}: {
  value: string[];
  onChange: (v: string[]) => void;
}) {
  return (
    <div className="space-y-1">
      {value.map((item, idx) => (
        <div key={idx} className="flex gap-1">
          <input
            type="text"
            value={item}
            onChange={(e) => {
              const next = [...value];
              next[idx] = e.target.value;
              onChange(next);
            }}
            className="flex-1 px-2 py-1.5 border border-gray-300 rounded text-sm focus:outline-none focus:ring-1 focus:ring-blue-500"
          />
          <button
            type="button"
            onClick={() => onChange(value.filter((_, i) => i !== idx))}
            className="p-1.5 text-red-400 hover:text-red-600"
          >
            <TrashIcon className="w-4 h-4" />
          </button>
        </div>
      ))}
      <button
        type="button"
        onClick={() => onChange([...value, ''])}
        className="flex items-center gap-1 text-xs text-blue-600 hover:text-blue-800 mt-1"
      >
        <PlusIcon className="w-3 h-3" /> Add item
      </button>
    </div>
  );
}

function ObjectField({
  value,
  onChange,
}: {
  value: Record<string, string>;
  onChange: (v: Record<string, string>) => void;
}) {
  return (
    <div className="space-y-2 pl-2 border-l-2 border-gray-200">
      {Object.entries(value).map(([k, v]) => (
        <div key={k}>
          <label className="block text-xs font-medium text-gray-500 mb-0.5">{formatLabel(k)}</label>
          <FieldInput
            fieldKey={k}
            value={v}
            onChange={(newVal) => onChange({ ...value, [k]: newVal as string })}
          />
        </div>
      ))}
    </div>
  );
}

function ArrayOfObjectsField({
  value,
  onChange,
}: {
  value: Record<string, unknown>[];
  onChange: (v: Record<string, unknown>[]) => void;
}) {
  return (
    <div className="space-y-3">
      {value.map((item, idx) => (
        <div key={idx} className="border border-gray-200 rounded p-3 bg-gray-50">
          <div className="flex justify-between items-center mb-2">
            <span className="text-xs font-semibold text-gray-500">Item {idx + 1}</span>
            <button
              type="button"
              onClick={() => onChange(value.filter((_, i) => i !== idx))}
              className="p-1 text-red-400 hover:text-red-600"
            >
              <TrashIcon className="w-4 h-4" />
            </button>
          </div>
          <div className="space-y-2">
            {Object.entries(item).map(([k, v]) => (
              <div key={k}>
                <label className="block text-xs font-medium text-gray-500 mb-0.5">{formatLabel(k)}</label>
                {isStringArray(v) ? (
                  <StringArrayField
                    value={v}
                    onChange={(newVal) => {
                      const next = [...value];
                      next[idx] = { ...item, [k]: newVal };
                      onChange(next);
                    }}
                  />
                ) : (
                  <FieldInput
                    fieldKey={k}
                    value={v}
                    onChange={(newVal) => {
                      const next = [...value];
                      next[idx] = { ...item, [k]: newVal };
                      onChange(next);
                    }}
                  />
                )}
              </div>
            ))}
          </div>
        </div>
      ))}
      <button
        type="button"
        onClick={() => {
          const shape = value.length > 0 ? emptyShapeOf(value[value.length - 1]) : {};
          onChange([...value, shape]);
        }}
        className="flex items-center gap-1 text-sm text-blue-600 hover:text-blue-800"
      >
        <PlusIcon className="w-4 h-4" /> Add Item
      </button>
    </div>
  );
}

function SectionEditor({
  sectionKey,
  sectionData,
  onChange,
}: {
  sectionKey: string;
  sectionData: Record<string, unknown>;
  onChange: (v: Record<string, unknown>) => void;
}) {
  const [open, setOpen] = useState(false);

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <CollapsibleTrigger className="flex w-full items-center justify-between px-4 py-2.5 bg-gray-50 hover:bg-gray-100 border border-gray-200 rounded-md text-sm font-semibold text-gray-700">
        {formatLabel(sectionKey)}
        <ChevronDownIcon
          className={`w-4 h-4 text-gray-500 transition-transform ${open ? 'rotate-180' : ''}`}
        />
      </CollapsibleTrigger>
      <CollapsibleContent className="border border-t-0 border-gray-200 rounded-b-md p-4 space-y-4 bg-white">
        {Object.entries(sectionData).map(([fieldKey, fieldVal]) => (
          <div key={fieldKey}>
            <label className="block text-xs font-semibold text-gray-600 mb-1">{formatLabel(fieldKey)}</label>
            {isStringArray(fieldVal) ? (
              <StringArrayField
                value={fieldVal}
                onChange={(v) => onChange({ ...sectionData, [fieldKey]: v })}
              />
            ) : isArrayOfObjects(fieldVal) ? (
              <ArrayOfObjectsField
                value={fieldVal}
                onChange={(v) => onChange({ ...sectionData, [fieldKey]: v })}
              />
            ) : isPlainObjectOfStrings(fieldVal) ? (
              <ObjectField
                value={fieldVal}
                onChange={(v) => onChange({ ...sectionData, [fieldKey]: v })}
              />
            ) : (
              <FieldInput
                fieldKey={fieldKey}
                value={fieldVal}
                onChange={(v) => onChange({ ...sectionData, [fieldKey]: v })}
              />
            )}
          </div>
        ))}
      </CollapsibleContent>
    </Collapsible>
  );
}

export function StructuredContentEditor({ value, onChange }: StructuredContentEditorProps) {
  const [mode, setMode] = useState<'visual' | 'raw'>('visual');
  const [rawError, setRawError] = useState<string | null>(null);

  let parsed: Record<string, unknown> = {};
  try {
    if (value) parsed = JSON.parse(value);
  } catch {
    // ignore
  }

  const handleSectionChange = useCallback(
    (sectionKey: string, sectionData: Record<string, unknown>) => {
      const next = { ...parsed, [sectionKey]: sectionData };
      onChange(JSON.stringify(next, null, 2));
    },
    [parsed, onChange]
  );

  const handleModeToggle = () => {
    if (mode === 'raw') {
      try {
        if (value) JSON.parse(value);
        setRawError(null);
        setMode('visual');
      } catch {
        setRawError('Invalid JSON — fix before switching to visual mode.');
      }
    } else {
      setRawError(null);
      setMode('raw');
    }
  };

  const topLevelKeys = Object.keys(parsed);

  return (
    <div>
      <div className="flex items-center justify-between mb-2">
        <label className="block text-sm font-medium text-gray-700">Content Blocks</label>
        <button
          type="button"
          onClick={handleModeToggle}
          className="text-xs px-2.5 py-1 rounded-full border border-gray-300 text-gray-600 hover:bg-gray-50"
        >
          {mode === 'visual' ? 'Raw JSON' : 'Visual'}
        </button>
      </div>

      {rawError && (
        <p className="text-xs text-red-600 mb-2">{rawError}</p>
      )}

      {mode === 'visual' ? (
        <div className="space-y-2">
          {topLevelKeys.length === 0 ? (
            <p className="text-xs text-gray-400 py-4 text-center border border-dashed border-gray-300 rounded">
              No content blocks. Switch to Raw JSON to add structured data.
            </p>
          ) : (
            topLevelKeys.map((key) => {
              const sectionData = parsed[key];
              if (typeof sectionData !== 'object' || sectionData === null || Array.isArray(sectionData)) {
                // top-level primitive or array — render as a single field
                return (
                  <div key={key}>
                    <label className="block text-xs font-semibold text-gray-600 mb-1">{formatLabel(key)}</label>
                    <FieldInput
                      fieldKey={key}
                      value={sectionData}
                      onChange={(v) => {
                        const next = { ...parsed, [key]: v };
                        onChange(JSON.stringify(next, null, 2));
                      }}
                    />
                  </div>
                );
              }
              return (
                <SectionEditor
                  key={key}
                  sectionKey={key}
                  sectionData={sectionData as Record<string, unknown>}
                  onChange={(v) => handleSectionChange(key, v)}
                />
              );
            })
          )}
        </div>
      ) : (
        <textarea
          rows={12}
          value={value}
          onChange={(e) => {
            setRawError(null);
            onChange(e.target.value);
          }}
          className="w-full px-3 py-2 border border-gray-300 rounded-md shadow-sm font-mono text-xs focus:outline-none focus:ring-blue-500 focus:border-blue-500"
          placeholder='{"hero": {"title": "...", ...}}'
        />
      )}
    </div>
  );
}
