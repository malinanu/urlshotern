'use client';

import { useEffect, useState } from 'react';
import Layout from '@/components/layout/Layout';

interface PageData {
  title: string;
  content: string;
  meta_description?: string;
  meta_keywords?: string;
  slug: string;
}

export default function PrivacyPage() {
  const [pageData, setPageData] = useState<PageData | null>(null);
  const [loading, setLoading] = useState(true);
  const [useFallback, setUseFallback] = useState(false);

  useEffect(() => {
    fetchPageData();
  }, []);

  const fetchPageData = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/api/v1/pages/privacy');
      
      if (response.ok) {
        const data = await response.json();
        setPageData(data.page);
      } else {
        setUseFallback(true);
      }
    } catch (error) {
      console.error('Error fetching page:', error);
      setUseFallback(true);
    } finally {
      setLoading(false);
    }
  };

  const fallbackContent = {
    title: "Privacy Policy",
    content: `
      <div class="space-y-8">
        <div class="text-center border-b border-gray-200 pb-8">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Privacy Policy</h1>
          <p class="text-lg text-gray-600 max-w-2xl mx-auto">
            Your privacy is important to us. This Privacy Policy explains how Trunc collects, uses, 
            and protects your information when you use our URL shortener service.
          </p>
          <p class="text-sm text-gray-500 mt-4">
            <strong>Last updated:</strong> September 4, 2025
          </p>
        </div>

        <div class="space-y-10">
          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">1. Information We Collect</h2>
            
            <h3 class="text-xl font-medium text-gray-800 mb-3">1.1 Information You Provide</h3>
            <ul class="list-disc pl-6 space-y-2 text-gray-700 mb-6">
              <li>Account information (name, email address, password)</li>
              <li>URLs you submit for shortening</li>
              <li>Custom aliases you create</li>
              <li>Contact form submissions and support communications</li>
              <li>Payment information (processed securely by our payment providers)</li>
            </ul>

            <h3 class="text-xl font-medium text-gray-800 mb-3">1.2 Information We Collect Automatically</h3>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>IP address and geolocation data</li>
              <li>Browser type and version</li>
              <li>Device information and operating system</li>
              <li>Referring website and pages visited</li>
              <li>Click data and analytics for your short links</li>
              <li>Usage patterns and feature interactions</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">2. How We Use Your Information</h2>
            <p class="text-gray-700 mb-4">We use the information we collect to:</p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>Provide and maintain our URL shortening service</li>
              <li>Generate analytics and insights for your links</li>
              <li>Prevent fraud and abuse of our service</li>
              <li>Communicate with you about your account and our services</li>
              <li>Improve our service and develop new features</li>
              <li>Comply with legal obligations and enforce our terms</li>
              <li>Process payments and manage billing</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">3. Information Sharing and Disclosure</h2>
            
            <h3 class="text-xl font-medium text-gray-800 mb-3">We do NOT sell your personal information.</h3>
            <p class="text-gray-700 mb-4">We may share your information in the following circumstances:</p>
            
            <ul class="list-disc pl-6 space-y-3 text-gray-700">
              <li><strong>Service Providers:</strong> With trusted third-party service providers who assist us in operating our service (cloud hosting, analytics, payment processing)</li>
              <li><strong>Legal Requirements:</strong> When required by law, court order, or government request</li>
              <li><strong>Safety and Security:</strong> To protect the rights, property, or safety of Trunc, our users, or the public</li>
              <li><strong>Business Transfers:</strong> In connection with a merger, acquisition, or sale of assets</li>
              <li><strong>With Consent:</strong> When you explicitly consent to sharing your information</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">4. Data Security</h2>
            <p class="text-gray-700 mb-4">We implement industry-standard security measures to protect your information:</p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>SSL/TLS encryption for data transmission</li>
              <li>Encrypted storage of sensitive information</li>
              <li>Regular security audits and monitoring</li>
              <li>Access controls and authentication systems</li>
              <li>Secure backup and disaster recovery procedures</li>
            </ul>
            <div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mt-4">
              <p class="text-blue-800">
                <strong>Note:</strong> While we implement strong security measures, no method of transmission 
                over the Internet is 100% secure. We cannot guarantee absolute security of your information.
              </p>
            </div>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">5. Your Privacy Rights</h2>
            <p class="text-gray-700 mb-4">Depending on your location, you may have the following rights:</p>
            
            <div class="grid md:grid-cols-2 gap-6">
              <div>
                <h4 class="font-semibold text-gray-800 mb-2">Access and Portability</h4>
                <p class="text-gray-600 text-sm">Request a copy of your personal information and data export</p>
              </div>
              <div>
                <h4 class="font-semibold text-gray-800 mb-2">Correction</h4>
                <p class="text-gray-600 text-sm">Update or correct inaccurate personal information</p>
              </div>
              <div>
                <h4 class="font-semibold text-gray-800 mb-2">Deletion</h4>
                <p class="text-gray-600 text-sm">Request deletion of your personal information</p>
              </div>
              <div>
                <h4 class="font-semibold text-gray-800 mb-2">Opt-out</h4>
                <p class="text-gray-600 text-sm">Opt out of certain data processing activities</p>
              </div>
            </div>

            <p class="text-gray-700 mt-6">
              To exercise these rights, please contact us at <a href="mailto:privacy@3logiq.com" class="text-blue-600 hover:underline">privacy@3logiq.com</a>. 
              We will respond to your request within 30 days.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">6. Cookies and Tracking</h2>
            <p class="text-gray-700 mb-4">We use cookies and similar technologies to:</p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4">
              <li>Maintain your login session</li>
              <li>Remember your preferences and settings</li>
              <li>Analyze usage patterns and improve our service</li>
              <li>Provide personalized content and features</li>
            </ul>
            <p class="text-gray-700">
              You can control cookie settings through your browser preferences. However, disabling 
              cookies may affect the functionality of our service.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">7. Data Retention</h2>
            <p class="text-gray-700 mb-4">We retain your information for as long as necessary to:</p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>Provide our services to you</li>
              <li>Comply with legal obligations</li>
              <li>Resolve disputes and enforce our agreements</li>
              <li>Maintain business records and analytics</li>
            </ul>
            <p class="text-gray-700 mt-4">
              When you delete your account, we will delete or anonymize your personal information within 90 days, 
              except where we are required to retain it for legal purposes.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">8. International Data Transfers</h2>
            <p class="text-gray-700 mb-4">
              Trunc is based in the United States. If you are accessing our service from outside the U.S., 
              your information may be transferred to and processed in the United States and other countries 
              where our service providers operate.
            </p>
            <p class="text-gray-700">
              We ensure appropriate safeguards are in place for international data transfers in compliance 
              with applicable privacy laws.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">9. Children's Privacy</h2>
            <p class="text-gray-700">
              Our service is not intended for children under 13 years of age. We do not knowingly collect 
              personal information from children under 13. If we become aware that we have collected personal 
              information from a child under 13, we will take steps to delete such information promptly.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">10. Changes to This Privacy Policy</h2>
            <p class="text-gray-700 mb-4">
              We may update this Privacy Policy from time to time. We will notify you of any material changes by:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>Posting the updated policy on our website</li>
              <li>Sending an email notification to your registered email address</li>
              <li>Displaying a notice in our service</li>
            </ul>
            <p class="text-gray-700 mt-4">
              Your continued use of our service after any changes indicates your acceptance of the updated Privacy Policy.
            </p>
          </section>

          <section class="bg-gray-50 p-6 rounded-lg">
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">11. Contact Us</h2>
            <p class="text-gray-700 mb-4">
              If you have any questions, concerns, or requests regarding this Privacy Policy or our privacy practices, 
              please contact us:
            </p>
            <div class="space-y-2 text-gray-700">
              <p><strong>Email:</strong> <a href="mailto:privacy@3logiq.com" class="text-blue-600 hover:underline">privacy@3logiq.com</a></p>
              <p><strong>Address:</strong> 3logiq Technologies, 123 Tech Boulevard, San Francisco, CA 94105</p>
              <p><strong>Phone:</strong> <a href="tel:+1-555-0123" class="text-blue-600 hover:underline">+1 (555) 012-3456</a></p>
            </div>
          </section>
        </div>
      </div>
    `
  };

  if (loading) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  const displayContent = useFallback ? fallbackContent : pageData;

  return (
    <Layout>
      <div className="min-h-screen bg-white py-12">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div 
            className="prose prose-lg max-w-none"
            dangerouslySetInnerHTML={{ __html: displayContent?.content || '' }}
          />
        </div>
      </div>
    </Layout>
  );
}