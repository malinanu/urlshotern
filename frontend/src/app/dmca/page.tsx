'use client';

import { useState } from 'react';
import Layout from '@/components/layout/Layout';
import { Input } from '@/components/ui/input';
import { Textarea } from '@/components/ui/textarea';
import { Button } from '@/components/ui/button';
import { API_ENDPOINTS } from '@/lib/api-config';
import { CheckCircleIcon, ClipboardDocumentIcon } from '@heroicons/react/24/outline';

// Designated Agent under 17 U.S.C. §512(c)(2). Keep in sync with the agent
// registered at dmca.copyright.gov and with DMCAAgentEmail in the backend.
const AGENT = {
  title: 'Copyright Agent',
  company: '3logiq Technologies, Inc.',
  street: '123 Tech Boulevard',
  cityStateZip: 'San Francisco, CA 94105',
  country: 'USA',
  email: 'dmca@trunc.site',
};

const NOTICE_TEMPLATE = `To: ${AGENT.title}, ${AGENT.company}
${AGENT.street}, ${AGENT.cityStateZip}, ${AGENT.country}
Email: ${AGENT.email}

Subject: DMCA Takedown Notice

1. Copyrighted work: [Describe the copyrighted work you claim is infringed. If several works are covered by one notice, list them or give a representative list.]

2. Location of the original work (optional): [URL or other description of where the authorized work can be found.]

3. Infringing material: [List each trunc.site short link (for example https://trunc.site/abc123) that points to or distributes the infringing material.]

4. Contact information:
   Name:
   Company (if any):
   Postal address:
   Telephone:
   Email:

5. I have a good faith belief that use of the material in the manner complained of is not authorized by the copyright owner, its agent, or the law.

6. I swear, under penalty of perjury, that the information in this notification is accurate and that I am the copyright owner or am authorized to act on behalf of the owner of an exclusive right that is allegedly infringed.

Signature: [Your full legal name, typed, serves as an electronic signature]
Date:
`;

interface NoticeForm {
  name: string;
  company: string;
  role: 'owner' | 'agent';
  email: string;
  phone: string;
  address: string;
  copyrighted_work: string;
  original_url: string;
  infringing_urls: string;
  good_faith: boolean;
  accuracy: boolean;
  signature: string;
}

const EMPTY_FORM: NoticeForm = {
  name: '',
  company: '',
  role: 'owner',
  email: '',
  phone: '',
  address: '',
  copyrighted_work: '',
  original_url: '',
  infringing_urls: '',
  good_faith: false,
  accuracy: false,
  signature: '',
};

function Section({ id, title, children }: { id?: string; title: string; children: React.ReactNode }) {
  return (
    <section id={id} className="mb-10">
      <h2 className="text-2xl font-semibold text-foreground mb-4">{title}</h2>
      <div className="space-y-4 text-muted-foreground leading-relaxed">{children}</div>
    </section>
  );
}

function Field({ label, htmlFor, required, hint, children }: {
  label: string;
  htmlFor: string;
  required?: boolean;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label htmlFor={htmlFor} className="block text-sm font-medium text-foreground mb-1">
        {label}
        {required && <span className="text-destructive"> *</span>}
      </label>
      {children}
      {hint && <p className="mt-1 text-xs text-muted-foreground">{hint}</p>}
    </div>
  );
}

export default function DMCAPage() {
  const [form, setForm] = useState<NoticeForm>(EMPTY_FORM);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [referenceId, setReferenceId] = useState<string | null>(null);
  const [copied, setCopied] = useState(false);

  const set = <K extends keyof NoticeForm>(key: K, value: NoticeForm[K]) =>
    setForm((f) => ({ ...f, [key]: value }));

  const copyTemplate = async () => {
    try {
      await navigator.clipboard.writeText(NOTICE_TEMPLATE);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      setCopied(false);
    }
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    if (!form.good_faith || !form.accuracy) {
      setError('A valid notice must include both the good-faith and the accuracy statements.');
      return;
    }
    setSubmitting(true);
    try {
      const res = await fetch(API_ENDPOINTS.legal.dmca, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(form),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setError(data.error || `We couldn't submit your notice. Please email it to ${AGENT.email}.`);
        return;
      }
      setReferenceId(data.reference_id || 'received');
      setForm(EMPTY_FORM);
    } catch {
      setError(`We couldn't reach the server. Please email your notice to ${AGENT.email}.`);
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Layout>
      <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
        <h1 className="text-4xl font-bold text-foreground mb-4">DMCA &amp; Copyright Policy</h1>
        <p className="text-muted-foreground mb-10 leading-relaxed">
          Trunc respects the intellectual property rights of others and expects its users to do the same.
          We respond to notices of alleged copyright infringement that comply with the Digital Millennium
          Copyright Act (17 U.S.C. §512). Trunc short links redirect to content hosted on other websites:
          when we receive a valid notice we disable the short link, but the content itself stays on the
          destination site, so you may also want to notify that site&apos;s host.
        </p>

        <Section title="Designated Copyright Agent">
          <p>Send copyright notices and counter-notices to our Designated Agent:</p>
          <address className="not-italic rounded-lg border border-border bg-muted/30 p-4 text-foreground">
            {AGENT.title}
            <br />
            {AGENT.company}
            <br />
            {AGENT.street}
            <br />
            {AGENT.cityStateZip}
            <br />
            {AGENT.country}
            <br />
            Email:{' '}
            <a href={`mailto:${AGENT.email}`} className="text-primary hover:text-primary/80">
              {AGENT.email}
            </a>
          </address>
          <p>
            You can also use the <a href="#notice-form" className="text-primary hover:text-primary/80">online form</a>{' '}
            below. This address is for copyright matters only; other messages won&apos;t receive a response.
          </p>
        </Section>

        <Section title="What a takedown notice must include">
          <p>Under 17 U.S.C. §512(c)(3), your notice must be in writing and include:</p>
          <ol className="list-decimal pl-6 space-y-2">
            <li>A physical or electronic signature of a person authorized to act on behalf of the owner of the exclusive right allegedly infringed.</li>
            <li>Identification of the copyrighted work claimed to have been infringed (or, for multiple works, a representative list).</li>
            <li>Identification of the infringing material and information reasonably sufficient for us to locate it, such as the specific trunc.site short links.</li>
            <li>Your contact information: postal address, telephone number and, if available, email address.</li>
            <li>A statement that you have a good faith belief that the use of the material is not authorized by the copyright owner, its agent, or the law.</li>
            <li>A statement that the information in the notice is accurate and, under penalty of perjury, that you are authorized to act on behalf of the owner of the exclusive right allegedly infringed.</li>
          </ol>
          <p>Notices that are missing these elements may not be acted on.</p>
        </Section>

        <Section title="What happens after we receive a notice">
          <p>
            We review each notice and, where it is valid, expeditiously disable the identified short links. We
            notify the account holder who created the links and may give them a copy of your notice, including your
            contact details, so they can respond.
          </p>
        </Section>

        <Section title="Counter-notices">
          <p>
            If your link was disabled and you believe this was a mistake or misidentification, you may send our
            Designated Agent a counter-notice under 17 U.S.C. §512(g)(3) that includes:
          </p>
          <ol className="list-decimal pl-6 space-y-2">
            <li>Your physical or electronic signature.</li>
            <li>Identification of the short link that was disabled and the location (URL) where it appeared before it was disabled.</li>
            <li>A statement under penalty of perjury that you have a good faith belief the material was disabled as a result of mistake or misidentification.</li>
            <li>
              Your name, address and telephone number, and a statement that you consent to the jurisdiction of the
              Federal District Court for the judicial district in which your address is located (or, if your address
              is outside the United States, any judicial district in which Trunc may be found), and that you will
              accept service of process from the person who submitted the original notice or their agent.
            </li>
          </ol>
          <p>
            We forward valid counter-notices to the person who sent the original notice. Unless they tell us they
            have filed a court action seeking to restrain the alleged infringement, we may restore the link 10 to 14
            business days after receiving the counter-notice.
          </p>
        </Section>

        <Section title="Repeat infringers">
          <p>
            In appropriate circumstances, Trunc disables and terminates the accounts of users who are repeat
            infringers.
          </p>
        </Section>

        <Section title="Misrepresentations">
          <p>
            Under 17 U.S.C. §512(f), anyone who knowingly materially misrepresents that material is infringing, or that
            it was removed or disabled by mistake, may be liable for damages, including costs and attorneys&apos; fees.
            If you aren&apos;t sure whether material infringes your copyright, consider consulting a lawyer first.
          </p>
        </Section>

        <Section title="Takedown notice template">
          <p>Copy this template, fill in every bracketed part, and email it to {AGENT.email}, or use the form below.</p>
          <div className="relative">
            <pre className="whitespace-pre-wrap rounded-lg border border-border bg-muted/30 p-4 text-sm text-foreground">
              {NOTICE_TEMPLATE}
            </pre>
            <button
              type="button"
              onClick={copyTemplate}
              className="absolute top-2 right-2 inline-flex items-center gap-1 rounded-md border border-border bg-background px-2 py-1 text-xs text-foreground hover:bg-muted"
            >
              <ClipboardDocumentIcon className="h-4 w-4" />
              {copied ? 'Copied' : 'Copy'}
            </button>
          </div>
        </Section>

        <Section id="notice-form" title="Submit a takedown notice online">
          {referenceId ? (
            <div className="rounded-lg border border-border bg-muted/30 p-6 text-center">
              <CheckCircleIcon className="h-12 w-12 text-green-600 dark:text-green-500 mx-auto mb-3" />
              <p className="text-foreground font-medium">Your notice was sent to our Designated Agent.</p>
              <p className="text-sm mt-1">
                Reference: <span className="font-mono text-foreground">{referenceId}</span>. We emailed you a receipt.
              </p>
              <button
                type="button"
                onClick={() => setReferenceId(null)}
                className="mt-4 text-sm text-primary hover:text-primary/80"
              >
                Submit another notice
              </button>
            </div>
          ) : (
            <form onSubmit={submit} className="space-y-5 rounded-lg border border-border p-6">
              <div className="grid gap-5 sm:grid-cols-2">
                <Field label="Full legal name" htmlFor="dmca-name" required>
                  <Input id="dmca-name" required value={form.name} onChange={(e) => set('name', e.target.value)} />
                </Field>
                <Field label="Company or organization" htmlFor="dmca-company">
                  <Input id="dmca-company" value={form.company} onChange={(e) => set('company', e.target.value)} />
                </Field>
                <Field label="Email" htmlFor="dmca-email" required>
                  <Input id="dmca-email" type="email" required value={form.email} onChange={(e) => set('email', e.target.value)} />
                </Field>
                <Field label="Telephone" htmlFor="dmca-phone">
                  <Input id="dmca-phone" type="tel" value={form.phone} onChange={(e) => set('phone', e.target.value)} />
                </Field>
              </div>

              <Field label="Postal address" htmlFor="dmca-address" required>
                <Textarea id="dmca-address" required rows={2} value={form.address} onChange={(e) => set('address', e.target.value)} />
              </Field>

              <fieldset>
                <legend className="block text-sm font-medium text-foreground mb-2">
                  I am <span className="text-destructive">*</span>
                </legend>
                <div className="flex flex-col gap-2 sm:flex-row sm:gap-6 text-sm text-foreground">
                  <label className="inline-flex items-center gap-2">
                    <input type="radio" name="role" checked={form.role === 'owner'} onChange={() => set('role', 'owner')} />
                    The copyright owner
                  </label>
                  <label className="inline-flex items-center gap-2">
                    <input type="radio" name="role" checked={form.role === 'agent'} onChange={() => set('role', 'agent')} />
                    Authorized to act for the owner
                  </label>
                </div>
              </fieldset>

              <Field
                label="Copyrighted work"
                htmlFor="dmca-work"
                required
                hint="Describe the work you own (title, type, registration number if any)."
              >
                <Textarea id="dmca-work" required rows={3} minLength={10} value={form.copyrighted_work} onChange={(e) => set('copyrighted_work', e.target.value)} />
              </Field>

              <Field label="Where the original work can be found" htmlFor="dmca-original" hint="Optional URL of the authorized copy.">
                <Input id="dmca-original" value={form.original_url} onChange={(e) => set('original_url', e.target.value)} />
              </Field>

              <Field
                label="Infringing trunc.site short links"
                htmlFor="dmca-urls"
                required
                hint="One link per line, for example https://trunc.site/abc123."
              >
                <Textarea id="dmca-urls" required rows={3} value={form.infringing_urls} onChange={(e) => set('infringing_urls', e.target.value)} />
              </Field>

              <div className="space-y-3 text-sm text-foreground">
                <label className="flex items-start gap-3">
                  <input type="checkbox" className="mt-1" checked={form.good_faith} onChange={(e) => set('good_faith', e.target.checked)} />
                  <span>
                    I have a good faith belief that use of the material in the manner complained of is not authorized by
                    the copyright owner, its agent, or the law.
                  </span>
                </label>
                <label className="flex items-start gap-3">
                  <input type="checkbox" className="mt-1" checked={form.accuracy} onChange={(e) => set('accuracy', e.target.checked)} />
                  <span>
                    I swear, under penalty of perjury, that the information in this notification is accurate and that I am
                    the copyright owner or am authorized to act on behalf of the owner of an exclusive right that is
                    allegedly infringed.
                  </span>
                </label>
              </div>

              <Field
                label="Electronic signature"
                htmlFor="dmca-signature"
                required
                hint="Type your full legal name. This serves as your electronic signature."
              >
                <Input id="dmca-signature" required value={form.signature} onChange={(e) => set('signature', e.target.value)} />
              </Field>

              {error && <p className="text-sm text-destructive">{error}</p>}

              <Button type="submit" disabled={submitting || !form.good_faith || !form.accuracy}>
                {submitting ? 'Submitting…' : 'Submit notice'}
              </Button>
            </form>
          )}
        </Section>
      </div>
    </Layout>
  );
}
