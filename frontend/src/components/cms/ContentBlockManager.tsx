'use client'

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import RichTextEditor from '../editor/RichTextEditor';
import { Modal } from '../ui/Modal';
import { Button } from '../ui/Button';
import { Input } from '../ui/Input';
import { Select } from '../ui/Select';
import { Textarea } from '../ui/Textarea';
import { Badge } from '../ui/Badge';
import { Card } from '../ui/Card';
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
  MagnifyingGlassIcon
} from '@heroicons/react/24/outline';

interface ContentBlock {
  id: number;
  name: string;
  type: 'text' | 'image' | 'video' | 'form' | 'list' | 'html';
  content: string;
  variables?: Record<string, string>;
  is_global: boolean;
  category: string;
  description: string;
  created_at: string;
  updated_at: string;
  usage_count?: number;
}

interface ContentBlockFormData {
  name: string;
  type: 'text' | 'image' | 'video' | 'form' | 'list' | 'html';
  content: string;
  variables: Record<string, string>;
  is_global: boolean;
  category: string;
  description: string;
}

const blockTypes = [
  { value: 'text', label: 'Text Content', icon: DocumentTextIcon },
  { value: 'image', label: 'Image Block', icon: PhotoIcon },
  { value: 'video', label: 'Video Block', icon: VideoCameraIcon },
  { value: 'form', label: 'Form Block', icon: DocumentTextIcon },
  { value: 'list', label: 'List Block', icon: ListBulletIcon },
  { value: 'html', label: 'HTML Block', icon: CodeBracketIcon }
];

const categories = [
  'general',
  'header',
  'footer',
  'sidebar',
  'hero',
  'testimonial',
  'pricing',
  'contact',
  'custom'
];

export function ContentBlockManager() {
  const { user, token } = useAuth();
  const [blocks, setBlocks] = useState<ContentBlock[]>([]);
  const [filteredBlocks, setFilteredBlocks] = useState<ContentBlock[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingBlock, setEditingBlock] = useState<ContentBlock | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewBlock, setPreviewBlock] = useState<ContentBlock | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterType, setFilterType] = useState('all');
  const [filterCategory, setFilterCategory] = useState('all');
  const [formData, setFormData] = useState<ContentBlockFormData>({
    name: '',
    type: 'text',
    content: '',
    variables: {},
    is_global: true,
    category: 'general',
    description: ''
  });

  useEffect(() => {
    fetchBlocks();
  }, []);

  useEffect(() => {
    filterBlocks();
  }, [blocks, searchTerm, filterType, filterCategory]);

  const fetchBlocks = async () => {
    try {
      const response = await fetch('/api/v1/admin/content-blocks', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setBlocks(data);
      } else {
        toast.error('Failed to fetch content blocks');
      }
    } catch (error) {
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
        block.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        block.category.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Type filter
    if (filterType !== 'all') {
      filtered = filtered.filter(block => block.type === filterType);
    }

    // Category filter
    if (filterCategory !== 'all') {
      filtered = filtered.filter(block => block.category === filterCategory);
    }

    setFilteredBlocks(filtered);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const url = editingBlock 
        ? `/api/v1/admin/content-blocks/${editingBlock.id}`
        : '/api/v1/admin/content-blocks';
      
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
        toast.error(error.error || 'Failed to save content block');
      }
    } catch (error) {
      toast.error('Error saving content block');
    }
  };

  const handleEdit = (block: ContentBlock) => {
    setEditingBlock(block);
    setFormData({
      name: block.name,
      type: block.type,
      content: block.content,
      variables: block.variables || {},
      is_global: block.is_global,
      category: block.category,
      description: block.description
    });
    setShowForm(true);
  };

  const handleDelete = async (block: ContentBlock) => {
    if (!confirm('Are you sure you want to delete this content block?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/admin/content-blocks/${block.id}`, {
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
      toast.error('Error deleting content block');
    }
  };

  const handleDuplicate = async (block: ContentBlock) => {
    setFormData({
      name: `${block.name} (Copy)`,
      type: block.type,
      content: block.content,
      variables: block.variables || {},
      is_global: block.is_global,
      category: block.category,
      description: block.description
    });
    setShowForm(true);
  };

  const handlePreview = async (block: ContentBlock) => {
    try {
      const response = await fetch(`/api/v1/admin/content-blocks/${block.id}/render`, {
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
      toast.error('Error rendering content block');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      type: 'text',
      content: '',
      variables: {},
      is_global: true,
      category: 'general',
      description: ''
    });
  };

  const addVariable = () => {
    const key = prompt('Enter variable name:');
    if (key) {
      setFormData(prev => ({
        ...prev,
        variables: {
          ...prev.variables,
          [key]: ''
        }
      }));
    }
  };

  const updateVariable = (key: string, value: string) => {
    setFormData(prev => ({
      ...prev,
      variables: {
        ...prev.variables,
        [key]: value
      }
    }));
  };

  const removeVariable = (key: string) => {
    setFormData(prev => {
      const newVariables = { ...prev.variables };
      delete newVariables[key];
      return {
        ...prev,
        variables: newVariables
      };
    });
  };

  const getBlockIcon = (type: string) => {
    const blockType = blockTypes.find(bt => bt.value === type);
    const Icon = blockType?.icon || DocumentTextIcon;
    return <Icon className="h-5 w-5" />;
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
          onClick={() => setShowForm(true)}
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
            value={filterCategory}
            onChange={(e) => setFilterCategory(e.target.value)}
          >
            <option value="all">All Categories</option>
            {categories.map(category => (
              <option key={category} value={category}>
                {category.charAt(0).toUpperCase() + category.slice(1)}
              </option>
            ))}
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
                  <p className="text-sm text-gray-500">{block.category}</p>
                </div>
              </div>
              <div className="flex items-center space-x-2">
                <Badge variant={block.is_global ? 'success' : 'secondary'}>
                  {block.is_global ? 'Global' : 'Local'}
                </Badge>
              </div>
            </div>

            <p className="text-sm text-gray-600 mb-4 line-clamp-2">
              {block.description}
            </p>

            {block.usage_count !== undefined && (
              <div className="text-xs text-gray-500 mb-4">
                Used {block.usage_count} time{block.usage_count !== 1 ? 's' : ''}
              </div>
            )}

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
            {searchTerm || filterType !== 'all' || filterCategory !== 'all'
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
                placeholder="Enter block name"
              />
            </div>

            <div>
              <label htmlFor="type" className="block text-sm font-medium text-gray-700 mb-2">
                Block Type *
              </label>
              <Select
                id="type"
                value={formData.type}
                onChange={(e) => setFormData(prev => ({ ...prev, type: e.target.value as any }))}
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
              <label htmlFor="category" className="block text-sm font-medium text-gray-700 mb-2">
                Category *
              </label>
              <Select
                id="category"
                value={formData.category}
                onChange={(e) => setFormData(prev => ({ ...prev, category: e.target.value }))}
                required
              >
                {categories.map(category => (
                  <option key={category} value={category}>
                    {category.charAt(0).toUpperCase() + category.slice(1)}
                  </option>
                ))}
              </Select>
            </div>

            <div className="flex items-center">
              <label className="flex items-center">
                <input
                  type="checkbox"
                  checked={formData.is_global}
                  onChange={(e) => setFormData(prev => ({ ...prev, is_global: e.target.checked }))}
                  className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
                />
                <span className="ml-2 text-sm text-gray-700">Global Block</span>
              </label>
            </div>
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Brief description of this content block"
              rows={2}
            />
          </div>

          {/* Variables */}
          <div>
            <div className="flex justify-between items-center mb-3">
              <label className="block text-sm font-medium text-gray-700">
                Template Variables
              </label>
              <Button
                type="button"
                variant="ghost"
                size="sm"
                onClick={addVariable}
              >
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Variable
              </Button>
            </div>
            {Object.keys(formData.variables).length > 0 ? (
              <div className="space-y-3">
                {Object.entries(formData.variables).map(([key, value]) => (
                  <div key={key} className="flex items-center space-x-3">
                    <div className="flex-1">
                      <Input
                        type="text"
                        value={key}
                        disabled
                        className="font-mono text-sm"
                      />
                    </div>
                    <div className="flex-2">
                      <Input
                        type="text"
                        value={value}
                        onChange={(e) => updateVariable(key, e.target.value)}
                        placeholder="Default value"
                      />
                    </div>
                    <Button
                      type="button"
                      variant="ghost"
                      size="sm"
                      onClick={() => removeVariable(key)}
                      className="text-red-600 hover:text-red-700"
                    >
                      <TrashIcon className="h-4 w-4" />
                    </Button>
                  </div>
                ))}
              </div>
            ) : (
              <p className="text-sm text-gray-500 italic">
                No variables defined. Variables allow you to make content dynamic.
              </p>
            )}
          </div>

          {/* Content */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Content *
            </label>
            {formData.type === 'html' ? (
              <Textarea
                value={formData.content}
                onChange={(e) => setFormData(prev => ({ ...prev, content: e.target.value }))}
                placeholder="Enter HTML content..."
                rows={10}
                className="font-mono text-sm"
                required
              />
            ) : (
              <RichTextEditor
                value={formData.content}
                onChange={(content) => setFormData(prev => ({ ...prev, content }))}
                placeholder="Enter your content here..."
              />
            )}
          </div>

          <div className="flex justify-end space-x-3">
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
              <Badge variant={previewBlock.is_global ? 'success' : 'secondary'}>
                {previewBlock.is_global ? 'Global' : 'Local'}
              </Badge>
              <span className="text-sm text-gray-500">{previewBlock.category}</span>
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