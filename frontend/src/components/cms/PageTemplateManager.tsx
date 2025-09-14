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
  DocumentTextIcon,
  MagnifyingGlassIcon,
  CubeIcon,
  SparklesIcon
} from '@heroicons/react/24/outline';

interface PageTemplate {
  id: number;
  name: string;
  slug: string;
  content: string;
  variables?: Record<string, string>;
  is_default: boolean;
  category: string;
  description: string;
  thumbnail?: string;
  created_at: string;
  updated_at: string;
  usage_count?: number;
}

interface PageTemplateFormData {
  name: string;
  slug: string;
  content: string;
  variables: Record<string, string>;
  is_default: boolean;
  category: string;
  description: string;
  thumbnail?: string;
}

const templateCategories = [
  'general',
  'landing',
  'blog',
  'product',
  'about',
  'contact',
  'portfolio',
  'documentation',
  'custom'
];

export function PageTemplateManager() {
  const { user, token } = useAuth();
  const [templates, setTemplates] = useState<PageTemplate[]>([]);
  const [filteredTemplates, setFilteredTemplates] = useState<PageTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<PageTemplate | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewTemplate, setPreviewTemplate] = useState<PageTemplate | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  const [filterCategory, setFilterCategory] = useState('all');
  const [formData, setFormData] = useState<PageTemplateFormData>({
    name: '',
    slug: '',
    content: '',
    variables: {},
    is_default: false,
    category: 'general',
    description: '',
    thumbnail: ''
  });

  useEffect(() => {
    fetchTemplates();
  }, []);

  useEffect(() => {
    filterTemplates();
  }, [templates, searchTerm, filterCategory]);

  const fetchTemplates = async () => {
    try {
      const response = await fetch('/api/v1/admin/page-templates', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setTemplates(data);
      } else {
        toast.error('Failed to fetch page templates');
      }
    } catch (error) {
      toast.error('Error fetching page templates');
    } finally {
      setLoading(false);
    }
  };

  const filterTemplates = () => {
    let filtered = [...templates];

    // Search filter
    if (searchTerm) {
      filtered = filtered.filter(template =>
        template.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.category.toLowerCase().includes(searchTerm.toLowerCase()) ||
        template.slug.toLowerCase().includes(searchTerm.toLowerCase())
      );
    }

    // Category filter
    if (filterCategory !== 'all') {
      filtered = filtered.filter(template => template.category === filterCategory);
    }

    setFilteredTemplates(filtered);
  };

  const generateSlug = (name: string) => {
    return name
      .toLowerCase()
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-|-$/g, '');
  };

  const handleNameChange = (name: string) => {
    setFormData(prev => ({
      ...prev,
      name,
      slug: editingTemplate ? prev.slug : generateSlug(name)
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const url = editingTemplate 
        ? `/api/v1/admin/page-templates/${editingTemplate.id}`
        : '/api/v1/admin/page-templates';
      
      const method = editingTemplate ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        toast.success(`Template ${editingTemplate ? 'updated' : 'created'} successfully`);
        setShowForm(false);
        setEditingTemplate(null);
        resetForm();
        fetchTemplates();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to save template');
      }
    } catch (error) {
      toast.error('Error saving template');
    }
  };

  const handleEdit = (template: PageTemplate) => {
    setEditingTemplate(template);
    setFormData({
      name: template.name,
      slug: template.slug,
      content: template.content,
      variables: template.variables || {},
      is_default: template.is_default,
      category: template.category,
      description: template.description,
      thumbnail: template.thumbnail || ''
    });
    setShowForm(true);
  };

  const handleDelete = async (template: PageTemplate) => {
    if (!confirm('Are you sure you want to delete this template?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/admin/page-templates/${template.id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        toast.success('Template deleted successfully');
        fetchTemplates();
      } else {
        toast.error('Failed to delete template');
      }
    } catch (error) {
      toast.error('Error deleting template');
    }
  };

  const handleDuplicate = async (template: PageTemplate) => {
    setFormData({
      name: `${template.name} (Copy)`,
      slug: generateSlug(`${template.name}-copy`),
      content: template.content,
      variables: template.variables || {},
      is_default: false,
      category: template.category,
      description: template.description,
      thumbnail: template.thumbnail || ''
    });
    setShowForm(true);
  };

  const handlePreview = async (template: PageTemplate) => {
    try {
      const response = await fetch(`/api/v1/admin/page-templates/${template.id}/render`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const rendered = await response.json();
        setPreviewTemplate({ ...template, content: rendered.content });
        setShowPreview(true);
      } else {
        toast.error('Failed to render template');
      }
    } catch (error) {
      toast.error('Error rendering template');
    }
  };

  const handleSetDefault = async (template: PageTemplate) => {
    try {
      const response = await fetch(`/api/v1/admin/page-templates/${template.id}/set-default`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        toast.success('Default template updated');
        fetchTemplates();
      } else {
        toast.error('Failed to set default template');
      }
    } catch (error) {
      toast.error('Error setting default template');
    }
  };

  const resetForm = () => {
    setFormData({
      name: '',
      slug: '',
      content: '',
      variables: {},
      is_default: false,
      category: 'general',
      description: '',
      thumbnail: ''
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
          <h2 className="text-2xl font-bold text-gray-900">Page Templates</h2>
          <p className="text-gray-600">Create and manage page templates with dynamic content</p>
        </div>
        <Button
          onClick={() => setShowForm(true)}
          className="flex items-center space-x-2"
        >
          <PlusIcon className="h-5 w-5" />
          <span>New Template</span>
        </Button>
      </div>

      {/* Filters */}
      <Card className="p-6">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
          <div className="relative">
            <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-3 text-gray-400" />
            <Input
              type="text"
              placeholder="Search templates..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          <Select
            value={filterCategory}
            onChange={(e) => setFilterCategory(e.target.value)}
          >
            <option value="all">All Categories</option>
            {templateCategories.map(category => (
              <option key={category} value={category}>
                {category.charAt(0).toUpperCase() + category.slice(1)}
              </option>
            ))}
          </Select>
          <div className="text-sm text-gray-500 flex items-center">
            {filteredTemplates.length} template{filteredTemplates.length !== 1 ? 's' : ''}
          </div>
        </div>
      </Card>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredTemplates.map((template) => (
          <Card key={template.id} className="p-6">
            {template.thumbnail && (
              <div className="mb-4">
                <img
                  src={template.thumbnail}
                  alt={template.name}
                  className="w-full h-32 object-cover rounded-lg"
                />
              </div>
            )}
            
            <div className="flex items-start justify-between mb-4">
              <div className="flex items-center space-x-3">
                <div className="flex-shrink-0 p-2 bg-purple-100 rounded-lg">
                  <DocumentTextIcon className="h-5 w-5 text-purple-600" />
                </div>
                <div>
                  <h3 className="font-semibold text-gray-900">{template.name}</h3>
                  <p className="text-sm text-gray-500">{template.slug}</p>
                </div>
              </div>
              <div className="flex flex-col items-end space-y-2">
                {template.is_default && (
                  <Badge variant="warning" className="flex items-center">
                    <SparklesIcon className="h-3 w-3 mr-1" />
                    Default
                  </Badge>
                )}
                <Badge variant="secondary">
                  {template.category}
                </Badge>
              </div>
            </div>

            <p className="text-sm text-gray-600 mb-4 line-clamp-2">
              {template.description}
            </p>

            {template.usage_count !== undefined && (
              <div className="text-xs text-gray-500 mb-4">
                Used {template.usage_count} time{template.usage_count !== 1 ? 's' : ''}
              </div>
            )}

            <div className="flex justify-between items-center">
              <div className="flex space-x-2">
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handlePreview(template)}
                  title="Preview template"
                >
                  <EyeIcon className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDuplicate(template)}
                  title="Duplicate template"
                >
                  <DocumentDuplicateIcon className="h-4 w-4" />
                </Button>
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleEdit(template)}
                  title="Edit template"
                >
                  <PencilIcon className="h-4 w-4" />
                </Button>
              </div>
              <div className="flex space-x-2">
                {!template.is_default && (
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => handleSetDefault(template)}
                    title="Set as default"
                    className="text-yellow-600 hover:text-yellow-700"
                  >
                    <SparklesIcon className="h-4 w-4" />
                  </Button>
                )}
                <Button
                  variant="ghost"
                  size="sm"
                  onClick={() => handleDelete(template)}
                  title="Delete template"
                  className="text-red-600 hover:text-red-700"
                >
                  <TrashIcon className="h-4 w-4" />
                </Button>
              </div>
            </div>
          </Card>
        ))}
      </div>

      {filteredTemplates.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-500 text-lg mb-2">No templates found</div>
          <p className="text-gray-400">
            {searchTerm || filterCategory !== 'all'
              ? 'Try adjusting your filters'
              : 'Create your first template to get started'}
          </p>
        </div>
      )}

      {/* Create/Edit Form Modal */}
      <Modal
        isOpen={showForm}
        onClose={() => {
          setShowForm(false);
          setEditingTemplate(null);
          resetForm();
        }}
        title={editingTemplate ? 'Edit Page Template' : 'Create New Page Template'}
        size="xl"
      >
        <form onSubmit={handleSubmit} className="space-y-6">
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Template Name *
              </label>
              <Input
                id="name"
                type="text"
                value={formData.name}
                onChange={(e) => handleNameChange(e.target.value)}
                required
                placeholder="Enter template name"
              />
            </div>

            <div>
              <label htmlFor="slug" className="block text-sm font-medium text-gray-700 mb-2">
                Template Slug *
              </label>
              <Input
                id="slug"
                type="text"
                value={formData.slug}
                onChange={(e) => setFormData(prev => ({ ...prev, slug: e.target.value }))}
                required
                placeholder="template-slug"
              />
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
                {templateCategories.map(category => (
                  <option key={category} value={category}>
                    {category.charAt(0).toUpperCase() + category.slice(1)}
                  </option>
                ))}
              </Select>
            </div>

            <div>
              <label htmlFor="thumbnail" className="block text-sm font-medium text-gray-700 mb-2">
                Thumbnail URL
              </label>
              <Input
                id="thumbnail"
                type="url"
                value={formData.thumbnail}
                onChange={(e) => setFormData(prev => ({ ...prev, thumbnail: e.target.value }))}
                placeholder="https://example.com/thumbnail.jpg"
              />
            </div>
          </div>

          <div className="flex items-center">
            <label className="flex items-center">
              <input
                type="checkbox"
                checked={formData.is_default}
                onChange={(e) => setFormData(prev => ({ ...prev, is_default: e.target.checked }))}
                className="rounded border-gray-300 text-blue-600 shadow-sm focus:border-blue-300 focus:ring focus:ring-blue-200 focus:ring-opacity-50"
              />
              <span className="ml-2 text-sm text-gray-700">Set as Default Template</span>
            </label>
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-700 mb-2">
              Description
            </label>
            <Textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData(prev => ({ ...prev, description: e.target.value }))}
              placeholder="Brief description of this template"
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
                No variables defined. Variables allow you to make templates dynamic.
              </p>
            )}
          </div>

          {/* Content */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Template Content *
            </label>
            <RichTextEditor
              value={formData.content}
              onChange={(content) => setFormData(prev => ({ ...prev, content }))}
              placeholder="Enter your template content here. Use variables with ${variable_name} syntax."
            />
            <p className="mt-2 text-sm text-gray-500">
              Use <code className="bg-gray-100 px-1 rounded">$&#123;variable_name&#125;</code> to insert variables in your content.
            </p>
          </div>

          <div className="flex justify-end space-x-3">
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setShowForm(false);
                setEditingTemplate(null);
                resetForm();
              }}
            >
              Cancel
            </Button>
            <Button type="submit">
              {editingTemplate ? 'Update Template' : 'Create Template'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Preview Modal */}
      <Modal
        isOpen={showPreview}
        onClose={() => {
          setShowPreview(false);
          setPreviewTemplate(null);
        }}
        title={`Preview: ${previewTemplate?.name}`}
        size="xl"
      >
        {previewTemplate && (
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <Badge variant="secondary">{previewTemplate.category}</Badge>
              {previewTemplate.is_default && (
                <Badge variant="warning">Default</Badge>
              )}
              <span className="text-sm text-gray-500">{previewTemplate.slug}</span>
            </div>
            <div className="border rounded-lg p-6 bg-white">
              <div
                dangerouslySetInnerHTML={{ __html: previewTemplate.content }}
                className="prose prose-lg max-w-none"
              />
            </div>
          </div>
        )}
      </Modal>
    </div>
  );
}