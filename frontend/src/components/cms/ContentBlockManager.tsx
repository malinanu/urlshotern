'use client'

import React, { useState, useEffect } from 'react';
import { adminAuthService } from '@/lib/admin-auth-service';
import RichTextEditor from '../editor/RichTextEditor';
import { Modal } from '../ui/modal';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Select } from '../ui/select';
import { Textarea } from '../ui/textarea';
import { Badge } from '../ui/badge';
import { Card } from '../ui/card';
import { toast } from 'react-toastify';
import {
  PlusIcon,
  PencilIcon,
  TrashIcon,
  DocumentDuplicateIcon,
  EyeIcon,
  CodeBracketIcon,
  PhotoIcon,
  VideoCameraIcon,
  DocumentTextIcon,
  ListBulletIcon,
  MagnifyingGlassIcon,
  CheckCircleIcon,
  XCircleIcon
} from '@heroicons/react/24/outline';

// Match the backend ContentBlock model
interface ContentBlock {
  id: number;
  name: string;
  identifier: string;
  content: string;
  type: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

// Match the backend ContentBlockRequest model
interface ContentBlockFormData {
  name: string;
  identifier: string;
  content: string;
  type: string;
  is_active: boolean;
  sort_order: number;
}

// Common block types
const blockTypes = [
  { value: 'callout', label: 'Callout Box', icon: DocumentTextIcon },
  { value: 'cta', label: 'Call to Action', icon: PhotoIcon },
  { value: 'warning', label: 'Warning Box', icon: VideoCameraIcon },
  { value: 'info', label: 'Info Box', icon: DocumentTextIcon },
  { value: 'policy', label: 'Policy Section', icon: ListBulletIcon },
  { value: 'hero', label: 'Hero Section', icon: PhotoIcon },
  { value: 'testimonial', label: 'Testimonial', icon: DocumentTextIcon },
  { value: 'pricing', label: 'Pricing Component', icon: ListBulletIcon },
  { value: 'footer', label: 'Footer Block', icon: ListBulletIcon },
  { value: 'html', label: 'Custom HTML', icon: CodeBracketIcon }
];

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function ContentBlockManager() {
  const [blocks, setBlocks] = useState<ContentBlock[]>([]);
  const [filteredBlocks, setFilteredBlocks] = useState<ContentBlock[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingBlock, setEditingBlock] = useState<ContentBlock | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewBlock, setPreviewBlock] = useState<ContentBlock | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [filterActive, setFilterActive] = useState('all');
  const [formData, setFormData] = useState<ContentBlockFormData>({
    name: '',
    identifier: '',
    content: '',
    type: 'callout',
    is_active: true,
    sort_order: 0
  });

  useEffect(() => {
    fetchBlocks();
  }, []);

  useEffect(() => {
    filterBlocks();
  }, [blocks, searchTerm, filterType, filterActive]);

  const fetchBlocks = async () => {
    try {
      const token = adminAuthService.getAdminToken();
      if (!token) {
        toast.error('Authentication required');
        return;
      }

      const response = await fetch(`${API_BASE_URL}/api/v1/admin/content-blocks`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setBlocks(data || []);
      } else {
        toast.error('Failed to fetch content blocks');
      }
    } catch (error) {
      console.error('Error fetching content blocks:', error);
      toast.error('Error fetching content blocks');
    } finally {
      setLoading(false);
    }
  };

  const filterBlocks = () => {
    let filtered = [...blocks];

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(block =>
        block.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        block.identifier.toLowerCase().includes(searchTerm.toLowerCase()) ||
        block.type.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Type filter
    if (filterType !== 'all') {
      filtered = filtered.filter(block => block.type === filterType);
    }

    // Active status filter
    if (filterActive !== 'all') {
      const isActive = filterActive === 'active';
      filtered = filtered.filter(block => block.is_active === isActive);
    }

    setFilteredBlocks(filtered);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Validate identifier format (lowercase, hyphens only)
    if (!/^[a-z0-9-]+$/.test(formData.identifier)) {
      toast.error('Identifier must be lowercase letters, numbers, and hyphens only');
      return;
    }

    try {
      const token = adminAuthService.getAdminToken();
      const url = editingBlock
        ? `${API_BASE_URL}/api/v1/admin/content-blocks/${editingBlock.id}`
        : `${API_BASE_URL}/api/v1/admin/content-blocks`;

      const method = editingBlock ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        toast.success(`Content block ${editingBlock ? 'updated' : 'created'} successfully`);
        setShowForm(false);
        setEditingBlock(null);
        resetForm();
        fetchBlocks();
      } else {
        const error = await response.json();
        toast.error(error.message || 'Failed to save content block');
      }
    } catch (error) {
      console.error('Error saving content block:', error);
      toast.error('Error saving content block');
    }
  };

  const handleEdit = (block: ContentBlock) => {
    setEditingBlock(block);
    setFormData({
      name: block.name,
      identifier: block.identifier,
      content: block.content,
      type: block.type,
      is_active: block.is_active,
      sort_order: block.sort_order
    });
    setShowForm(true);
  };

  const handleDelete = async (block: ContentBlock) => {
    if (!confirm(`Are you sure you want to delete the content block "${block.name}"?`)) {
      return;
    }

    try {
      const token = adminAuthService.getAdminToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/content-blocks/${block.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        toast.success('Content block deleted successfully');
        fetchBlocks();
      } else {
        toast.error('Failed to delete content block');
      }
    } catch (error) {
      console.error('Error deleting content block:', error);
      toast.error('Error deleting content block');
    }
  };

  const handleDuplicate = (block: ContentBlock) => {
    setFormData({
      name: `${block.name} (Copy)`,
      identifier: `${block.identifier}-copy`,
      content: block.content,
      type: block.type,
      is_active: false, // Start as inactive
      sort_order: block.sort_order + 1
    });
    setShowForm(true);
  };

  const handlePreview = async (block: ContentBlock) => {
    try {
      const token = adminAuthService.getAdminToken();
      const response = await fetch(`${API_BASE_URL}/api/v1/admin/content-blocks/${block.id}/render`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const rendered = await response.json();
        setPreviewBlock({ ...block, content: rendered.content });
        setShowPreview(true);
      } else {
        toast.error('Failed to render content block');
      }
    } catch (error) {
      console.error('Error rendering content block:', error);
      toast.error('Error rendering content block');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      identifier: '',
      content: '',
      type: 'callout',
      is_active: true,
      sort_order: 0
    });
  };

  const getBlockIcon = (type: string) => {
    const blockType = blockTypes.find(bt => bt.value === type);
    const Icon = blockType?.icon || DocumentTextIcon;
    return <Icon className="h-5 w-5" />;
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Content Blocks</h2>
          <p className="text-gray-600">Create and manage reusable content components</p>
        </div>
        <Button
          onClick={() => {
            setEditingBlock(null);
            resetForm();
            setShowForm(true);
          }}
          className="flex items-center space-x-2"
        >
          <PlusIcon className="h-5 w-5" />
          <span>New Block</span>
        </Button>
      </div>

      {/* Filters */}
      <Card className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
          <div className="relative">
            <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-3 text-gray-400" />
            <Input
              type="text"
              placeholder="Search blocks..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select
            value={filterType}
            onChange={(e) => setFilterType(e.target.value)}
          >
            <option value="all">All Types</option>
            {blockTypes.map(type => (
              <option key={type.value} value={type.value}>
                {type.label}
              </option>
            ))}
          </Select>
          <Select
            value={filterActive}
            onChange={(e) => setFilterActive(e.target.value)}
          >
            <option value="all">All Status</option>
            <option value="active">Active Only</option>
            <option value="inactive">Inactive Only</option>
          </Select>
          <div className="text-sm text-gray-500 flex items-center">
            {filteredBlocks.length} block{filteredBlocks.length !== 1 ? 's' : ''}
          </div>
        </div>
      </Card>

      {/* Content Blocks Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredBlocks.map((block) => (
          <Card key={block.id} className="p-6">
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0 p-2 bg-blue-100 rounded-lg">
                  {getBlockIcon(block.type)}
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{block.name}</h3>
                  <p className="text-xs text-gray-500 font-mono">{block.identifier}</p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant={block.is_active ? 'default' : 'secondary'}>
                  {block.is_active ? (
                    <span className="flex items-center space-x-1">
                      <CheckCircleIcon className="h-3 w-3" />
                      <span>Active</span>
                    </span>
                  ) : (
                    <span className="flex items-center space-x-1">
                      <XCircleIcon className="h-3 w-3" />
                      <span>Inactive</span>
                    </span>
                  )}
                </Badge>
              </div>
            </div>

            <div className="space-y-2 mb-4">
              <div className="text-xs text-gray-500">
                <span className="font-medium">Type:</span> {block.type}
              </div>
              <div className="text-xs text-gray-500">
                <span className="font-medium">Order:</span> {block.sort_order}
              </div>
              <div className="text-xs text-gray-500">
                <span className="font-medium">Updated:</span> {formatDate(block.updated_at)}
              </div>
            </div>

            <div className="flex justify-end space-x-2">
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handlePreview(block)}
                title="Preview block"
              >
                <EyeIcon className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDuplicate(block)}
                title="Duplicate block"
              >
                <DocumentDuplicateIcon className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleEdit(block)}
                title="Edit block"
              >
                <PencilIcon className="h-4 w-4" />
              </Button>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDelete(block)}
                title="Delete block"
                className="text-red-600 hover:text-red-700"
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            </div>
          </Card>
        ))}
      </div>

      {filteredBlocks.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-500 text-lg mb-2">No content blocks found</div>
          <p className="text-gray-400">
            {searchTerm || filterType !== 'all' || filterActive !== 'all'
              ? 'Try adjusting your filters'
              : 'Create your first content block to get started'}
          </p>
        </div>
      )}

      {/* Create/Edit Form Modal */}
      <Modal
        isOpen={showForm}
        onClose={() => {
          setShowForm(false);
          setEditingBlock(null);
          resetForm();
        }}
        title={editingBlock ? 'Edit Content Block' : 'Create New Content Block'}
        size="xl"
      >
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Block Name *
              </label>
              <Input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) => setFormData(prev => ({ ...prev, name: e.target.value }))}
                required
                placeholder="Privacy - Data Collection Summary"
              />
            </div>

            <div>
              <label htmlFor="identifier" className="block text-sm font-medium text-gray-700 mb-2">
                Identifier * <span className="text-xs text-gray-500">(lowercase, hyphens only)</span>
              </label>
              <Input
                id="identifier"
                type="text"
                value={formData.identifier}
                onChange={(e) => setFormData(prev => ({ ...prev, identifier: e.target.value.toLowerCase() }))}
                required
                disabled={!!editingBlock} // Can't change identifier on edit
                placeholder="privacy-data-collection-summary"
                pattern="[a-z0-9-]+"
                title="Only lowercase letters, numbers, and hyphens"
              />
              {editingBlock && (
                <p className="text-xs text-gray-500 mt-1">Identifier cannot be changed after creation</p>
              )}
            </div>

            <div>
              <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-2">
                Block Type *
              </label>
              <Select
                id="type"
                value={formData.type}
                onChange={(e) => setFormData(prev => ({ ...prev, type: e.target.value }))}
                required
              >
                {blockTypes.map(type => (
                  <option key={type.value} value={type.value}>
                    {type.label}
                  </option>
                ))}
              </Select>
            </div>

            <div>
              <label htmlFor="sort_order" className="block text-sm font-medium text-gray-700 mb-2">
                Sort Order
              </label>
              <Input
                id="sort_order"
                type="number"
                value={formData.sort_order}
                onChange={(e) => setFormData(prev => ({ ...prev, sort_order: parseInt(e.target.value) || 0 }))}
                placeholder="0"
              />
            </div>
          </div>

          <div className="flex items-center">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.is_active}
                onChange={(e) => setFormData(prev => ({ ...prev, is_active: e.target.checked }))}
                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">Active (visible on website)</span>
            </label>
          </div>

          {/* Content */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Content * <span className="text-xs text-gray-500">(HTML content)</span>
            </label>
            <Textarea
              value={formData.content}
              onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
              placeholder="<div>Your HTML content here...</div>"
              rows={15}
              className="font-mono text-sm"
              required
            />
            <p className="text-xs text-gray-500 mt-1">
              You can use {{`{{variable}}`}} placeholders for dynamic content
            </p>
          </div>

          <div className="flex justify-end space-x-3 pt-4 border-t">
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setShowForm(false);
                setEditingBlock(null);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button type="submit">
              {editingBlock ? 'Update Block' : 'Create Block'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Preview Modal */}
      <Modal
        isOpen={showPreview}
        onClose={() => {
          setShowPreview(false);
          setPreviewBlock(null);
        }}
        title={`Preview: ${previewBlock?.name}`}
        size="lg"
      >
        {previewBlock && (
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <Badge variant="secondary">{previewBlock.type}</Badge>
              <Badge variant={previewBlock.is_active ? 'default' : 'secondary'}>
                {previewBlock.is_active ? 'Active' : 'Inactive'}
              </Badge>
              <span className="text-sm text-gray-500 font-mono">{previewBlock.identifier}</span>
            </div>
            <div className="border rounded-lg p-6 bg-white">
              <div
                dangerouslySetInnerHTML={{ __html: previewBlock.content }}
                className="prose prose-sm max-w-none"
              />
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}
