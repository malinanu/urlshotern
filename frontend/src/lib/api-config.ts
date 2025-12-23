/**
 * API Configuration
 *
 * Centralized API endpoint configuration for the URL shortener application.
 * All API routes are based on the backend routes defined in internal/routes/routes.go
 */

// Get the API base URL from environment or default to localhost
const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

/**
 * API endpoint configuration object
 */
export const API_ENDPOINTS = {
  // Authentication endpoints
  auth: {
    register: `${API_BASE_URL}/api/v1/auth/register`,
    login: `${API_BASE_URL}/api/v1/auth/login`,
    logout: `${API_BASE_URL}/api/v1/auth/logout`,
    refresh: `${API_BASE_URL}/api/v1/auth/refresh`,
    profile: `${API_BASE_URL}/api/v1/auth/profile`,
    updateProfile: `${API_BASE_URL}/api/v1/auth/profile`,
    changePassword: `${API_BASE_URL}/api/v1/auth/change-password`,
    verifyEmail: `${API_BASE_URL}/api/v1/auth/verify-email`,
    resendEmail: `${API_BASE_URL}/api/v1/auth/resend-email`,
    otp: {
      send: `${API_BASE_URL}/api/v1/auth/otp/send`,
      verify: `${API_BASE_URL}/api/v1/auth/otp/verify`,
      resend: `${API_BASE_URL}/api/v1/auth/otp/resend`,
    },
  },

  // URL shortening endpoints
  urls: {
    shorten: `${API_BASE_URL}/api/v1/shorten`,
    batchShorten: `${API_BASE_URL}/api/v1/batch-shorten`,
    myUrls: `${API_BASE_URL}/api/v1/my-urls`,
    deleteUrl: (shortCode: string) => `${API_BASE_URL}/api/v1/my-urls/${shortCode}`,
    updateUrl: (shortCode: string) => `${API_BASE_URL}/api/v1/my-urls/${shortCode}`,
  },

  // Analytics endpoints
  analytics: {
    basic: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}`,
    trends: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/trends`,
    advanced: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/advanced`,
    geographic: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/geographic`,
    time: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/time`,
    device: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/device`,
    referrers: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/referrers`,
    referrersEnhanced: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/referrers/enhanced`,
    utmCampaigns: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/utm-campaigns`,
    referrerInsights: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/referrers/insights`,
    heatmap: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/heatmap`,
    map: (shortCode: string) => `${API_BASE_URL}/api/v1/analytics/${shortCode}/map`,
  },

  // Dashboard endpoints
  dashboard: {
    stats: `${API_BASE_URL}/api/v1/dashboard/stats`,
  },

  // Billing endpoints (Polar integration)
  billing: {
    subscription: `${API_BASE_URL}/api/v1/billing/subscription`,
    plans: `${API_BASE_URL}/api/v1/billing/plans`,
    products: `${API_BASE_URL}/api/v1/billing/products`,      // Protected - requires auth
    publicProducts: `${API_BASE_URL}/api/v1/products`,         // Public - no auth required
    usage: `${API_BASE_URL}/api/v1/billing/usage`,
    checkout: `${API_BASE_URL}/api/v1/billing/checkout`,
    cancel: `${API_BASE_URL}/api/v1/billing/cancel-subscription`,
    reactivate: `${API_BASE_URL}/api/v1/billing/reactivate-subscription`,
    portal: `${API_BASE_URL}/api/v1/billing/portal`,
    paymentMethod: `${API_BASE_URL}/api/v1/billing/payment-method`,
    history: `${API_BASE_URL}/api/v1/billing/history`,
  },

  // Admin endpoints (temporarily disabled in backend)
  admin: {
    stats: `${API_BASE_URL}/api/v1/admin/stats`,
    users: `${API_BASE_URL}/api/v1/admin/users`,
    activity: `${API_BASE_URL}/api/v1/admin/activity`,
    userStatus: (userId: string) => `${API_BASE_URL}/api/v1/admin/users/${userId}/status`,
    deleteUser: (userId: string) => `${API_BASE_URL}/api/v1/admin/users/${userId}`,
  },

  // CMS endpoints (temporarily disabled in backend)
  cms: {
    terms: `${API_BASE_URL}/api/v1/pages/terms`,
    privacy: `${API_BASE_URL}/api/v1/pages/privacy`,
    page: (slug: string) => `${API_BASE_URL}/api/v1/pages/${slug}`,
  },

  // Conversion tracking endpoints
  conversions: {
    goals: `${API_BASE_URL}/api/v1/conversions/goals`,
    goal: (goalId: string) => `${API_BASE_URL}/api/v1/conversions/goals/${goalId}`,
    track: `${API_BASE_URL}/api/v1/conversions/track`,
    stats: (goalId: string) => `${API_BASE_URL}/api/v1/conversions/goals/${goalId}/stats`,
    attribution: (goalId: string) => `${API_BASE_URL}/api/v1/conversions/goals/${goalId}/attribution`,
  },

  // A/B Testing endpoints
  abTesting: {
    tests: `${API_BASE_URL}/api/v1/ab-tests`,
    test: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}`,
    start: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/start`,
    stop: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/stop`,
    variant: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/variant`,
    conversion: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/conversion`,
    results: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/results`,
    sequential: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/sequential`,
    power: (testId: string) => `${API_BASE_URL}/api/v1/ab-tests/${testId}/power`,
    sampleSize: `${API_BASE_URL}/api/v1/ab-tests/sample-size-calculator`,
  },

  // Attribution endpoints
  attribution: {
    touchpoints: `${API_BASE_URL}/api/v1/attribution/touchpoints`,
    journey: (conversionId: string) => `${API_BASE_URL}/api/v1/attribution/journey/${conversionId}`,
    report: (conversionId: string) => `${API_BASE_URL}/api/v1/attribution/report/${conversionId}`,
    calculate: (conversionId: string) => `${API_BASE_URL}/api/v1/attribution/calculate/${conversionId}`,
    channels: (shortCode: string) => `${API_BASE_URL}/api/v1/attribution/channels/${shortCode}`,
    compare: (shortCode: string) => `${API_BASE_URL}/api/v1/attribution/compare/${shortCode}`,
    insights: (shortCode: string) => `${API_BASE_URL}/api/v1/attribution/insights/${shortCode}`,
  },

  // Real-time analytics endpoints
  realtime: {
    ws: `${API_BASE_URL}/api/v1/realtime/ws`,
    stats: `${API_BASE_URL}/api/v1/realtime/stats`,
  },

  // User analytics endpoints
  userAnalytics: {
    summary: `${API_BASE_URL}/api/v1/user-analytics/summary`,
    engagement: `${API_BASE_URL}/api/v1/user-analytics/engagement`,
    dashboard: (period: string) => `${API_BASE_URL}/api/v1/user-analytics/dashboard/${period}`,
    activityLog: `${API_BASE_URL}/api/v1/user-analytics/activity-log`,
    logActivity: `${API_BASE_URL}/api/v1/user-analytics/log-activity`,
    startSession: `${API_BASE_URL}/api/v1/user-analytics/start-session`,
    endSession: `${API_BASE_URL}/api/v1/user-analytics/end-session`,
  },

  // API Key management endpoints
  apiKeys: {
    create: `${API_BASE_URL}/api/v1/api-keys`,
    list: `${API_BASE_URL}/api/v1/api-keys`,
    get: (id: string) => `${API_BASE_URL}/api/v1/api-keys/${id}`,
    revoke: (id: string) => `${API_BASE_URL}/api/v1/api-keys/${id}/revoke`,
    delete: (id: string) => `${API_BASE_URL}/api/v1/api-keys/${id}`,
    stats: (id: string) => `${API_BASE_URL}/api/v1/api-keys/${id}/stats`,
  },

  // Health check endpoints
  health: {
    check: `${API_BASE_URL}/health`,
    dbPool: `${API_BASE_URL}/health/db-pool`,
    goroutines: `${API_BASE_URL}/health/goroutines`,
  },
} as const;

/**
 * Helper function to build API URLs with query parameters
 * @param baseUrl - The base URL
 * @param params - Query parameters as key-value pairs
 * @returns Complete URL with query parameters
 */
export function buildApiUrl(baseUrl: string, params?: Record<string, any>): string {
  if (!params || Object.keys(params).length === 0) {
    return baseUrl;
  }

  const queryString = Object.entries(params)
    .filter(([_, value]) => value !== undefined && value !== null)
    .map(([key, value]) => `${encodeURIComponent(key)}=${encodeURIComponent(String(value))}`)
    .join('&');

  return queryString ? `${baseUrl}?${queryString}` : baseUrl;
}

/**
 * Get WebSocket URL from HTTP URL
 * @param httpUrl - HTTP/HTTPS URL
 * @returns WebSocket URL (ws:// or wss://)
 */
export function getWebSocketUrl(httpUrl: string): string {
  return httpUrl.replace(/^http/, 'ws');
}

export default API_ENDPOINTS;
