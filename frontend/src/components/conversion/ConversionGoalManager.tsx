'use client';

import { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { 
  PlusIcon, 
  PencilIcon, 
  TrashIcon,
  ChartBarIcon,
  CurrencyDollarIcon,
  ClipboardDocumentListIcon,
  LinkIcon,
  BeakerIcon
} from '@heroicons/react/24/outline';

interface ConversionGoal {
  id: number;
  goal_name: string;
  goal_type: 'url_visit' | 'custom_event' | 'form_submit' | 'purchase';
  target_url?: string;
  custom_event_name?: string;
  goal_value: number;
  attribution_window: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface CreateGoalRequest {
  goal_name: string;
  goal_type: string;
  target_url?: string;
  custom_event_name?: string;
  goal_value?: number;
  attribution_window?: number;
}

const ConversionGoalManager: React.FC = () => {
  const { isAuthenticated } = useAuth();
  const [goals, setGoals] = useState<ConversionGoal[]>([]);
  const [loading, setLoading] = useState(true);
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingGoal, setEditingGoal] = useState<ConversionGoal | null>(null);
  const [formData, setFormData] = useState<CreateGoalRequest>({
    goal_name: '',
    goal_type: 'url_visit',
    attribution_window: 30,
    goal_value: 0,
  });

  useEffect(() => {
    if (isAuthenticated) {
      fetchGoals();
    }
  }, [isAuthenticated]);

  const fetchGoals = async () => {
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('http://localhost:8080/api/v1/conversions/goals', {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        const data = await response.json();
        setGoals(data.goals || []);
      }
    } catch (error) {
      console.error('Error fetching conversion goals:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateGoal = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch('http://localhost:8080/api/v1/conversions/goals', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        await fetchGoals();
        setShowCreateModal(false);
        resetForm();
      }
    } catch (error) {
      console.error('Error creating conversion goal:', error);
    }
  };

  const handleUpdateGoal = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!editingGoal) return;

    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`http://localhost:8080/api/v1/conversions/goals/${editingGoal.id}`, {
        method: 'PUT',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      if (response.ok) {
        await fetchGoals();
        setEditingGoal(null);
        resetForm();
      }
    } catch (error) {
      console.error('Error updating conversion goal:', error);
    }
  };

  const handleDeleteGoal = async (goalId: number) => {
    if (!confirm('Are you sure you want to delete this conversion goal?')) {
      return;
    }

    try {
      const token = localStorage.getItem('access_token');
      const response = await fetch(`http://localhost:8080/api/v1/conversions/goals/${goalId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        await fetchGoals();
      }
    } catch (error) {
      console.error('Error deleting conversion goal:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      goal_name: '',
      goal_type: 'url_visit',
      attribution_window: 30,
      goal_value: 0,
    });
    setShowCreateModal(false);
    setEditingGoal(null);
  };

  const openEditModal = (goal: ConversionGoal) => {
    setFormData({
      goal_name: goal.goal_name,
      goal_type: goal.goal_type,
      target_url: goal.target_url,
      custom_event_name: goal.custom_event_name,
      goal_value: goal.goal_value,
      attribution_window: goal.attribution_window,
    });
    setEditingGoal(goal);
  };

  const getGoalTypeIcon = (type: string) => {
    switch (type) {
      case 'url_visit':
        return <LinkIcon className="h-5 w-5" />;
      case 'form_submit':
        return <ClipboardDocumentListIcon className="h-5 w-5" />;
      case 'purchase':
        return <CurrencyDollarIcon className="h-5 w-5" />;
      case 'custom_event':
        return <BeakerIcon className="h-5 w-5" />;
      default:
        return <ChartBarIcon className="h-5 w-5" />;
    }
  };

  const getGoalTypeLabel = (type: string) => {
    switch (type) {
      case 'url_visit':
        return 'URL Visit';
      case 'form_submit':
        return 'Form Submit';
      case 'purchase':
        return 'Purchase';
      case 'custom_event':
        return 'Custom Event';
      default:
        return type;
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">Conversion Goals</h2>
          <p className="text-gray-600">Track and optimize your conversion funnel</p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
        >
          <PlusIcon className="h-5 w-5" />
          Create Goal
        </button>
      </div>

      {/* Goals Grid */}
      {goals.length === 0 ? (
        <div className="text-center py-12">
          <ChartBarIcon className="h-16 w-16 text-gray-400 mx-auto mb-4" />
          <h3 className="text-xl font-semibold text-gray-900 mb-2">No Conversion Goals</h3>
          <p className="text-gray-600 mb-6">
            Create your first conversion goal to start tracking user actions and optimizing your funnel.
          </p>
          <button
            onClick={() => setShowCreateModal(true)}
            className="px-6 py-3 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition-colors"
          >
            Create Your First Goal
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {goals.map((goal) => (
            <div key={goal.id} className="bg-white rounded-lg border border-gray-200 p-6 hover:shadow-md transition-shadow">
              {/* Goal Header */}
              <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                  <div className="p-2 bg-primary-100 rounded-lg">
                    {getGoalTypeIcon(goal.goal_type)}
                  </div>
                  <div>
                    <h3 className="font-semibold text-gray-900">{goal.goal_name}</h3>
                    <p className="text-sm text-gray-600">{getGoalTypeLabel(goal.goal_type)}</p>
                  </div>
                </div>
                <div className="flex items-center gap-2">
                  <button
                    onClick={() => openEditModal(goal)}
                    className="p-2 text-gray-400 hover:text-gray-600 transition-colors"
                  >
                    <PencilIcon className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => handleDeleteGoal(goal.id)}
                    className="p-2 text-gray-400 hover:text-red-600 transition-colors"
                  >
                    <TrashIcon className="h-4 w-4" />
                  </button>
                </div>
              </div>

              {/* Goal Details */}
              <div className="space-y-2 text-sm">
                {goal.target_url && (
                  <div>
                    <span className="font-medium">Target URL:</span>
                    <span className="ml-2 text-gray-600 truncate block">{goal.target_url}</span>
                  </div>
                )}
                {goal.custom_event_name && (
                  <div>
                    <span className="font-medium">Event Name:</span>
                    <span className="ml-2 text-gray-600">{goal.custom_event_name}</span>
                  </div>
                )}
                {goal.goal_value > 0 && (
                  <div>
                    <span className="font-medium">Value:</span>
                    <span className="ml-2 text-gray-600">${goal.goal_value}</span>
                  </div>
                )}
                <div>
                  <span className="font-medium">Attribution Window:</span>
                  <span className="ml-2 text-gray-600">{goal.attribution_window} days</span>
                </div>
              </div>

              {/* Status */}
              <div className="flex items-center justify-between mt-4 pt-4 border-t">
                <span className={`px-2 py-1 text-xs rounded-full ${
                  goal.is_active 
                    ? 'bg-green-100 text-green-800' 
                    : 'bg-gray-100 text-gray-800'
                }`}>
                  {goal.is_active ? 'Active' : 'Inactive'}
                </span>
                <button className="text-sm text-primary-600 hover:text-primary-700">
                  View Analytics
                </button>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* Create/Edit Modal */}
      {(showCreateModal || editingGoal) && (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 w-full max-w-md max-h-[90vh] overflow-y-auto">
            <h3 className="text-lg font-semibold mb-4">
              {editingGoal ? 'Edit Conversion Goal' : 'Create Conversion Goal'}
            </h3>

            <form onSubmit={editingGoal ? handleUpdateGoal : handleCreateGoal} className="space-y-4">
              {/* Goal Name */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Goal Name *
                </label>
                <input
                  type="text"
                  value={formData.goal_name}
                  onChange={(e) => setFormData({ ...formData, goal_name: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="e.g., Newsletter Signup"
                  required
                />
              </div>

              {/* Goal Type */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Goal Type *
                </label>
                <select
                  value={formData.goal_type}
                  onChange={(e) => setFormData({ ...formData, goal_type: e.target.value })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                >
                  <option value="url_visit">URL Visit</option>
                  <option value="form_submit">Form Submit</option>
                  <option value="purchase">Purchase</option>
                  <option value="custom_event">Custom Event</option>
                </select>
              </div>

              {/* Conditional Fields */}
              {formData.goal_type === 'url_visit' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Target URL
                  </label>
                  <input
                    type="url"
                    value={formData.target_url || ''}
                    onChange={(e) => setFormData({ ...formData, target_url: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    placeholder="https://example.com/thank-you"
                  />
                </div>
              )}

              {formData.goal_type === 'custom_event' && (
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-1">
                    Custom Event Name
                  </label>
                  <input
                    type="text"
                    value={formData.custom_event_name || ''}
                    onChange={(e) => setFormData({ ...formData, custom_event_name: e.target.value })}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                    placeholder="e.g., button_click_signup"
                  />
                </div>
              )}

              {/* Goal Value */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Goal Value ($)
                </label>
                <input
                  type="number"
                  step="0.01"
                  value={formData.goal_value || 0}
                  onChange={(e) => setFormData({ ...formData, goal_value: parseFloat(e.target.value) || 0 })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                  placeholder="0.00"
                />
              </div>

              {/* Attribution Window */}
              <div>
                <label className="block text-sm font-medium text-gray-700 mb-1">
                  Attribution Window (days)
                </label>
                <input
                  type="number"
                  min="1"
                  max="365"
                  value={formData.attribution_window || 30}
                  onChange={(e) => setFormData({ ...formData, attribution_window: parseInt(e.target.value) || 30 })}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-2 focus:ring-primary-500 focus:border-primary-500"
                />
                <p className="text-sm text-gray-500 mt-1">
                  How many days after a click can a conversion be attributed
                </p>
              </div>

              {/* Actions */}
              <div className="flex justify-end gap-3 pt-4">
                <button
                  type="button"
                  onClick={resetForm}
                  className="px-4 py-2 text-gray-700 border border-gray-300 rounded-md hover:bg-gray-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 bg-primary-600 text-white rounded-md hover:bg-primary-700 transition-colors"
                >
                  {editingGoal ? 'Update Goal' : 'Create Goal'}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default ConversionGoalManager;