'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import RichTextEditor from '@/components/editor/RichTextEditor';
import {
  CogIcon,
  GlobeAltIcon,
  PaintBrushIcon,
  EyeIcon,
  ShieldCheckIcon,
  ChartBarIcon,
  EnvelopeIcon,
  PlusIcon,
  TrashIcon,
  XMarkIcon,
  CheckIcon,
  ExclamationTriangleIcon,
} from '@heroicons/react/24/outline';

interface GlobalSetting {
  id: number;
  key: string;
  value: string;
  type: string;
  category: string;
  display_name: string;
  description?: string;
  is_public: boolean;
  sort_order: number;
}

interface SettingsByCategory {
  [category: string]: GlobalSetting[];
}

interface NewSettingData {
  key: string;
  value: string;
  type: string;
  category: string;
  display_name: string;
  description: string;
  is_public: boolean;
  sort_order: number;
}

export default function GlobalSettingsManager() {
  const { getAccessToken } = useAuth();
  const [settings, setSettings] = useState<SettingsByCategory>({});
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [activeCategory, setActiveCategory] = useState('general');
  const [editingSettings, setEditingSettings] = useState<{[key: string]: string}>({});
  const [showNewSettingForm, setShowNewSettingForm] = useState(false);
  const [newSetting, setNewSetting] = useState<NewSettingData>({
    key: '',
    value: '',
    type: 'text',
    category: 'general',
    display_name: '',
    description: '',
    is_public: false,
    sort_order: 0,
  });

  const categories = [
    { key: 'general', label: 'General', icon: CogIcon, description: 'Basic site settings' },
    { key: 'contact', label: 'Contact', icon: EnvelopeIcon, description: 'Contact information' },
    { key: 'social', label: 'Social', icon: GlobeAltIcon, description: 'Social media links' },
    { key: 'seo', label: 'SEO', icon: ChartBarIcon, description: 'Search engine optimization' },
    { key: 'tracking', label: 'Tracking', icon: EyeIcon, description: 'Analytics and tracking' },
    { key: 'footer', label: 'Footer', icon: PaintBrushIcon, description: 'Footer content' },
    { key: 'header', label: 'Header', icon: PaintBrushIcon, description: 'Header content' },
    { key: 'system', label: 'System', icon: ShieldCheckIcon, description: 'System settings' },
  ];

  const inputTypes = [
    { value: 'text', label: 'Text' },
    { value: 'textarea', label: 'Textarea' },
    { value: 'html', label: 'HTML' },
    { value: 'url', label: 'URL' },
    { value: 'email', label: 'Email' },
    { value: 'number', label: 'Number' },
    { value: 'boolean', label: 'Boolean' },
    { value: 'json', label: 'JSON' },
  ];

  useEffect(() => {
    fetchSettings();
  }, []);

  const fetchSettings = async () => {
    const token = getAccessToken();
    if (!token) return;

    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/api/v1/admin/settings', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSettings(data.settings || {});
      } else {
        console.error('Failed to fetch settings');
      }
    } catch (error) {
      console.error('Error fetching settings:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSettingChange = (key: string, value: string) => {
    setEditingSettings(prev => ({
      ...prev,
      [key]: value,
    }));
  };

  const saveSetting = async (key: string) => {
    const token = getAccessToken();
    if (!token) return;

    const value = editingSettings[key];
    if (value === undefined) return;

    try {
      setSaving(true);
      const response = await fetch(`http://localhost:8080/api/v1/admin/settings/${key}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ value }),
      });

      if (response.ok) {
        // Update local state
        setSettings(prev => {
          const updated = { ...prev };
          Object.keys(updated).forEach(category => {
            updated[category] = updated[category].map(setting =>
              setting.key === key ? { ...setting, value } : setting
            );
          });
          return updated;
        });

        // Remove from editing state
        setEditingSettings(prev => {
          const updated = { ...prev };
          delete updated[key];
          return updated;
        });
      } else {
        console.error('Failed to save setting');
      }
    } catch (error) {
      console.error('Error saving setting:', error);
    } finally {
      setSaving(false);
    }
  };

  const bulkSaveSettings = async () => {
    const token = getAccessToken();
    if (!token || Object.keys(editingSettings).length === 0) return;

    try {
      setSaving(true);
      const response = await fetch('http://localhost:8080/api/v1/admin/settings/bulk', {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(editingSettings),
      });

      if (response.ok) {
        // Update local state
        setSettings(prev => {
          const updated = { ...prev };
          Object.keys(updated).forEach(category => {
            updated[category] = updated[category].map(setting => {
              const newValue = editingSettings[setting.key];
              return newValue !== undefined ? { ...setting, value: newValue } : setting;
            });
          });
          return updated;
        });

        // Clear editing state
        setEditingSettings({});
      } else {
        console.error('Failed to bulk save settings');
      }
    } catch (error) {
      console.error('Error bulk saving settings:', error);
    } finally {
      setSaving(false);
    }
  };

  const createSetting = async () => {
    const token = getAccessToken();
    if (!token) return;

    try {
      setSaving(true);
      const response = await fetch('http://localhost:8080/api/v1/admin/settings', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(newSetting),
      });

      if (response.ok) {
        await fetchSettings(); // Refresh settings
        setShowNewSettingForm(false);
        setNewSetting({
          key: '',
          value: '',
          type: 'text',
          category: 'general',
          display_name: '',
          description: '',
          is_public: false,
          sort_order: 0,
        });
      } else {
        const error = await response.json();
        alert(error.error || 'Failed to create setting');
      }
    } catch (error) {
      console.error('Error creating setting:', error);
    } finally {
      setSaving(false);
    }
  };

  const deleteSetting = async (key: string) => {
    if (!confirm('Are you sure you want to delete this setting?')) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(`http://localhost:8080/api/v1/admin/settings/${key}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        await fetchSettings(); // Refresh settings
      } else {
        const error = await response.json();
        alert(error.error || 'Failed to delete setting');
      }
    } catch (error) {
      console.error('Error deleting setting:', error);
    }
  };

  const renderSettingInput = (setting: GlobalSetting) => {
    const value = editingSettings[setting.key] !== undefined 
      ? editingSettings[setting.key] 
      : setting.value;

    const onChange = (newValue: string) => handleSettingChange(setting.key, newValue);
    const isEditing = editingSettings[setting.key] !== undefined;

    switch (setting.type) {
      case 'textarea':
        return (
          <textarea
            value={value}
            onChange={(e) => onChange(e.target.value)}
            rows={4}
            className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 ${
              isEditing ? 'border-blue-300 bg-blue-50' : 'border-gray-300'
            }`}
          />
        );

      case 'html':
        return (
          <RichTextEditor
            content={value}
            onChange={onChange}
            height="200px"
            placeholder={`Enter ${setting.display_name.toLowerCase()}...`}
          />
        );

      case 'boolean':
        return (
          <select
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 ${
              isEditing ? 'border-blue-300 bg-blue-50' : 'border-gray-300'
            }`}
          >
            <option value="true">True</option>
            <option value="false">False</option>
          </select>
        );

      case 'number':
        return (
          <input
            type="number"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 ${
              isEditing ? 'border-blue-300 bg-blue-50' : 'border-gray-300'
            }`}
          />
        );

      default:
        return (
          <input
            type={setting.type === 'url' ? 'url' : setting.type === 'email' ? 'email' : 'text'}
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className={`mt-1 block w-full px-3 py-2 border rounded-md shadow-sm focus:outline-none focus:ring-blue-500 focus:border-blue-500 ${
              isEditing ? 'border-blue-300 bg-blue-50' : 'border-gray-300'
            }`}
          />
        );
    }
  };

  const hasUnsavedChanges = Object.keys(editingSettings).length > 0;

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-12 w-12 border-4 border-blue-500 border-t-transparent"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Global Site Settings</h2>
          <p className="text-gray-600">Configure site-wide settings and content</p>
        </div>
        <div className="flex items-center space-x-3">
          {hasUnsavedChanges && (
            <div className="flex items-center text-sm text-orange-600">
              <ExclamationTriangleIcon className="h-4 w-4 mr-1" />
              {Object.keys(editingSettings).length} unsaved changes
            </div>
          )}
          <button
            onClick={() => setShowNewSettingForm(true)}
            className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-lg flex items-center space-x-2"
          >
            <PlusIcon className="h-5 w-5" />
            <span>New Setting</span>
          </button>
          {hasUnsavedChanges && (
            <button
              onClick={bulkSaveSettings}
              disabled={saving}
              className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-lg flex items-center space-x-2"
            >
              <CheckIcon className="h-5 w-5" />
              <span>Save All</span>
            </button>
          )}
        </div>
      </div>

      <div className="flex space-x-6">
        {/* Categories Sidebar */}
        <div className="w-64 space-y-2">
          {categories.map((category) => {
            const Icon = category.icon;
            const categorySettings = settings[category.key] || [];
            
            return (
              <button
                key={category.key}
                onClick={() => setActiveCategory(category.key)}
                className={`w-full text-left p-3 rounded-lg transition-colors ${
                  activeCategory === category.key
                    ? 'bg-blue-50 text-blue-700 border border-blue-200'
                    : 'hover:bg-gray-50'
                }`}
              >
                <div className="flex items-center space-x-3">
                  <Icon className="h-5 w-5" />
                  <div>
                    <div className="font-medium">{category.label}</div>
                    <div className="text-sm text-gray-500">
                      {categorySettings.length} settings
                    </div>
                  </div>
                </div>
              </button>
            );
          })}
        </div>

        {/* Settings Content */}
        <div className="flex-1">
          <div className="bg-white border border-gray-200 rounded-lg">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">
                {categories.find(c => c.key === activeCategory)?.label} Settings
              </h3>
              <p className="text-gray-600 text-sm mt-1">
                {categories.find(c => c.key === activeCategory)?.description}
              </p>
            </div>

            <div className="p-6 space-y-6">
              {(settings[activeCategory] || []).map((setting) => {
                const isEditing = editingSettings[setting.key] !== undefined;
                
                return (
                  <div key={setting.key} className="space-y-2">
                    <div className="flex items-center justify-between">
                      <div>
                        <label className="block text-sm font-medium text-gray-700">
                          {setting.display_name}
                          {setting.is_public && (
                            <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-green-100 text-green-800">
                              Public
                            </span>
                          )}
                        </label>
                        {setting.description && (
                          <p className="text-xs text-gray-500 mt-1">{setting.description}</p>
                        )}
                      </div>
                      <div className="flex items-center space-x-2">
                        {isEditing && (
                          <button
                            onClick={() => saveSetting(setting.key)}
                            disabled={saving}
                            className="text-green-600 hover:text-green-800"
                            title="Save"
                          >
                            <CheckIcon className="h-4 w-4" />
                          </button>
                        )}
                        <button
                          onClick={() => deleteSetting(setting.key)}
                          className="text-red-600 hover:text-red-800"
                          title="Delete"
                        >
                          <TrashIcon className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                    
                    {renderSettingInput(setting)}
                  </div>
                );
              })}

              {(!settings[activeCategory] || settings[activeCategory].length === 0) && (
                <div className="text-center py-8 text-gray-500">
                  No settings found for this category
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* New Setting Modal */}
      {showNewSettingForm && (
        <div className="fixed inset-0 bg-gray-600 bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg shadow-xl max-w-md w-full m-4">
            <div className="p-6">
              <div className="flex items-center justify-between mb-4">
                <h3 className="text-lg font-medium text-gray-900">Create New Setting</h3>
                <button
                  onClick={() => setShowNewSettingForm(false)}
                  className="text-gray-400 hover:text-gray-600"
                >
                  <XMarkIcon className="h-5 w-5" />
                </button>
              </div>

              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700">Key *</label>
                  <input
                    type="text"
                    value={newSetting.key}
                    onChange={(e) => setNewSetting({ ...newSetting, key: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    placeholder="setting_key"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">Display Name *</label>
                  <input
                    type="text"
                    value={newSetting.display_name}
                    onChange={(e) => setNewSetting({ ...newSetting, display_name: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    placeholder="Setting Name"
                  />
                </div>

                <div className="grid grid-cols-2 gap-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700">Type</label>
                    <select
                      value={newSetting.type}
                      onChange={(e) => setNewSetting({ ...newSetting, type: e.target.value })}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      {inputTypes.map(type => (
                        <option key={type.value} value={type.value}>{type.label}</option>
                      ))}
                    </select>
                  </div>

                  <div>
                    <label className="block text-sm font-medium text-gray-700">Category</label>
                    <select
                      value={newSetting.category}
                      onChange={(e) => setNewSetting({ ...newSetting, category: e.target.value })}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    >
                      {categories.map(category => (
                        <option key={category.key} value={category.key}>{category.label}</option>
                      ))}
                    </select>
                  </div>
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">Description</label>
                  <textarea
                    value={newSetting.description}
                    onChange={(e) => setNewSetting({ ...newSetting, description: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    rows={2}
                    placeholder="Optional description"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700">Default Value</label>
                  <input
                    type="text"
                    value={newSetting.value}
                    onChange={(e) => setNewSetting({ ...newSetting, value: e.target.value })}
                    className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                  />
                </div>

                <div className="flex items-center space-x-4">
                  <label className="flex items-center">
                    <input
                      type="checkbox"
                      checked={newSetting.is_public}
                      onChange={(e) => setNewSetting({ ...newSetting, is_public: e.target.checked })}
                      className="rounded border-gray-300 text-blue-600"
                    />
                    <span className="ml-2 text-sm text-gray-700">Public setting</span>
                  </label>

                  <div className="flex-1">
                    <label className="block text-sm font-medium text-gray-700">Sort Order</label>
                    <input
                      type="number"
                      value={newSetting.sort_order}
                      onChange={(e) => setNewSetting({ ...newSetting, sort_order: parseInt(e.target.value) || 0 })}
                      className="mt-1 block w-full px-3 py-2 border border-gray-300 rounded-md"
                    />
                  </div>
                </div>
              </div>

              <div className="flex justify-end space-x-3 mt-6">
                <button
                  onClick={() => setShowNewSettingForm(false)}
                  className="px-4 py-2 text-sm font-medium text-gray-700 bg-white border border-gray-300 rounded-md hover:bg-gray-50"
                >
                  Cancel
                </button>
                <button
                  onClick={createSetting}
                  disabled={!newSetting.key || !newSetting.display_name || saving}
                  className="px-4 py-2 text-sm font-medium text-white bg-blue-600 border border-transparent rounded-md hover:bg-blue-700 disabled:opacity-50"
                >
                  {saving ? 'Creating...' : 'Create Setting'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}