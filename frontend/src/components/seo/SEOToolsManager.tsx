'use client'

import React, { useState, useEffect } from 'react';
import { useAuth } from '@/hooks/useAuth';
import { Modal } from '../ui/modal';
import { Button } from '../ui/button';
import { Input } from '../ui/input';
import { Textarea } from '../ui/textarea';
import { Card } from '../ui/card';
import { Badge } from '../ui/badge';
import { toast } from 'react-toastify';
import {
  MagnifyingGlassIcon,
  ChartBarIcon,
  EyeIcon,
  CheckCircleIcon,
  ExclamationTriangleIcon,
  XCircleIcon,
  LinkIcon,
  ClockIcon,
  GlobeAltIcon,
  DocumentTextIcon,
  PhotoIcon,
  TagIcon,
  BeakerIcon
} from '@heroicons/react/24/outline';

interface SEOAnalysis {
  page_id?: number;
  url: string;
  title: string;
  meta_description: string;
  h1_tags: string[];
  h2_tags: string[];
  image_alt_count: number;
  word_count: number;
  internal_links: number;
  external_links: number;
  mobile_friendly: boolean;
  loading_speed: number;
  ssl_enabled: boolean;
  schema_markup: any;
  score: number;
  issues: SEOIssue[];
  recommendations: string[];
  created_at: string;
}

interface SEOIssue {
  type: 'error' | 'warning' | 'info';
  category: string;
  description: string;
  element?: string;
}

interface SitemapEntry {
  url: string;
  lastmod: string;
  changefreq: string;
  priority: number;
}

interface MetaTag {
  id?: number;
  page_id?: number;
  name: string;
  content: string;
  property?: string;
}

export function SEOToolsManager() {
  const { user, token } = useAuth();
  const [activeTab, setActiveTab] = useState<'analysis' | 'sitemap' | 'meta' | 'schema'>('analysis');
  const [analyses, setAnalyses] = useState<SEOAnalysis[]>([]);
  const [sitemap, setSitemap] = useState<SitemapEntry[]>([]);
  const [metaTags, setMetaTags] = useState<MetaTag[]>([]);
  const [loading, setLoading] = useState(false);
  const [analyzing, setAnalyzing] = useState(false);
  const [showAnalysisForm, setShowAnalysisForm] = useState(false);
  const [showMetaForm, setShowMetaForm] = useState(false);
  const [analysisUrl, setAnalysisUrl] = useState('');
  const [selectedAnalysis, setSelectedAnalysis] = useState<SEOAnalysis | null>(null);
  const [metaFormData, setMetaFormData] = useState<MetaTag>({
    name: '',
    content: '',
    property: ''
  });

  useEffect(() => {
    fetchAnalyses();
    fetchSitemap();
    fetchMetaTags();
  }, []);

  const fetchAnalyses = async () => {
    try {
      const response = await fetch('/api/v1/admin/seo/analyses', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setAnalyses(data);
      }
    } catch (error) {
      console.error('Error fetching SEO analyses:', error);
    }
  };

  const fetchSitemap = async () => {
    try {
      const response = await fetch('/api/v1/admin/seo/sitemap', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setSitemap(data);
      }
    } catch (error) {
      console.error('Error fetching sitemap:', error);
    }
  };

  const fetchMetaTags = async () => {
    try {
      const response = await fetch('/api/v1/admin/seo/meta-tags', {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        const data = await response.json();
        setMetaTags(data);
      }
    } catch (error) {
      console.error('Error fetching meta tags:', error);
    }
  };

  const handleAnalyze = async (e: React.FormEvent) => {
    e.preventDefault();
    setAnalyzing(true);

    try {
      const response = await fetch('/api/v1/admin/seo/analyze', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify({ url: analysisUrl }),
      });

      if (response.ok) {
        const analysis = await response.json();
        toast.success('SEO analysis completed');
        setShowAnalysisForm(false);
        setAnalysisUrl('');
        fetchAnalyses();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to analyze URL');
      }
    } catch (error) {
      toast.error('Error analyzing URL');
    } finally {
      setAnalyzing(false);
    }
  };

  const handleCreateMetaTag = async (e: React.FormEvent) => {
    e.preventDefault();

    try {
      const response = await fetch('/api/v1/admin/seo/meta-tags', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${token}`,
        },
        body: JSON.stringify(metaFormData),
      });

      if (response.ok) {
        toast.success('Meta tag created successfully');
        setShowMetaForm(false);
        setMetaFormData({ name: '', content: '', property: '' });
        fetchMetaTags();
      } else {
        const error = await response.json();
        toast.error(error.error || 'Failed to create meta tag');
      }
    } catch (error) {
      toast.error('Error creating meta tag');
    }
  };

  const handleDeleteMetaTag = async (id: number) => {
    if (!confirm('Are you sure you want to delete this meta tag?')) {
      return;
    }

    try {
      const response = await fetch(`/api/v1/admin/seo/meta-tags/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        toast.success('Meta tag deleted successfully');
        fetchMetaTags();
      } else {
        toast.error('Failed to delete meta tag');
      }
    } catch (error) {
      toast.error('Error deleting meta tag');
    }
  };

  const generateSitemap = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/admin/seo/sitemap/generate', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      });

      if (response.ok) {
        toast.success('Sitemap generated successfully');
        fetchSitemap();
      } else {
        toast.error('Failed to generate sitemap');
      }
    } catch (error) {
      toast.error('Error generating sitemap');
    } finally {
      setLoading(false);
    }
  };

  const getScoreColor = (score: number) => {
    if (score >= 90) return 'text-green-600';
    if (score >= 70) return 'text-yellow-600';
    return 'text-red-600';
  };

  const getScoreBadge = (score: number) => {
    if (score >= 90) return 'success';
    if (score >= 70) return 'warning';
    return 'danger';
  };

  const getIssueIcon = (type: string) => {
    switch (type) {
      case 'error':
        return <XCircleIcon className="h-5 w-5 text-red-500" />;
      case 'warning':
        return <ExclamationTriangleIcon className="h-5 w-5 text-yellow-500" />;
      default:
        return <CheckCircleIcon className="h-5 w-5 text-blue-500" />;
    }
  };

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h2 className="text-2xl font-bold text-gray-900">SEO Tools</h2>
          <p className="text-gray-600">Analyze and optimize your site's SEO performance</p>
        </div>
        <div className="flex space-x-3">
          {activeTab === 'analysis' && (
            <Button
              onClick={() => setShowAnalysisForm(true)}
              className="flex items-center space-x-2"
            >
              <BeakerIcon className="h-5 w-5" />
              <span>Analyze URL</span>
            </Button>
          )}
          {activeTab === 'meta' && (
            <Button
              onClick={() => setShowMetaForm(true)}
              className="flex items-center space-x-2"
            >
              <TagIcon className="h-5 w-5" />
              <span>Add Meta Tag</span>
            </Button>
          )}
          {activeTab === 'sitemap' && (
            <Button
              onClick={generateSitemap}
              disabled={loading}
              className="flex items-center space-x-2"
            >
              <GlobeAltIcon className="h-5 w-5" />
              <span>{loading ? 'Generating...' : 'Generate Sitemap'}</span>
            </Button>
          )}
        </div>
      </div>

      {/* Tabs */}
      <div className="border-b border-gray-200">
        <nav className="-mb-px flex space-x-8" aria-label="Tabs">
          <button
            onClick={() => setActiveTab('analysis')}
            className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'analysis'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <ChartBarIcon className="h-5 w-5 inline-block mr-2" />
            SEO Analysis
          </button>
          <button
            onClick={() => setActiveTab('sitemap')}
            className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'sitemap'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <GlobeAltIcon className="h-5 w-5 inline-block mr-2" />
            Sitemap
          </button>
          <button
            onClick={() => setActiveTab('meta')}
            className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'meta'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <TagIcon className="h-5 w-5 inline-block mr-2" />
            Meta Tags
          </button>
          <button
            onClick={() => setActiveTab('schema')}
            className={`whitespace-nowrap py-2 px-1 border-b-2 font-medium text-sm ${
              activeTab === 'schema'
                ? 'border-blue-500 text-blue-600'
                : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
            }`}
          >
            <DocumentTextIcon className="h-5 w-5 inline-block mr-2" />
            Schema Markup
          </button>
        </nav>
      </div>

      {/* SEO Analysis Tab */}
      {activeTab === 'analysis' && (
        <div className="space-y-6">
          {analyses.length === 0 ? (
            <div className="text-center py-12">
              <ChartBarIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No SEO analyses yet</h3>
              <p className="text-gray-600 mb-6">
                Start by analyzing your website's SEO performance
              </p>
              <Button onClick={() => setShowAnalysisForm(true)}>
                Analyze Your First URL
              </Button>
            </div>
          ) : (
            <div className="grid grid-cols-1 gap-6">
              {analyses.map((analysis, index) => (
                <Card key={index} className="p-6">
                  <div className="flex items-start justify-between mb-4">
                    <div className="flex-1">
                      <h3 className="text-lg font-semibold text-gray-900 mb-1">
                        {analysis.title}
                      </h3>
                      <a
                        href={analysis.url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-blue-600 hover:text-blue-800 text-sm flex items-center"
                      >
                        <LinkIcon className="h-4 w-4 mr-1" />
                        {analysis.url}
                      </a>
                    </div>
                    <div className="flex items-center space-x-4">
                      <div className="text-right">
                        <div className={`text-2xl font-bold ${getScoreColor(analysis.score)}`}>
                          {analysis.score}
                        </div>
                        <Badge variant={getScoreBadge(analysis.score)}>
                          SEO Score
                        </Badge>
                      </div>
                      <Button
                        variant="ghost"
                        size="sm"
                        onClick={() => setSelectedAnalysis(analysis)}
                      >
                        <EyeIcon className="h-4 w-4 mr-1" />
                        View Details
                      </Button>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                    <div className="text-center">
                      <div className="text-2xl font-semibold text-gray-900">
                        {analysis.word_count}
                      </div>
                      <div className="text-sm text-gray-500">Words</div>
                    </div>
                    <div className="text-center">
                      <div className="text-2xl font-semibold text-gray-900">
                        {analysis.h1_tags.length}
                      </div>
                      <div className="text-sm text-gray-500">H1 Tags</div>
                    </div>
                    <div className="text-center">
                      <div className="text-2xl font-semibold text-gray-900">
                        {analysis.internal_links}
                      </div>
                      <div className="text-sm text-gray-500">Internal Links</div>
                    </div>
                    <div className="text-center">
                      <div className="text-2xl font-semibold text-gray-900">
                        {analysis.external_links}
                      </div>
                      <div className="text-sm text-gray-500">External Links</div>
                    </div>
                  </div>

                  {analysis.issues.length > 0 && (
                    <div className="border-t pt-4">
                      <h4 className="text-sm font-medium text-gray-900 mb-2">
                        Top Issues ({analysis.issues.length})
                      </h4>
                      <div className="space-y-2">
                        {analysis.issues.slice(0, 3).map((issue, issueIndex) => (
                          <div key={issueIndex} className="flex items-center space-x-2 text-sm">
                            {getIssueIcon(issue.type)}
                            <span className="text-gray-700">{issue.description}</span>
                          </div>
                        ))}
                        {analysis.issues.length > 3 && (
                          <div className="text-sm text-gray-500">
                            +{analysis.issues.length - 3} more issues
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  <div className="flex justify-between items-center mt-4 pt-4 border-t text-sm text-gray-500">
                    <div className="flex items-center">
                      <ClockIcon className="h-4 w-4 mr-1" />
                      {new Date(analysis.created_at).toLocaleDateString()}
                    </div>
                    <div className="flex items-center space-x-4">
                      {analysis.mobile_friendly && (
                        <span className="text-green-600">Mobile Friendly</span>
                      )}
                      {analysis.ssl_enabled && (
                        <span className="text-green-600">SSL Enabled</span>
                      )}
                    </div>
                  </div>
                </Card>
              ))}
            </div>
          )}
        </div>
      )}

      {/* Sitemap Tab */}
      {activeTab === 'sitemap' && (
        <div className="space-y-6">
          <div className="bg-white rounded-lg shadow overflow-hidden">
            <div className="px-6 py-4 border-b border-gray-200">
              <h3 className="text-lg font-medium text-gray-900">XML Sitemap</h3>
              <p className="text-sm text-gray-600">
                {sitemap.length} URL{sitemap.length !== 1 ? 's' : ''} in sitemap
              </p>
            </div>
            {sitemap.length > 0 ? (
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200">
                  <thead className="bg-gray-50">
                    <tr>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        URL
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Last Modified
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Change Frequency
                      </th>
                      <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                        Priority
                      </th>
                    </tr>
                  </thead>
                  <tbody className="bg-white divide-y divide-gray-200">
                    {sitemap.map((entry, index) => (
                      <tr key={index}>
                        <td className="px-6 py-4 whitespace-nowrap">
                          <a
                            href={entry.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            className="text-blue-600 hover:text-blue-800 text-sm"
                          >
                            {entry.url}
                          </a>
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {new Date(entry.lastmod).toLocaleDateString()}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {entry.changefreq}
                        </td>
                        <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900">
                          {entry.priority}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            ) : (
              <div className="text-center py-8">
                <p className="text-gray-500">No sitemap entries found</p>
              </div>
            )}
          </div>
        </div>
      )}

      {/* Meta Tags Tab */}
      {activeTab === 'meta' && (
        <div className="space-y-6">
          <div className="grid grid-cols-1 gap-4">
            {metaTags.map((tag) => (
              <Card key={tag.id} className="p-4">
                <div className="flex items-center justify-between">
                  <div>
                    <div className="font-medium text-gray-900">
                      {tag.property ? `property="${tag.property}"` : `name="${tag.name}"`}
                    </div>
                    <div className="text-sm text-gray-600 mt-1">
                      {tag.content}
                    </div>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => tag.id && handleDeleteMetaTag(tag.id)}
                    className="text-red-600 hover:text-red-700"
                  >
                    Delete
                  </Button>
                </div>
              </Card>
            ))}
          </div>

          {metaTags.length === 0 && (
            <div className="text-center py-12">
              <TagIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
              <h3 className="text-lg font-medium text-gray-900 mb-2">No meta tags configured</h3>
              <p className="text-gray-600 mb-6">
                Add meta tags to improve your site's SEO
              </p>
              <Button onClick={() => setShowMetaForm(true)}>
                Add Your First Meta Tag
              </Button>
            </div>
          )}
        </div>
      )}

      {/* Schema Markup Tab */}
      {activeTab === 'schema' && (
        <div className="space-y-6">
          <div className="text-center py-12">
            <DocumentTextIcon className="h-12 w-12 text-gray-400 mx-auto mb-4" />
            <h3 className="text-lg font-medium text-gray-900 mb-2">Schema Markup</h3>
            <p className="text-gray-600">
              Schema markup tools coming soon
            </p>
          </div>
        </div>
      )}

      {/* Analysis Form Modal */}
      <Modal
        isOpen={showAnalysisForm}
        onClose={() => setShowAnalysisForm(false)}
        title="Analyze URL for SEO"
        size="md"
      >
        <form onSubmit={handleAnalyze} className="space-y-4">
          <div>
            <label htmlFor="url" className="block text-sm font-medium text-gray-700 mb-2">
              URL to Analyze
            </label>
            <Input
              id="url"
              type="url"
              value={analysisUrl}
              onChange={(e) => setAnalysisUrl(e.target.value)}
              placeholder="https://example.com/page"
              required
            />
          </div>
          <div className="flex justify-end space-x-3">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowAnalysisForm(false)}
            >
              Cancel
            </Button>
            <Button type="submit" disabled={analyzing}>
              {analyzing ? 'Analyzing...' : 'Analyze'}
            </Button>
          </div>
        </form>
      </Modal>

      {/* Meta Tag Form Modal */}
      <Modal
        isOpen={showMetaForm}
        onClose={() => setShowMetaForm(false)}
        title="Add Meta Tag"
        size="md"
      >
        <form onSubmit={handleCreateMetaTag} className="space-y-4">
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-2">
                Name
              </label>
              <Input
                id="name"
                type="text"
                value={metaFormData.name}
                onChange={(e) => setMetaFormData(prev => ({ ...prev, name: e.target.value }))}
                placeholder="description"
              />
            </div>
            <div>
              <label htmlFor="property" className="block text-sm font-medium text-gray-700 mb-2">
                Property (Optional)
              </label>
              <Input
                id="property"
                type="text"
                value={metaFormData.property || ''}
                onChange={(e) => setMetaFormData(prev => ({ ...prev, property: e.target.value }))}
                placeholder="og:title"
              />
            </div>
          </div>
          <div>
            <label htmlFor="content" className="block text-sm font-medium text-gray-700 mb-2">
              Content *
            </label>
            <Textarea
              id="content"
              value={metaFormData.content}
              onChange={(e) => setMetaFormData(prev => ({ ...prev, content: e.target.value }))}
              placeholder="Meta tag content"
              required
              rows={3}
            />
          </div>
          <div className="flex justify-end space-x-3">
            <Button
              type="button"
              variant="ghost"
              onClick={() => setShowMetaForm(false)}
            >
              Cancel
            </Button>
            <Button type="submit">
              Add Meta Tag
            </Button>
          </div>
        </form>
      </Modal>

      {/* Analysis Details Modal */}
      <Modal
        isOpen={!!selectedAnalysis}
        onClose={() => setSelectedAnalysis(null)}
        title={`SEO Analysis: ${selectedAnalysis?.title}`}
        size="xl"
      >
        {selectedAnalysis && (
          <div className="space-y-6">
            {/* Score Overview */}
            <div className="flex items-center justify-between p-4 bg-gray-50 rounded-lg">
              <div>
                <h3 className="text-lg font-semibold text-gray-900">Overall SEO Score</h3>
                <p className="text-sm text-gray-600">{selectedAnalysis.url}</p>
              </div>
              <div className={`text-4xl font-bold ${getScoreColor(selectedAnalysis.score)}`}>
                {selectedAnalysis.score}
              </div>
            </div>

            {/* Metrics */}
            <div className="grid grid-cols-2 md:grid-cols-4 gap-4">
              <div className="text-center p-4 bg-white border rounded-lg">
                <div className="text-2xl font-semibold text-gray-900">
                  {selectedAnalysis.word_count}
                </div>
                <div className="text-sm text-gray-500">Total Words</div>
              </div>
              <div className="text-center p-4 bg-white border rounded-lg">
                <div className="text-2xl font-semibold text-gray-900">
                  {selectedAnalysis.h1_tags.length}
                </div>
                <div className="text-sm text-gray-500">H1 Tags</div>
              </div>
              <div className="text-center p-4 bg-white border rounded-lg">
                <div className="text-2xl font-semibold text-gray-900">
                  {selectedAnalysis.image_alt_count}
                </div>
                <div className="text-sm text-gray-500">Images with Alt</div>
              </div>
              <div className="text-center p-4 bg-white border rounded-lg">
                <div className="text-2xl font-semibold text-gray-900">
                  {selectedAnalysis.loading_speed.toFixed(1)}s
                </div>
                <div className="text-sm text-gray-500">Load Time</div>
              </div>
            </div>

            {/* Issues */}
            {selectedAnalysis.issues.length > 0 && (
              <div>
                <h4 className="text-lg font-semibold text-gray-900 mb-4">Issues & Warnings</h4>
                <div className="space-y-3">
                  {selectedAnalysis.issues.map((issue, index) => (
                    <div key={index} className="flex items-start space-x-3 p-3 border rounded-lg">
                      {getIssueIcon(issue.type)}
                      <div>
                        <div className="font-medium text-gray-900">{issue.category}</div>
                        <div className="text-sm text-gray-600">{issue.description}</div>
                        {issue.element && (
                          <div className="text-xs text-gray-500 mt-1 font-mono bg-gray-100 px-2 py-1 rounded">
                            {issue.element}
                          </div>
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Recommendations */}
            {selectedAnalysis.recommendations.length > 0 && (
              <div>
                <h4 className="text-lg font-semibold text-gray-900 mb-4">Recommendations</h4>
                <div className="space-y-2">
                  {selectedAnalysis.recommendations.map((rec, index) => (
                    <div key={index} className="flex items-start space-x-2">
                      <CheckCircleIcon className="h-5 w-5 text-green-500 mt-0.5" />
                      <span className="text-sm text-gray-700">{rec}</span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
}