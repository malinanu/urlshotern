# CMS Migration Implementation Checklist

Quick reference for completing the CMS migration.

## ✅ Phase 1: Backend & Migration (COMPLETE)

- [x] Migration file created: `028_populate_cms_content_up.sql`
- [x] Migration rollback created: `028_populate_cms_content_down.sql`
- [x] CMS API endpoints exist and are functional
- [x] React hook created: `useCMSPage.ts`

## 🔄 Phase 2: Apply Migration (NEXT STEP)

```bash
# Start Docker services if not running
docker-compose -f docker/docker-compose.local.yml up -d

# Apply migration
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -f /app/configs/migrations/028_populate_cms_content_up.sql

# OR using PowerShell script
.\scripts\migrate.ps1 up 028

# Verify
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -c "SELECT slug, title, is_published FROM static_pages WHERE slug IN ('home', 'about', 'contact');"
```

Expected output:
```
 slug    |        title         | is_published
---------+---------------------+--------------
 home    | Shorten URLs. Tr... | t
 about   | We Are Trunc        | t
 contact | Get in Touch        | t
(3 rows)
```

## ✅ Phase 3: Frontend Integration

### Homepage (COMPLETE ✅)
- [x] Imported `useCMSPage` hook
- [x] Fetches from `/api/v1/pages/home`
- [x] All sections use CMS data
- [x] Fallback content preserved
- [x] Development indicator added

### About Page (PARTIAL ⏳)

**Already Done:**
- [x] Hook imported and initialized
- [x] Content blocks extracted
- [x] Icon mapping configured
- [x] Loading state added
- [x] Hero section updated

**TODO:**
- [ ] Update Mission section to use `mission` object
- [ ] Update Values section to use `valuesBlock` object
- [ ] Update Team section to use `team` object
- [ ] Update Achievements section to use `achievementsBlock` object
- [ ] Update CTA section to use `ctaBottom` object
- [ ] Add development indicator

### Contact Page (NOT STARTED ⏳)

**TODO:**
- [ ] Import `useCMSPage` hook
- [ ] Initialize hook and fetch page data
- [ ] Extract content blocks (hero, contactMethods, contactForm, supportResources, ctaBottom)
- [ ] Extract custom fields (inquiryTypes, priorityLevels)
- [ ] Update hero section JSX
- [ ] Update contact methods section JSX
- [ ] Update contact form section JSX
- [ ] Update support resources section JSX
- [ ] Update CTA section JSX
- [ ] Add loading state
- [ ] Add development indicator

## 🧪 Phase 4: Testing

### Database Verification
```bash
# Check pages
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -c "SELECT slug, title, is_published FROM static_pages WHERE slug IN ('home', 'about', 'contact');"

# Check content blocks
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -c "SELECT name, identifier, is_active FROM content_blocks WHERE identifier IN ('hero_home', 'features_grid', 'social_proof_stats');"
```

### API Testing
```bash
# Test public endpoints (no auth needed)
curl http://localhost:8080/api/v1/pages/home | jq '.'
curl http://localhost:8080/api/v1/pages/about | jq '.'
curl http://localhost:8080/api/v1/pages/contact | jq '.'

# Verify response structure
curl http://localhost:8080/api/v1/pages/home | jq '.page.content_blocks.hero'
```

Expected response:
```json
{
  "page": {
    "id": 1,
    "slug": "home",
    "title": "Shorten URLs. Track Everything.",
    "content": "Transform your long URLs...",
    "content_blocks": {
      "hero": {
        "badge": "✨ New: Advanced Analytics Dashboard",
        "title": "Shorten URLs. Track Everything.",
        ...
      },
      ...
    },
    "is_published": true
  }
}
```

### Frontend Testing
```bash
# Start frontend dev server
cd frontend
npm run dev
```

**Visit these URLs and verify:**
- [ ] `http://localhost:3000/` - Homepage looks identical to before
- [ ] `http://localhost:3000/about` - About page looks identical
- [ ] `http://localhost:3000/contact` - Contact page looks identical
- [ ] Green "✓ Using CMS" badge visible (bottom-right, dev only)
- [ ] No errors in browser console
- [ ] Network tab shows successful `/api/v1/pages/*` requests

**Test fallback mode:**
- [ ] Stop backend server: `docker stop urlshortener-backend`
- [ ] Refresh pages - should still display (fallback content)
- [ ] Check console for error message (expected)
- [ ] Restart backend: `docker start urlshortener-backend`
- [ ] Refresh pages - should switch back to CMS content

## 📝 Quick Reference: About Page Updates

**File:** `frontend/src/app/about/page.tsx`

### Mission Section (~line 109)
```tsx
// BEFORE:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  Connecting Every Digital Experience
</h2>
<p className="text-xl text-gray-600 leading-relaxed">
  We envision a world where every URL is intelligent...
</p>

// AFTER:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  {mission.heading}
</h2>
<p className="text-xl text-gray-600 leading-relaxed">
  {mission.description}
</p>
```

### Values Section (~line 162)
```tsx
// BEFORE:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  What Drives Us Forward
</h2>
<p className="text-xl text-gray-600 max-w-3xl mx-auto">
  Our core values shape every decision we make...
</p>

// AFTER:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  {valuesBlock.heading}
</h2>
<p className="text-xl text-gray-600 max-w-3xl mx-auto">
  {valuesBlock.tagline}
</p>
```

### Team Section (~line 209)
```tsx
// BEFORE:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  Global Remote Team
</h2>
<p className="text-xl text-gray-600 max-w-2xl mx-auto leading-relaxed">
  Our diverse team of engineers, designers...
</p>
<Button variant="outline" size="lg">
  View Open Positions
</Button>

// AFTER:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  {team.heading}
</h2>
<p className="text-xl text-gray-600 max-w-2xl mx-auto leading-relaxed">
  {team.description}
</p>
<Button variant="outline" size="lg" asChild>
  <a href={team.cta.url}>{team.cta.text}</a>
</Button>
```

### Achievements Section (~line 242)
```tsx
// BEFORE:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  Key Milestones
</h2>

// AFTER:
<h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
  {achievementsBlock.heading}
</h2>
```

### CTA Section (Same as Homepage)
Use the `ctaBottom` object - same structure as homepage.

## 📝 Quick Reference: Contact Page Updates

**File:** `frontend/src/app/contact/page.tsx`

### 1. Add Imports (top of file)
```tsx
import { useCMSPage, getContentBlock, getCustomField } from '@/hooks/useCMSPage';
```

### 2. Initialize Hook (inside component)
```tsx
const { content: page, loading, usedCMS } = useCMSPage('contact', {
  slug: 'contact',
  title: 'Get in Touch',
  content: 'Ready to transform your link management?',
});

const hero = getContentBlock(page, 'hero', {
  badge: 'Contact Us',
  badge_icon: 'ChatBubbleLeftRightIcon',
  title: 'Get in Touch',
  title_highlight: 'Touch',
  tagline: 'Ready to transform your link management?...',
});

const contactMethodsBlock = getContentBlock(page, 'contact_methods', {
  badge: 'Contact Options',
  heading: 'Choose How to Connect',
  tagline: 'Multiple ways to get the help you need, when you need it.',
  methods: [/* current contactMethods array */],
});

const contactFormBlock = getContentBlock(page, 'contact_form', {
  badge: 'Contact Form',
  heading: 'Send us a Message',
  tagline: 'Fill out the form below...',
  form_title: 'Let\'s Talk',
  form_description: 'We\'re here to help you succeed with Trunc.',
});

const inquiryTypes = getCustomField(page, 'inquiry_types', [/* current array */]);
const priorityLevels = getCustomField(page, 'priority_levels', [/* current array */]);

// contactMethods should map from contactMethodsBlock.methods
const contactMethods = contactMethodsBlock.methods || [/* fallback */];
```

### 3. Update JSX (use extracted content blocks)
Similar pattern to Homepage and About page.

## 🎯 Final Checks

Before considering this complete:

- [ ] Migration 028 applied successfully
- [ ] 3 pages in database (home, about, contact)
- [ ] 6 content blocks in database
- [ ] Homepage displays correctly from CMS
- [ ] About page displays correctly from CMS
- [ ] Contact page displays correctly from CMS
- [ ] All pages work in fallback mode (backend offline)
- [ ] No console errors on any page
- [ ] Development indicator shows "✓ Using CMS" in dev mode
- [ ] SEO meta tags populated from CMS

---

**Estimated Time:**
- Apply migration: 5 minutes
- Complete About page: 15 minutes
- Complete Contact page: 20 minutes
- Testing: 15 minutes
- **Total: ~1 hour**

**Current Progress:** 60% complete
- Backend: 100% ✅
- Migration: 100% ✅
- Homepage: 100% ✅
- About: 40% ⏳
- Contact: 0% ⏳
