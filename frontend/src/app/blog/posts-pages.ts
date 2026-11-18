import type { BlogPost } from './posts';

/**
 * Trunc Pages content cluster: digital business cards, vCard QR codes,
 * link in bio and small-business pages. Kept separate from the URL-shortener
 * posts so the cluster is easy to review and extend.
 *
 * Rule for this file: describe only shipped features and never invent
 * customers, statistics or prices (see PRODUCT.md, "Evidence on Hand").
 */
export const pagesPosts: BlogPost[] = [
  {
    slug: 'how-to-create-a-free-digital-business-card',
    title: 'How to Create a Free Digital Business Card (with QR Code) in 2026',
    seoTitle: 'How to Make a Free Digital Business Card (2026)',
    description:
      'Create a free digital business card with a QR code and a one-tap Save Contact button. Step-by-step guide, what to include, and how to share it at events.',
    publishedAt: '2026-11-18',
    updatedAt: '2026-11-18',
    readTime: 7,
    category: 'Guide',
    image: '/blog/how-to-create-a-free-digital-business-card.webp',
    imageAlt: 'A phone showing a digital business card with a Save Contact button next to a QR code',
    author: 'Alex Rivera',
    authorSlug: 'alex-rivera',
    content: [
      { type: 'p', text: 'A digital business card is a mobile web page with your name, role, photo and contact details, shared through a link or a QR code. The person you meet scans the code, sees your card in their browser, and taps "Save Contact" to add you to their phone. Nobody needs to install an app, and you can update the details any time without reprinting anything. With Trunc you can build one for free in about five minutes.' },

      { type: 'h2', text: 'Why Use a Digital Business Card Instead of Paper?' },
      { type: 'ul', items: [
        'It actually gets saved. A paper card goes into a pocket; a digital card goes straight into the phone\'s contacts with your photo, number, email and website filled in.',
        'You can change it. New phone number, new job title, new website: edit the card and everyone who scans the same QR code sees the update.',
        'It does more than a card can. Add buttons to call, WhatsApp, email or book a meeting, plus links to your portfolio, LinkedIn or latest work.',
        'You can see if it works. A digital card can show how many people viewed it, saved your contact or tapped a link. Paper can\'t.',
        'It never runs out. No reprints, no "sorry, I\'m out of cards" at the end of a conference.',
      ]},

      { type: 'h2', text: 'What to Put on a Digital Business Card' },
      { type: 'p', text: 'Keep it focused. Someone scanning your card has a few seconds and is usually standing in front of you. Put the one thing you want them to do (save your contact) at the top, and keep everything else short.' },
      { type: 'table', headers: ['Element', 'Include?', 'Tip'], rows: [
        ['Name and job title', 'Always', 'Use the title people will search for, not an internal one'],
        ['Photo or logo', 'Yes', 'A clear headshot helps people remember who you were'],
        ['Save Contact button', 'Always', 'This should be the first button on the screen'],
        ['Phone, email, WhatsApp', 'Yes', 'Use tap-to-call and tap-to-message buttons, not plain text'],
        ['Company and website', 'Yes', 'One link to your site, not five'],
        ['Social links', 'Optional', 'Only the profiles you keep up to date'],
        ['Short bio', 'Optional', 'One or two lines: what you do and who you help'],
        ['Booking link', 'Optional', 'Useful for sales, consultants and recruiters'],
      ]},

      { type: 'h2', text: 'How to Create a Free Digital Business Card with Trunc (5 Steps)' },
      { type: 'ol', items: [
        'Create a free Trunc account at trunc.site/register. No credit card is needed.',
        'Open Pages in the dashboard and choose the Digital Business Card template.',
        'In the Build tab, fill in your profile (name, title, photo) and the Save Contact block with the details you want saved: phone, email, company, website and address.',
        'In the Design tab, set your colours and font so the card matches your brand. The live preview shows exactly what people will see on their phone.',
        'Publish. Your card gets its own link and a QR code you can download from the editor. The QR code is tagged so scans show up separately from link clicks in the Track tab.',
      ]},
      { type: 'callout', text: 'The free plan includes one published page, which is all you need for a personal business card. Pro and Ultra add more pages and contact forms, and Ultra lets you remove the small "Made with Trunc" badge.' },

      { type: 'h2', text: 'How the "Save Contact" Button Works' },
      { type: 'p', text: 'When someone taps Save Contact on a Trunc card, their phone downloads a standard vCard (.vcf) file, the same format phones use internally for contacts. iPhone and Android both recognise it and show an "Add to contacts" screen with your details already filled in. Because it is a real vCard file rather than a screenshot or a link, your name, numbers, email, company and website land in the right fields. If you want to understand the format itself, read our guide to vCard QR codes.' },

      { type: 'h2', text: 'How to Share Your Digital Business Card' },
      { type: 'h3', text: 'In person' },
      { type: 'ul', items: [
        'Show the QR code on your phone\'s screen (save it as a photo, or add it to your lock screen).',
        'Print the QR code on the back of a paper card, a lanyard badge or a table stand at your booth.',
        'Put the link on an NFC tag or card if you use one. Most NFC business cards just open a URL.',
      ]},
      { type: 'h3', text: 'Online' },
      { type: 'ul', items: [
        'Add the link to your email signature and LinkedIn "Contact info" section.',
        'Paste it in your Instagram, TikTok or X bio.',
        'Include it in proposals, invoices and Zoom chat after a call.',
      ]},

      { type: 'h2', text: 'Static QR Code vs. QR Code to a Card Page' },
      { type: 'p', text: 'There are two ways to put contact details behind a QR code. A static vCard QR code stores your details inside the code itself: it works offline but can\'t be edited, can\'t be tracked, and gets dense and harder to scan as you add fields. A QR code that points to a card page stores only a short link: the code stays simple, you can edit the page any time, and you can see scans and saves. For most professionals the second option is better, because job titles and numbers change.' },

      { type: 'h2', text: 'Mistakes to Avoid' },
      { type: 'ul', items: [
        'Too many buttons. Five clear actions beat fifteen links.',
        'A tiny printed QR code. Print it at least 2 × 2 cm (about 0.8 inches) and leave a white margin around it.',
        'Not testing. Scan your code with both an iPhone and an Android phone before you print anything.',
        'Forgetting the follow-up. A saved contact is a start, so follow up within 24 hours while they still remember you.',
      ]},

      { type: 'h2', text: 'Frequently Asked Questions' },
      { type: 'h3', text: 'Is a digital business card really free?' },
      { type: 'p', text: 'With Trunc, yes: the free plan includes one published page with Save Contact, a QR code and basic analytics, with no time limit. Paid plans add more pages, forms and badge removal.' },
      { type: 'h3', text: 'Does the other person need an app?' },
      { type: 'p', text: 'No. The card opens in the phone\'s normal browser and the contact is saved with the phone\'s own contacts app.' },
      { type: 'h3', text: 'Can I update my card after printing the QR code?' },
      { type: 'p', text: 'Yes. The QR code points to your card\'s link, not to the details themselves, so any edit you publish appears the next time someone scans it.' },

      { type: 'cta', text: 'Make your free digital business card now. It takes about five minutes and no credit card is required.', href: 'https://trunc.site/register', label: 'Create My Free Card' },
      { type: 'cta', text: 'Want to know what is inside a contact QR code?', href: '/blog/vcard-qr-code', label: 'Read: vCard QR Codes Explained' },
    ],
  },

  {
    slug: 'vcard-qr-code',
    title: 'vCard QR Code: What It Is, How It Works and How to Make One',
    seoTitle: 'vCard QR Code: What It Is and How to Make One',
    description:
      'Learn what a vCard QR code is, what data it holds, static vs. dynamic codes, and how to make one that saves your contact in one scan on iPhone and Android.',
    publishedAt: '2026-11-18',
    updatedAt: '2026-11-18',
    readTime: 6,
    category: 'Guide',
    image: '/blog/vcard-qr-code.webp',
    imageAlt: 'A QR code beside a contact card showing name, phone and email fields from a vCard file',
    author: 'Alex Rivera',
    authorSlug: 'alex-rivera',
    content: [
      { type: 'p', text: 'A vCard QR code is a QR code that saves contact details to a phone when scanned. "vCard" is the standard file format for contacts, the same format your phone uses when you share a contact or export your address book. When the camera reads a vCard QR code, iPhone and Android offer to add the person to contacts with name, phone, email and company already filled in.' },

      { type: 'h2', text: 'What Is Inside a vCard?' },
      { type: 'p', text: 'A vCard is plain text with one field per line. A minimal version 3.0 vCard (the most widely supported version) looks like this:' },
      { type: 'ul', items: [
        'BEGIN:VCARD',
        'VERSION:3.0',
        'N:Perera;Nimal;;;',
        'FN:Nimal Perera',
        'ORG:Example Co.',
        'TITLE:Sales Manager',
        'TEL;TYPE=CELL:+94771234567',
        'EMAIL:nimal@example.com',
        'URL:https://example.com',
        'END:VCARD',
      ]},
      { type: 'p', text: 'N holds the structured name (last; first), FN the display name, and TEL, EMAIL and URL the contact methods. Phone numbers should always use the international format with the country code, so they work for people abroad.' },

      { type: 'h2', text: 'Two Ways to Make a vCard QR Code' },
      { type: 'h3', text: '1. Static vCard QR code' },
      { type: 'p', text: 'The whole vCard text is encoded inside the QR code. Many free generators do this. It works without internet, but every field you add makes the code denser (more, smaller squares), which makes it harder to scan when printed small. And once printed, it can never be changed or tracked.' },
      { type: 'h3', text: '2. QR code to a contact page (dynamic)' },
      { type: 'p', text: 'The QR code holds only a short link to a contact page. The page has a "Save Contact" button that downloads the vCard file. The code stays small and easy to scan, you can edit the details any time, and you can see how many people scanned and saved.' },
      { type: 'table', headers: ['', 'Static vCard QR', 'QR to contact page'], rows: [
        ['Edit after printing', 'No', 'Yes'],
        ['Scan tracking', 'No', 'Yes'],
        ['Code density', 'Grows with every field', 'Always small (short link)'],
        ['Works offline', 'Yes', 'Needs internet to open the page'],
        ['Extra content (photo, links, booking)', 'No', 'Yes'],
        ['Best for', 'One-off use, offline venues', 'Business cards, badges, anything printed in bulk'],
      ]},
      { type: 'callout', text: 'Rule of thumb: if the code will be printed on something you can\'t easily reprint, such as cards, badges, signage or packaging, point it to a page you can edit.' },

      { type: 'h2', text: 'How to Make a vCard QR Code with Trunc' },
      { type: 'ol', items: [
        'Sign up free at trunc.site/register.',
        'Go to Pages and create a page from the Digital Business Card template.',
        'Fill in the Save Contact block: name, title, company, phones, email, website and address. These become the fields in the vCard.',
        'Publish, then open the QR code from the editor and download it.',
        'Scan it with your own phone, tap Save Contact, and check every field landed in the right place.',
      ]},
      { type: 'p', text: 'Trunc generates a standard vCard 3.0 file when someone taps Save Contact, and counts each save in the Track tab, so you can see how many scans turned into saved contacts.' },

      { type: 'h2', text: 'Does a vCard QR Code Work on iPhone and Android?' },
      { type: 'p', text: 'Yes. The built-in camera on iPhone recognises vCard QR codes and shows a contact preview. Most Android phones do the same through the camera app or Google Lens. With a contact-page QR code, the phone opens the page in the browser and the Save Contact button hands the .vcf file to the contacts app, which works the same way on both.' },

      { type: 'h2', text: 'Printing Tips' },
      { type: 'ul', items: [
        'Minimum size: about 2 × 2 cm (0.8 in) for a short-link code on a business card; larger for static vCard codes.',
        'Keep a quiet zone: leave a white margin around the code at least as wide as four of its small squares.',
        'Use dark on light. Inverted or low-contrast codes fail on some phones.',
        'Export as SVG for print so the code stays sharp at any size.',
        'Add a short label like "Scan to save my contact" so people know what happens.',
      ]},

      { type: 'h2', text: 'Frequently Asked Questions' },
      { type: 'h3', text: 'Are vCard QR codes free?' },
      { type: 'p', text: 'Static vCard QR codes are free with most generators. Editable, trackable codes usually require an account; Trunc includes one page with a QR code on the free plan.' },
      { type: 'h3', text: 'Is my data safe in a vCard QR code?' },
      { type: 'p', text: 'Anything in the code or on the page is visible to whoever scans it, so only include details you are happy to share with anyone who sees the code. Trunc pages don\'t use third-party trackers, and the analytics are first-party.' },

      { type: 'cta', text: 'Create an editable, trackable vCard QR code for free.', href: 'https://trunc.site/register', label: 'Make My vCard QR Code' },
      { type: 'cta', text: 'Need the full walkthrough for building the card itself?', href: '/blog/how-to-create-a-free-digital-business-card', label: 'Read: Create a Free Digital Business Card' },
    ],
  },

  {
    slug: 'best-linktree-alternatives',
    title: 'Best Linktree Alternatives in 2026: An Honest Comparison',
    seoTitle: 'Best Linktree Alternatives in 2026 (Honest Guide)',
    description:
      'The best Linktree alternatives for creators, small businesses and professionals in 2026: what each tool does best, where it falls short, and how to choose.',
    publishedAt: '2026-11-18',
    updatedAt: '2026-11-18',
    readTime: 9,
    category: 'Comparison',
    image: '/blog/best-linktree-alternatives.webp',
    imageAlt: 'Several phone screens showing different link-in-bio page layouts side by side',
    author: 'Alex Rivera',
    authorSlug: 'alex-rivera',
    content: [
      { type: 'p', text: 'Linktree made "link in bio" a category, and it is still a solid choice. But it isn\'t the best fit for everyone. Some people need to sell products, some need a real business card, some want their short links and QR codes in the same place, and some just want more control over design. This guide compares the main alternatives by what they are actually good at, so you can pick based on your situation rather than on whoever ranks themselves first.' },
      { type: 'callout', text: 'Disclosure: Trunc is our product and it is on this list. We have described every tool, including ours, by its main purpose and limitations. Prices and plans change often, so check each vendor\'s pricing page before deciding.' },

      { type: 'h2', text: 'How We Compared Them' },
      { type: 'ul', items: [
        'Main use case: what the tool is designed for (creators, selling, small business, professionals).',
        'Free plan: whether you can publish a usable page without paying.',
        'Analytics: whether you can see views, clicks and traffic sources.',
        'QR codes and short links: whether you can share the page offline and track it.',
        'Design control: templates only, or real layout control.',
      ]},

      { type: 'h2', text: 'Quick Comparison' },
      { type: 'table', headers: ['Tool', 'Best for', 'Stand-out strength', 'Watch out for'], rows: [
        ['Linktree', 'Creators who want the default', 'Most integrations, very familiar to visitors', 'Advanced analytics and customisation sit on paid tiers'],
        ['Beacons', 'Creators who sell', 'Built-in store, media kit and email tools', 'Feature-heavy if you only need links'],
        ['Stan Store', 'Selling digital products and coaching', 'Checkout built into the bio page', 'Paid only; overkill for a simple link list'],
        ['Carrd', 'People who want a real one-page site', 'Flexible layouts, custom domains on paid plans', 'Not built around link-in-bio analytics'],
        ['Taplink', 'Small businesses and WhatsApp sellers', 'Messenger buttons, forms and simple payments', 'Design is more template-bound'],
        ['Bitly', 'Marketing teams already using Bitly links', 'Link-in-bio page next to an established shortener', 'Page builder is a side feature of a link platform'],
        ['Trunc', 'Professionals, small shops and event organisers who also use short links and QR codes', 'Page, short link, QR code and first-party analytics in one place; real vCard Save Contact', 'Newer, smaller ecosystem of integrations than Linktree'],
      ]},

      { type: 'h2', text: 'The Alternatives in Detail' },
      { type: 'h3', text: 'Beacons: best for creators who sell' },
      { type: 'p', text: 'Beacons combines a link-in-bio page with a store for digital products, a media kit for brand deals and email collection. If you are a creator who monetises directly, it replaces several tools. If you only need a clean list of links, it is more than you need.' },
      { type: 'h3', text: 'Stan Store: best for courses, coaching and downloads' },
      { type: 'p', text: 'Stan is a storefront first and a link page second. Visitors can buy a course, book a call or download a product without leaving the page. It suits coaches and educators. It has no free plan, so it only makes sense if you are selling.' },
      { type: 'h3', text: 'Carrd: best for a real one-page website' },
      { type: 'p', text: 'Carrd is a lightweight site builder. Many people use a Carrd page as their bio link because they get full layout control and custom domains. You give up the purpose-built link analytics and simplicity of a link-in-bio tool.' },
      { type: 'h3', text: 'Taplink: best for small businesses taking orders in chat' },
      { type: 'p', text: 'Taplink leans toward small businesses: messenger buttons (WhatsApp, Telegram), forms and simple payment blocks. It is a good fit if your sales happen in conversations rather than on a checkout page.' },
      { type: 'h3', text: 'Bitly: best if your team already lives in Bitly' },
      { type: 'p', text: 'Bitly offers link-in-bio pages alongside its short links and QR codes. For teams already paying for Bitly, keeping everything in one account is convenient. As a page builder it is fairly basic compared with dedicated tools.' },
      { type: 'h3', text: 'Trunc: best when the page, short link and QR code should live together' },
      { type: 'p', text: 'Trunc started as a URL shortener with first-party analytics, and Trunc Pages adds mobile pages built from blocks: digital business cards with a real Save Contact (vCard) button, link-in-bio, product showcases with buy buttons, and event pages with agenda, map, countdown and "Add to calendar". Every page gets a short link and a QR code, and the Track tab shows views, saves, taps and whether visitors came from a QR scan or a social profile, without third-party trackers. Pages are server-rendered and work without JavaScript, so they load quickly on slow mobile connections. The free plan includes one page; Pro adds up to 10 pages and contact or newsletter forms; Ultra is unlimited and removes the badge. The trade-off is a smaller integration ecosystem than Linktree\'s.' },

      { type: 'h2', text: 'Which Linktree Alternative Should You Choose?' },
      { type: 'table', headers: ['If you are…', 'Look at'], rows: [
        ['A creator selling digital products', 'Beacons or Stan Store'],
        ['A professional who networks in person', 'Trunc (digital business card + QR)'],
        ['A small shop taking orders on WhatsApp', 'Taplink or Trunc'],
        ['An event organiser', 'Trunc (event page with calendar file + QR)'],
        ['Someone who wants a full mini-website', 'Carrd'],
        ['A marketing team standardised on one link tool', 'Bitly or Trunc'],
        ['Happy with Linktree and just want links', 'Stay on Linktree'],
      ]},

      { type: 'h2', text: 'How to Switch from Linktree Without Losing Traffic' },
      { type: 'ol', items: [
        'Build your new page and copy your links across, putting the most important one first.',
        'Publish and test the page on your own phone, including every button.',
        'Replace the link in each social bio (Instagram, TikTok, YouTube, X, LinkedIn).',
        'Keep your old Linktree live for a few weeks with a single link to the new page, so old shares and printed codes still work.',
        'Compare clicks in your new analytics against your old numbers before closing the old account.',
      ]},

      { type: 'h2', text: 'Frequently Asked Questions' },
      { type: 'h3', text: 'What is the best free Linktree alternative?' },
      { type: 'p', text: 'It depends on what you need for free. For a simple link list with analytics, several tools have workable free plans. If you also want a QR code and a digital business card, Trunc\'s free plan covers one page with both.' },
      { type: 'h3', text: 'Does a link-in-bio tool affect my Instagram reach?' },
      { type: 'p', text: 'No. Instagram treats the bio link the same regardless of which tool hosts it. What matters is whether visitors find what they came for quickly.' },

      { type: 'cta', text: 'Try Trunc Pages free: one page, a short link, a QR code and first-party analytics.', href: 'https://trunc.site/register', label: 'Build My Page Free' },
      { type: 'cta', text: 'Selling through WhatsApp or Instagram DMs?', href: '/blog/link-in-bio-for-small-business', label: 'Read: Link in Bio for Small Business' },
    ],
  },

  {
    slug: 'link-in-bio-for-small-business',
    title: 'Link in Bio for Small Business: Turn Followers into WhatsApp Orders',
    seoTitle: 'Link in Bio for Small Business (WhatsApp Orders)',
    description:
      'Set up a link in bio page for your small business: products, prices and a WhatsApp order button in one link, plus wa.me setup and a QR code for your shop.',
    publishedAt: '2026-11-18',
    updatedAt: '2026-11-18',
    readTime: 7,
    category: 'Guide',
    image: '/blog/link-in-bio-for-small-business.webp',
    imageAlt: 'A small shop link-in-bio page on a phone with product cards and an Order on WhatsApp button',
    author: 'Alex Rivera',
    authorSlug: 'alex-rivera',
    content: [
      { type: 'p', text: 'If your small business sells through Instagram, Facebook or TikTok, your bio link is your shop window. Most people who tap it are on a phone, have a few seconds of patience, and want three answers: what do you sell, how much is it, and how do I order? A good link-in-bio page answers all three on the first screen and ends with one button: order on WhatsApp.' },

      { type: 'h2', text: 'What a Small Business Bio Page Needs' },
      { type: 'table', headers: ['Section', 'Why it matters', 'Example'], rows: [
        ['Logo and one-line description', 'Visitors know they are in the right place', '"Handmade soy candles, delivered island-wide"'],
        ['Order on WhatsApp button', 'The main action, at the top', 'Opens a chat with a pre-filled message'],
        ['3–6 products with prices', 'Answers "how much?" without a DM', 'Product cards with photo, price, Buy button'],
        ['Opening hours and location', 'Builds trust, helps local customers', 'Map with directions'],
        ['Delivery and payment info', 'Removes the most common questions', '"Cash on delivery and bank transfer"'],
        ['Social links', 'Lets people follow you elsewhere', 'Instagram, Facebook, TikTok icons'],
      ]},

      { type: 'h2', text: 'How to Create a WhatsApp Order Link (wa.me)' },
      { type: 'p', text: 'WhatsApp supports "click to chat" links that open a conversation with your number. The format is https://wa.me/<number>?text=<message>. Use your full number in international format without the plus sign, spaces or leading zero.' },
      { type: 'ul', items: [
        'Phone number +94 77 123 4567 becomes 94771234567.',
        'Basic link: https://wa.me/94771234567',
        'With a pre-filled message: https://wa.me/94771234567?text=Hi%2C%20I%27d%20like%20to%20order',
        'Spaces become %20, commas %2C and apostrophes %27 (URL encoding).',
      ]},
      { type: 'callout', text: 'A pre-filled message like "Hi, I\'d like to order the lavender candle" saves your customer typing and tells you instantly what they want. Use a different message per product so you know which product brought each chat.' },

      { type: 'h2', text: 'Build Your Shop Page with Trunc (Step by Step)' },
      { type: 'ol', items: [
        'Create a free account at trunc.site/register and open Pages.',
        'Choose the Product Showcase template (or Link-in-Bio if you sell services).',
        'Add your logo and a one-line description in the profile block.',
        'Add a contact button for WhatsApp with your number, and place it at the top.',
        'Add product cards or a product carousel: photo, name, price and a buy link. The buy link can be your wa.me link with a product-specific message, or your online store.',
        'Add a map block with your shop location if customers can visit, and a text block with delivery and payment details.',
        'Publish, then paste the page link into your Instagram, Facebook and TikTok bios and your WhatsApp Business profile.',
      ]},

      { type: 'h2', text: 'Use the Same Page Offline with a QR Code' },
      { type: 'p', text: 'Your bio page doesn\'t have to stay online. Download the page\'s QR code from the Trunc editor and print it on packaging, receipts, flyers, your shop counter and market-stall signage. Customers who scan it land on the same page and can reorder on WhatsApp later. Because the QR code is tagged, the Track tab shows how many visits came from scans versus social profiles, which tells you whether your printed materials are earning their cost.' },

      { type: 'h2', text: 'What to Measure' },
      { type: 'ul', items: [
        'Page views by source: Instagram vs. TikTok vs. QR scans. Put more effort into the channel that sends people.',
        'Taps on the WhatsApp button: this is your real conversion rate from visit to conversation.',
        'Taps per product: which items get interest. Move the best performers to the top.',
        'Views by day: line up social posts and promotions with the days your page is busiest.',
      ]},

      { type: 'h2', text: 'Common Mistakes' },
      { type: 'ul', items: [
        'No prices. "DM for price" loses customers who won\'t message a stranger to find out.',
        'Too many links. A page with 20 buttons hides the one that matters.',
        'Wrong number format in the wa.me link, which opens WhatsApp with an error. Always test it.',
        'Forgetting to update stock. Remove sold-out items or mark them clearly.',
      ]},

      { type: 'h2', text: 'Frequently Asked Questions' },
      { type: 'h3', text: 'Can I use a normal WhatsApp number, or do I need WhatsApp Business?' },
      { type: 'p', text: 'wa.me links work with both. WhatsApp Business adds a catalogue, quick replies and a business profile, which help once orders pick up.' },
      { type: 'h3', text: 'Do I need a website as well?' },
      { type: 'p', text: 'Not to start. Many small businesses run entirely on a bio page plus WhatsApp. When you outgrow it, link the page to your online store.' },
      { type: 'h3', text: 'Is it free?' },
      { type: 'p', text: 'Trunc\'s free plan includes one published page with products, WhatsApp buttons, a map, a QR code and analytics. Pro adds more pages and contact or newsletter forms.' },

      { type: 'cta', text: 'Put your shop on one link, with orders straight to WhatsApp. Free to start.', href: 'https://trunc.site/register', label: 'Build My Shop Page' },
      { type: 'cta', text: 'Comparing tools first?', href: '/blog/best-linktree-alternatives', label: 'Read: Best Linktree Alternatives' },
    ],
  },
];
