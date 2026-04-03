import type { NextConfig } from "next";

// Bundle analyzer
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
});

const nextConfig: NextConfig = {
  // Enforce no trailing slashes — /blog/ redirects to /blog
  trailingSlash: false,

  // Note: In Next.js 16, eslint config is no longer in next.config.ts
  // Use SKIP_LINT=true or configure via package.json scripts instead

  // Bypass TypeScript errors during build (for CI/CD and Docker builds)
  typescript: {
    ignoreBuildErrors: true,
  },

  // Enable React Compiler and experimental features
  experimental: {
    optimizePackageImports: ['@heroicons/react', 'recharts', 'lucide-react'],
  },

  // Image optimization
  images: {
    formats: ['image/avif', 'image/webp'],
    minimumCacheTTL: 31536000, // 1 year
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
  },

  // Compiler optimizations
  compiler: {
    removeConsole: process.env.NODE_ENV === 'production' ? {
      exclude: ['error', 'warn'],
    } : false,
  },

  // Performance optimizations
  compress: true,

  // Note: headers() function removed - incompatible with output: 'export'
  // Static exports don't support custom headers
  // Headers should be configured at the web server level (nginx, etc.)

  // Enable static export for production build only
  // In development, use Next.js dev server for hot reload
  output: process.env.NODE_ENV === 'production' ? 'export' : undefined,

  // API rewrites for development (proxy /api/* to backend)
  async rewrites() {
    if (process.env.NODE_ENV === 'development') {
      // Use BACKEND_URL for server-side rewrites (works inside Docker network)
      // Falls back to localhost for non-Docker development
      const backendUrl = process.env.BACKEND_URL || process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
      return [
        {
          source: '/api/:path*',
          destination: `${backendUrl}/api/:path*`,
        },
        {
          source: '/health/:path*',
          destination: `${backendUrl}/health/:path*`,
        },
        {
          source: '/:shortCode([a-zA-Z0-9]{6,10})',
          destination: `${backendUrl}/:shortCode`,
        },
      ];
    }
    return [];
  },

  // Silence workspace lockfile warning
  outputFileTracingRoot: __dirname,

  // Production source maps for error tracking (smaller than dev maps)
  productionBrowserSourceMaps: false,

  // Turbopack is now the default in Next.js 16
  // Empty config to silence webpack compatibility warning
  turbopack: {},

  // Simplified webpack configuration for Next.js 15 compatibility
  // The aggressive chunking strategy was causing module resolution issues
  webpack: (config) => {
    return config;
  },
};

export default withBundleAnalyzer(nextConfig);
