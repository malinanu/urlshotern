-- Migration 035: Sync Privacy and Terms Content (Down)
-- Revert content to simple summaries

UPDATE static_pages 
SET content = 'Your privacy is important to us. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our URL shortening service.'
WHERE slug = 'privacy';

UPDATE static_pages 
SET content = 'Please read these Terms of Service carefully before using our URL shortening service. By accessing or using our service, you agree to be bound by these terms.'
WHERE slug = 'terms';
