'use client';

import { useState } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import { 
  PlusIcon, 
  TrashIcon, 
  ArrowUpTrayIcon,
  CheckCircleIcon,
  XCircleIcon,
  ExclamationTriangleIcon,
  ClipboardDocumentIcon,
  ArrowDownTrayIcon
} from '@heroicons/react/24/outline';

interface URLInput {
  id: string;
  url: string;
  customCode?: string;
}

interface BulkResult {
  successful: number;
  failed: number;
  results: {
    short_code: string;
    short_url: string;
    original_url: string;
    created_at: string;
  }[];
  errors: {
    error: string;
    message: string;
  }[];
}

export default function BulkURLShortener() {
  const [urls, setUrls] = useState<URLInput[]>([
    { id: '1', url: '', customCode: '' }
  ]);
  const [isLoading, setIsLoading] = useState(false);
  const [results, setResults] = useState<BulkResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const { getAccessToken, isAuthenticated } = useAuth();

  const addUrlInput = () => {
    const newId = (urls.length + 1).toString();
    setUrls([...urls, { id: newId, url: '', customCode: '' }]);
  };

  const removeUrlInput = (id: string) => {
    if (urls.length === 1) return; // Keep at least one input
    setUrls(urls.filter(url => url.id !== id));
  };

  const updateUrl = (id: string, field: keyof URLInput, value: string) => {
    setUrls(urls.map(url => 
      url.id === id ? { ...url, [field]: value } : url
    ));
  };

  const importFromText = (text: string) => {
    const lines = text.split('\n').filter(line => line.trim());
    const newUrls: URLInput[] = lines.map((line, index) => ({
      id: (index + 1).toString(),
      url: line.trim(),
      customCode: ''
    }));
    
    if (newUrls.length > 0) {
      setUrls(newUrls);
    }
  };

  const handleFileImport = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = (e) => {
      const text = e.target?.result as string;
      importFromText(text);
    };
    reader.readAsText(file);
  };

  const exportResults = () => {
    if (!results || !results.results.length) return;

    const csvContent = [
      ['Original URL', 'Short URL', 'Short Code', 'Created At'].join(','),
      ...results.results.map(result => [
        result.original_url,
        result.short_url,
        result.short_code,
        result.created_at
      ].join(','))
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `bulk-urls-${Date.now()}.csv`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  };

  const copyAllUrls = async () => {
    if (!results || !results.results.length) return;

    const urlList = results.results.map(result => result.short_url).join('\n');
    
    try {
      await navigator.clipboard.writeText(urlList);
      // You could add a toast notification here
    } catch (err) {
      console.error('Failed to copy URLs:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);
    setResults(null);

    // Filter out empty URLs
    const validUrls = urls.filter(url => url.url.trim());
    
    if (validUrls.length === 0) {
      setError('Please enter at least one URL');
      setIsLoading(false);
      return;
    }

    try {
      const token = getAccessToken();
      const payload = validUrls.map(url => ({
        url: url.url.trim(),
        custom_code: url.customCode?.trim() || undefined
      }));

      const headers: Record<string, string> = {
        'Content-Type': 'application/json',
      };
      
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }

      const response = await fetch('http://localhost:8080/api/v1/batch-shorten', {
        method: 'POST',
        headers,
        body: JSON.stringify(payload),
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URLs');
      }

      const result = await response.json();
      setResults(result);
      
      // Clear the form if all URLs were successful
      if (result.failed === 0) {
        setUrls([{ id: '1', url: '', customCode: '' }]);
      }
      
    } catch (err) {
      setError(err instanceof Error ? err.message : 'An error occurred');
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto">
      <div className="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h2 className="text-2xl font-bold text-gray-900">Bulk URL Shortener</h2>
            <p className="text-gray-600 mt-1">Create multiple short URLs at once</p>
          </div>
          
          {/* Import Options */}
          <div className="flex items-center gap-2">
            <label className="cursor-pointer bg-gray-100 text-gray-700 px-3 py-2 rounded-lg hover:bg-gray-200 transition-colors text-sm font-medium flex items-center gap-2">
              <ArrowUpTrayIcon className="h-4 w-4" />
              Import File
              <input
                type="file"
                accept=".txt,.csv"
                onChange={handleFileImport}
                className="hidden"
              />
            </label>
            
            <button
              type="button"
              onClick={() => {
                const text = prompt('Paste URLs (one per line):');
                if (text) importFromText(text);
              }}
              className="bg-gray-100 text-gray-700 px-3 py-2 rounded-lg hover:bg-gray-200 transition-colors text-sm font-medium"
            >
              Paste URLs
            </button>
          </div>
        </div>

        <form onSubmit={handleSubmit} className="space-y-4">
          {/* URL Inputs */}
          <div className="space-y-3">
            {urls.map((urlInput, index) => (
              <div key={urlInput.id} className="flex items-center gap-3">
                <span className="text-sm text-gray-500 w-8">{index + 1}.</span>
                
                <div className="flex-1">
                  <input
                    type="url"
                    placeholder="https://example.com/long-url"
                    value={urlInput.url}
                    onChange={(e) => updateUrl(urlInput.id, 'url', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 text-sm"
                  />
                </div>
                
                <div className="w-32">
                  <input
                    type="text"
                    placeholder="Custom code"
                    value={urlInput.customCode || ''}
                    onChange={(e) => updateUrl(urlInput.id, 'customCode', e.target.value)}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-primary-500 focus:border-primary-500 text-sm"
                  />
                </div>
                
                <button
                  type="button"
                  onClick={() => removeUrlInput(urlInput.id)}
                  disabled={urls.length === 1}
                  className="p-2 text-red-600 hover:text-red-800 disabled:text-gray-400 disabled:cursor-not-allowed transition-colors"
                  title="Remove URL"
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
            ))}
          </div>

          {/* Add More Button */}
          <div className="flex items-center justify-between">
            <button
              type="button"
              onClick={addUrlInput}
              className="flex items-center gap-2 text-primary-600 hover:text-primary-700 text-sm font-medium"
            >
              <PlusIcon className="h-4 w-4" />
              Add Another URL
            </button>
            
            <div className="text-sm text-gray-500">
              {urls.filter(url => url.url.trim()).length} URLs ready
            </div>
          </div>

          {/* Error Display */}
          {error && (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
              <div className="flex items-center gap-2">
                <XCircleIcon className="h-5 w-5 text-red-500" />
                <p className="text-sm text-red-600">{error}</p>
              </div>
            </div>
          )}

          {/* Submit Button */}
          <div className="pt-4">
            <button
              type="submit"
              disabled={isLoading || !urls.some(url => url.url.trim())}
              className="w-full bg-primary-600 text-white px-6 py-3 rounded-lg hover:bg-primary-700 focus:ring-2 focus:ring-primary-500 focus:ring-offset-2 disabled:opacity-50 disabled:cursor-not-allowed transition-colors font-medium"
            >
              {isLoading ? 'Creating Short URLs...' : 'Create All Short URLs'}
            </button>
          </div>
        </form>

        {/* Results */}
        {results && (
          <div className="mt-8 border-t pt-6">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-lg font-semibold text-gray-900">Results</h3>
              
              <div className="flex items-center gap-2">
                <button
                  onClick={copyAllUrls}
                  className="flex items-center gap-2 text-primary-600 hover:text-primary-700 text-sm font-medium"
                >
                  <ClipboardDocumentIcon className="h-4 w-4" />
                  Copy All URLs
                </button>
                
                <button
                  onClick={exportResults}
                  className="flex items-center gap-2 bg-primary-600 text-white px-3 py-2 rounded-lg hover:bg-primary-700 transition-colors text-sm"
                >
                  <ArrowDownTrayIcon className="h-4 w-4" />
                  Export CSV
                </button>
              </div>
            </div>

            {/* Summary Stats */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
              <div className="bg-green-50 p-4 rounded-lg">
                <div className="flex items-center gap-2">
                  <CheckCircleIcon className="h-6 w-6 text-green-600" />
                  <div>
                    <p className="text-sm font-medium text-green-800">Successful</p>
                    <p className="text-2xl font-bold text-green-900">{results.successful}</p>
                  </div>
                </div>
              </div>
              
              <div className="bg-red-50 p-4 rounded-lg">
                <div className="flex items-center gap-2">
                  <XCircleIcon className="h-6 w-6 text-red-600" />
                  <div>
                    <p className="text-sm font-medium text-red-800">Failed</p>
                    <p className="text-2xl font-bold text-red-900">{results.failed}</p>
                  </div>
                </div>
              </div>
              
              <div className="bg-blue-50 p-4 rounded-lg">
                <div className="flex items-center gap-2">
                  <ExclamationTriangleIcon className="h-6 w-6 text-blue-600" />
                  <div>
                    <p className="text-sm font-medium text-blue-800">Total</p>
                    <p className="text-2xl font-bold text-blue-900">{results.successful + results.failed}</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Successful Results */}
            {results.results.length > 0 && (
              <div className="mb-6">
                <h4 className="font-medium text-gray-900 mb-3">Successfully Created URLs</h4>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {results.results.map((result, index) => (
                    <div key={index} className="flex items-center justify-between bg-green-50 p-3 rounded-lg">
                      <div className="flex-1 min-w-0">
                        <p className="text-sm font-medium text-green-900 truncate">
                          {result.short_url}
                        </p>
                        <p className="text-xs text-green-600 truncate">
                          {result.original_url}
                        </p>
                      </div>
                      <button
                        onClick={() => navigator.clipboard.writeText(result.short_url)}
                        className="ml-2 p-1 text-green-700 hover:text-green-800 transition-colors"
                        title="Copy URL"
                      >
                        <ClipboardDocumentIcon className="h-4 w-4" />
                      </button>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Errors */}
            {results.errors.length > 0 && (
              <div>
                <h4 className="font-medium text-gray-900 mb-3">Errors</h4>
                <div className="space-y-2 max-h-32 overflow-y-auto">
                  {results.errors.map((error, index) => (
                    <div key={index} className="bg-red-50 p-3 rounded-lg">
                      <p className="text-sm text-red-600">{error.message}</p>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
}