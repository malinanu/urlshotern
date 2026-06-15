'use client';

import { useEffect, useState } from 'react';
import Layout from '@/components/layout/Layout';
import { API_ENDPOINTS } from '@/lib/api-config';
import { toast } from 'react-toastify';

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
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    fetchPageData();
  }, []);

  const fetchPageData = async () => {
    try {
      setLoading(true);
      setError(null); // Clear previous errors

      const response = await fetch(API_ENDPOINTS.cms.privacy);

      if (!response.ok) {
        const errorMsg = `Failed to load page (${response.status})`;
        setError(errorMsg);
        setUseFallback(true);
        toast.error(errorMsg); // Show user-facing error
        return;
      }

      const data = await response.json();
      const content = data.page?.content || '';
      // Block JSON from the CMS builder is not renderable as HTML
      if (content.trim().startsWith('[') || content.trim().startsWith('{')) {
        setUseFallback(true);
        return;
      }
      setPageData(data.page);
    } catch (error) {
      const errorMsg = error instanceof Error
        ? `Network error: ${error.message}`
        : 'Failed to load page content';

      console.error('Error fetching page:', error);
      setError(errorMsg);
      setUseFallback(true);
      toast.error(errorMsg); // Show user-facing error
    } finally {
      setLoading(false);
    }
  };

  const fallbackContent = {
    title: "Privacy Policy",
    content: `
      <div class="space-y-8">
        <div class="text-center border-b border-border pb-8">
          <h1 class="text-4xl font-bold text-foreground mb-4">Privacy Policy</h1>
          <p class="text-lg text-muted-foreground max-w-2xl mx-auto">
            This Privacy Policy explains how Trunc collects, uses, shares, and protects
            personal data when you use our URL shortening service at trunc.site, and the
            rights you have over your data.
          </p>
          <p class="text-sm text-muted-foreground mt-4">
            <strong>Last updated:</strong> 16 June 2026
          </p>
        </div>

        <div class="space-y-10">
          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">1. Who We Are (Data Controller)</h2>
            <p class="text-muted-foreground mb-4">
              Trunc ("Trunc", "we", "us", "our") is a URL shortening service. The data
              controller responsible for your personal data is
              <strong>[LEGAL ENTITY NAME — to be completed]</strong>, registered in
              <strong>[country of incorporation, e.g. Sri Lanka]</strong>
              (registration no. <strong>[company number]</strong>), registered address
              <strong>[registered address]</strong>.
            </p>
            <p class="text-muted-foreground mb-4">
              You can contact us about privacy at
              <a href="mailto:hello@trunc.site" class="text-primary hover:underline">hello@trunc.site</a>.
              <strong>[If a Data Protection Officer is appointed, add their contact here.]</strong>
            </p>
            <p class="text-muted-foreground">
              <strong>EU/EEA representative (GDPR Art. 27):</strong> [name &amp; EU address — to be appointed].
              <strong>UK representative (UK GDPR):</strong> [name &amp; UK address — to be appointed].
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">2. Personal Data We Collect</h2>

            <h3 class="text-xl font-medium text-foreground mb-3">2.1 Information you provide</h3>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground mb-6">
              <li>Account data: your name, email address, and a securely hashed password.</li>
              <li>Optional phone number, used for one-time-passcode (OTP) verification.</li>
              <li>If you sign in with Google or GitHub: your name, email, and avatar from that provider.</li>
              <li>The URLs you shorten, custom aliases, and link settings.</li>
              <li>Contact-form and support messages.</li>
              <li>Billing details, which are processed by our payment provider (Polar) — we do not store full card numbers.</li>
              <li>Your communication preferences, including whether you opt in to marketing emails.</li>
            </ul>

            <h3 class="text-xl font-medium text-foreground mb-3">2.2 Information collected automatically</h3>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground mb-6">
              <li>Device and usage data: browser type, operating system, and how you interact with the service.</li>
              <li>Authentication/session data, including IP address and user-agent tied to your sessions, for security.</li>
              <li>Audit logs of key account actions (with IP and user-agent) for security and fraud prevention.</li>
              <li>With your consent, analytics and session-insight data via the tools listed in our <a href="/cookies" class="text-primary hover:underline">Cookie Policy</a>.</li>
            </ul>

            <h3 class="text-xl font-medium text-foreground mb-3">2.3 Data about people who click short links</h3>
            <p class="text-muted-foreground mb-3">
              When anyone clicks a link shortened with Trunc, we log limited data to produce
              click analytics for the link owner and to detect bots and abuse:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground">
              <li>IP address, user-agent, and referring page;</li>
              <li>approximate location (country) derived from the IP address;</li>
              <li>a one-way hash of the IP (with a daily-rotating salt) used to count unique visitors without retaining the raw IP.</li>
            </ul>
            <p class="text-muted-foreground mt-3">
              Raw IP addresses for link clicks are deleted (set to null) within 24 hours
              — see Section 8 for full retention periods.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">3. How We Use Your Data &amp; Legal Bases</h2>
            <p class="text-muted-foreground mb-4">
              Where the GDPR / UK GDPR applies, we rely on the following legal bases (Art. 6):
            </p>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground">
              <li><strong>Performance of a contract:</strong> creating and operating your account, shortening links, generating your analytics, processing payments.</li>
              <li><strong>Legitimate interests:</strong> securing the service, preventing fraud/abuse and bot traffic, maintaining audit logs, and basic service improvement — balanced against your rights.</li>
              <li><strong>Consent:</strong> non-essential analytics/session cookies and marketing communications. You may withdraw consent at any time.</li>
              <li><strong>Legal obligation:</strong> complying with law, tax/accounting rules, and lawful requests.</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">4. How We Share Data (Processors &amp; Sub-processors)</h2>
            <p class="text-muted-foreground mb-2"><strong>We do not sell your personal information.</strong></p>
            <p class="text-muted-foreground mb-4">
              We share data with the service providers below, who process it on our behalf
              under contract. Categories and providers:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground mb-4">
              <li><strong>Hosting &amp; database:</strong> Oracle Cloud Infrastructure (region: [confirm region]).</li>
              <li><strong>Payments &amp; subscriptions:</strong> Polar (polar.sh) — your email and subscription data.</li>
              <li><strong>Analytics &amp; session insights (consent-based):</strong> Google Analytics, Ahrefs Analytics, and ContentSquare. ContentSquare may record interactions with pages (e.g. clicks, scrolls, navigation) to help us improve usability; sensitive fields such as passwords are masked.</li>
              <li><strong>Email delivery:</strong> our SMTP email provider — for verification, account, and (where you opted in) marketing emails.</li>
              <li><strong>SMS / WhatsApp (OTP &amp; notifications):</strong> text.lk and Twilio — your phone number and message content.</li>
              <li><strong>AI insights:</strong> OpenRouter — receives <em>aggregated</em> link analytics to generate performance summaries.</li>
              <li><strong>Sign-in providers (if you use them):</strong> Google and GitHub OAuth.</li>
            </ul>
            <p class="text-muted-foreground">
              We may also disclose data where required by law or lawful request, to protect
              the rights, safety, or property of Trunc, our users, or the public, and in
              connection with a merger, acquisition, or sale of assets (with notice where required).
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">5. International Data Transfers</h2>
            <p class="text-muted-foreground mb-4">
              We and our providers operate outside the European Economic Area (EEA) and the UK
              — including in <strong>[country of establishment, e.g. Sri Lanka]</strong> and other
              countries where our sub-processors are located. When we transfer personal data of
              EEA/UK individuals abroad, we use appropriate safeguards such as the European
              Commission's Standard Contractual Clauses (SCCs) and the UK International Data
              Transfer Addendum (IDTA).
            </p>
            <p class="text-muted-foreground">
              You can request details of these safeguards by contacting
              <a href="mailto:hello@trunc.site" class="text-primary hover:underline">hello@trunc.site</a>.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">6. Cookies &amp; Tracking</h2>
            <p class="text-muted-foreground">
              We use strictly necessary cookies to run the site (login session, theme) and,
              <strong>only with your consent</strong>, analytics/session cookies. Non-essential
              cookies are off until you opt in, and you can change or withdraw your choice at any
              time via the "Cookie Preferences" link in the footer. For the full list of cookies
              and providers, see our <a href="/cookies" class="text-primary hover:underline">Cookie Policy</a>.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">7. Data Security</h2>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground">
              <li>TLS/HTTPS encryption in transit and passwords hashed with bcrypt.</li>
              <li>Access controls, authentication, and security monitoring.</li>
              <li>Backups and disaster-recovery procedures.</li>
            </ul>
            <div class="bg-primary/10 border border-primary/20 rounded-lg p-4 mt-4">
              <p class="text-primary">
                <strong>Note:</strong> No method of transmission or storage is 100% secure, so we
                cannot guarantee absolute security. We will notify you and the relevant authority
                of a personal-data breach where the law requires.
              </p>
            </div>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">8. Data Retention</h2>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground">
              <li><strong>Account data:</strong> kept while your account is active; deleted or anonymised after you close your account, except where we must keep records for legal/tax purposes.</li>
              <li><strong>Raw click IP addresses:</strong> removed within 24 hours.</li>
              <li><strong>Click analytics:</strong> retained for 30 days (Free plans), 1 year (Premium plans), or for the life of the account (Enterprise), then deleted.</li>
              <li><strong>Verification/OTP &amp; password-reset tokens:</strong> short-lived (minutes) and then invalid.</li>
              <li><strong>Audit and security logs:</strong> kept for a limited period for security and legal compliance.</li>
            </ul>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">9. Your Rights</h2>
            <p class="text-muted-foreground mb-4">
              Subject to your location and applicable law (including the GDPR and UK GDPR), you may have the right to:
            </p>
            <ul class="list-disc pl-6 space-y-2 text-muted-foreground mb-4">
              <li>access a copy of your personal data;</li>
              <li>rectify inaccurate data;</li>
              <li>erase your data ("right to be forgotten");</li>
              <li>restrict or object to certain processing;</li>
              <li>data portability;</li>
              <li>withdraw consent at any time (without affecting prior processing);</li>
              <li>lodge a complaint with a supervisory authority.</li>
            </ul>
            <p class="text-muted-foreground mb-4">
              To exercise your rights, email
              <a href="mailto:hello@trunc.site" class="text-primary hover:underline">hello@trunc.site</a>.
              We respond within one month (extendable where the law allows). You can delete your
              account at any time from your dashboard, which removes your links and associated analytics.
            </p>
            <p class="text-muted-foreground">
              <strong>Supervisory authorities:</strong> in the UK, the Information Commissioner's
              Office (ico.org.uk); in the EEA, your local data protection authority; in Sri Lanka,
              the Data Protection Authority established under the Personal Data Protection Act, No. 9 of 2022.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">10. California Privacy Rights (CCPA/CPRA)</h2>
            <p class="text-muted-foreground mb-4">
              If you are a California resident, you have the right to know, delete, and correct your
              personal information, and to opt out of any "sale" or "sharing" of it. In the past 12
              months we have collected the categories described in Section 2 (identifiers, internet/
              network activity, approximate geolocation, and commercial/account information), for the
              purposes in Section 3.
            </p>
            <p class="text-muted-foreground mb-4">
              <strong>We do not sell your personal information</strong> for money. We use analytics/
              session tools (Section 4) only with your consent; to the extent any such use is treated
              as "sharing" for cross-context behavioral advertising under California law, you can opt
              out by rejecting analytics cookies via "Cookie Preferences" in the footer, or by emailing
              <a href="mailto:hello@trunc.site" class="text-primary hover:underline">hello@trunc.site</a>.
              We will not discriminate against you for exercising your rights. You may use an authorized agent.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">11. Children's Privacy</h2>
            <p class="text-muted-foreground">
              Trunc is not directed to children. We do not knowingly collect personal data from
              children under 16 in the EEA/UK or under 13 in the United States. If you believe a
              child has provided us data, contact us and we will delete it.
            </p>
          </section>

          <section>
            <h2 class="text-2xl font-semibold text-foreground mb-4">12. Changes to This Policy</h2>
            <p class="text-muted-foreground">
              We may update this Privacy Policy from time to time. We will post the updated version
              here and, for material changes, notify you by email or an in-service notice. The
              "Last updated" date above reflects the latest version.
            </p>
          </section>

          <section class="bg-muted p-6 rounded-lg">
            <h2 class="text-2xl font-semibold text-foreground mb-4">13. Contact Us</h2>
            <p class="text-muted-foreground mb-4">
              Questions or requests about this Policy or your data:
            </p>
            <div class="space-y-2 text-muted-foreground">
              <p><strong>Email:</strong> <a href="mailto:hello@trunc.site" class="text-primary hover:underline">hello@trunc.site</a></p>
              <p><strong>Entity:</strong> [LEGAL ENTITY NAME — to be completed]</p>
              <p><strong>Registered address:</strong> [registered address — to be completed]</p>
            </div>
          </section>
        </div>
      </div>
    `
  };

  if (loading) {
    return (
      <Layout>
        <div className="min-h-[60vh] flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  const displayContent = useFallback ? fallbackContent : pageData;

  return (
    <Layout>
      <div className="min-h-screen py-12">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          {error && (
            <div className="mb-6 p-4 bg-destructive/10 border border-destructive/20 rounded-lg">
              <p className="text-destructive mb-2">{error}</p>
              <button
                onClick={fetchPageData}
                className="text-sm text-destructive hover:text-destructive/80 underline"
              >
                Retry
              </button>
            </div>
          )}
          <div
            className="prose prose-lg max-w-none"
            dangerouslySetInnerHTML={{ __html: displayContent?.content || '' }}
          />
        </div>
      </div>
    </Layout>
  );
}