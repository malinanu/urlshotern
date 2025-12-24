import type { NextConfig } from "next";

// Bundle analyzer
const withBundleAnalyzer = require('@next/bundle-analyzer')({
  enabled: process.env.ANALYZE === 'true',
});

const nextConfig: NextConfig = {
  // Note: eslint and typescript configs moved to next.config.js or removed in Next.js 16
  // Use next lint --no-eslint and tsc --noEmit for type checking instead

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

  // Enable static export for production build
  output: 'export',

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
