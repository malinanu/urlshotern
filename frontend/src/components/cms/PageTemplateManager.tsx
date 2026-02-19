'use client'

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import RichTextEditor from '../editor/RichTextEditor';
import { Modal } from '../ui/modal';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
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
  DocumentTextIcon,
  MagnifyingGlassIcon,
} from '@heroicons/react/24/outline';

interface PageTemplate {
  id: number;
  name: string;
  slug: string;
  content: string;
  description?: string;
  fields?: string;
  preview_url?: string;
  is_active: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
}

interface PageTemplateFormData {
  name: string;
  slug: string;
  content: string;
  description: string;
  fields: string;
  preview_url: string;
  is_active: boolean;
  sort_order: number;
}

// Parse fields JSON string into key-value pairs for display
function parseFields(fields?: string): Record<string, string> {
  if (!fields) return {};
  try {
    const parsed = JSON.parse(fields);
    if (typeof parsed === 'object' && !Array.isArray(parsed)) return parsed;
  } catch {
    // ignore
  }
  return {};
}

export function PageTemplateManager() {
  const { user, getAccessToken } = useAuth();
  const [templates, setTemplates] = useState<PageTemplate[]>([]);
  const [filteredTemplates, setFilteredTemplates] = useState<PageTemplate[]>([]);
  const [loading, setLoading] = useState(true);
  const [showForm, setShowForm] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<PageTemplate | null>(null);
  const [showPreview, setShowPreview] = useState(false);
  const [previewTemplate, setPreviewTemplate] = useState<PageTemplate | null>(null);
  const [searchTerm, setSearchTerm] = useState('');
  // Local variable pairs for the fields editor
  const [variablePairs, setVariablePairs] = useState<Record<string, string>>({});
  const [formData, setFormData] = useState<PageTemplateFormData>({
    name: '',
    slug: '',
    content: '',
    description: '',
    fields: '',
    preview_url: '',
    is_active: true,
    sort_order: 0,
  });

  useEffect(() => {
    fetchTemplates();
  }, []);

  useEffect(() => {
    filterTemplates();
  }, [templates, searchTerm]);

  const fetchTemplates = async () => {
    try {
      const response = await fetch('/api/v1/admin/page-templates', {
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setTemplates(data.templates || []);
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
    if (!searchTerm) {
      setFilteredTemplates(templates);
      return;
    }
    const lower = searchTerm.toLowerCase();
    setFilteredTemplates(templates.filter(t =>
      t.name.toLowerCase().includes(lower) ||
      (t.description || '').toLowerCase().includes(lower) ||
      t.slug.toLowerCase().includes(lower)
    ));
  };

  const generateSlug = (name: string) =>
    name.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');

  const handleNameChange = (name: string) => {
    setFormData(prev => ({
      ...prev,
      name,
      slug: editingTemplate ? prev.slug : generateSlug(name),
    }));
  };

  const syncFieldsFromPairs = (pairs: Record<string, string>) => {
    setFormData(prev => ({
      ...prev,
      fields: Object.keys(pairs).length > 0 ? JSON.stringify(pairs) : '',
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const payload = {
      ...formData,
      description: formData.description || undefined,
      fields: formData.fields || undefined,
      preview_url: formData.preview_url || undefined,
    };

    try {
      const url = editingTemplate
        ? `/api/v1/admin/page-templates/${editingTemplate.id}`
        : '/api/v1/admin/page-templates';
      const method = editingTemplate ? 'PUT' : 'POST';

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${getAccessToken()}`,
        },
        body: JSON.stringify(payload),
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
    const pairs = parseFields(template.fields);
    setVariablePairs(pairs);
    setFormData({
      name: template.name,
      slug: template.slug,
      content: template.content,
      description: template.description || '',
      fields: template.fields || '',
      preview_url: template.preview_url || '',
      is_active: template.is_active,
      sort_order: template.sort_order,
    });
    setShowForm(true);
  };

  const handleDelete = async (template: PageTemplate) => {
    if (!confirm('Are you sure you want to delete this template?')) return;

    try {
      const response = await fetch(`/api/v1/admin/page-templates/${template.id}`, {
        method: 'DELETE',
        headers: { 'Authorization': `Bearer ${getAccessToken()}` },
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

  const handleDuplicate = (template: PageTemplate) => {
    const pairs = parseFields(template.fields);
    setVariablePairs(pairs);
    setFormData({
      name: `${template.name} (Copy)`,
      slug: generateSlug(`${template.name}-copy`),
      content: template.content,
      description: template.description || '',
      fields: template.fields || '',
      preview_url: template.preview_url || '',
      is_active: true,
      sort_order: 0,
    });
    setShowForm(true);
  };

  const handlePreview = async (template: PageTemplate) => {
    try {
      const response = await fetch(`/api/v1/admin/page-templates/${template.id}/render`, {
        headers: { 'Authorization': `Bearer ${getAccessToken()}` },
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

  const resetForm = () => {
    setVariablePairs({});
    setFormData({
      name: '',
      slug: '',
      content: '',
      description: '',
      fields: '',
      preview_url: '',
      is_active: true,
      sort_order: 0,
    });
  };

  const addVariable = () => {
    const key = prompt('Enter variable name:');
    if (key) {
      const newPairs = { ...variablePairs, [key]: '' };
      setVariablePairs(newPairs);
      syncFieldsFromPairs(newPairs);
    }
  };

  const updateVariable = (key: string, value: string) => {
    const newPairs = { ...variablePairs, [key]: value };
    setVariablePairs(newPairs);
    syncFieldsFromPairs(newPairs);
  };

  const removeVariable = (key: string) => {
    const newPairs = { ...variablePairs };
    delete newPairs[key];
    setVariablePairs(newPairs);
    syncFieldsFromPairs(newPairs);
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
        <Button onClick={() => setShowForm(true)} className="flex items-center space-x-2">
          <PlusIcon className="h-5 w-5" />
          <span>New Template</span>
        </Button>
      </div>

      {/* Search */}
      <Card className="p-4">
        <div className="flex items-center gap-4">
          <div className="relative flex-1">
            <MagnifyingGlassIcon className="h-5 w-5 absolute left-3 top-3 text-gray-400" />
            <Input
              type="text"
              placeholder="Search templates..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-10"
            />
          </div>
          <div className="text-sm text-gray-500">
            {filteredTemplates.length} template{filteredTemplates.length !== 1 ? 's' : ''}
          </div>
        </div>
      </Card>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {filteredTemplates.map((template) => (
          <Card key={template.id} className="p-6">
            {template.preview_url && (
              <div className="mb-4">
                <img
                  src={template.preview_url}
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
              <Badge variant={template.is_active ? 'default' : 'secondary'}>
                {template.is_active ? 'Active' : 'Inactive'}
              </Badge>
            </div>

            {template.description && (
              <p className="text-sm text-gray-600 mb-4 line-clamp-2">{template.description}</p>
            )}

            <div className="flex justify-between items-center">
              <div className="flex space-x-2">
                <Button variant="ghost" size="sm" onClick={() => handlePreview(template)} title="Preview">
                  <EyeIcon className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="sm" onClick={() => handleDuplicate(template)} title="Duplicate">
                  <DocumentDuplicateIcon className="h-4 w-4" />
                </Button>
                <Button variant="ghost" size="sm" onClick={() => handleEdit(template)} title="Edit">
                  <PencilIcon className="h-4 w-4" />
                </Button>
              </div>
              <Button
                variant="ghost"
                size="sm"
                onClick={() => handleDelete(template)}
                title="Delete"
                className="text-red-600 hover:text-red-700"
              >
                <TrashIcon className="h-4 w-4" />
              </Button>
            </div>
          </Card>
        ))}
      </div>

      {filteredTemplates.length === 0 && (
        <div className="text-center py-12">
          <div className="text-gray-500 text-lg mb-2">No templates found</div>
          <p className="text-gray-400">
            {searchTerm ? 'Try adjusting your search' : 'Create your first template to get started'}
          </p>
        </div>
      )}

      {/* Create/Edit Form Modal */}
      <Modal
        isOpen={showForm}
        onClose={() => { setShowForm(false); setEditingTemplate(null); resetForm(); }}
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
              <label htmlFor="preview_url" className="block text-sm font-medium text-gray-700 mb-2">
                Preview URL
              </label>
              <Input
                id="preview_url"
                type="url"
                value={formData.preview_url}
                onChange={(e) => setFormData(prev => ({ ...prev, preview_url: e.target.value }))}
                placeholder="https://example.com/preview.jpg"
              />
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
              <span className="ml-2 text-sm text-gray-700">Active</span>
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

          {/* Variables (stored as JSON in fields) */}
          <div>
            <div className="flex justify-between items-center mb-3">
              <label className="block text-sm font-medium text-gray-700">
                Template Variables
              </label>
              <Button type="button" variant="ghost" size="sm" onClick={addVariable}>
                <PlusIcon className="h-4 w-4 mr-1" />
                Add Variable
              </Button>
            </div>
            {Object.keys(variablePairs).length > 0 ? (
              <div className="space-y-3">
                {Object.entries(variablePairs).map(([key, value]) => (
                  <div key={key} className="flex items-center space-x-3">
                    <div className="flex-1">
                      <Input type="text" value={key} disabled className="font-mono text-sm" />
                    </div>
                    <div className="flex-1">
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
              content={formData.content}
              onChange={(content) => setFormData(prev => ({ ...prev, content }))}
              placeholder="Enter your template content here. Use {{variable_name}} syntax for variables."
            />
            <p className="mt-2 text-sm text-gray-500">
              Use <code className="bg-gray-100 px-1 rounded">&#123;&#123;variable_name&#125;&#125;</code> to insert variables.
            </p>
          </div>

          <div className="flex justify-end space-x-3">
            <Button
              type="button"
              variant="ghost"
              onClick={() => { setShowForm(false); setEditingTemplate(null); resetForm(); }}
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
        onClose={() => { setShowPreview(false); setPreviewTemplate(null); }}
        title={`Preview: ${previewTemplate?.name}`}
        size="xl"
      >
        {previewTemplate && (
          <div className="space-y-4">
            <div className="flex items-center space-x-4">
              <Badge variant={previewTemplate.is_active ? 'default' : 'secondary'}>
                {previewTemplate.is_active ? 'Active' : 'Inactive'}
              </Badge>
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
