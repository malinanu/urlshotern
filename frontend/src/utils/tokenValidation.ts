/**
 * Token Validation Utilities
 *
 * Centralized token validation with proper error handling for JWT tokens.
 * Addresses issues #35, #36, #37 related to authentication token handling.
 *
 * ## Key Functions:
 *
 * **Token Expiry Checking (Issue #37 - Performance Optimization):**
 * - `shouldRefreshToken()` - **PREFERRED**: Uses stored timestamp from backend's `expires_in`
 * - `getTokenRemainingTime()` - Get remaining seconds using stored timestamp
 *
 * **Token Validation (Issue #35 - Structure & Error Handling):**
 * - `validateToken()` - Comprehensive validation including structure, expiry, and payload
 * - `storeAuthTokens()` - Validates and stores tokens with calculated expiry timestamp
 *
 * **Low-Level Utilities:**
 * - `isValidJWTStructure()` - Checks JWT format (3 dot-separated base64url parts)
 * - `decodeJWTPayload()` - Safely decodes JWT payload with error handling
 */

// Constants
export const TOKEN_EXPIRY_BUFFER_SECONDS = 300; // 5 minutes
export const TOKEN_PARTS_COUNT = 3; // JWT structure: header.payload.signature

// LocalStorage keys
const STORAGE_KEYS = {
  ACCESS_TOKEN: 'access_token',
  REFRESH_TOKEN: 'refresh_token',
  USER_DATA: 'user_data',
  TOKEN_EXPIRY_AT: 'token_expiry_at',
} as const;

/**
 * JWT payload interface
 */
interface JWTPayload {
  exp: number;
  iat?: number;
  [key: string]: any;
}

/**
 * Token validation result
 */
export interface TokenValidationResult {
  isValid: boolean;
  error?: string;
  payload?: JWTPayload;
}

/**
 * Validates JWT structure (3 parts separated by dots)
 * @param token - JWT token string
 * @returns true if token has valid JWT structure
 */
export function isValidJWTStructure(token: string | null | undefined): boolean {
  if (!token || typeof token !== 'string') {
    return false;
  }

  const parts = token.split('.');
  if (parts.length !== TOKEN_PARTS_COUNT) {
    return false;
  }

  // Each part should be non-empty and base64url encoded
  return parts.every(part => {
    if (!part || part.length === 0) {
      return false;
    }
    // Base64url regex: alphanumeric, dash, underscore
    return /^[A-Za-z0-9_-]+$/.test(part);
  });
}

/**
 * Safely decodes JWT payload with error handling
 * @param token - JWT token string
 * @returns Decoded payload or null if decode fails
 */
export function decodeJWTPayload(token: string): JWTPayload | null {
  try {
    if (!isValidJWTStructure(token)) {
      return null;
    }

    const payload = token.split('.')[1];
    const decoded = JSON.parse(atob(payload));

    // Validate that payload has required exp field
    if (!decoded.exp || typeof decoded.exp !== 'number') {
      return null;
    }

    return decoded as JWTPayload;
  } catch (error) {
    console.error('[TokenValidation] Error decoding JWT payload:', error);
    return null;
  }
}

/**
 * Comprehensive token validation
 * @param token - JWT token string
 * @returns Validation result with details
 */
export function validateToken(token: string | null | undefined): TokenValidationResult {
  if (!token) {
    return {
      isValid: false,
      error: 'Token is null or undefined',
    };
  }

  if (!isValidJWTStructure(token)) {
    return {
      isValid: false,
      error: 'Invalid JWT structure',
    };
  }

  const payload = decodeJWTPayload(token);
  if (!payload) {
    return {
      isValid: false,
      error: 'Failed to decode JWT payload',
    };
  }

  const currentTime = Math.floor(Date.now() / 1000);
  if (payload.exp < currentTime) {
    return {
      isValid: false,
      error: 'Token has expired',
      payload,
    };
  }

  return {
    isValid: true,
    payload,
  };
}

/**
 * Validates and stores authentication tokens in localStorage
 * @param accessToken - Access token
 * @param refreshToken - Refresh token
 * @param expiresIn - Token expiry time in seconds (from backend)
 * @throws Error if tokens are invalid
 */
export function storeAuthTokens(
  accessToken: string,
  refreshToken: string,
  expiresIn: number
): void {
  // Validate access token structure
  const accessValidation = validateToken(accessToken);
  if (!accessValidation.isValid) {
    throw new Error(`Invalid access token: ${accessValidation.error}`);
  }

  // Validate refresh token structure
  if (!isValidJWTStructure(refreshToken)) {
    throw new Error('Invalid refresh token structure');
  }

  // Calculate and store expiry timestamp using backend's expires_in
  // This is more accurate than decoding JWT (addresses issue #37)
  const expiryAt = Date.now() + (expiresIn * 1000);

  try {
    localStorage.setItem(STORAGE_KEYS.ACCESS_TOKEN, accessToken);
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
    localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRY_AT, expiryAt.toString());

    // Also store token in a cookie so nginx auth_request can forward it
    // to the backend for server-side route protection (e.g. /admin).
    const maxAge = Math.floor(expiresIn);
    document.cookie = `access_token=${accessToken}; Max-Age=${maxAge}; Path=/; SameSite=Strict; Secure`;
  } catch (error) {
    console.error('[TokenValidation] Error storing tokens:', error);
    throw new Error('Failed to store tokens in localStorage');
  }
}

/**
 * Checks if token should be refreshed based on stored expiry timestamp.
 *
 * **PREFERRED METHOD** for checking token expiry. Uses the backend's `expires_in` field
 * stored in localStorage, avoiding JWT decoding and providing better performance and accuracy.
 *
 * Benefits:
 * - No JWT decoding overhead
 * - More accurate (uses server time, avoids clock skew)
 * - 5-minute buffer prevents edge-case expiry issues
 *
 * @returns true if token should be refreshed (expires within buffer time)
 */
export function shouldRefreshToken(): boolean {
  try {
    const expiryAtStr = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY_AT);
    if (!expiryAtStr) {
      return true; // No expiry stored, should refresh
    }

    const expiryAt = parseInt(expiryAtStr, 10);
    if (isNaN(expiryAt)) {
      return true; // Invalid expiry, should refresh
    }

    const currentTime = Date.now();
    const bufferMs = TOKEN_EXPIRY_BUFFER_SECONDS * 1000;

    // Refresh if expiring within buffer time
    return expiryAt < (currentTime + bufferMs);
  } catch (error) {
    console.error('[TokenValidation] Error checking token expiry:', error);
    return true; // On error, assume should refresh
  }
}

/**
 * Gets remaining time until token expiry in seconds
 * @returns Remaining seconds or 0 if expired/invalid
 */
export function getTokenRemainingTime(): number {
  try {
    const expiryAtStr = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY_AT);
    if (!expiryAtStr) {
      return 0;
    }

    const expiryAt = parseInt(expiryAtStr, 10);
    if (isNaN(expiryAt)) {
      return 0;
    }

    const currentTime = Date.now();
    const remainingMs = expiryAt - currentTime;

    return Math.max(0, Math.floor(remainingMs / 1000));
  } catch (error) {
    console.error('[TokenValidation] Error getting remaining time:', error);
    return 0;
  }
}

/**
 * Clears all authentication data from localStorage
 */
export function clearAuthTokens(): void {
  try {
    localStorage.removeItem(STORAGE_KEYS.ACCESS_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
    localStorage.removeItem(STORAGE_KEYS.USER_DATA);
    localStorage.removeItem(STORAGE_KEYS.TOKEN_EXPIRY_AT);

    // Expire the auth cookie set for nginx auth_request
    document.cookie = 'access_token=; Max-Age=0; Path=/; SameSite=Strict; Secure';
  } catch (error) {
    console.error('[TokenValidation] Error clearing auth tokens:', error);
  }
}

/**
 * Gets stored access token
 * @returns Access token or null
 */
export function getAccessToken(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEYS.ACCESS_TOKEN);
  } catch (error) {
    console.error('[TokenValidation] Error getting access token:', error);
    return null;
  }
}

/**
 * Gets stored refresh token
 * @returns Refresh token or null
 */
export function getRefreshToken(): string | null {
  try {
    return localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
  } catch (error) {
    console.error('[TokenValidation] Error getting refresh token:', error);
    return null;
  }
}
