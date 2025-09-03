'use client';

import { useState } from 'react';
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
  {
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
  "short_url": "http://localhost:8080/my-link",
  "original_url": "https://example.com/very-long-url",
  "created_at": "2023-01-01T12:00:00Z"
}`
    }
  },
  {
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
    method: 'POST',
    path: '/api/v1/auth/register',
    description: 'Register a new user',
    requestBody: {
      contentType: 'application/json',
      schema: `{
  "name": "string (required)",
  "email": "string (required)",
  "password": "string (required)",
  "phone": "string (required)",
  "terms_accepted": "boolean (required)",
  "marketing_consent": "boolean (optional)"
}`
    },
    responses: [
      {
        status: 201,
        description: 'User registered successfully',
        schema: `{
  "user": {
    "id": "number",
    "name": "string",
    "email": "string",
    "phone": "string",
    "account_type": "string",
    "created_at": "string"
  },
  "access_token": "string",
  "refresh_token": "string",
  "expires_in": "number"
}`
      },
      {
        status: 400,
        description: 'Invalid request data'
      },
      {
        status: 409,
        description: 'Email already registered'
      }
    ]
  },
  {
    method: 'POST',
    path: '/api/v1/auth/login',
    description: 'Login user',
    requestBody: {
      contentType: 'application/json',
      schema: `{
  "email": "string (required)",
  "password": "string (required)",
  "remember_me": "boolean (optional)"
}`
    },
    responses: [
      {
        status: 200,
        description: 'Login successful',
        schema: `{
  "user": "User object",
  "access_token": "string",
  "refresh_token": "string",
  "expires_in": "number"
}`
      },
      {
        status: 401,
        description: 'Invalid credentials'
      }
    ]
  },
  {
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
    method: 'GET',
    path: '/api/v1/analytics/{shortCode}',
    description: 'Get analytics for a URL',
    parameters: [
      {
        name: 'shortCode',
        type: 'string',
        required: true,
        description: 'The short code for the URL'
      },
      {
        name: 'days',
        type: 'number',
        required: false,
        description: 'Number of days to retrieve (default: 30, max: 365)'
      }
    ],
    responses: [
      {
        status: 200,
        description: 'Analytics data retrieved successfully',
        schema: `{
  "short_code": "string",
  "original_url": "string",
  "total_clicks": "number",
  "created_at": "string",
  "last_click_at": "string",
  "daily_clicks": "Array of daily click data",
  "country_stats": "Array of country statistics"
}`
      },
      {
        status: 404,
        description: 'URL not found'
      }
    ]
  }
];

export default function APIDocsPage() {
  const [selectedEndpoint, setSelectedEndpoint] = useState<APIEndpoint | null>(null);
  const [copiedCode, setCopiedCode] = useState<string | null>(null);

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
              Complete reference for the URL Shortener API endpoints and usage.
            </p>
          </div>

          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Endpoint List */}
            <div className="lg:col-span-1">
              <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6 sticky top-8">
                <h2 className="text-lg font-semibold text-gray-900 mb-4">Endpoints</h2>
                <div className="space-y-2">
                  {apiEndpoints.map((endpoint, index) => (
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
                      <div className="text-sm font-mono text-gray-700 mb-1">
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
                  <div className="flex items-center gap-3 mb-4">
                    <span className={`px-3 py-1 text-sm font-medium rounded ${getMethodColor(selectedEndpoint.method)}`}>
                      {selectedEndpoint.method}
                    </span>
                    <code className="text-lg font-mono text-gray-900">{selectedEndpoint.path}</code>
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
                  {selectedEndpoint.parameters && (
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
                <code className="block bg-gray-100 p-2 rounded text-sm">http://localhost:8080</code>
              </div>
              
              <div>
                <h3 className="text-lg font-semibold text-gray-900 mb-2">Authentication</h3>
                <p className="text-sm text-gray-600">
                  Include your JWT token in the Authorization header:
                </p>
                <code className="block bg-gray-100 p-2 rounded text-sm mt-1">
                  Authorization: Bearer YOUR_JWT_TOKEN
                </code>
              </div>
            </div>

            <div className="mt-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-2">Rate Limiting</h3>
              <p className="text-sm text-gray-600">
                API requests are rate limited to prevent abuse. Different endpoints have different limits:
              </p>
              <ul className="text-sm text-gray-600 mt-2 space-y-1">
                <li>• General endpoints: Standard rate limiting applies</li>
                <li>• Authentication endpoints: More restrictive limits</li>
                <li>• OTP endpoints: Most restrictive limits</li>
              </ul>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}