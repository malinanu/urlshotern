import Link from "next/link";
import Layout from "@/components/layout/Layout";
import { CookiePreferencesButton } from "@/components/CookiePreferencesButton";

export default function CookiesPage() {
  return (
    <Layout>
      <div className="min-h-screen py-12">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="space-y-8">
            <div className="text-center border-b border-border pb-8">
              <h1 className="text-4xl font-bold text-foreground mb-4">Cookie Policy</h1>
              <p className="text-lg text-muted-foreground max-w-2xl mx-auto">
                This policy explains how Trunc uses cookies and similar technologies, and how
                you can control them. It supplements our{" "}
                <Link href="/privacy" className="text-primary hover:underline">
                  Privacy Policy
                </Link>
                .
              </p>
              <p className="text-sm text-muted-foreground mt-4">
                <strong>Last updated:</strong> 16 June 2026
              </p>
            </div>

            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-foreground">1. What are cookies?</h2>
              <p className="text-muted-foreground">
                Cookies are small files stored on your device. We also use similar technologies
                such as browser <code>localStorage</code>. We group these into{" "}
                <strong>strictly necessary</strong> (always on) and{" "}
                <strong>analytics</strong> (used only with your consent).
              </p>
            </section>

            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-foreground">2. Your choices</h2>
              <p className="text-muted-foreground">
                Analytics cookies are <strong>off by default</strong>. When you first visit, our
                banner lets you Accept All, Reject Non-Essential, or choose. You can change or
                withdraw your choice at any time:
              </p>
              <p>
                <CookiePreferencesButton className="inline-flex items-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 transition-colors" />
              </p>
              <p className="text-muted-foreground">
                You can also block or delete cookies in your browser settings, though strictly
                necessary cookies are needed for the site to work.
              </p>
            </section>

            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-foreground">
                3. Strictly necessary (always active)
              </h2>
              <div className="overflow-x-auto">
                <table className="w-full text-sm text-left border border-border rounded-lg">
                  <thead className="bg-muted text-foreground">
                    <tr>
                      <th className="px-4 py-2 font-semibold">Name</th>
                      <th className="px-4 py-2 font-semibold">Provider</th>
                      <th className="px-4 py-2 font-semibold">Purpose</th>
                      <th className="px-4 py-2 font-semibold">Storage</th>
                    </tr>
                  </thead>
                  <tbody className="text-muted-foreground">
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">Auth / session tokens</td>
                      <td className="px-4 py-2">Trunc</td>
                      <td className="px-4 py-2">Keeps you signed in</td>
                      <td className="px-4 py-2">localStorage</td>
                    </tr>
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">theme</td>
                      <td className="px-4 py-2">Trunc</td>
                      <td className="px-4 py-2">Remembers light/dark preference</td>
                      <td className="px-4 py-2">localStorage</td>
                    </tr>
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">cookie_consent</td>
                      <td className="px-4 py-2">Trunc</td>
                      <td className="px-4 py-2">Stores your cookie choices</td>
                      <td className="px-4 py-2">localStorage</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </section>

            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-foreground">
                4. Analytics (set only with your consent)
              </h2>
              <div className="overflow-x-auto">
                <table className="w-full text-sm text-left border border-border rounded-lg">
                  <thead className="bg-muted text-foreground">
                    <tr>
                      <th className="px-4 py-2 font-semibold">Provider</th>
                      <th className="px-4 py-2 font-semibold">Purpose</th>
                      <th className="px-4 py-2 font-semibold">More info</th>
                    </tr>
                  </thead>
                  <tbody className="text-muted-foreground">
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">Google Analytics (GA4)</td>
                      <td className="px-4 py-2">Aggregated usage and traffic measurement</td>
                      <td className="px-4 py-2">
                        <a
                          href="https://policies.google.com/privacy"
                          className="text-primary hover:underline"
                          rel="nofollow noopener"
                          target="_blank"
                        >
                          Google
                        </a>
                      </td>
                    </tr>
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">Ahrefs Analytics</td>
                      <td className="px-4 py-2">Privacy-friendly traffic analytics</td>
                      <td className="px-4 py-2">
                        <a
                          href="https://ahrefs.com/privacy"
                          className="text-primary hover:underline"
                          rel="nofollow noopener"
                          target="_blank"
                        >
                          Ahrefs
                        </a>
                      </td>
                    </tr>
                    <tr className="border-t border-border">
                      <td className="px-4 py-2">ContentSquare</td>
                      <td className="px-4 py-2">
                        Session insights (clicks, scrolls, navigation) to improve usability;
                        sensitive fields are masked
                      </td>
                      <td className="px-4 py-2">
                        <a
                          href="https://contentsquare.com/privacy-center/"
                          className="text-primary hover:underline"
                          rel="nofollow noopener"
                          target="_blank"
                        >
                          ContentSquare
                        </a>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p className="text-muted-foreground">
                These tools and their cookies do not load at all until you grant analytics
                consent. If you reject or withdraw consent, they are not used.
              </p>
            </section>

            <section className="space-y-4">
              <h2 className="text-2xl font-semibold text-foreground">5. Changes</h2>
              <p className="text-muted-foreground">
                We may update this Cookie Policy as our tools change. The &ldquo;Last
                updated&rdquo; date above shows the latest version. Questions? Email{" "}
                <a href="mailto:hello@trunc.site" className="text-primary hover:underline">
                  hello@trunc.site
                </a>
                .
              </p>
            </section>
          </div>
        </div>
      </div>
    </Layout>
  );
}
