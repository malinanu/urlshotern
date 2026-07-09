'use client';

import { useEffect, useState } from 'react';
import Layout from '@/components/layout/Layout';
import { sanitizeHtml } from '@/utils/sanitize';

interface BuilderBlock {
  id: string;
  type: string;
  content: any;
  styles: any;
  resolved_html?: string;
}

interface PageData {
  id: number;
  title: string;
  slug: string;
  meta_description?: string;
}

function renderBlock(block: BuilderBlock) {
  const style: React.CSSProperties = {
    padding: block.styles?.padding,
    margin: block.styles?.margin,
    backgroundColor: block.styles?.backgroundColor,
    color: block.styles?.textColor,
    borderRadius: block.styles?.borderRadius,
    border: block.styles?.border,
    textAlign: block.styles?.textAlign,
    fontSize: block.styles?.fontSize,
    fontWeight: block.styles?.fontWeight,
  };

  switch (block.type) {
    case 'text':
      return (
        <div key={block.id} style={style} dangerouslySetInnerHTML={{ __html: sanitizeHtml(block.content?.html) }} />
      );
    case 'image':
      return block.content?.src ? (
        <div key={block.id} style={style}>
          <img
            src={block.content.src}
            alt={block.content.alt || ''}
            style={{ width: block.content.width || '100%', height: block.content.height || 'auto' }}
          />
        </div>
      ) : null;
    case 'button':
      return (
        <div key={block.id} style={style}>
          <a href={block.content?.url || '#'} style={{ display: 'inline-block', padding: '12px 24px', borderRadius: block.styles?.borderRadius }}>
            {block.content?.text || 'Button'}
          </a>
        </div>
      );
    case 'spacer':
      return <div key={block.id} style={{ height: block.content?.height || '50px' }} />;
    case 'divider':
      return (
        <hr
          key={block.id}
          style={{
            borderStyle: block.content?.style || 'solid',
            borderColor: block.content?.color || '#e5e7eb',
            borderWidth: block.content?.thickness || '1px',
            margin: '20px 0',
          }}
        />
      );
    case 'columns':
      return (
        <div key={block.id} style={{ ...style, display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
          {(block.content?.columns || []).map((col: any, i: number) => (
            <div key={i} style={{ flex: 1, minWidth: '200px' }} dangerouslySetInnerHTML={{ __html: sanitizeHtml(col.content) }} />
          ))}
        </div>
      );
    case 'content-block':
      return block.resolved_html ? (
        <div key={block.id} style={style} dangerouslySetInnerHTML={{ __html: sanitizeHtml(block.resolved_html) }} />
      ) : null;
    default:
      return null;
  }
}

export default function PublicBuilderPageClient({ slug }: { slug: string }) {
  const [page, setPage] = useState<PageData | null>(null);
  const [blocks, setBlocks] = useState<BuilderBlock[]>([]);
  const [loading, setLoading] = useState(true);
  const [notFound, setNotFound] = useState(false);

  useEffect(() => {
    fetch(`/api/v1/pages/${slug}`)
      .then((res) => {
        if (!res.ok) { setNotFound(true); setLoading(false); return null; }
        return res.json();
      })
      .then((data) => {
        if (!data) return;
        setPage(data.page);
        setBlocks(data.blocks || []);
        setLoading(false);
      })
      .catch(() => { setNotFound(true); setLoading(false); });
  }, [slug]);

  if (loading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-gray-500">Loading...</div>
        </div>
      </Layout>
    );
  }

  if (notFound || !page) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="text-center">
            <h1 className="text-4xl font-bold text-gray-900 mb-4">404</h1>
            <p className="text-gray-500">Page not found</p>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen">
        {blocks.length > 0 ? (
          <div>{blocks.map((block) => renderBlock(block))}</div>
        ) : (
          <div className="max-w-4xl mx-auto px-4 py-16">
            <h1 className="text-3xl font-bold mb-8">{page.title}</h1>
          </div>
        )}
      </div>
    </Layout>
  );
}
