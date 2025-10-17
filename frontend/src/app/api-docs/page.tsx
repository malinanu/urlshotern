'use client';

import { useState, useMemo, useEffect } from 'react';
import Layout from '@/components/layout/Layout';
import {
  CodeBracketIcon,
  DocumentTextIcon,
  ChevronRightIcon,
  ClipboardDocumentIcon,
  CheckIcon
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

  // ============ API Keys ============
  {
    category: 'API Keys',
    method: 'POST',
    path: '/api/v1/api-keys',
    description: 'Create a new API key for programmatic access',
    auth: 'Required',
    requestBody: {
      contentType: 'application/json',
      schema: `{
  "name": "string (required)",
  "description": "string (optional)",
  "permissions": "array of strings (optional)",
  "rate_limit": "number (optional, default: 1000 requests/hour)",
  "expires_at": "string (optional, ISO 8601 format)"
}`
    },
    responses: [
      {
        status: 201,
        description: 'API key created successfully',
        schema: `{
  "api_key": {
    "id": "uuid",
    "key_prefix": "string",
    "name": "string",
    "description": "string",
    "permissions": "array",
    "rate_limit": "number",
    "expires_at": "string | null",
    "created_at": "string"
  },
  "plain_key": "string",
  "key_prefix": "string"
}`
      },
      {
        status: 401,
        description: 'Authentication required'
      }
    ],
    example: {
      request: `{
  "name": "My Website Integration",
  "description": "API key for example.com",
  "permissions": ["urls:create", "urls:read", "analytics:read"],
  "rate_limit": 5000
}`,
      response: `{
  "api_key": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "key_prefix": "us_a1b2c3...x4y5z6",
    "name": "My Website Integration",
    "description": "API key for example.com",
    "permissions": ["urls:create", "urls:read", "analytics:read"],
    "rate_limit": 5000,
    "expires_at": null,
    "created_at": "2023-01-01T12:00:00Z"
  },
  "plain_key": "us_a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4",
  "key_prefix": "us_a1b2c3...x4y5z6"
}`
    }
  },
  {
    category: 'API Keys',
    method: 'GET',
    path: '/api/v1/api-keys',
    description: 'List all API keys for authenticated user',
    auth: 'Required',
    parameters: [
      {
        name: 'limit',
        type: 'number',
        required: false,
        description: 'Items per page (default: 20, max: 100)'
      },
      {
        name: 'offset',
        type: 'number',
        required: false,
        description: 'Pagination offset (default: 0)'
      },
      {
        name: 'active',
        type: 'boolean',
        required: false,
        description: 'Filter by active status'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'API keys retrieved successfully',
        schema: `{
  "data": "Array of API key objects",
  "total": "number",
  "limit": "number",
  "offset": "number",
  "has_more": "boolean"
}`
      }
    ]
  },
  {
    category: 'API Keys',
    method: 'GET',
    path: '/api/v1/api-keys/{id}',
    description: 'Get a specific API key',
    auth: 'Required',
    parameters: [
      {
        name: 'id',
        type: 'uuid',
        required: true,
        description: 'The API key UUID (e.g., 550e8400-e29b-41d4-a716-446655440000)'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'API key retrieved successfully'
      },
      {
        status: 404,
        description: 'API key not found'
      }
    ]
  },
  {
    category: 'API Keys',
    method: 'GET',
    path: '/api/v1/api-keys/{id}/stats',
    description: 'Get usage statistics for an API key',
    auth: 'Required',
    parameters: [
      {
        name: 'id',
        type: 'uuid',
        required: true,
        description: 'The API key UUID'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'Statistics retrieved successfully',
        schema: `{
  "api_key_id": "uuid",
  "total_requests": "number",
  "today_requests": "number",
  "week_requests": "number",
  "month_requests": "number",
  "success_rate": "number",
  "avg_response_time": "number"
}`
      }
    ]
  },
  {
    category: 'API Keys',
    method: 'PUT',
    path: '/api/v1/api-keys/{id}/revoke',
    description: 'Revoke (deactivate) an API key',
    auth: 'Required',
    parameters: [
      {
        name: 'id',
        type: 'uuid',
        required: true,
        description: 'The API key UUID'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'API key revoked successfully'
      },
      {
        status: 404,
        description: 'API key not found'
      }
    ]
  },
  {
    category: 'API Keys',
    method: 'DELETE',
    path: '/api/v1/api-keys/{id}',
    description: 'Permanently delete an API key',
    auth: 'Required',
    parameters: [
      {
        name: 'id',
        type: 'uuid',
        required: true,
        description: 'The API key UUID'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'API key deleted successfully'
      },
      {
        status: 404,
        description: 'API key not found'
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
  const [apiUrl, setApiUrl] = useState<string>('http://localhost:8080');

  // Set API URL on client side only to avoid hydration mismatch
  useEffect(() => {
    setApiUrl(getApiUrl());
  }, []);

  // Get unique categories
  const categories = useMemo(() => {
    const cats = ['All', ...new Set(apiEndpoints.map(e => e.category || 'Other'))];
    return cats;
  }, []);

  // Filter endpoints by category
  const filteredEndpoints = useMemo(() => {
    if (selectedCategory === 'All') return apiEndpoints;
    return apiEndpoints.filter(e => e.category === selectedCategory);
  }, [selectedCategory]);

  const copyToClipboard = (text: string, id: string) => {
    navigator.clipboard.writeText(text).then(() => {
      setCopiedCode(id);
      setTimeout(() => setCopiedCode(null), 2000);
    });
  };

  const getMethodColor = (method: string) => {
    switch (method) {
      case 'GET': return 'bg-green-100 text-green-800';
      case 'POST': return 'bg-blue-100 text-blue-800';
      case 'PUT': return 'bg-yellow-100 text-yellow-800';
      case 'DELETE': return 'bg-red-100 text-red-800';
      default: return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <DocumentTextIcon className="h-8 w-8 text-primary-600" />
              <h1 className="text-3xl font-bold text-gray-900">API Documentation</h1>
            </div>
            <p className="text-gray-600">
              Complete reference for the Trunc API endpoints and usage.
            </p>
          </div>

          {/* Category Filter */}
          <div className="mb-6 flex flex-wrap gap-2">
            {categories.map((category) => (
              <button
                key={category}
                onClick={() => setSelectedCategory(category)}
                className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
                  selectedCategory === category
                    ? 'bg-primary-600 text-white'
                    : 'bg-white text-gray-700 hover:bg-gray-50 border border-gray-200'
                }`}
              >
                {category}
              </button>
            ))}
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Endpoint List */}
            <div className="lg:col-span-1">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 sticky top-8 max-h-[calc(100vh-8rem)] overflow-y-auto">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">Endpoints ({filteredEndpoints.length})</h2>
                <div className="space-y-2">
                  {filteredEndpoints.map((endpoint, index) => (
                    <button
                      key={index}
                      onClick={() => setSelectedEndpoint(endpoint)}
                      className={`w-full text-left p-3 rounded-lg transition-colors ${
                        selectedEndpoint === endpoint
                          ? 'bg-primary-50 border border-primary-200'
                          : 'hover:bg-gray-50 border border-transparent'
                      }`}
                    >
                      <div className="flex items-center justify-between mb-1">
                        <span className={`px-2 py-1 text-xs font-medium rounded ${getMethodColor(endpoint.method)}`}>
                          {endpoint.method}
                        </span>
                        {endpoint.auth && (
                          <span className="text-xs text-gray-500">{endpoint.auth}</span>
                        )}
                      </div>
                      <div className="text-sm font-mono text-gray-700 mb-1 break-all">
                        {endpoint.path}
                      </div>
                      <div className="text-xs text-gray-500">
                        {endpoint.description}
                      </div>
                    </button>
                  ))}
                </div>
              </div>
            </div>

            {/* Endpoint Details */}
            <div className="lg:col-span-2">
              {selectedEndpoint ? (
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
                  {selectedEndpoint.category && (
                    <div className="mb-2">
                      <span className="text-xs font-medium text-primary-600 bg-primary-50 px-2 py-1 rounded">
                        {selectedEndpoint.category}
                      </span>
                    </div>
                  )}
                  <div className="flex items-center gap-3 mb-4">
                    <span className={`px-3 py-1 text-sm font-medium rounded ${getMethodColor(selectedEndpoint.method)}`}>
                      {selectedEndpoint.method}
                    </span>
                    <code className="text-lg font-mono text-gray-900 break-all">{selectedEndpoint.path}</code>
                  </div>

                  <p className="text-gray-600 mb-6">{selectedEndpoint.description}</p>

                  {/* Authentication */}
                  {selectedEndpoint.auth && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-2">Authentication</h3>
                      <p className="text-sm text-gray-600">
                        {selectedEndpoint.auth === 'Required'
                          ? 'This endpoint requires authentication. Include the Bearer token in the Authorization header.'
                          : 'Authentication is optional for this endpoint. Include the Bearer token for enhanced features.'
                        }
                      </p>
                    </div>
                  )}

                  {/* Parameters */}
                  {selectedEndpoint.parameters && selectedEndpoint.parameters.length > 0 && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-2">Parameters</h3>
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="border-b">
                              <th className="text-left py-2">Name</th>
                              <th className="text-left py-2">Type</th>
                              <th className="text-left py-2">Required</th>
                              <th className="text-left py-2">Description</th>
                            </tr>
                          </thead>
                          <tbody>
                            {selectedEndpoint.parameters.map((param, index) => (
                              <tr key={index} className="border-b">
                                <td className="py-2 font-mono">{param.name}</td>
                                <td className="py-2">{param.type}</td>
                                <td className="py-2">
                                  <span className={`px-2 py-1 text-xs rounded ${
                                    param.required ? 'bg-red-100 text-red-800' : 'bg-gray-100 text-gray-800'
                                  }`}>
                                    {param.required ? 'Yes' : 'No'}
                                  </span>
                                </td>
                                <td className="py-2">{param.description}</td>
                              </tr>
                            ))}
                          </tbody>
                        </table>
                      </div>
                    </div>
                  )}

                  {/* Request Body */}
                  {selectedEndpoint.requestBody && (
                    <div className="mb-6">
                      <h3 className="text-lg font-semibold text-gray-900 mb-2">Request Body</h3>
                      <p className="text-sm text-gray-600 mb-2">Content-Type: {selectedEndpoint.requestBody.contentType}</p>
                      <div className="bg-gray-900 rounded-lg p-4 relative">
                        <button
                          onClick={() => copyToClipboard(selectedEndpoint.requestBody!.schema, 'request-schema')}
                          className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                        >
                          {copiedCode === 'request-schema' ? (
                            <CheckIcon className="h-4 w-4" />
                          ) : (
                            <ClipboardDocumentIcon className="h-4 w-4" />
                          )}
                        </button>
                        <pre className="text-green-400 text-sm overflow-x-auto">
                          <code>{selectedEndpoint.requestBody.schema}</code>
                        </pre>
                      </div>
                    </div>
                  )}

                  {/* Responses */}
                  <div className="mb-6">
                    <h3 className="text-lg font-semibold text-gray-900 mb-2">Responses</h3>
                    <div className="space-y-4">
                      {selectedEndpoint.responses.map((response, index) => (
                        <div key={index} className="border border-gray-200 rounded-lg p-4">
                          <div className="flex items-center gap-2 mb-2">
                            <span className={`px-2 py-1 text-xs font-medium rounded ${
                              response.status < 300 ? 'bg-green-100 text-green-800' :
                              response.status < 400 ? 'bg-blue-100 text-blue-800' :
                              response.status < 500 ? 'bg-yellow-100 text-yellow-800' :
                              'bg-red-100 text-red-800'
                            }`}>
                              {response.status}
                            </span>
                            <span className="text-sm text-gray-600">{response.description}</span>
                          </div>
                          {response.schema && (
                            <div className="bg-gray-900 rounded p-3 relative">
                              <button
                                onClick={() => copyToClipboard(response.schema!, `response-${index}`)}
                                className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                              >
                                {copiedCode === `response-${index}` ? (
                                  <CheckIcon className="h-4 w-4" />
                                ) : (
                                  <ClipboardDocumentIcon className="h-4 w-4" />
                                )}
                              </button>
                              <pre className="text-green-400 text-xs overflow-x-auto">
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
                      <h3 className="text-lg font-semibold text-gray-900 mb-2">Example</h3>

                      {selectedEndpoint.example.request && (
                        <div className="mb-4">
                          <h4 className="text-sm font-medium text-gray-700 mb-2">Request:</h4>
                          <div className="bg-gray-900 rounded-lg p-4 relative">
                            <button
                              onClick={() => copyToClipboard(selectedEndpoint.example!.request!, 'example-request')}
                              className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                            >
                              {copiedCode === 'example-request' ? (
                                <CheckIcon className="h-4 w-4" />
                              ) : (
                                <ClipboardDocumentIcon className="h-4 w-4" />
                              )}
                            </button>
                            <pre className="text-green-400 text-sm overflow-x-auto">
                              <code>{selectedEndpoint.example.request}</code>
                            </pre>
                          </div>
                        </div>
                      )}

                      <div>
                        <h4 className="text-sm font-medium text-gray-700 mb-2">Response:</h4>
                        <div className="bg-gray-900 rounded-lg p-4 relative">
                          <button
                            onClick={() => copyToClipboard(selectedEndpoint.example.response, 'example-response')}
                            className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                          >
                            {copiedCode === 'example-response' ? (
                              <CheckIcon className="h-4 w-4" />
                            ) : (
                              <ClipboardDocumentIcon className="h-4 w-4" />
                            )}
                          </button>
                          <pre className="text-green-400 text-sm overflow-x-auto">
                            <code>{selectedEndpoint.example.response}</code>
                          </pre>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              ) : (
                <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-12 text-center">
                  <CodeBracketIcon className="h-16 w-16 text-gray-400 mx-auto mb-4" />
                  <h3 className="text-xl font-semibold text-gray-900 mb-2">
                    Select an Endpoint
                  </h3>
                  <p className="text-gray-600">
                    Choose an endpoint from the list to view detailed documentation.
                  </p>
                </div>
              )}
            </div>
          </div>

          {/* Getting Started Section */}
          <div className="mt-12 bg-white rounded-lg shadow-sm border border-gray-200 p-6">
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Getting Started</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Base URL</h3>
                <code className="block bg-gray-100 p-2 rounded text-sm">{apiUrl}</code>
                <p className="text-xs text-gray-500 mt-1">
                  Production: <code className="text-primary-600">https://api.trunc.site</code>
                </p>
                <p className="text-xs text-gray-500">
                  Local: <code className="text-gray-600">http://localhost:8080</code>
                </p>
              </div>

              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Authentication</h3>
                <p className="text-sm text-gray-600 mb-2">
                  Include your API key in the Authorization header:
                </p>
                <code className="block bg-gray-100 p-2 rounded text-sm">
                  Authorization: Bearer YOUR_API_KEY
                </code>
                <p className="text-xs text-gray-500 mt-2">
                  API keys start with <code className="text-primary-600">us_</code> prefix.
                </p>
              </div>
            </div>

            <div className="mt-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Rate Limiting</h3>
              <p className="text-sm text-gray-600">
                API requests are rate limited to prevent abuse. Rate limits are configured per API key and can vary based on your subscription plan.
              </p>
              <ul className="text-sm text-gray-600 mt-2 space-y-1">
                <li>• Standard endpoints: Default rate limiting applies</li>
                <li>• Custom rate limits: Can be configured for your API key in settings</li>
                <li>• Rate limit headers: Check response headers for current limit status</li>
              </ul>
            </div>
          </div>

          {/* Integration Guide Section */}
          <div className="mt-12 bg-gradient-to-br from-primary-50 to-blue-50 rounded-lg shadow-sm border border-primary-200 p-8">
            <div className="flex items-center gap-3 mb-4">
              <CodeBracketIcon className="h-8 w-8 text-primary-600" />
              <h2 className="text-3xl font-bold text-gray-900">Integration Guide</h2>
            </div>
            <p className="text-gray-700 mb-8">
              Learn how to integrate Trunc URL Shortener into your website or application using API keys.
            </p>

            {/* Prerequisites */}
            <div className="bg-amber-50 border border-amber-200 rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">⚠️ Prerequisites</h3>
              <p className="text-sm text-gray-700 mb-4">
                Before you can use the API, you need a Trunc account and an API key. <strong>You cannot register or login via the API.</strong> Account creation and API key generation must be done through the web interface.
              </p>
              <div className="space-y-3">
                <div className="flex items-start gap-3">
                  <span className="text-lg">1️⃣</span>
                  <div>
                    <h4 className="font-semibold text-gray-900">Register on the Website</h4>
                    <p className="text-sm text-gray-700">Visit <a href="/" className="text-primary-600 hover:underline">trunc.site</a> and create an account if you haven't already.</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <span className="text-lg">2️⃣</span>
                  <div>
                    <h4 className="font-semibold text-gray-900">Login to Dashboard</h4>
                    <p className="text-sm text-gray-700">Sign in to your account and access the dashboard.</p>
                  </div>
                </div>
                <div className="flex items-start gap-3">
                  <span className="text-lg">3️⃣</span>
                  <div>
                    <h4 className="font-semibold text-gray-900">Generate API Key</h4>
                    <p className="text-sm text-gray-700">Navigate to Settings → API Keys and create a new API key with the required permissions.</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Quick Start */}
            <div className="bg-white rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">🚀 Quick Start</h3>
              <p className="text-sm text-gray-600 mb-4">
                Once you have your API key from the prerequisites above, follow these steps:
              </p>
              <div className="space-y-4">
                <div className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                    1
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 mb-1">Copy Your API Key</h4>
                    <p className="text-sm text-gray-600">After creating an API key in the dashboard, copy it immediately. You won't be able to see it again!</p>
                  </div>
                </div>
                <div className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                    2
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 mb-1">Store Securely</h4>
                    <p className="text-sm text-gray-600">Save your API key in a secure location (environment variables, secrets manager, etc.). Never commit it to version control!</p>
                  </div>
                </div>
                <div className="flex gap-4">
                  <div className="flex-shrink-0 w-8 h-8 bg-primary-600 text-white rounded-full flex items-center justify-center font-bold">
                    3
                  </div>
                  <div>
                    <h4 className="font-semibold text-gray-900 mb-1">Make Your First Request</h4>
                    <p className="text-sm text-gray-600">Use the code examples below to start shortening URLs programmatically!</p>
                  </div>
                </div>
              </div>
            </div>

            {/* JavaScript/React Example */}
            <div className="bg-white rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">JavaScript / React Integration</h3>

              <div className="mb-6">
                <h4 className="text-sm font-semibold text-gray-700 mb-2">Vanilla JavaScript:</h4>
                <div className="bg-gray-900 rounded-lg p-4 relative">
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
                    className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                  >
                    {copiedCode === 'js-example' ? (
                      <CheckIcon className="h-4 w-4" />
                    ) : (
                      <ClipboardDocumentIcon className="h-4 w-4" />
                    )}
                  </button>
                  <pre className="text-green-400 text-xs overflow-x-auto">
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

              <div className="mb-6">
                <h4 className="text-sm font-semibold text-gray-700 mb-2">React Custom Hook:</h4>
                <div className="bg-gray-900 rounded-lg p-4 relative">
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
}

// Usage in component:
function MyComponent() {
  const { shortenURL, loading, error } = useURLShortener(process.env.REACT_APP_API_KEY);
  const [shortUrl, setShortUrl] = useState('');

  const handleShorten = async () => {
    try {
      const result = await shortenURL('https://example.com/long-url');
      setShortUrl(result);
    } catch (err) {
      console.error('Failed to shorten:', err);
    }
  };

  return (
    <div>
      <button onClick={handleShorten} disabled={loading}>
        {loading ? 'Shortening...' : 'Shorten URL'}
      </button>
      {shortUrl && <p>Short URL: {shortUrl}</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}`, 'react-hook')}
                    className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                  >
                    {copiedCode === 'react-hook' ? (
                      <CheckIcon className="h-4 w-4" />
                    ) : (
                      <ClipboardDocumentIcon className="h-4 w-4" />
                    )}
                  </button>
                  <pre className="text-green-400 text-xs overflow-x-auto">
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
}

// Usage in component:
function MyComponent() {
  const { shortenURL, loading, error } = useURLShortener(process.env.REACT_APP_API_KEY);
  const [shortUrl, setShortUrl] = useState('');

  const handleShorten = async () => {
    try {
      const result = await shortenURL('https://example.com/long-url');
      setShortUrl(result);
    } catch (err) {
      console.error('Failed to shorten:', err);
    }
  };

  return (
    <div>
      <button onClick={handleShorten} disabled={loading}>
        {loading ? 'Shortening...' : 'Shorten URL'}
      </button>
      {shortUrl && <p>Short URL: {shortUrl}</p>}
      {error && <p className="error">{error}</p>}
    </div>
  );
}`}</code>
                  </pre>
                </div>
              </div>
            </div>

            {/* Node.js Example */}
            <div className="bg-white rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">Node.js Backend Integration</h3>
              <div className="bg-gray-900 rounded-lg p-4 relative">
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
                  className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                >
                  {copiedCode === 'node-example' ? (
                    <CheckIcon className="h-4 w-4" />
                  ) : (
                    <ClipboardDocumentIcon className="h-4 w-4" />
                  )}
                </button>
                <pre className="text-green-400 text-xs overflow-x-auto">
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

            {/* Python Example */}
            <div className="bg-white rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">Python Integration</h3>
              <div className="bg-gray-900 rounded-lg p-4 relative">
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
print(f"Shortened URL: {short_url}")

# Flask example:
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api/shorten', methods=['POST'])
def shorten():
    original_url = request.json.get('url')
    api_key = os.getenv('TRUNC_API_KEY')

    try:
        short_url = shorten_url(original_url, api_key)
        return jsonify({'success': True, 'shortUrl': short_url})
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)}), 500`, 'python-example')}
                  className="absolute top-2 right-2 p-1 text-gray-400 hover:text-white transition-colors"
                >
                  {copiedCode === 'python-example' ? (
                    <CheckIcon className="h-4 w-4" />
                  ) : (
                    <ClipboardDocumentIcon className="h-4 w-4" />
                  )}
                </button>
                <pre className="text-green-400 text-xs overflow-x-auto">
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
print(f"Shortened URL: {short_url}")

# Flask example:
from flask import Flask, request, jsonify

app = Flask(__name__)

@app.route('/api/shorten', methods=['POST'])
def shorten():
    original_url = request.json.get('url')
    api_key = os.getenv('TRUNC_API_KEY')

    try:
        short_url = shorten_url(original_url, api_key)
        return jsonify({'success': True, 'shortUrl': short_url})
    except Exception as e:
        return jsonify({'success': False, 'error': str(e)}), 500`}</code>
                </pre>
              </div>
            </div>

            {/* Security Best Practices */}
            <div className="bg-white rounded-lg p-6 mb-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">🔐 Security Best Practices</h3>
              <div className="space-y-4">
                <div className="border-l-4 border-yellow-500 bg-yellow-50 p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">⚠️ Never Expose API Keys in Frontend Code</h4>
                  <p className="text-sm text-gray-700">
                    API keys should NEVER be included directly in your frontend JavaScript code or committed to version control.
                  </p>
                </div>

                <div className="border-l-4 border-green-500 bg-green-50 p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">✅ Use Environment Variables</h4>
                  <p className="text-sm text-gray-700 mb-2">
                    Store API keys in environment variables:
                  </p>
                  <code className="block bg-gray-900 text-green-400 p-2 rounded text-xs">
                    TRUNC_API_KEY=us_your_api_key_here
                  </code>
                </div>

                <div className="border-l-4 border-blue-500 bg-blue-50 p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">✅ Implement a Backend Proxy</h4>
                  <p className="text-sm text-gray-700">
                    Create a backend endpoint that calls the Trunc API. Your frontend calls your backend, which securely stores the API key.
                  </p>
                </div>

                <div className="border-l-4 border-purple-500 bg-purple-50 p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">✅ Use Permission Scopes</h4>
                  <p className="text-sm text-gray-700">
                    Create API keys with only the permissions they need. Don't give full access if you only need URL shortening.
                  </p>
                </div>

                <div className="border-l-4 border-red-500 bg-red-50 p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">✅ Rotate Keys Regularly</h4>
                  <p className="text-sm text-gray-700">
                    Regenerate API keys periodically and immediately revoke keys that may have been compromised.
                  </p>
                </div>
              </div>
            </div>

            {/* Common Use Cases */}
            <div className="bg-white rounded-lg p-6">
              <h3 className="text-xl font-bold text-gray-900 mb-4">💡 Common Use Cases</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">📝 Blog Post Sharing</h4>
                  <p className="text-sm text-gray-600">
                    Automatically generate short URLs for blog posts to share on social media.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">📧 Email Newsletters</h4>
                  <p className="text-sm text-gray-600">
                    Create trackable short links for email campaigns with custom UTM parameters.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">🎯 Marketing Campaigns</h4>
                  <p className="text-sm text-gray-600">
                    Generate branded short URLs for marketing materials and track click performance.
                  </p>
                </div>
                <div className="border border-gray-200 rounded-lg p-4">
                  <h4 className="font-semibold text-gray-900 mb-2">📱 QR Code Generation</h4>
                  <p className="text-sm text-gray-600">
                    Create short URLs that are perfect for encoding in QR codes.
                  </p>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}
