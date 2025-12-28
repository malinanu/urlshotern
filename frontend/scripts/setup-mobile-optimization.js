#!/usr/bin/env node

/**
 * Mobile Optimization Setup Script
 * Verifies installation and configuration
 */

const fs = require('fs');
const path = require('path');

console.log('🚀 Mobile Optimization Setup\n');

// Check dependencies
const packageJson = require('../package.json');
const requiredDeps = [
  'idb',
  'react-intersection-observer',
  '@tanstack/react-virtual',
  '@use-gesture/react',
];

console.log('📦 Checking dependencies...');
let missingDeps = [];

requiredDeps.forEach(dep => {
  if (!packageJson.dependencies[dep]) {
    missingDeps.push(dep);
    console.log(`  ❌ ${dep} - Missing`);
  } else {
    console.log(`  ✅ ${dep} - Installed`);
  }
});

if (missingDeps.length > 0) {
  console.log('\n⚠️  Missing dependencies. Run:');
  console.log(`npm install --legacy-peer-deps ${missingDeps.join(' ')}`);
  process.exit(1);
}

// Check required files
console.log('\n📁 Checking required files...');
const requiredFiles = [
  'src/utils/networkDetection.ts',
  'src/utils/offlineSync.ts',
  'src/utils/performanceMonitoring.ts',
  'src/hooks/useNetworkStatus.ts',
  'src/hooks/useIntersectionObserver.ts',
  'src/hooks/useTouchGestures.ts',
  'src/components/ResourceHints.tsx',
  'src/components/MobileAppShell.tsx',
  'src/components/mobile/VirtualScrollList.tsx',
  'src/components/mobile/MobileOptimizedAnalytics.tsx',
  'public/manifest.json',
];

let missingFiles = [];

requiredFiles.forEach(file => {
  const filePath = path.join(__dirname, '..', file);
  if (fs.existsSync(filePath)) {
    console.log(`  ✅ ${file}`);
  } else {
    console.log(`  ❌ ${file} - Missing`);
    missingFiles.push(file);
  }
});

if (missingFiles.length > 0) {
  console.log('\n⚠️  Some files are missing. Please ensure all files are created.');
  process.exit(1);
}

// Check service worker registration
console.log('\n🔧 Checking service worker registration...');
const layoutPath = path.join(__dirname, '..', 'src/app/layout.tsx');
if (fs.existsSync(layoutPath)) {
  const layoutContent = fs.readFileSync(layoutPath, 'utf8');
  if (layoutContent.includes('serviceWorker.register')) {
    console.log('  ✅ Service worker registration found');
  } else {
    console.log('  ⚠️  Service worker registration not found in layout.tsx');
    console.log('  Add service worker registration to layout.tsx');
  }
} else {
  console.log('  ❌ layout.tsx not found');
}

// Check manifest.json
console.log('\n📱 Checking PWA manifest...');
const manifestPath = path.join(__dirname, '..', 'public/manifest.json');
if (fs.existsSync(manifestPath)) {
  const manifest = JSON.parse(fs.readFileSync(manifestPath, 'utf8'));

  const checks = [
    { key: 'name', label: 'App name' },
    { key: 'short_name', label: 'Short name' },
    { key: 'start_url', label: 'Start URL' },
    { key: 'display', label: 'Display mode' },
    { key: 'icons', label: 'Icons' },
    { key: 'theme_color', label: 'Theme color' },
  ];

  checks.forEach(check => {
    if (manifest[check.key]) {
      console.log(`  ✅ ${check.label}: ${JSON.stringify(manifest[check.key]).substring(0, 50)}`);
    } else {
      console.log(`  ❌ ${check.label} - Missing`);
    }
  });
} else {
  console.log('  ❌ manifest.json not found');
}

// Performance recommendations
console.log('\n⚡ Performance Recommendations:');
console.log('  1. Build production bundle: npm run build');
console.log('  2. Test with Lighthouse: lighthouse http://localhost:3000 --view');
console.log('  3. Test offline: DevTools → Network → Offline');
console.log('  4. Test on real devices (iOS and Android)');
console.log('  5. Monitor Core Web Vitals');

// Next steps
console.log('\n✨ Next Steps:');
console.log('  1. Review MOBILE_OPTIMIZATION_GUIDE.md for implementation details');
console.log('  2. Review MOBILE_TESTING_CHECKLIST.md for testing procedures');
console.log('  3. Integrate MobileAppShell into your layout');
console.log('  4. Replace list components with VirtualScrollList');
console.log('  5. Add touch gestures to interactive elements');
console.log('  6. Run performance audits and optimize');

console.log('\n✅ Setup verification complete!\n');
