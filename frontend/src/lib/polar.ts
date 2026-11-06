/**
 * Polar Integration Library
 *
 * Client-side utilities for interacting with Polar billing system through the backend API.
 * Polar (polar.sh) is used for subscription management and billing.
 */

import { API_ENDPOINTS, buildApiUrl } from './api-config';

/**
 * Polar Product (Plan) - Matches backend PolarProduct model
 */
export interface PolarProduct {
  id: string;
  polar_product_id: string;
  name: string;
  description: string;
  plan_type: 'free' | 'premium' | 'growth' | 'enterprise';
  is_recurring: boolean;
  prices: PolarPrice[];
  features: string[];
  benefits: PolarBenefit[];
  organization_id: string;
  is_archived: boolean;
  created_at: string;
  updated_at: string;
  last_synced_at: string;
}

/**
 * Polar Price - Price point for a product
 */
export interface PolarPrice {
  id: string;
  price_amount: number; // Amount in cents
  price_currency: string; // USD, EUR, etc.
  recurring_interval?: 'month' | 'year';
  is_archived: boolean;
}

/**
 * Polar Benefit - Benefit granted to a subscription
 */
export interface PolarBenefit {
  id: string;
  type: string; // custom, articles, discord, github_repository, etc.
  description: string;
  properties?: Record<string, any>;
}

/**
 * Polar Subscription - Matches backend PolarSubscription model
 */
export interface PolarSubscription {
  id: number;
  user_id: number;
  polar_customer_id: string;
  polar_subscription_id?: string;
  polar_product_id: string;
  polar_price_id?: string;
  polar_checkout_id?: string;
  status: 'active' | 'cancelled' | 'incomplete' | 'incomplete_expired' | 'trialing' | 'past_due';
  plan_type: 'free' | 'premium' | 'growth' | 'enterprise';
  billing_cycle?: 'month' | 'year';
  current_period_start?: string;
  current_period_end?: string;
  cancel_at?: string;
  cancel_at_period_end: boolean;
  cancelled_at?: string;
  ended_at?: string;
  trial_start?: string;
  trial_end?: string;
  benefits: PolarBenefit[];
  metadata?: Record<string, any>;
  created_at: string;
  updated_at: string;
}

/**
 * Usage information for the current billing period
 */
export interface UsageInfo {
  current_usage: number;
  limit: number;
  percentage: number;
  period_start: string;
  period_end: string;
}

/**
 * Checkout session response
 */
export interface CheckoutSessionResponse {
  checkout_url: string;
  checkout_id: string;
}

/**
 * Fetches available Polar products/plans from the backend
 * Uses public endpoint - no authentication required
 * @returns Promise with array of Polar products
 */
export async function getProducts(): Promise<PolarProduct[]> {
  try {
    // Use public endpoint - no authentication required
    const response = await fetch(API_ENDPOINTS.billing.publicProducts, {
      method: 'GET',
      headers: {
        'Content-Type': 'application/json',
      },
      // No credentials needed for public endpoint
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch products: ${response.statusText}`);
    }

    const data = await response.json();
    return data.products || [];
  } catch (error) {
    console.error('[Polar] Error fetching products:', error);
    throw error;
  }
}

/**
 * Fetches available pricing plans (alias for getProducts for backward compatibility)
 * @returns Promise with array of Polar products
 */
export async function getPlans(): Promise<PolarProduct[]> {
  return getProducts();
}

/**
 * Redirects to Polar checkout for a specific product
 *
 * @param productId - Polar product ID (required)
 * @param accessToken - User's JWT access token (required)
 * @param priceId - Optional specific price ID (for products with multiple prices)
 * @returns Promise with checkout URL
 * @throws Error if authentication fails or checkout creation fails
 *
 * @example
 * const token = getAccessToken();
 * await redirectToCheckout('prod_abc123', token);
 *
 * // With specific price
 * await redirectToCheckout('prod_abc123', token, 'price_xyz789');
 */
export async function redirectToCheckout(
  productId: string,
  accessToken: string,
  priceId?: string
): Promise<string> {
  try {
    // Validate token
    if (!accessToken || accessToken.trim() === '') {
      console.error('[Polar] redirectToCheckout called without access token');
      throw new Error('Authentication required. Please log in to continue.');
    }

    console.log('[Polar] Creating checkout session for product:', productId);

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${accessToken}`,
    };

    const response = await fetch(API_ENDPOINTS.billing.checkout, {
      method: 'POST',
      headers,
      credentials: 'include',
      body: JSON.stringify({
        product_id: productId,
        price_id: priceId,
      }),
    });

    if (!response.ok) {
      const errorData = await response.json();

      // Enhanced error handling for authentication failures
      if (response.status === 401) {
        console.error('[Polar] Authentication failed - token may be expired');
        throw new Error('Your session has expired. Please log in again.');
      }

      console.error('[Polar] Checkout failed:', {
        status: response.status,
        error: errorData
      });

      throw new Error(errorData.error || 'Failed to create checkout session');
    }

    const data: CheckoutSessionResponse = await response.json();

    console.log('[Polar] Checkout session created:', data.checkout_id);

    // Redirect to Polar checkout page
    if (data.checkout_url) {
      window.location.href = data.checkout_url;
    }

    return data.checkout_url;
  } catch (error) {
    console.error('[Polar] Error creating checkout session:', error);
    throw error;
  }
}

/**
 * Fetches the current user's subscription information
 * @param accessToken - User's JWT access token
 * @returns Promise with subscription data
 */
export async function getSubscription(accessToken: string): Promise<PolarSubscription | null> {
  try {
    const response = await fetch(API_ENDPOINTS.billing.subscription, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (response.status === 404) {
      // No subscription found
      return null;
    }

    if (!response.ok) {
      throw new Error(`Failed to fetch subscription: ${response.statusText}`);
    }

    const data = await response.json();
    return data.subscription || null;
  } catch (error) {
    console.error('[Polar] Error fetching subscription:', error);
    throw error;
  }
}

/**
 * Fetches current usage information for the billing period
 * @param accessToken - User's JWT access token
 * @returns Promise with usage data
 */
export async function getUsage(accessToken: string): Promise<UsageInfo> {
  try {
    const response = await fetch(API_ENDPOINTS.billing.usage, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to fetch usage: ${response.statusText}`);
    }

    const data = await response.json();
    return data.usage;
  } catch (error) {
    console.error('[Polar] Error fetching usage:', error);
    throw error;
  }
}

/**
 * Cancels the current subscription (will remain active until end of billing period)
 * @param accessToken - User's JWT access token
 * @returns Promise with updated subscription
 */
export async function cancelSubscription(accessToken: string): Promise<PolarSubscription> {
  try {
    const response = await fetch(API_ENDPOINTS.billing.cancel, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to cancel subscription');
    }

    const data = await response.json();
    return data.subscription;
  } catch (error) {
    console.error('[Polar] Error cancelling subscription:', error);
    throw error;
  }
}

/**
 * Reactivates a cancelled subscription (must be called before period end)
 * @param accessToken - User's JWT access token
 * @returns Promise with updated subscription
 */
export async function reactivateSubscription(accessToken: string): Promise<PolarSubscription> {
  try {
    const response = await fetch(API_ENDPOINTS.billing.reactivate, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to reactivate subscription');
    }

    const data = await response.json();
    return data.subscription;
  } catch (error) {
    console.error('[Polar] Error reactivating subscription:', error);
    throw error;
  }
}

/**
 * Gets the Polar customer portal URL for managing subscription and payment methods
 * @param accessToken - User's JWT access token
 * @returns Promise with portal URL
 */
export async function getCustomerPortalUrl(accessToken: string): Promise<string> {
  try {
    const response = await fetch(API_ENDPOINTS.billing.portal, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${accessToken}`,
        'Content-Type': 'application/json',
      },
      credentials: 'include',
    });

    if (!response.ok) {
      throw new Error(`Failed to get portal URL: ${response.statusText}`);
    }

    const data = await response.json();
    return data.portal_url;
  } catch (error) {
    console.error('[Polar] Error getting portal URL:', error);
    throw error;
  }
}

/**
 * Helper function to format price from cents to display format
 * @param priceInCents - Price amount in cents
 * @param currency - Currency code (default: 'USD')
 * @returns Formatted price string (e.g., "$9.99")
 */
export function formatPrice(priceInCents: number, currency: string = 'USD'): string {
  const amount = priceInCents / 100;

  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: currency,
  }).format(amount);
}

/**
 * Helper function to get the display name for billing cycle
 * @param interval - Billing interval ('month' or 'year')
 * @returns Display name (e.g., "Monthly" or "Annually")
 */
export function getBillingCycleName(interval?: string): string {
  switch (interval) {
    case 'month':
      return 'Monthly';
    case 'year':
      return 'Annually';
    default:
      return 'One-time';
  }
}

/**
 * Helper function to check if subscription is active
 * @param subscription - Polar subscription object
 * @returns true if subscription is active or trialing
 */
export function isSubscriptionActive(subscription: PolarSubscription | null): boolean {
  if (!subscription) return false;
  return subscription.status === 'active' || subscription.status === 'trialing';
}

/**
 * Helper function to check if subscription is cancelled but still active
 * @param subscription - Polar subscription object
 * @returns true if subscription is cancelled but hasn't ended yet
 */
export function isSubscriptionCancelling(subscription: PolarSubscription | null): boolean {
  if (!subscription) return false;
  return subscription.cancel_at_period_end && subscription.status === 'active';
}

export default {
  getProducts,
  getPlans,
  redirectToCheckout,
  getSubscription,
  getUsage,
  cancelSubscription,
  reactivateSubscription,
  getCustomerPortalUrl,
  formatPrice,
  getBillingCycleName,
  isSubscriptionActive,
  isSubscriptionCancelling,
};
