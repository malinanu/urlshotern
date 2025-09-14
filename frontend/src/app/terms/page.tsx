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

export default function TermsPage() {
  const [pageData, setPageData] = useState<PageData | null>(null);
  const [loading, setLoading] = useState(true);
  const [useFallback, setUseFallback] = useState(false);

  useEffect(() => {
    fetchPageData();
  }, []);

  const fetchPageData = async () => {
    try {
      setLoading(true);
      const response = await fetch('http://localhost:8080/api/v1/pages/terms');
      
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
    title: "Terms of Service",
    content: `
      <div class="space-y-8">
        <div class="text-center border-b border-gray-200 pb-8">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Terms of Service</h1>
          <p class="text-lg text-gray-600 max-w-2xl mx-auto">
            These Terms of Service govern your use of Trunc's URL shortener service. 
            By using our service, you agree to these terms.
          </p>
          <p class="text-sm text-gray-500 mt-4">
            <strong>Last updated:</strong> September 4, 2025
          </p>
        </div>

        <div class="space-y-10">
          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">1. Acceptance of Terms</h2>
            <p class="text-gray-700 mb-4">
              By accessing and using Trunc's URL shortening service ("Service"), you accept and agree to 
              be bound by the terms and provision of this agreement. If you do not agree to abide by the 
              above, please do not use this service.
            </p>
            <p class="text-gray-700">
              These Terms of Service, along with our Privacy Policy, constitute the entire agreement 
              between you and 3logiq Technologies, Inc. ("3logiq", "we", "us", "our"), the developer of Trunc.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">2. Description of Service</h2>
            <p class="text-gray-700 mb-4">
              Trunc provides a URL shortening service that allows users to create shortened versions of 
              long URLs, track link analytics, and manage their shortened links through our platform.
            </p>
            <h3 class="text-xl font-medium text-gray-800 mb-3">Our Service includes:</h3>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>URL shortening and custom alias creation</li>
              <li>Link analytics and click tracking</li>
              <li>Custom domains and branding options</li>
              <li>API access for developers</li>
              <li>Account management and dashboard</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">3. User Accounts and Registration</h2>
            <p class="text-gray-700 mb-4">
              To access certain features of our Service, you may be required to create an account. 
              You agree to provide accurate, current, and complete information during the registration process.
            </p>
            <h3 class="text-xl font-medium text-gray-800 mb-3">Account Responsibilities:</h3>
            <ul class="list-disc pl-6 space-y-2 text-gray-700">
              <li>You are responsible for safeguarding your account credentials</li>
              <li>You must notify us immediately of any unauthorized use of your account</li>
              <li>You are responsible for all activities that occur under your account</li>
              <li>You must keep your account information up to date</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">4. Acceptable Use Policy</h2>
            <p class="text-gray-700 mb-4">You agree not to use our Service to:</p>
            
            <div class="bg-red-50 border border-red-200 rounded-lg p-6 mb-6">
              <h3 class="text-lg font-semibold text-red-800 mb-3">Prohibited Activities</h3>
              <ul class="list-disc pl-6 space-y-2 text-red-700">
                <li>Distribute malware, viruses, or other harmful code</li>
                <li>Engage in phishing, spam, or fraudulent activities</li>
                <li>Share content that is illegal, hateful, or violates others' rights</li>
                <li>Circumvent or interfere with security features</li>
                <li>Use automated means to access the service excessively</li>
                <li>Impersonate others or provide false information</li>
                <li>Violate any applicable laws or regulations</li>
              </ul>
            </div>

            <p class="text-gray-700">
              We reserve the right to suspend or terminate accounts that violate this Acceptable Use Policy 
              or any other provision of these Terms.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">5. Service Availability and Modifications</h2>
            <p class="text-gray-700 mb-4">
              While we strive to maintain high availability, we cannot guarantee uninterrupted access 
              to our Service. We may modify, suspend, or discontinue any part of our Service at any time.
            </p>
            <div class="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <p class="text-blue-800">
                <strong>Service Level:</strong> We target 99.9% uptime but are not liable for any downtime 
                or service interruptions. We will provide reasonable notice for planned maintenance when possible.
              </p>
            </div>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">6. Intellectual Property Rights</h2>
            <h3 class="text-xl font-medium text-gray-800 mb-3">Our Rights</h3>
            <p class="text-gray-700 mb-4">
              All content, features, and functionality of our Service are owned by 3logiq Technologies, Inc. and are 
              protected by international copyright, trademark, and other intellectual property laws.
            </p>
            
            <h3 class="text-xl font-medium text-gray-800 mb-3">Your Rights</h3>
            <p class="text-gray-700 mb-4">
              You retain ownership of any content you submit to our Service. By using our Service, 
              you grant us a limited license to use, store, and display your content as necessary 
              to provide our Service.
            </p>
            
            <h3 class="text-xl font-medium text-gray-800 mb-3">License to Use</h3>
            <p class="text-gray-700">
              We grant you a limited, non-exclusive, non-transferable license to use our Service 
              in accordance with these Terms.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">7. Payment Terms and Billing</h2>
            <p class="text-gray-700 mb-4">
              Certain features of our Service require payment. By subscribing to a paid plan, you agree to pay all charges associated with your selected plan.
            </p>
            
            <div class="grid md:grid-cols-2 gap-6 mb-6">
              <div class="bg-green-50 p-4 rounded-lg">
                <h4 class="font-semibold text-green-800 mb-2">Payment Terms</h4>
                <ul class="text-sm text-green-700 space-y-1">
                  <li>• Bills are due upon receipt</li>
                  <li>• Subscription fees are non-refundable</li>
                  <li>• Auto-renewal unless cancelled</li>
                  <li>• Price changes with 30-day notice</li>
                </ul>
              </div>
              <div class="bg-orange-50 p-4 rounded-lg">
                <h4 class="font-semibold text-orange-800 mb-2">Cancellation</h4>
                <ul class="text-sm text-orange-700 space-y-1">
                  <li>• Cancel anytime from your dashboard</li>
                  <li>• Service continues until period end</li>
                  <li>• No partial refunds for unused time</li>
                  <li>• Data export available before deletion</li>
                </ul>
              </div>
            </div>

            <p class="text-gray-700">
              We reserve the right to suspend accounts with overdue payments and delete data from 
              accounts that remain unpaid for more than 60 days.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">8. Data and Privacy</h2>
            <p class="text-gray-700 mb-4">
              Your privacy is important to us. Please review our Privacy Policy to understand 
              how we collect, use, and protect your information.
            </p>
            
            <h3 class="text-xl font-medium text-gray-800 mb-3">Data Ownership and Portability</h3>
            <ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4">
              <li>You own your data and can export it at any time</li>
              <li>We may use aggregated, anonymized data for service improvement</li>
              <li>You can delete your data by closing your account</li>
              <li>We comply with applicable data protection laws (GDPR, CCPA, etc.)</li>
            </ul>

            <p class="text-gray-700">
              By using our Service, you consent to our data practices as outlined in our Privacy Policy.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">9. Disclaimers and Limitations</h2>
            
            <div class="bg-yellow-50 border border-yellow-200 rounded-lg p-6 mb-6">
              <h3 class="text-lg font-semibold text-yellow-800 mb-3">Service Disclaimers</h3>
              <p class="text-yellow-700 text-sm">
                OUR SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND. 
                WE DISCLAIM ALL WARRANTIES, EXPRESS OR IMPLIED, INCLUDING MERCHANTABILITY, 
                FITNESS FOR A PARTICULAR PURPOSE, AND NON-INFRINGEMENT.
              </p>
            </div>

            <h3 class="text-xl font-medium text-gray-800 mb-3">Limitation of Liability</h3>
            <p class="text-gray-700 mb-4">
              TO THE MAXIMUM EXTENT PERMITTED BY LAW, 3LOGIQ SHALL NOT BE LIABLE FOR ANY INDIRECT, 
              INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES ARISING FROM YOUR USE OF THE SERVICE.
            </p>
            
            <p class="text-gray-700">
              Our total liability to you for any claims related to the Service shall not exceed 
              the amount you paid us in the twelve (12) months preceding the claim.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">10. Indemnification</h2>
            <p class="text-gray-700">
              You agree to indemnify and hold harmless 3logiq Technologies, Inc., its officers, directors, employees, 
              and agents from any claims, damages, or expenses arising from your use of the Service, 
              violation of these Terms, or infringement of any third-party rights.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">11. Termination</h2>
            <p class="text-gray-700 mb-4">
              Either party may terminate this agreement at any time. Upon termination:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4">
              <li>Your access to the Service will be discontinued</li>
              <li>Your short links will continue to redirect for 30 days</li>
              <li>You may export your data within 30 days</li>
              <li>We may delete your account and data after 90 days</li>
            </ul>
            <p class="text-gray-700">
              Sections relating to intellectual property, disclaimers, and indemnification 
              shall survive termination of this agreement.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">12. Governing Law and Jurisdiction</h2>
            <p class="text-gray-700 mb-4">
              These Terms shall be governed by and construed in accordance with the laws of the 
              State of California, United States, without regard to its conflict of law provisions.
            </p>
            <p class="text-gray-700">
              Any disputes arising from these Terms or your use of the Service shall be resolved 
              through binding arbitration in San Francisco, California, in accordance with the 
              Commercial Arbitration Rules of the American Arbitration Association.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">13. Changes to Terms</h2>
            <p class="text-gray-700 mb-4">
              We may update these Terms from time to time. We will notify you of material changes by:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4">
              <li>Posting the updated Terms on our website</li>
              <li>Sending an email notification to your registered email</li>
              <li>Displaying a notice in our Service</li>
            </ul>
            <p class="text-gray-700">
              Your continued use of the Service after changes become effective constitutes 
              acceptance of the updated Terms.
            </p>
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