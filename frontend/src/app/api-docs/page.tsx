'use client';

import { useState, useMemo, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import {
  CodeBracketIcon,
  DocumentTextIcon,
  ChevronRightIcon,
  ClipboardDocumentIcon,
  CheckIcon,
  MagnifyingGlassIcon,
  GlobeAltIcon,
  KeyIcon,
  BoltIcon,
  ServerIcon,
  LockClosedIcon,
  SparklesIcon
} from '@heroicons/react/24/outline';

interface APIEndpoint {
  method: string;
  path: string;
  description: string;
  auth?: string;
  parameters?: Parameter[];
  requestBody?: RequestBody;
  responses: Response[];
  example?: {
    request?: string;
    response: string;
  };
  category?: string;
}

interface Parameter {
  name: string;
  type: string;
  required: boolean;
  description: string;
}

interface RequestBody {
  contentType: string;
  schema: string;
}

interface Response {
  status: number;
  description: string;
  schema?: string;
}

const apiEndpoints: APIEndpoint[] = [
  // ============ URL Shortening ============
  {
    category: 'URL Shortening',
    method: 'POST',
    path: '/api/v1/shorten',
    description: 'Shorten a URL',
    auth: 'Optional',
    requestBody: {
      contentType: 'application/json',
      schema: `{
  "url": "string (required)",
  "custom_code": "string (optional)",
  "expires_at": "string (optional, ISO 8601 format)"
}`
    },
    responses: [
      {
        status: 201,
        description: 'URL shortened successfully',
        schema: `{
  "short_code": "string",
  "short_url": "string",
  "original_url": "string",
  "created_at": "string",
  "expires_at": "string | null"
}`
      },
      {
        status: 400,
        description: 'Invalid request'
      },
      {
        status: 409,
        description: 'Custom code already exists'
      }
    ],
    example: {
      request: `{
  "url": "https://example.com/very-long-url",
  "custom_code": "my-link"
}`,
      response: `{
  "short_code": "my-link",
  "short_url": "${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/my-link",
  "original_url": "https://example.com/very-long-url",
  "created_at": "2023-01-01T12:00:00Z"
}`
    }
  },
  {
    category: 'URL Shortening',
    method: 'GET',
    path: '/{shortCode}',
    description: 'Redirect to original URL',
    parameters: [
      {
        name: 'shortCode',
        type: 'string',
        required: true,
        description: 'The short code for the URL'
      }
    ],
    responses: [
      {
        status: 301,
        description: 'Redirects to original URL'
      },
      {
        status: 404,
        description: 'URL not found'
      },
      {
        status: 410,
        description: 'URL expired'
      }
    ]
  },
  {
    category: 'URL Shortening',
    method: 'POST',
    path: '/api/v1/batch-shorten',
    description: 'Shorten multiple URLs at once',
    auth: 'Required',
    requestBody: {
      contentType: 'application/json',
      schema: `[
  {
    "url": "string (required)",
    "custom_code": "string (optional)"
  }
]`
    },
    responses: [
      {
        status: 200,
        description: 'Batch processing completed',
        schema: `{
  "successful": "number",
  "failed": "number",
  "results": "Array of shortened URLs",
  "errors": "Array of errors"
}`
      },
      {
        status: 401,
        description: 'Authentication required'
      }
    ]
  },
  {
    category: 'URL Shortening',
    method: 'GET',
    path: '/api/v1/my-urls',
    description: 'Get user\'s URLs',
    auth: 'Required',
    parameters: [
      {
        name: 'page',
        type: 'number',
        required: false,
        description: 'Page number (default: 1)'
      },
      {
        name: 'limit',
        type: 'number',
        required: false,
        description: 'Items per page (default: 20, max: 100)'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'User URLs retrieved successfully',
        schema: `{
  "urls": "Array of URL objects",
  "total": "number",
  "page": "number",
  "limit": "number",
  "total_pages": "number"
}`
      },
      {
        status: 401,
        description: 'Authentication required'
      }
    ]
  },
  {
    category: 'URL Shortening',
    method: 'PUT',
    path: '/api/v1/my-urls/{shortCode}',
    description: 'Update a shortened URL',
    auth: 'Required',
    parameters: [
      {
        name: 'shortCode',
        type: 'string',
        required: true,
        description: 'The short code to update'
      }
    ],
    requestBody: {
      contentType: 'application/json',
      schema: `{
  "original_url": "string (optional)",
  "expires_at": "string (optional)"
}`
    },
    responses: [
      {
        status: 200,
        description: 'URL updated successfully'
      },
      {
        status: 404,
        description: 'URL not found'
      }
    ]
  },
  {
    category: 'URL Shortening',
    method: 'DELETE',
    path: '/api/v1/my-urls/{shortCode}',
    description: 'Delete a shortened URL',
    auth: 'Required',
    parameters: [
      {
        name: 'shortCode',
        type: 'string',
        required: true,
        description: 'The short code to delete'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'URL deleted successfully'
      },
      {
        status: 404,
        description: 'URL not found'
      }
    ]
  },

  // ============ System ============
  {
    category: 'System',
    method: 'GET',
    path: '/health',
    description: 'Health check endpoint',
    responses: [
      {
        status: 200,
        description: 'System is healthy',
        schema: `{
  "status": "healthy",
  "timestamp": "string"
}`
      }
    ]
  }
];

// Get API URL from environment with fallback to localhost (outside component to avoid hydration issues)
const getApiUrl = () => {
  if (typeof window !== 'undefined') {
    return window.location.origin.includes('localhost')
      ? 'http://localhost:8080'
      : process.env.NEXT_PUBLIC_API_URL || 'https://api.trunc.site';
  }
  return process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
};

export default function APIDocsPage() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(null);
  const [copiedCode, setCopiedCode] = useState<string | null>(null);
  const [selectedCategory, setSelectedCategory] = useState<string>('All');
  const [searchQuery, setSearchQuery] = useState<string>('');
  const [apiUrl, setApiUrl] = useState<string>('http://localhost:8080');
  const [selectedLanguage, setSelectedLanguage] = useState<string>('javascript');

  // Set API URL on client side only to avoid hydration mismatch
  useEffect(() => {
    setApiUrl(getApiUrl());
  }, []);

  // Get unique categories
  const categories = useMemo(() => {
    const cats = ['All', ...new Set(apiEndpoints.map(e => e.category || 'Other'))];
    return cats;
  }, []);

  // Filter endpoints by category and search
  const filteredEndpoints = useMemo(() => {
    let filtered = apiEndpoints;

    if (selectedCategory !== 'All') {
      filtered = filtered.filter(e => e.category === selectedCategory);
    }

    if (searchQuery) {
      filtered = filtered.filter(e =>
        e.path.toLowerCase().includes(searchQuery.toLowerCase()) ||
        e.description.toLowerCase().includes(searchQuery.toLowerCase()) ||
        e.method.toLowerCase().includes(searchQuery.toLowerCase())
      );
    }

    return filtered;
  }, [selectedCategory, searchQuery]);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedCode(id);
      setTimeout(() => setCopiedCode(null), 2000);
    });
  };

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET': return 'bg-green-500 text-white';
      case 'POST': return 'bg-blue-500 text-white';
      case 'PUT': return 'bg-yellow-500 text-white';
      case 'DELETE': return 'bg-red-500 text-white';
      default: return 'bg-gray-500 text-white';
    }
  };

  const getCategoryIcon = (category: string) => {
    switch (category) {
      case 'URL Shortening': return <BoltIcon className="h-4 w-4" />;
      case 'API Keys': return <KeyIcon className="h-4 w-4" />;
      case 'System': return <ServerIcon className="h-4 w-4" />;
      default: return <DocumentTextIcon className="h-4 w-4" />;
    }
  };

  return (
    <Layout>
      <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-50">
        {/* Hero Section with Quick Reference */}
        <div className="bg-gradient-to-br from-primary-600 via-primary-700 to-blue-700 text-white">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
            <div className="flex items-center gap-4 mb-6">
              <div className="p-3 bg-white/10 backdrop-blur-sm rounded-xl">
                <DocumentTextIcon className="h-10 w-10" />
              </div>
              <div>
                <h1 className="text-4xl font-bold mb-2">API Documentation</h1>
                <p className="text-primary-100 text-lg">
                  Complete reference for integrating Trunc into your applications
                </p>
              </div>
            </div>

            {/* Quick Reference Cards */}
            <div className="grid md:grid-cols-3 gap-4 mt-8">
              <div className="bg-white/10 backdrop-blur-sm rounded-xl p-6 border border-white/20 hover:bg-white/15 transition-all duration-300">
                <div className="flex items-center gap-3 mb-3">
                  <GlobeAltIcon className="h-6 w-6 text-primary-100" />
                  <h3 className="font-semibold text-lg">Base URL</h3>
                </div>
                <code className="block bg-black/20 backdrop-blur-sm px-4 py-3 rounded-lg text-sm font-mono text-white">
                  {apiUrl}
                </code>
                <p className="text-xs text-primary-100 mt-3">
                  Production: <code className="bg-black/20 px-2 py-1 rounded">https://api.trunc.site</code>
                </p>
              </div>

              <div className="bg-white/10 backdrop-blur-sm rounded-xl p-6 border border-white/20 hover:bg-white/15 transition-all duration-300">
                <div className="flex items-center gap-3 mb-3">
                  <LockClosedIcon className="h-6 w-6 text-primary-100" />
                  <h3 className="font-semibold text-lg">Authentication</h3>
                </div>
                <code className="block bg-black/20 backdrop-blur-sm px-4 py-3 rounded-lg text-sm font-mono text-white break-all">
                  Authorization: Bearer YOUR_KEY
                </code>
                <p className="text-xs text-primary-100 mt-3">
                  API keys start with <code className="bg-black/20 px-2 py-1 rounded">us_</code> prefix
                </p>
              </div>

              <div className="bg-white/10 backdrop-blur-sm rounded-xl p-6 border border-white/20 hover:bg-white/15 transition-all duration-300">
                <div className="flex items-center gap-3 mb-3">
                  <SparklesIcon className="h-6 w-6 text-primary-100" />
                  <h3 className="font-semibold text-lg">Rate Limits</h3>
                </div>
                <div className="space-y-2 text-sm">
                  <div className="flex justify-between items-center bg-black/20 backdrop-blur-sm px-4 py-2 rounded-lg">
                    <span>Standard</span>
                    <code className="font-mono">1000/hour</code>
                  </div>
                  <p className="text-xs text-primary-100 mt-2">
                    Configurable per API key
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Search and Filter */}
          <div className="mb-8 space-y-4">
            {/* Search Bar */}
            <div className="relative">
              <MagnifyingGlassIcon className="absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-gray-400" />
              <input
                type="text"
                placeholder="Search endpoints by name, path, or method..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full pl-12 pr-4 py-4 rounded-xl border-2 border-gray-200 focus:border-primary-500 focus:ring-4 focus:ring-primary-100 transition-all outline-none text-gray-900 placeholder-gray-400"
              />
            </div>

            {/* Category Filter */}
            <div className="flex flex-wrap gap-3">
              {categories.map((category) => (
                <button
                  key={category}
                  onClick={() => setSelectedCategory(category)}
                  className={`group px-5 py-2.5 rounded-xl text-sm font-medium transition-all duration-300 flex items-center gap-2 ${
                    selectedCategory === category
                      ? 'bg-primary-600 text-white shadow-lg shadow-primary-200 scale-105'
                      : 'bg-white text-gray-700 hover:bg-gray-50 border-2 border-gray-200 hover:border-primary-300'
                  }`}
                >
                  {category !== 'All' && getCategoryIcon(category)}
                  {category}
                  <span className={`px-2 py-0.5 rounded-full text-xs font-semibold ${
                    selectedCategory === category
                      ? 'bg-white/20 text-white'
                      : 'bg-gray-100 text-gray-600 group-hover:bg-primary-50 group-hover:text-primary-700'
                  }`}>
                    {apiEndpoints.filter(e => category === 'All' || e.category === category).length}
                  </span>
                </button>
              ))}
            </div>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Endpoint List */}
            <div className="lg:col-span-1">
              <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-6 sticky top-8 max-h-[calc(100vh-8rem)] overflow-y-auto">
                <div className="flex items-center justify-between mb-6">
                  <h2 className="text-xl font-bold text-gray-900 flex items-center gap-2">
                    <CodeBracketIcon className="h-6 w-6 text-primary-600" />
                    Endpoints
                  </h2>
                  <span className="px-3 py-1 bg-primary-100 text-primary-700 rounded-full text-sm font-semibold">
                    {filteredEndpoints.length}
                  </span>
                </div>

                <div className="space-y-2">
                  {filteredEndpoints.map((endpoint, index) => (
                    <button
                      key={index}
                      onClick={() => setSelectedEndpoint(endpoint)}
                      className={`w-full text-left p-4 rounded-xl transition-all duration-300 group ${
                        selectedEndpoint === endpoint
                          ? 'bg-gradient-to-r from-primary-50 to-blue-50 border-2 border-primary-300 shadow-md scale-[1.02]'
                          : 'hover:bg-gray-50 border-2 border-transparent hover:border-gray-200'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-2">
                        <span className={`px-3 py-1 text-xs font-bold rounded-lg ${getMethodColor(endpoint.method)} shadow-sm`}>
                          {endpoint.method}
                        </span>
                        {endpoint.auth && (
                          <span className="text-xs text-gray-500 font-medium flex items-center gap-1">
                            <LockClosedIcon className="h-3 w-3" />
                            {endpoint.auth}
                          </span>
                        )}
                      </div>
                      <div className="text-sm font-mono text-gray-700 mb-2 break-all font-semibold">
                        {endpoint.path}
                      </div>
                      <div className="text-xs text-gray-600 line-clamp-2">
                        {endpoint.description}
                      </div>
                    </button>
                  ))}

                  {filteredEndpoints.length === 0 && (
                    <div className="text-center py-12 text-gray-400">
                      <MagnifyingGlassIcon className="h-12 w-12 mx-auto mb-3 opacity-50" />
                      <p className="text-sm">No endpoints found</p>
                    </div>
                  )}
                </div>
              </div>
            </div>

            {/* Endpoint Details */}
            <div className="lg:col-span-2">
              {selectedEndpoint ? (
                <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-8">
                  {selectedEndpoint.category && (
                    <div className="mb-4">
                      <span className="inline-flex items-center gap-2 text-sm font-semibold text-primary-700 bg-primary-50 px-4 py-2 rounded-lg border border-primary-200">
                        {getCategoryIcon(selectedEndpoint.category)}
                        {selectedEndpoint.category}
                      </span>
                    </div>
                  )}

                  <div className="flex items-center gap-4 mb-6 pb-6 border-b-2 border-gray-100">
                    <span className={`px-4 py-2 text-sm font-bold rounded-lg ${getMethodColor(selectedEndpoint.method)} shadow-md`}>
                      {selectedEndpoint.method}
                    </span>
                    <code className="text-xl font-mono text-gray-900 break-all font-bold">{selectedEndpoint.path}</code>
                  </div>

                  <p className="text-gray-700 mb-8 text-lg leading-relaxed">{selectedEndpoint.description}</p>

                  {/* Authentication */}
                  {selectedEndpoint.auth && (
                    <div className="mb-8 bg-gradient-to-r from-amber-50 to-orange-50 border-l-4 border-amber-400 rounded-r-xl p-6">
                      <h3 className="text-lg font-bold text-gray-900 mb-3 flex items-center gap-2">
                        <LockClosedIcon className="h-5 w-5 text-amber-600" />
                        Authentication
                      </h3>
                      <p className="text-sm text-gray-700">
                        {selectedEndpoint.auth === 'Required'
                          ? '🔒 This endpoint requires authentication. Include the Bearer token in the Authorization header.'
                          : '🔓 Authentication is optional for this endpoint. Include the Bearer token for enhanced features.'
                        }
                      </p>
                    </div>
                  )}

                  {/* Parameters */}
                  {selectedEndpoint.parameters && selectedEndpoint.parameters.length > 0 && (
                    <div className="mb-8">
                      <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <ChevronRightIcon className="h-5 w-5 text-primary-600" />
                        Parameters
                      </h3>
                      <div className="overflow-hidden rounded-xl border-2 border-gray-200">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="bg-gradient-to-r from-gray-50 to-gray-100 border-b-2 border-gray-200">
                              <th className="text-left py-4 px-4 font-bold text-gray-700">Name</th>
                              <th className="text-left py-4 px-4 font-bold text-gray-700">Type</th>
                              <th className="text-left py-4 px-4 font-bold text-gray-700">Required</th>
                              <th className="text-left py-4 px-4 font-bold text-gray-700">Description</th>
                            </tr>
                          </thead>
                          <tbody>
                            {selectedEndpoint.parameters.map((param, index) => (
                              <tr key={index} className={`border-b border-gray-100 ${index % 2 === 0 ? 'bg-white' : 'bg-gray-50'} hover:bg-primary-50 transition-colors`}>
                                <td className="py-4 px-4 font-mono font-semibold text-primary-700">{param.name}</td>
                                <td className="py-4 px-4 text-gray-600">{param.type}</td>
                                <td className="py-4 px-4">
                                  <span className={`px-3 py-1 text-xs font-bold rounded-full ${
                                    param.required ? 'bg-red-100 text-red-700 border border-red-200' : 'bg-gray-100 text-gray-600 border border-gray-200'
                                  }`}>
                                    {param.required ? 'Yes' : 'No'}
                                  </span>
                                </td>
                                <td className="py-4 px-4 text-gray-700">{param.description}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Request Body */}
                  {selectedEndpoint.requestBody && (
                    <div className="mb-8">
                      <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <ChevronRightIcon className="h-5 w-5 text-primary-600" />
                        Request Body
                      </h3>
                      <p className="text-sm text-gray-600 mb-3 font-medium">
                        Content-Type: <code className="bg-gray-100 px-2 py-1 rounded text-primary-700">{selectedEndpoint.requestBody.contentType}</code>
                      </p>
                      <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                        <div className="absolute top-4 right-4 flex items-center gap-2">
                          <span className="text-xs text-gray-400 font-mono">JSON</span>
                          <button
                            onClick={() => copyToClipboard(selectedEndpoint.requestBody!.schema, 'request-schema')}
                            className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                          >
                            {copiedCode === 'request-schema' ? (
                              <CheckIcon className="h-5 w-5 text-green-400" />
                            ) : (
                              <ClipboardDocumentIcon className="h-5 w-5" />
                            )}
                          </button>
                        </div>
                        <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8">
                          <code>{selectedEndpoint.requestBody.schema}</code>
                        </pre>
                      </div>
                    </div>
                  )}

                  {/* Responses */}
                  <div className="mb-8">
                    <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                      <ChevronRightIcon className="h-5 w-5 text-primary-600" />
                      Responses
                    </h3>
                    <div className="space-y-4">
                      {selectedEndpoint.responses.map((response, index) => (
                        <div key={index} className="border-2 border-gray-200 rounded-xl p-6 hover:border-primary-300 transition-all hover:shadow-lg">
                          <div className="flex items-center gap-3 mb-4">
                            <span className={`px-4 py-2 text-sm font-bold rounded-lg shadow-md ${
                              response.status < 300 ? 'bg-green-500 text-white' :
                              response.status < 400 ? 'bg-blue-500 text-white' :
                              response.status < 500 ? 'bg-yellow-500 text-white' :
                              'bg-red-500 text-white'
                            }`}>
                              {response.status}
                            </span>
                            <span className="text-sm text-gray-700 font-medium">{response.description}</span>
                          </div>
                          {response.schema && (
                            <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-lg p-4 relative shadow-xl border border-gray-700">
                              <button
                                onClick={() => copyToClipboard(response.schema!, `response-${index}`)}
                                className="absolute top-3 right-3 p-1.5 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                              >
                                {copiedCode === `response-${index}` ? (
                                  <CheckIcon className="h-4 w-4 text-green-400" />
                                ) : (
                                  <ClipboardDocumentIcon className="h-4 w-4" />
                                )}
                              </button>
                              <pre className="text-green-400 text-xs overflow-x-auto font-mono leading-relaxed pr-8">
                                <code>{response.schema}</code>
                              </pre>
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Example */}
                  {selectedEndpoint.example && (
                    <div className="mb-6">
                      <h3 className="text-lg font-bold text-gray-900 mb-4 flex items-center gap-2">
                        <SparklesIcon className="h-5 w-5 text-primary-600" />
                        Example
                      </h3>

                      {selectedEndpoint.example.request && (
                        <div className="mb-6">
                          <h4 className="text-sm font-bold text-gray-700 mb-3 flex items-center gap-2">
                            📤 Request:
                          </h4>
                          <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                            <div className="absolute top-4 right-4 flex items-center gap-2">
                              <span className="text-xs text-gray-400 font-mono">JSON</span>
                              <button
                                onClick={() => copyToClipboard(selectedEndpoint.example!.request!, 'example-request')}
                                className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                              >
                                {copiedCode === 'example-request' ? (
                                  <CheckIcon className="h-5 w-5 text-green-400" />
                                ) : (
                                  <ClipboardDocumentIcon className="h-5 w-5" />
                                )}
                              </button>
                            </div>
                            <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8">
                              <code>{selectedEndpoint.example.request}</code>
                            </pre>
                          </div>
                        </div>
                      )}

                      <div>
                        <h4 className="text-sm font-bold text-gray-700 mb-3 flex items-center gap-2">
                          📥 Response:
                        </h4>
                        <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                          <div className="absolute top-4 right-4 flex items-center gap-2">
                            <span className="text-xs text-gray-400 font-mono">JSON</span>
                            <button
                              onClick={() => copyToClipboard(selectedEndpoint.example.response, 'example-response')}
                              className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                            >
                              {copiedCode === 'example-response' ? (
                                <CheckIcon className="h-5 w-5 text-green-400" />
                              ) : (
                                <ClipboardDocumentIcon className="h-5 w-5" />
                              )}
                            </button>
                          </div>
                          <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8">
                            <code>{selectedEndpoint.example.response}</code>
                          </pre>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="bg-white rounded-2xl shadow-xl border border-gray-100 p-16 text-center">
                  <div className="max-w-md mx-auto">
                    <div className="w-20 h-20 bg-gradient-to-br from-primary-100 to-blue-100 rounded-2xl flex items-center justify-center mx-auto mb-6">
                      <CodeBracketIcon className="h-10 w-10 text-primary-600" />
                    </div>
                    <h3 className="text-2xl font-bold text-gray-900 mb-3">
                      Select an Endpoint
                    </h3>
                    <p className="text-gray-600 text-lg">
                      Choose an endpoint from the list to view detailed documentation and examples.
                    </p>
                  </div>
                </div>
              )}
            </div>
          </div>

          {/* Integration Guide Section */}
          <div className="mt-16 bg-gradient-to-br from-primary-50 via-blue-50 to-purple-50 rounded-3xl shadow-2xl border-2 border-primary-200 p-10 relative overflow-hidden">
            {/* Decorative Elements */}
            <div className="absolute top-0 right-0 w-64 h-64 bg-gradient-to-br from-primary-200/20 to-blue-200/20 rounded-full blur-3xl"></div>
            <div className="absolute bottom-0 left-0 w-64 h-64 bg-gradient-to-tr from-purple-200/20 to-pink-200/20 rounded-full blur-3xl"></div>

            <div className="relative z-10">
              <div className="flex items-center gap-4 mb-6">
                <div className="p-4 bg-gradient-to-br from-primary-600 to-blue-600 rounded-2xl shadow-lg">
                  <CodeBracketIcon className="h-10 w-10 text-white" />
                </div>
                <div>
                  <h2 className="text-4xl font-bold text-gray-900">Integration Guide</h2>
                  <p className="text-gray-700 text-lg mt-1">
                    Learn how to integrate Trunc into your applications
                  </p>
                </div>
              </div>

              {/* Prerequisites */}
              <div className="bg-gradient-to-br from-amber-50 to-orange-50 border-2 border-amber-300 rounded-2xl p-8 mb-8 shadow-lg">
                <h3 className="text-2xl font-bold text-gray-900 mb-5 flex items-center gap-2">
                  <span className="text-3xl">⚠️</span> Prerequisites
                </h3>
                <p className="text-sm text-gray-700 mb-6 leading-relaxed">
                  Before you can use the API, you need a Trunc account and an API key. <strong>You cannot register or login via the API.</strong> Account creation and API key generation must be done through the web interface.
                </p>
                <div className="space-y-5">
                  <div className="flex gap-4 items-start group hover:translate-x-1 transition-transform">
                    <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg">
                      1
                    </div>
                    <div>
                      <h4 className="font-bold text-gray-900 text-lg mb-1">Register on the Website</h4>
                      <p className="text-sm text-gray-700">Visit <a href="/" className="text-primary-600 hover:text-primary-700 underline font-semibold">trunc.site</a> and create an account if you haven't already.</p>
                    </div>
                  </div>
                  <div className="flex gap-4 items-start group hover:translate-x-1 transition-transform">
                    <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg">
                      2
                    </div>
                    <div>
                      <h4 className="font-bold text-gray-900 text-lg mb-1">Login to Dashboard</h4>
                      <p className="text-sm text-gray-700">Sign in to your account and access the dashboard.</p>
                    </div>
                  </div>
                  <div className="flex gap-4 items-start group hover:translate-x-1 transition-transform">
                    <div className="flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg">
                      3
                    </div>
                    <div>
                      <h4 className="font-bold text-gray-900 text-lg mb-1">Generate API Key</h4>
                      <p className="text-sm text-gray-700">Navigate to Settings → API Keys and create a new API key with the required permissions.</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Quick Start */}
              <div className="bg-white rounded-2xl p-8 mb-8 shadow-xl border-2 border-gray-100">
                <h3 className="text-2xl font-bold text-gray-900 mb-5 flex items-center gap-2">
                  <span className="text-3xl">🚀</span> Quick Start
                </h3>
                <p className="text-sm text-gray-600 mb-6 leading-relaxed">
                  Once you have your API key from the prerequisites above, follow these steps:
                </p>
                <div className="space-y-5">
                  <div className="flex gap-5 items-start group">
                    <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-green-500 to-emerald-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg">
                      1
                    </div>
                    <div className="flex-1">
                      <h4 className="font-bold text-gray-900 mb-2 text-lg">Copy Your API Key</h4>
                      <p className="text-sm text-gray-600">After creating an API key in the dashboard, copy it immediately. You won't be able to see it again!</p>
                    </div>
                  </div>
                  <div className="flex gap-5 items-start group">
                    <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg">
                      2
                    </div>
                    <div className="flex-1">
                      <h4 className="font-bold text-gray-900 mb-2 text-lg">Store Securely</h4>
                      <p className="text-sm text-gray-600">Save your API key in a secure location (environment variables, secrets manager, etc.). Never commit it to version control!</p>
                    </div>
                  </div>
                  <div className="flex gap-5 items-start group">
                    <div className="flex-shrink-0 w-12 h-12 bg-gradient-to-br from-purple-500 to-pink-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg">
                      3
                    </div>
                    <div className="flex-1">
                      <h4 className="font-bold text-gray-900 mb-2 text-lg">Make Your First Request</h4>
                      <p className="text-sm text-gray-600">Use the code examples below to start shortening URLs programmatically!</p>
                    </div>
                  </div>
                </div>
              </div>

              {/* Language Selector */}
              <div className="mb-8">
                <div className="flex items-center gap-3 mb-6">
                  <h3 className="text-2xl font-bold text-gray-900">Code Examples</h3>
                  <div className="flex-1"></div>
                  <div className="flex gap-2">
                    {['javascript', 'react', 'nodejs', 'python'].map((lang) => (
                      <button
                        key={lang}
                        onClick={() => setSelectedLanguage(lang)}
                        className={`px-5 py-2.5 rounded-xl font-semibold text-sm transition-all duration-300 ${
                          selectedLanguage === lang
                            ? 'bg-gradient-to-r from-primary-600 to-blue-600 text-white shadow-lg scale-105'
                            : 'bg-white text-gray-700 border-2 border-gray-200 hover:border-primary-300 hover:text-primary-700'
                        }`}
                      >
                        {lang === 'javascript' ? 'JavaScript' :
                         lang === 'react' ? 'React' :
                         lang === 'nodejs' ? 'Node.js' :
                         'Python'}
                      </button>
                    ))}
                  </div>
                </div>

                {/* JavaScript Example */}
                {selectedLanguage === 'javascript' && (
                  <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
                    <h4 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                      <span className="text-2xl">📜</span> Vanilla JavaScript
                    </h4>
                    <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                      <div className="absolute top-4 right-4 flex items-center gap-3">
                        <span className="text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg">JavaScript</span>
                        <button
                          onClick={() => copyToClipboard(`async function shortenURL(originalUrl) {
  const response = await fetch('${apiUrl}/api/v1/shorten', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_API_KEY_HERE'
    },
    body: JSON.stringify({ url: originalUrl })
  });

  if (!response.ok) {
    throw new Error('Failed to shorten URL');
  }

  const data = await response.json();
  return data.short_url;
}

// Usage
shortenURL('https://example.com/very-long-url')
  .then(shortUrl => console.log('Shortened URL:', shortUrl))
  .catch(error => console.error('Error:', error));`, 'js-example')}
                          className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        >
                          {copiedCode === 'js-example' ? (
                            <CheckIcon className="h-5 w-5 text-green-400" />
                          ) : (
                            <ClipboardDocumentIcon className="h-5 w-5" />
                          )}
                        </button>
                      </div>
                      <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10">
                        <code>{`async function shortenURL(originalUrl) {
  const response = await fetch('${apiUrl}/api/v1/shorten', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_API_KEY_HERE'
    },
    body: JSON.stringify({ url: originalUrl })
  });

  if (!response.ok) {
    throw new Error('Failed to shorten URL');
  }

  const data = await response.json();
  return data.short_url;
}

// Usage
shortenURL('https://example.com/very-long-url')
  .then(shortUrl => console.log('Shortened URL:', shortUrl))
  .catch(error => console.error('Error:', error));`}</code>
                      </pre>
                    </div>
                  </div>
                )}

                {/* React Example */}
                {selectedLanguage === 'react' && (
                  <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
                    <h4 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                      <span className="text-2xl">⚛️</span> React Custom Hook
                    </h4>
                    <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                      <div className="absolute top-4 right-4 flex items-center gap-3">
                        <span className="text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg">React</span>
                        <button
                          onClick={() => copyToClipboard(`import { useState } from 'react';

export function useURLShortener(apiKey) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const shortenURL = async (originalUrl) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('${apiUrl}/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        },
        body: JSON.stringify({ url: originalUrl })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URL');
      }

      const data = await response.json();
      return data.short_url;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { shortenURL, loading, error };
}`, 'react-hook')}
                          className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        >
                          {copiedCode === 'react-hook' ? (
                            <CheckIcon className="h-5 w-5 text-green-400" />
                          ) : (
                            <ClipboardDocumentIcon className="h-5 w-5" />
                          )}
                        </button>
                      </div>
                      <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10">
                        <code>{`import { useState } from 'react';

export function useURLShortener(apiKey) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const shortenURL = async (originalUrl) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('${apiUrl}/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        },
        body: JSON.stringify({ url: originalUrl })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URL');
      }

      const data = await response.json();
      return data.short_url;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { shortenURL, loading, error };
}`}</code>
                      </pre>
                    </div>
                  </div>
                )}

                {/* Node.js Example */}
                {selectedLanguage === 'nodejs' && (
                  <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
                    <h4 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                      <span className="text-2xl">🟢</span> Node.js Backend
                    </h4>
                    <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                      <div className="absolute top-4 right-4 flex items-center gap-3">
                        <span className="text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg">Node.js</span>
                        <button
                          onClick={() => copyToClipboard(`const axios = require('axios');

async function shortenURL(originalUrl, apiKey) {
  try {
    const response = await axios.post('${apiUrl}/api/v1/shorten',
      { url: originalUrl },
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        }
      }
    );

    return response.data.short_url;
  } catch (error) {
    console.error('Error shortening URL:', error.response?.data || error.message);
    throw error;
  }
}

// Express.js endpoint example:
app.post('/api/shorten', async (req, res) => {
  const { url } = req.body;
  const apiKey = process.env.TRUNC_API_KEY;

  try {
    const shortUrl = await shortenURL(url, apiKey);
    res.json({ success: true, shortUrl });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: 'Failed to shorten URL'
    });
  }
});`, 'node-example')}
                          className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        >
                          {copiedCode === 'node-example' ? (
                            <CheckIcon className="h-5 w-5 text-green-400" />
                          ) : (
                            <ClipboardDocumentIcon className="h-5 w-5" />
                          )}
                        </button>
                      </div>
                      <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10">
                        <code>{`const axios = require('axios');

async function shortenURL(originalUrl, apiKey) {
  try {
    const response = await axios.post('${apiUrl}/api/v1/shorten',
      { url: originalUrl },
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        }
      }
    );

    return response.data.short_url;
  } catch (error) {
    console.error('Error shortening URL:', error.response?.data || error.message);
    throw error;
  }
}

// Express.js endpoint example:
app.post('/api/shorten', async (req, res) => {
  const { url } = req.body;
  const apiKey = process.env.TRUNC_API_KEY;

  try {
    const shortUrl = await shortenURL(url, apiKey);
    res.json({ success: true, shortUrl });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: 'Failed to shorten URL'
    });
  }
});`}</code>
                      </pre>
                    </div>
                  </div>
                )}

                {/* Python Example */}
                {selectedLanguage === 'python' && (
                  <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
                    <h4 className="text-xl font-bold text-gray-900 mb-4 flex items-center gap-2">
                      <span className="text-2xl">🐍</span> Python Integration
                    </h4>
                    <div className="bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700">
                      <div className="absolute top-4 right-4 flex items-center gap-3">
                        <span className="text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg">Python</span>
                        <button
                          onClick={() => copyToClipboard(`import requests
import os

def shorten_url(original_url, api_key):
    """Shorten a URL using the Trunc API"""
    response = requests.post(
        '${apiUrl}/api/v1/shorten',
        json={'url': original_url},
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}'
        }
    )

    if response.status_code == 201:
        return response.json()['short_url']
    else:
        raise Exception(f"Failed to shorten URL: {response.text}")

# Usage
api_key = os.getenv('TRUNC_API_KEY')
short_url = shorten_url('https://example.com/long-url', api_key)
print(f"Shortened URL: {short_url}")`, 'python-example')}
                          className="p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all"
                        >
                          {copiedCode === 'python-example' ? (
                            <CheckIcon className="h-5 w-5 text-green-400" />
                          ) : (
                            <ClipboardDocumentIcon className="h-5 w-5" />
                          )}
                        </button>
                      </div>
                      <pre className="text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10">
                        <code>{`import requests
import os

def shorten_url(original_url, api_key):
    """Shorten a URL using the Trunc API"""
    response = requests.post(
        '${apiUrl}/api/v1/shorten',
        json={'url': original_url},
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}'
        }
    )

    if response.status_code == 201:
        return response.json()['short_url']
    else:
        raise Exception(f"Failed to shorten URL: {response.text}")

# Usage
api_key = os.getenv('TRUNC_API_KEY')
short_url = shorten_url('https://example.com/long-url', api_key)
print(f"Shortened URL: {short_url}")`}</code>
                      </pre>
                    </div>
                  </div>
                )}
              </div>

              {/* Security Best Practices */}
              <div className="bg-white rounded-2xl p-8 mb-8 shadow-xl border-2 border-gray-100">
                <h3 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
                  <span className="text-3xl">🔐</span> Security Best Practices
                </h3>
                <div className="space-y-5">
                  <div className="border-l-4 border-red-500 bg-red-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow">
                    <h4 className="font-bold text-gray-900 mb-2 text-lg flex items-center gap-2">
                      ⛔ Never Expose API Keys in Frontend Code
                    </h4>
                    <p className="text-sm text-gray-700">
                      API keys should NEVER be included directly in your frontend JavaScript code or committed to version control.
                    </p>
                  </div>

                  <div className="border-l-4 border-green-500 bg-green-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow">
                    <h4 className="font-bold text-gray-900 mb-3 text-lg flex items-center gap-2">
                      ✅ Use Environment Variables
                    </h4>
                    <p className="text-sm text-gray-700 mb-3">
                      Store API keys in environment variables:
                    </p>
                    <code className="block bg-gray-900 text-green-400 p-4 rounded-lg text-sm font-mono">
                      TRUNC_API_KEY=us_your_api_key_here
                    </code>
                  </div>

                  <div className="border-l-4 border-blue-500 bg-blue-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow">
                    <h4 className="font-bold text-gray-900 mb-2 text-lg flex items-center gap-2">
                      ✅ Implement a Backend Proxy
                    </h4>
                    <p className="text-sm text-gray-700">
                      Create a backend endpoint that calls the Trunc API. Your frontend calls your backend, which securely stores the API key.
                    </p>
                  </div>

                  <div className="border-l-4 border-purple-500 bg-purple-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow">
                    <h4 className="font-bold text-gray-900 mb-2 text-lg flex items-center gap-2">
                      ✅ Use Permission Scopes
                    </h4>
                    <p className="text-sm text-gray-700">
                      Create API keys with only the permissions they need. Don't give full access if you only need URL shortening.
                    </p>
                  </div>
                </div>
              </div>

              {/* Common Use Cases */}
              <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
                <h3 className="text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2">
                  <span className="text-3xl">💡</span> Common Use Cases
                </h3>
                <div className="grid md:grid-cols-2 gap-6">
                  <div className="border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group">
                    <div className="text-4xl mb-3 group-hover:scale-110 transition-transform">📝</div>
                    <h4 className="font-bold text-gray-900 mb-2 text-lg">Blog Post Sharing</h4>
                    <p className="text-sm text-gray-600">
                      Automatically generate short URLs for blog posts to share on social media.
                    </p>
                  </div>
                  <div className="border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group">
                    <div className="text-4xl mb-3 group-hover:scale-110 transition-transform">📧</div>
                    <h4 className="font-bold text-gray-900 mb-2 text-lg">Email Newsletters</h4>
                    <p className="text-sm text-gray-600">
                      Create trackable short links for email campaigns with custom UTM parameters.
                    </p>
                  </div>
                  <div className="border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group">
                    <div className="text-4xl mb-3 group-hover:scale-110 transition-transform">🎯</div>
                    <h4 className="font-bold text-gray-900 mb-2 text-lg">Marketing Campaigns</h4>
                    <p className="text-sm text-gray-600">
                      Generate branded short URLs for marketing materials and track click performance.
                    </p>
                  </div>
                  <div className="border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group">
                    <div className="text-4xl mb-3 group-hover:scale-110 transition-transform">📱</div>
                    <h4 className="font-bold text-gray-900 mb-2 text-lg">QR Code Generation</h4>
                    <p className="text-sm text-gray-600">
                      Create short URLs that are perfect for encoding in QR codes.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}
