'use client'

import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useAuth } from '@/hooks/useAuth';
import RichTextEditor from '../editor/RichTextEditor';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';
import { Card } from '../ui/Card';
import { Badge } from '../ui/Badge';
import { toast } from 'react-toastify';
import {
  PlusIcon,
  DocumentTextIcon,
  PhotoIcon,
  PlayIcon,
  ListBulletIcon,
  Bars3Icon,
  XMarkIcon,
  ArrowUpIcon,
  ArrowDownIcon,
  TrashIcon,
  PencilIcon,
  EyeIcon,
  CubeIcon,
  Cog6ToothIcon,
  DevicePhoneMobileIcon,
  ComputerDesktopIcon,
  DeviceTabletIcon
} from '@heroicons/react/24/outline';

interface BuilderBlock {
  id: string;
  type: 'text' | 'image' | 'video' | 'gallery' | 'button' | 'spacer' | 'divider' | 'columns' | 'content-block';
  content: any;
  styles: BlockStyles;
  settings: BlockSettings;
}

interface BlockStyles {
  padding: string;
  margin: string;
  backgroundColor: string;
  textColor: string;
  borderRadius: string;
  border: string;
  textAlign: 'left' | 'center' | 'right';
  fontSize: string;
  fontWeight: string;
}

interface BlockSettings {
  visible: boolean;
  animation?: string;
  customCSS?: string;
  responsive?: {
    mobile: Partial<BlockStyles>;
    tablet: Partial<BlockStyles>;
    desktop: Partial<BlockStyles>;
  };
}

interface ContentBlock {
  id: number;
  name: string;
  type: string;
  content: string;
  category: string;
}

const defaultBlockStyles: BlockStyles = {
  padding: '16px',
  margin: '0px',
  backgroundColor: 'transparent',
  textColor: '#000000',
  borderRadius: '0px',
  border: 'none',
  textAlign: 'left',
  fontSize: '16px',
  fontWeight: 'normal'
};

const defaultBlockSettings: BlockSettings = {
  visible: true,
  animation: 'none',
  customCSS: '',
  responsive: {
    mobile: {},
    tablet: {},
    desktop: {}
  }
};

const blockTypes = [
  { type: 'text', label: 'Text', icon: DocumentTextIcon },
  { type: 'image', label: 'Image', icon: PhotoIcon },
  { type: 'video', label: 'Video', icon: PlayIcon },
  { type: 'gallery', label: 'Gallery', icon: PhotoIcon },
  { type: 'button', label: 'Button', icon: CubeIcon },
  { type: 'spacer', label: 'Spacer', icon: Bars3Icon },
  { type: 'divider', label: 'Divider', icon: Bars3Icon },
  { type: 'columns', label: 'Columns', icon: ListBulletIcon },
  { type: 'content-block', label: 'Content Block', icon: CubeIcon }
];

export function VisualPageBuilder() {
  const { user, token } = useAuth();
  const [blocks, setBlocks] = useState<BuilderBlock[]>([]);
  const [selectedBlockId, setSelectedBlockId] = useState<string | null>(null);
  const [previewMode, setPreviewMode] = useState(false);
  const [devicePreview, setDevicePreview] = useState<'mobile' | 'tablet' | 'desktop'>('desktop');
  const [showBlockLibrary, setShowBlockLibrary] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [contentBlocks, setContentBlocks] = useState<ContentBlock[]>([]);
  const [draggedBlock, setDraggedBlock] = useState<string | null>(null);
  const [dragOverIndex, setDragOverIndex] = useState<number | null>(null);
  const [pageSettings, setPageSettings] = useState({
    title: '',
    slug: '',
    template: 'default',
    isPublished: false
  });

  const canvasRef = useRef<HTMLDivElement>(null);
  const sidebarRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    fetchContentBlocks();
  }, []);

  const fetchContentBlocks = async () => {
    try {
      const response = await fetch('/api/v1/admin/content-blocks', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const contentType = response.headers.get('content-type');
        if (contentType && contentType.includes('application/json')) {
          const data = await response.json();
          setContentBlocks(data);
        } else {
          console.error('Expected JSON response but got:', contentType);
          toast.error('Invalid response format from server');
        }
      } else {
        console.error('Failed to fetch content blocks:', response.status, response.statusText);
        toast.error('Failed to fetch content blocks');
      }
    } catch (error) {
      console.error('Error fetching content blocks:', error);
    }
  };

  const generateBlockId = () => {
    return 'block-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
  };

  const addBlock = (type: string, content?: any) => {
    const newBlock: BuilderBlock = {
      id: generateBlockId(),
      type: type as any,
      content: content || getDefaultContent(type),
      styles: { ...defaultBlockStyles },
      settings: { ...defaultBlockSettings }
    };

    setBlocks(prev => [...prev, newBlock]);
    setSelectedBlockId(newBlock.id);
    setShowBlockLibrary(false);
  };

  const getDefaultContent = (type: string) => {
    switch (type) {
      case 'text':
        return { html: '<p>Click to edit text</p>' };
      case 'image':
        return { src: '', alt: '', width: '100%', height: 'auto' };
      case 'video':
        return { src: '', autoplay: false, controls: true };
      case 'gallery':
        return { images: [], columns: 3 };
      case 'button':
        return { text: 'Button', url: '#', style: 'primary' };
      case 'spacer':
        return { height: '50px' };
      case 'divider':
        return { style: 'solid', color: '#e5e7eb', thickness: '1px' };
      case 'columns':
        return { columns: [{ content: '<p>Column 1</p>' }, { content: '<p>Column 2</p>' }] };
      case 'content-block':
        return { blockId: null, variables: {} };
      default:
        return {};
    }
  };

  const updateBlock = (blockId: string, updates: Partial<BuilderBlock>) => {
    setBlocks(prev => prev.map(block => 
      block.id === blockId ? { ...block, ...updates } : block
    ));
  };

  const deleteBlock = (blockId: string) => {
    setBlocks(prev => prev.filter(block => block.id !== blockId));
    if (selectedBlockId === blockId) {
      setSelectedBlockId(null);
    }
  };

  const moveBlock = (blockId: string, direction: 'up' | 'down') => {
    const currentIndex = blocks.findIndex(block => block.id === blockId);
    if (currentIndex === -1) return;

    const newIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;
    if (newIndex < 0 || newIndex >= blocks.length) return;

    const newBlocks = [...blocks];
    [newBlocks[currentIndex], newBlocks[newIndex]] = [newBlocks[newIndex], newBlocks[currentIndex]];
    
    setBlocks(newBlocks);
  };

  const duplicateBlock = (blockId: string) => {
    const block = blocks.find(b => b.id === blockId);
    if (!block) return;

    const newBlock = {
      ...block,
      id: generateBlockId()
    };

    const index = blocks.findIndex(b => b.id === blockId);
    const newBlocks = [...blocks];
    newBlocks.splice(index + 1, 0, newBlock);

    setBlocks(newBlocks);
  };

  const handleDragStart = (e: React.DragEvent, blockId: string) => {
    setDraggedBlock(blockId);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e: React.DragEvent, index: number) => {
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    setDragOverIndex(index);
  };

  const handleDrop = (e: React.DragEvent, dropIndex: number) => {
    e.preventDefault();
    
    if (!draggedBlock) return;

    const dragIndex = blocks.findIndex(block => block.id === draggedBlock);
    if (dragIndex === -1 || dragIndex === dropIndex) return;

    const newBlocks = [...blocks];
    const draggedItem = newBlocks.splice(dragIndex, 1)[0];
    newBlocks.splice(dropIndex, 0, draggedItem);

    setBlocks(newBlocks);
    setDraggedBlock(null);
    setDragOverIndex(null);
  };

  const savePage = async () => {
    try {
      const pageData = {
        title: pageSettings.title,
        slug: pageSettings.slug,
        content: JSON.stringify(blocks),
        template: pageSettings.template,
        is_published: pageSettings.isPublished,
        builder_data: {
          version: '1.0',
          blocks: blocks
        }
      };

      const response = await fetch('/api/v1/admin/cms/pages', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(pageData),
      });

      if (response.ok) {
        toast.success('Page saved successfully');
      } else {
        try {
          const contentType = response.headers.get('content-type');
          if (contentType && contentType.includes('application/json')) {
            const error = await response.json();
            toast.error(error.error || error.message || 'Failed to save page');
          } else {
            const text = await response.text();
            console.error('Non-JSON error response:', text);
            toast.error(`Server error: ${response.status} ${response.statusText}`);
          }
        } catch (parseError) {
          console.error('Error parsing error response:', parseError);
          toast.error(`Failed to save page: ${response.status} ${response.statusText}`);
        }
      }
    } catch (error) {
      toast.error('Error saving page');
    }
  };

  const getDeviceStyles = () => {
    switch (devicePreview) {
      case 'mobile':
        return { width: '375px', minHeight: '667px' };
      case 'tablet':
        return { width: '768px', minHeight: '1024px' };
      default:
        return { width: '100%', minHeight: '600px' };
    }
  };

  const renderBlock = (block: BuilderBlock, index: number) => {
    const isSelected = selectedBlockId === block.id;
    const blockStyles = {
      ...block.styles,
      ...(devicePreview !== 'desktop' && block.settings.responsive?.[devicePreview] || {})
    };

    const commonProps = {
      className: `builder-block ${isSelected ? 'selected' : ''} ${!previewMode ? 'editable' : ''}`,
      style: {
        padding: blockStyles.padding,
        margin: blockStyles.margin,
        backgroundColor: blockStyles.backgroundColor,
        color: blockStyles.textColor,
        borderRadius: blockStyles.borderRadius,
        border: blockStyles.border,
        textAlign: blockStyles.textAlign,
        fontSize: blockStyles.fontSize,
        fontWeight: blockStyles.fontWeight,
        position: 'relative' as const
      },
      onClick: !previewMode ? (e: React.MouseEvent) => {
        e.stopPropagation();
        setSelectedBlockId(block.id);
      } : undefined,
      draggable: !previewMode,
      onDragStart: !previewMode ? (e: React.DragEvent) => handleDragStart(e, block.id) : undefined,
      onDragOver: !previewMode ? (e: React.DragEvent) => handleDragOver(e, index) : undefined,
      onDrop: !previewMode ? (e: React.DragEvent) => handleDrop(e, index) : undefined
    };

    return (
      <div key={block.id} {...commonProps}>
        {!previewMode && isSelected && (
          <div className="block-toolbar absolute -top-8 left-0 flex items-center space-x-1 bg-blue-600 text-white px-2 py-1 rounded text-xs z-10">
            <button
              onClick={(e) => {
                e.stopPropagation();
                moveBlock(block.id, 'up');
              }}
              disabled={index === 0}
              className="hover:bg-blue-700 p-1 rounded disabled:opacity-50"
            >
              <ArrowUpIcon className="h-3 w-3" />
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                moveBlock(block.id, 'down');
              }}
              disabled={index === blocks.length - 1}
              className="hover:bg-blue-700 p-1 rounded disabled:opacity-50"
            >
              <ArrowDownIcon className="h-3 w-3" />
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                duplicateBlock(block.id);
              }}
              className="hover:bg-blue-700 p-1 rounded"
            >
              <CubeIcon className="h-3 w-3" />
            </button>
            <button
              onClick={(e) => {
                e.stopPropagation();
                deleteBlock(block.id);
              }}
              className="hover:bg-blue-700 p-1 rounded text-red-300"
            >
              <TrashIcon className="h-3 w-3" />
            </button>
          </div>
        )}

        {renderBlockContent(block)}

        {dragOverIndex === index && (
          <div className="drag-indicator absolute top-0 left-0 right-0 h-1 bg-blue-500 z-20"></div>
        )}
      </div>
    );
  };

  const renderBlockContent = (block: BuilderBlock) => {
    switch (block.type) {
      case 'text':
        return (
          <div
            dangerouslySetInnerHTML={{ __html: block.content.html }}
            contentEditable={!previewMode && selectedBlockId === block.id}
            onBlur={(e) => {
              if (!previewMode) {
                updateBlock(block.id, {
                  content: { ...block.content, html: e.currentTarget.innerHTML }
                });
              }
            }}
          />
        );

      case 'image':
        return block.content.src ? (
          <img
            src={block.content.src}
            alt={block.content.alt || ''}
            style={{
              width: block.content.width,
              height: block.content.height,
              objectFit: 'cover'
            }}
          />
        ) : (
          <div className="border-2 border-dashed border-gray-300 p-8 text-center">
            <PhotoIcon className="h-12 w-12 mx-auto text-gray-400 mb-2" />
            <p className="text-gray-500">Click to add image</p>
          </div>
        );

      case 'video':
        return block.content.src ? (
          <video
            src={block.content.src}
            controls={block.content.controls}
            autoPlay={block.content.autoplay}
            style={{ width: '100%', height: 'auto' }}
          />
        ) : (
          <div className="border-2 border-dashed border-gray-300 p-8 text-center">
            <PlayIcon className="h-12 w-12 mx-auto text-gray-400 mb-2" />
            <p className="text-gray-500">Click to add video</p>
          </div>
        );

      case 'button':
        return (
          <button
            className={`btn ${block.content.style}`}
            style={{
              padding: '12px 24px',
              borderRadius: block.styles.borderRadius,
              fontSize: block.styles.fontSize,
              fontWeight: block.styles.fontWeight
            }}
          >
            {block.content.text}
          </button>
        );

      case 'spacer':
        return (
          <div
            style={{ height: block.content.height }}
            className="border-2 border-dashed border-gray-200"
          >
            {!previewMode && (
              <div className="h-full flex items-center justify-center text-gray-400 text-sm">
                Spacer ({block.content.height})
              </div>
            )}
          </div>
        );

      case 'divider':
        return (
          <hr
            style={{
              borderStyle: block.content.style,
              borderColor: block.content.color,
              borderWidth: block.content.thickness,
              margin: '20px 0'
            }}
          />
        );

      case 'columns':
        return (
          <div
            className="flex gap-4"
            style={{
              flexDirection: devicePreview === 'mobile' ? 'column' : 'row'
            }}
          >
            {block.content.columns.map((column: any, index: number) => (
              <div
                key={index}
                className="flex-1"
                dangerouslySetInnerHTML={{ __html: column.content }}
              />
            ))}
          </div>
        );

      case 'content-block':
        const contentBlock = contentBlocks.find(cb => cb.id === block.content.blockId);
        return contentBlock ? (
          <div dangerouslySetInnerHTML={{ __html: contentBlock.content }} />
        ) : (
          <div className="border-2 border-dashed border-gray-300 p-4 text-center">
            <CubeIcon className="h-8 w-8 mx-auto text-gray-400 mb-2" />
            <p className="text-gray-500">Select content block</p>
          </div>
        );

      default:
        return <div>Unknown block type</div>;
    }
  };

  const selectedBlock = blocks.find(block => block.id === selectedBlockId);

  return (
    <div className="h-screen flex bg-gray-100">
      {/* Sidebar */}
      <div
        ref={sidebarRef}
        className={`bg-white shadow-lg transition-all duration-300 ${
          previewMode ? 'w-0 overflow-hidden' : 'w-80'
        }`}
      >
        <div className="p-4 border-b">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-lg font-semibold">Page Builder</h2>
            <Button
              variant="ghost"
              size="sm"
              onClick={() => setShowSettings(true)}
            >
              <Cog6ToothIcon className="h-5 w-5" />
            </Button>
          </div>

          {/* Page Settings */}
          <div className="space-y-3 mb-4">
            <Input
              placeholder="Page title"
              value={pageSettings.title}
              onChange={(e) => setPageSettings(prev => ({ ...prev, title: e.target.value }))}
            />
            <Input
              placeholder="page-slug"
              value={pageSettings.slug}
              onChange={(e) => setPageSettings(prev => ({ ...prev, slug: e.target.value }))}
            />
          </div>

          {/* Actions */}
          <div className="flex space-x-2">
            <Button
              onClick={() => setShowBlockLibrary(true)}
              className="flex-1"
              size="sm"
            >
              <PlusIcon className="h-4 w-4 mr-1" />
              Add Block
            </Button>
            <Button
              onClick={savePage}
              variant="ghost"
              size="sm"
            >
              Save
            </Button>
          </div>
        </div>

        {/* Block Properties */}
        {selectedBlock && !previewMode && (
          <div className="p-4 border-b">
            <h3 className="font-medium mb-3">Block Settings</h3>
            <div className="space-y-3">
              {/* Common style controls */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Padding
                </label>
                <Input
                  value={selectedBlock.styles.padding}
                  onChange={(e) => updateBlock(selectedBlockId!, {
                    styles: { ...selectedBlock.styles, padding: e.target.value }
                  })}
                  placeholder="16px"
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Background Color
                </label>
                <Input
                  type="color"
                  value={selectedBlock.styles.backgroundColor}
                  onChange={(e) => updateBlock(selectedBlockId!, {
                    styles: { ...selectedBlock.styles, backgroundColor: e.target.value }
                  })}
                />
              </div>

              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Text Alignment
                </label>
                <Select
                  value={selectedBlock.styles.textAlign}
                  onChange={(e) => updateBlock(selectedBlockId!, {
                    styles: { ...selectedBlock.styles, textAlign: e.target.value as any }
                  })}
                >
                  <option value="left">Left</option>
                  <option value="center">Center</option>
                  <option value="right">Right</option>
                </Select>
              </div>

              {/* Block-specific controls */}
              {selectedBlock.type === 'image' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Image URL
                  </label>
                  <Input
                    value={selectedBlock.content.src || ''}
                    onChange={(e) => updateBlock(selectedBlockId!, {
                      content: { ...selectedBlock.content, src: e.target.value }
                    })}
                    placeholder="https://example.com/image.jpg"
                  />
                </div>
              )}

              {selectedBlock.type === 'button' && (
                <>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Button Text
                    </label>
                    <Input
                      value={selectedBlock.content.text || ''}
                      onChange={(e) => updateBlock(selectedBlockId!, {
                        content: { ...selectedBlock.content, text: e.target.value }
                      })}
                      placeholder="Button"
                    />
                  </div>
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-1">
                      Button URL
                    </label>
                    <Input
                      value={selectedBlock.content.url || ''}
                      onChange={(e) => updateBlock(selectedBlockId!, {
                        content: { ...selectedBlock.content, url: e.target.value }
                      })}
                      placeholder="https://example.com"
                    />
                  </div>
                </>
              )}

              {selectedBlock.type === 'content-block' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Select Content Block
                  </label>
                  <Select
                    value={selectedBlock.content.blockId || ''}
                    onChange={(e) => updateBlock(selectedBlockId!, {
                      content: { ...selectedBlock.content, blockId: parseInt(e.target.value) }
                    })}
                  >
                    <option value="">Select a block...</option>
                    {contentBlocks.map(cb => (
                      <option key={cb.id} value={cb.id}>
                        {cb.name}
                      </option>
                    ))}
                  </Select>
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* Main Canvas */}
      <div className="flex-1 flex flex-col">
        {/* Toolbar */}
        <div className="bg-white border-b p-4 flex items-center justify-between">
          <div className="flex items-center space-x-4">
            <Button
              onClick={() => setPreviewMode(!previewMode)}
              variant={previewMode ? 'solid' : 'ghost'}
              className="flex items-center space-x-2"
            >
              <EyeIcon className="h-4 w-4" />
              <span>Preview</span>
            </Button>

            {/* Device Preview Toggle */}
            <div className="flex items-center border rounded-lg p-1">
              <button
                onClick={() => setDevicePreview('mobile')}
                className={`p-2 rounded ${devicePreview === 'mobile' ? 'bg-blue-100 text-blue-600' : 'text-gray-600'}`}
              >
                <DevicePhoneMobileIcon className="h-4 w-4" />
              </button>
              <button
                onClick={() => setDevicePreview('tablet')}
                className={`p-2 rounded ${devicePreview === 'tablet' ? 'bg-blue-100 text-blue-600' : 'text-gray-600'}`}
              >
                <DeviceTabletIcon className="h-4 w-4" />
              </button>
              <button
                onClick={() => setDevicePreview('desktop')}
                className={`p-2 rounded ${devicePreview === 'desktop' ? 'bg-blue-100 text-blue-600' : 'text-gray-600'}`}
              >
                <ComputerDesktopIcon className="h-4 w-4" />
              </button>
            </div>
          </div>

          <div className="text-sm text-gray-500">
            {blocks.length} block{blocks.length !== 1 ? 's' : ''}
          </div>
        </div>

        {/* Canvas */}
        <div className="flex-1 overflow-auto p-8 bg-gray-50">
          <div
            ref={canvasRef}
            className="mx-auto bg-white shadow-lg min-h-full"
            style={getDeviceStyles()}
            onClick={() => !previewMode && setSelectedBlockId(null)}
          >
            {blocks.length === 0 ? (
              <div className="h-64 flex items-center justify-center text-gray-500">
                <div className="text-center">
                  <CubeIcon className="h-12 w-12 mx-auto mb-4 text-gray-300" />
                  <h3 className="text-lg font-medium mb-2">Start Building</h3>
                  <p>Add your first block to begin creating your page</p>
                  <Button
                    onClick={() => setShowBlockLibrary(true)}
                    className="mt-4"
                  >
                    <PlusIcon className="h-4 w-4 mr-2" />
                    Add Block
                  </Button>
                </div>
              </div>
            ) : (
              <div className="space-y-0">
                {blocks.map((block, index) => renderBlock(block, index))}
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Block Library Modal */}
      <Modal
        isOpen={showBlockLibrary}
        onClose={() => setShowBlockLibrary(false)}
        title="Add Block"
        size="lg"
      >
        <div className="grid grid-cols-2 md:grid-cols-3 gap-4">
          {blockTypes.map(({ type, label, icon: Icon }) => (
            <button
              key={type}
              onClick={() => addBlock(type)}
              className="p-4 border-2 border-dashed border-gray-200 hover:border-blue-300 hover:bg-blue-50 rounded-lg transition-colors text-center"
            >
              <Icon className="h-8 w-8 mx-auto mb-2 text-gray-600" />
              <div className="font-medium text-gray-900">{label}</div>
            </button>
          ))}
        </div>

        {contentBlocks.length > 0 && (
          <div className="mt-6">
            <h3 className="text-lg font-medium mb-4">Content Blocks</h3>
            <div className="grid grid-cols-1 gap-2 max-h-48 overflow-y-auto">
              {contentBlocks.map((contentBlock) => (
                <button
                  key={contentBlock.id}
                  onClick={() => addBlock('content-block', { blockId: contentBlock.id })}
                  className="p-3 text-left border border-gray-200 hover:bg-gray-50 rounded-lg"
                >
                  <div className="font-medium">{contentBlock.name}</div>
                  <div className="text-sm text-gray-500">{contentBlock.category}</div>
                </button>
              ))}
            </div>
          </div>
        )}
      </Modal>

      {/* Page Settings Modal */}
      <Modal
        isOpen={showSettings}
        onClose={() => setShowSettings(false)}
        title="Page Settings"
        size="md"
      >
        <div className="space-y-4">
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Page Title
            </label>
            <Input
              value={pageSettings.title}
              onChange={(e) => setPageSettings(prev => ({ ...prev, title: e.target.value }))}
              placeholder="Enter page title"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              URL Slug
            </label>
            <Input
              value={pageSettings.slug}
              onChange={(e) => setPageSettings(prev => ({ ...prev, slug: e.target.value }))}
              placeholder="page-slug"
            />
          </div>

          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Template
            </label>
            <Select
              value={pageSettings.template}
              onChange={(e) => setPageSettings(prev => ({ ...prev, template: e.target.value }))}
            >
              <option value="default">Default</option>
              <option value="landing">Landing Page</option>
              <option value="blog">Blog</option>
            </Select>
          </div>

          <div className="flex items-center">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={pageSettings.isPublished}
                onChange={(e) => setPageSettings(prev => ({ ...prev, isPublished: e.target.checked }))}
                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">Publish immediately</span>
            </label>
          </div>

          <div className="flex justify-end space-x-3 pt-4">
            <Button
              variant="ghost"
              onClick={() => setShowSettings(false)}
            >
              Cancel
            </Button>
            <Button onClick={() => {
              setShowSettings(false);
              savePage();
            }}>
              Save Page
            </Button>
          </div>
        </div>
      </Modal>
    </div>
  );
}