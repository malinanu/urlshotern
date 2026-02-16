-- Migration 035: Sync Privacy and Terms Content
-- Created: 2026-02-15
-- Purpose: Update content column with full HTML for Privacy and Terms pages to support dangerouslySetInnerHTML

-- Update Privacy Policy content
UPDATE static_pages 
SET content = '
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
            <h2 class="text-2xl font-semibold text-gray-900 mb-4">9. Children''s Privacy</h2>
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
    '
WHERE slug = 'privacy';

-- Update Terms of Service content
UPDATE static_pages 
SET content = '
      <div class="space-y-8">
        <div class="text-center border-b border-gray-200 pb-8">
          <h1 class="text-4xl font-bold text-gray-900 mb-4">Terms of Service</h1>
          <p class="text-lg text-gray-600 max-w-2xl mx-auto">
            These Terms of Service govern your use of Trunc''s URL shortener service. 
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
              By accessing and using Trunc''s URL shortening service ("Service"), you accept and agree to 
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
                <li>Share content that is illegal, hateful, or violates others'' rights</li>
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
    '
WHERE slug = 'terms';
