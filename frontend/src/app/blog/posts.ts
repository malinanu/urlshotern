export type ContentSection =
  | { type: 'p'; text: string }
  | { type: 'h2'; text: string }
  | { type: 'h3'; text: string }
  | { type: 'ul'; items: string[] }
  | { type: 'ol'; items: string[] }
  | { type: 'table'; headers: string[]; rows: string[][] }
  | { type: 'callout'; text: string }
  | { type: 'cta'; text: string; href: string; label: string };

export interface BlogPost {
  slug: string;
  title: string;
  description: string;
  publishedAt: string;
  readTime: number;
  category: string;
  content: ContentSection[];
}

export const posts: BlogPost[] = [
  {
    slug: 'how-to-shorten-a-url',
    title: 'How to Shorten a URL: The Complete Step-by-Step Guide (2026)',
    description: 'Learn how to shorten a URL in seconds using Trunc. Free URL shortener with click analytics, custom domains, and QR codes. No account required to start.',
    publishedAt: '2026-03-20',
    readTime: 5,
    category: 'Guide',
    content: [
      { type: 'p', text: 'URL shortening converts a long web address into a compact, shareable link. When someone clicks the short link, a URL shortener like Trunc looks up the original URL in its database and instantly redirects the visitor — the whole process takes under 50 milliseconds. Short links are easier to share on social media, in emails, and in print materials, and they come with built-in click tracking so you can see exactly how many people used your link.' },
      { type: 'h2', text: 'How to Shorten a URL with Trunc (3 Steps)' },
      { type: 'ol', items: [
        'Go to trunc.site and paste your long URL into the input field on the homepage.',
        'Optionally, enter a custom short code to personalize your link (e.g. trunc.site/my-campaign). Leave it blank and Trunc will generate one automatically.',
        'Click "Shorten." Your short link is ready instantly — copy it and share it anywhere.',
      ]},
      { type: 'callout', text: 'Trunc generates short links instantly with no account required on the free plan. Every link includes click tracking by default.' },
      { type: 'h2', text: 'How to Shorten a URL on Your Phone' },
      { type: 'p', text: 'Trunc is fully mobile-optimised. Open trunc.site in your phone\'s browser, paste or type the URL you want to shorten, and tap "Shorten." The short link is copied to your clipboard in one tap. You can also install Trunc as a Progressive Web App (PWA) from your phone\'s browser menu for faster access.' },
      { type: 'h2', text: 'How to Shorten a URL and Track Who Clicks It' },
      { type: 'p', text: 'Every link shortened with a free Trunc account includes automatic click tracking. After shortening, sign in or create a free account to access your dashboard. There you can see:' },
      { type: 'ul', items: [
        'Total clicks over time with a trend chart',
        'Geographic breakdown — which countries and cities your clicks come from',
        'Device and browser breakdown — desktop vs mobile, Chrome vs Safari',
        'Referrer sources — which websites sent traffic to your link',
        'Click-by-click timestamps',
      ]},
      { type: 'h2', text: 'How to Shorten a URL with a Custom Domain' },
      { type: 'p', text: 'On Trunc\'s Premium plan, you can use your own branded domain for short links — for example, links.yourcompany.com/promo instead of trunc.site/abc123. This increases trust with your audience and reinforces your brand every time someone sees the link. Setup takes about 5 minutes: add your domain in the Trunc dashboard, update your DNS records, and all new links will use your brand domain automatically.' },
      { type: 'h2', text: 'How to Shorten a URL via the Trunc API' },
      { type: 'p', text: 'Developers can shorten URLs programmatically using the Trunc REST API. Send a POST request to the /api/v1/links endpoint with your long URL in the request body and your API key in the header. The API returns the short link and its analytics endpoint in under 100ms. Full documentation is available at trunc.site/api-docs.' },
      { type: 'h2', text: 'Why URL Shortening Matters for Marketing' },
      { type: 'p', text: 'Short links do more than save characters. When you share a shortened link in a campaign, every click is tracked automatically — you know exactly how many people responded to your ad, email, or social post. Combined with UTM parameters, short links give you a complete picture of which channels and messages drive the most traffic. Trunc\'s AI Insights feature goes a step further, analysing your click patterns and suggesting the best times to share your links for maximum reach.' },
      { type: 'cta', text: 'Ready to shorten your first link? It takes 10 seconds and no credit card is required.', href: 'https://trunc.site/register', label: 'Get Started Free' },
    ],
  },

  {
    slug: 'trunc-vs-bitly-tinyurl',
    title: 'Trunc vs Bitly vs TinyURL: Which URL Shortener is Best in 2026?',
    description: 'Honest comparison of Trunc, Bitly, and TinyURL. We compare pricing, analytics, custom domains, and AI features to help you choose the best URL shortener for your needs.',
    publishedAt: '2026-03-15',
    readTime: 8,
    category: 'Comparison',
    content: [
      { type: 'p', text: 'Choosing the right URL shortener affects your analytics, your brand, and your budget. Bitly is the household name, TinyURL is the oldest free option, and Trunc is the new generation — built with AI insights, real-time analytics, and A/B testing from day one. Here is an honest, feature-by-feature breakdown to help you pick the right tool.' },
      { type: 'h2', text: 'At a Glance: Trunc vs Bitly vs TinyURL' },
      { type: 'table', headers: ['Feature', 'Trunc', 'Bitly', 'TinyURL'], rows: [
        ['Free plan links', '10/month', '10/month', 'Unlimited (no analytics)'],
        ['Click analytics (free)', 'Yes', 'Limited (30 days)', 'No'],
        ['Custom domains', 'Premium', 'Paid ($35+/mo)', 'No'],
        ['QR codes', 'Yes (free)', 'Yes (paid)', 'No'],
        ['A/B link testing', 'Yes', 'No', 'No'],
        ['AI insights', 'Yes', 'No', 'No'],
        ['Team collaboration', 'Yes (Premium)', 'Yes (paid)', 'No'],
        ['API access', 'Yes (Premium)', 'Yes (paid)', 'No'],
        ['Password-protected links', 'Yes', 'No', 'No'],
        ['Link expiration', 'Yes (free)', 'No', 'No'],
        ['Starting paid price', 'See pricing', '$8/month', '$9.99/month'],
      ]},
      { type: 'h2', text: 'Trunc: Best for Marketers Who Want More Than a Short Link' },
      { type: 'p', text: 'Trunc was built with analytics as a first-class feature, not an afterthought. Every free account gets full click tracking — geographic data, device breakdown, referrer sources, and real-time stats — with no time limit on how long data is stored. The feature that sets Trunc apart from every competitor is AI Insights: the dashboard analyses your click patterns and tells you which links are underperforming, what times drive the most clicks, and what your audience profile looks like.' },
      { type: 'ul', items: [
        'AI Insights — unique to Trunc, not available in Bitly or TinyURL',
        'A/B link testing — split traffic between two destination URLs and see which performs better',
        'QR codes included free — no upsell required',
        'Link expiration — set links to stop working after a date or click count',
        'Password protection — gated links for private content',
      ]},
      { type: 'h2', text: 'Bitly: Best for Enterprise Teams Already Invested in the Ecosystem' },
      { type: 'p', text: 'Bitly is the most established URL shortener and has the deepest integration ecosystem — it connects natively with HubSpot, Hootsuite, Zapier, and dozens of other marketing tools. If your team is already using these integrations and switching cost is a concern, Bitly is a defensible choice. However, its pricing has increased significantly: the Core plan starts at $8/month for only 10 branded links, and you need the Growth plan ($29/month) for meaningful analytics depth. Custom domains require the Business tier at $199/month.' },
      { type: 'h2', text: 'TinyURL: Best for One-Off Links with No Analytics Needs' },
      { type: 'p', text: 'TinyURL is the simplest option: paste a URL, get a short link, done. There is no account required, no analytics, and no dashboard. For occasional personal use — sharing a long link in a text message or printed document — TinyURL is perfectly adequate. But for any marketing or business use where you need to know if the link was actually clicked, TinyURL is a dead end.' },
      { type: 'h2', text: 'Pricing Comparison' },
      { type: 'p', text: 'Trunc offers a free plan with real analytics, a transparent Premium tier with custom domains and team features, and an Enterprise plan for unlimited usage. Bitly\'s free tier has severely limited analytics (30-day data retention) and locks most useful features behind paid plans starting at $8/month. TinyURL Pro starts at $9.99/month and includes basic analytics, but no AI features or team collaboration.' },
      { type: 'callout', text: 'Trunc offers 50% off your first year on any annual plan. Use coupon code ANNUAL50 at checkout.' },
      { type: 'h2', text: 'Verdict: Which Should You Choose?' },
      { type: 'ul', items: [
        'Choose Trunc if you want full analytics on a free plan, AI-powered insights, A/B testing, or QR codes without paying extra.',
        'Choose Bitly if your team already relies on its third-party integrations and switching costs are high.',
        'Choose TinyURL only for personal, one-off link shortening where analytics are irrelevant.',
      ]},
      { type: 'cta', text: 'Try Trunc free — no credit card required. Get full analytics, QR codes, and AI insights from day one.', href: 'https://trunc.site/register', label: 'Start for Free' },
    ],
  },

  {
    slug: 'best-url-shorteners-2026',
    title: '7 Best URL Shorteners in 2026: Ranked and Reviewed',
    description: 'We tested 7 URL shorteners in 2026 and ranked them on analytics, pricing, custom domains, and ease of use. See which URL shortener is right for you.',
    publishedAt: '2026-03-25',
    readTime: 10,
    category: 'Roundup',
    content: [
      { type: 'p', text: 'URL shorteners have evolved from simple redirect tools into full marketing analytics platforms. The best ones in 2026 offer real-time click data, custom branded domains, QR code generation, A/B testing, and — increasingly — AI-powered insights. We tested seven of the most popular options across these dimensions and ranked them for different use cases.' },
      { type: 'h2', text: '1. Trunc — Best Overall for Analytics and AI Features' },
      { type: 'p', text: 'Trunc earns the top spot by delivering enterprise-grade analytics on a free plan and offering features that no competitor has matched: AI Insights and A/B link testing. AI Insights analyses your click data and surfaces actionable recommendations — the best times to share links, which audience segments are clicking, and which campaigns are underperforming. A/B testing lets you send 50% of clicks to one URL and 50% to another, so you can measure which landing page, offer, or headline actually converts better.' },
      { type: 'ul', items: [
        'Free plan: 10 links/month with full analytics (geographic, device, referrer, real-time)',
        'AI Insights: automatic pattern analysis across your links',
        'A/B testing: split traffic between two destinations',
        'QR codes included on all plans',
        'Custom domains on Premium',
        'Team collaboration and API access',
      ]},
      { type: 'callout', text: 'Trunc is the only URL shortener with built-in AI Insights. Get 50% off annual plans with code ANNUAL50.' },
      { type: 'h2', text: '2. Bitly — Best for Integration Ecosystem' },
      { type: 'p', text: 'Bitly remains the most widely integrated URL shortener, with native connectors for HubSpot, Hootsuite, Salesforce, Zapier, and hundreds of other tools. If your stack already revolves around these platforms, Bitly\'s deep integrations reduce friction. The downside is cost — meaningful analytics and custom domains are locked behind plans starting at $29-199/month. For analytics depth alone, Trunc delivers more at a lower price point.' },
      { type: 'h2', text: '3. Rebrandly — Best for Custom Domain Branding' },
      { type: 'p', text: 'Rebrandly is purpose-built for branded short domains. Every link uses your own domain name — go.yourcompany.com/sale — which increases click-through rates by up to 39% compared to generic short domains. The trade-off is that Rebrandly\'s analytics are shallower than Trunc or Bitly, and the free plan limits you to 10 links with one custom domain. Plans start at $13/month.' },
      { type: 'h2', text: '4. Short.io — Best for Developers and API Users' },
      { type: 'p', text: 'Short.io is developer-first: its REST API is well-documented, supports webhooks, and has native SDKs for JavaScript, Python, PHP, and Ruby. Custom domains are included from the first paid tier ($20/month). Analytics are solid but lack the depth of Trunc\'s AI-powered dashboard. Best suited to developers who need to embed link shortening into their own applications.' },
      { type: 'h2', text: '5. TinyURL — Best for Occasional Personal Use' },
      { type: 'p', text: 'TinyURL is the simplest option: no account required, no analytics, unlimited free links. It has been operating since 2002 and has a proven reliability record. For personal use — sharing a link in a text message or a printed flyer — it does the job. For any business or marketing purpose, the lack of analytics makes it unsuitable.' },
      { type: 'h2', text: '6. T2M — Best for Lifetime Deals and Budget Teams' },
      { type: 'p', text: 'T2M (track-to-monetize) offers a one-time lifetime purchase option — you pay once and use the service forever. For small teams that want to avoid monthly subscriptions, this is an attractive model. Analytics and custom domains are included, though the interface is less polished than Trunc or Bitly and the AI features are absent.' },
      { type: 'h2', text: '7. BL.INK — Best for Enterprise and Compliance' },
      { type: 'p', text: 'BL.INK targets large enterprises with SOC 2 compliance, SSO, and detailed audit logs. If your organisation has strict security requirements — financial services, healthcare, government — BL.INK covers the compliance boxes that consumer-grade shorteners do not. Pricing reflects the enterprise focus, starting at $48/month.' },
      { type: 'h2', text: 'How to Choose the Right URL Shortener' },
      { type: 'ul', items: [
        'Need analytics on a free plan? → Trunc',
        'Need AI insights and A/B testing? → Trunc',
        'Already deep in the HubSpot/Salesforce ecosystem? → Bitly',
        'Branding is the priority and analytics are secondary? → Rebrandly',
        'Building a product that needs link shortening via API? → Short.io or Trunc',
        'Occasional personal use, no analytics needed? → TinyURL',
        'Enterprise with compliance requirements? → BL.INK',
      ]},
      { type: 'h2', text: 'Final Verdict' },
      { type: 'p', text: 'For most marketers, developers, and small-to-medium businesses in 2026, Trunc offers the best balance of features, analytics depth, and price. Its free plan is genuinely useful — not a crippled trial — and features like AI Insights and A/B testing give it a material advantage over every competitor on this list.' },
      { type: 'cta', text: 'Start using Trunc free. No credit card required. 50% off annual plans with code ANNUAL50.', href: 'https://trunc.site/register', label: 'Try Trunc Free' },
    ],
  },
];

export function getPost(slug: string): BlogPost | undefined {
  return posts.find((p) => p.slug === slug);
}
