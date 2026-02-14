# CMS Migration Implementation Guide

This guide details the implementation plan for migrating hardcoded frontend content to the CMS system.

## 📋 Overview

We are migrating hardcoded content from the frontend pages (homepage, about, contact) to the CMS database, allowing content to be managed through the admin panel without code changes.

## ✅ Completed Steps

### 1. Migration Files Created

- ✅ `configs/migrations/028_populate_cms_content_up.sql` - Populates CMS with real content
- ✅ `configs/migrations/028_populate_cms_content_down.sql` - Rollback migration

**Content Migrated:**
- Homepage (hero, features, social proof stats, CTA)
- About page (hero, stats, mission, values, team, achievements, CTA)
- Contact page (hero, contact methods, form configuration, support resources, CTA)
- 6 reusable content blocks (hero sections, features grid, social proof, CTA)

### 2. React Hook Created

- ✅ `frontend/src/hooks/useCMSPage.ts` - Custom hook for fetching CMS content with fallback support
- Includes helper functions: `getContentBlock()`, `getCustomField()`
- Handles errors gracefully and falls back to hardcoded content

### 3. Homepage Updated

- ✅ `frontend/src/app/page.tsx` - Now fetches from CMS
- Integrated `useCMSPage` hook
- All sections use CMS data with fallback
- Development indicator shows when CMS is active

## 🔄 Remaining Steps

### Step 1: Apply Migration

**Option A: Using Docker (Recommended)**

```bash
# Start Docker services
docker-compose -f docker/docker-compose.local.yml up -d

# Run migration
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -f /app/configs/migrations/028_populate_cms_content_up.sql

# Verify migration
docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db -c "SELECT slug, title FROM static_pages WHERE slug IN ('home', 'about', 'contact');"
```

**Option B: Using PowerShell Script**

```powershell
# Apply migration using the migration script
.\scripts\migrate.ps1 up 028

# Verify
.\scripts\migrate.ps1 status
```

**Option C: Direct psql (if PostgreSQL installed locally)**

```bash
psql -h localhost -U urlshortener -d urlshortener_db -f configs/migrations/028_populate_cms_content_up.sql
```

### Step 2: Complete About Page Integration

The About page has been partially updated. You need to update the remaining JSX sections:

**File:** `frontend/src/app/about/page.tsx`

**Sections to Update:**

1. **Mission Section** (lines ~101-152)
   - Replace hardcoded text with `mission.heading`, `mission.description`
   - Use `mission.cta_primary` and `mission.cta_secondary` for buttons

2. **Values Section** (lines ~156-199)
   - Replace hardcoded heading/tagline with `valuesBlock.heading`, `valuesBlock.tagline`
   - The `values` array is already mapped from CMS

3. **Team Section** (lines ~203-232)
   - Use `team.heading`, `team.description`, `team.cta`

4. **Achievements Section** (lines ~236-263)
   - Use `achievementsBlock.heading`
   - The `achievements` array is already loaded from CMS

5. **CTA Section** (lines ~268-331)
   - Use `ctaBottom` object (same as homepage)

6. **Add Development Indicator** (at end of file)
   ```tsx
   {/* Development Indicator */}
   {process.env.NODE_ENV === 'development' && usedCMS && (
     <div className="fixed bottom-4 right-4 bg-green-500 text-white px-3 py-1 rounded-full text-xs font-medium shadow-lg">
       ✓ Using CMS
     </div>
   )}
   ```

### Step 3: Update Contact Page

**File:** `frontend/src/app/contact/page.tsx`

**Required Changes:**

1. **Add imports:**
   ```tsx
   import { useCMSPage, getContentBlock, getCustomField } from '@/hooks/useCMSPage';
   ```

2. **Fetch CMS content:**
   ```tsx
   const { content: page, loading, usedCMS } = useCMSPage('contact', {
     slug: 'contact',
     title: 'Get in Touch',
     content: 'Ready to transform your link management?',
   });

   const hero = getContentBlock(page, 'hero', { /* fallback */ });
   const contactMethods = getContentBlock(page, 'contact_methods', { /* fallback */ });
   const contactForm = getContentBlock(page, 'contact_form', { /* fallback */ });
   const supportResources = getContentBlock(page, 'support_resources', { /* fallback */ });
   const ctaBottom = getContentBlock(page, 'cta_bottom', { /* fallback */ });

   const inquiryTypes = getCustomField(page, 'inquiry_types', [/* fallback */]);
   const priorityLevels = getCustomField(page, 'priority_levels', [/* fallback */]);
   ```

3. **Update all JSX sections** to use extracted content blocks

4. **Add loading state** (same as homepage/about)

5. **Add development indicator**

### Step 4: Build Admin Panel CMS Editor (Optional - Future Enhancement)

This step is optional for the initial migration. You can manage content through direct database updates or API calls for now.

**Future Components to Create:**

1. **PageList Component** (`frontend/src/app/admin/components/cms/PageList.tsx`)
   - Data table showing all pages
   - Edit/Delete/Preview actions
   - Filters and search

2. **PageEditor Component** (`frontend/src/app/admin/components/cms/PageEditor.tsx`)
   - Rich text editor for content
   - JSON editor for content_blocks
   - Meta tag fields (title, description, keywords)
   - Publishing options (draft, scheduled, published)

3. **ContentBlockEditor Component** (`frontend/src/app/admin/components/cms/ContentBlockEditor.tsx`)
   - Editor for reusable content blocks
   - Type selector (hero, CTA, features, etc.)
   - Active/Inactive toggle

4. **Integration with Admin Panel**
   - Add "Content Management" tab to admin dashboard
   - Use existing CMS API endpoints

### Step 5: Testing

**1. Verify Migration:**
```sql
-- Check pages
SELECT slug, title, is_published
FROM static_pages
WHERE slug IN ('home', 'about', 'contact');

-- Check content blocks
SELECT name, identifier, type, is_active
FROM content_blocks
WHERE identifier IN ('hero_home', 'features_grid', 'social_proof_stats', 'cta_bottom', 'hero_about', 'hero_contact');
```

**2. Test Frontend Pages:**
```bash
cd frontend
npm run dev
```

Visit:
- `http://localhost:3000/` - Homepage should look identical
- `http://localhost:3000/about` - About page should look identical
- `http://localhost:3000/contact` - Contact page should look identical

**3. Check CMS Integration:**
- Look for green "✓ Using CMS" badge in bottom-right (development only)
- Open browser console - should see no errors
- Check Network tab - should see successful `/api/v1/pages/{slug}` requests

**4. Test Fallback:**
- Stop backend server temporarily
- Refresh frontend pages
- Pages should still display (using fallback content)
- Restart backend - pages should switch back to CMS content

**5. Test Admin API (Optional):**
```bash
# Get authentication token first
TOKEN="your-jwt-token-here"

# List all pages
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/admin/cms/pages

# Update a page
curl -X PUT \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title": "New Title", "content": "New content..."}' \
  http://localhost:8080/api/v1/admin/cms/pages/1

# List content blocks
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/admin/content-blocks
```

## 📚 Key Files Reference

### Backend (Complete)
- ✅ `internal/handlers/cms_handlers.go` - CMS API endpoints
- ✅ `internal/services/cms.go` - CMS business logic
- ✅ `internal/models/cms.go` - CMS data models
- ✅ `configs/migrations/017_cms_schema_up.sql` - CMS database schema
- ✅ `configs/migrations/028_populate_cms_content_up.sql` - Content migration

### Frontend (Partial)
- ✅ `frontend/src/hooks/useCMSPage.ts` - CMS fetch hook
- ✅ `frontend/src/app/page.tsx` - Homepage (complete)
- ⏳ `frontend/src/app/about/page.tsx` - About page (partial)
- ⏳ `frontend/src/app/contact/page.tsx` - Contact page (not started)
- ⏳ Admin CMS components (future enhancement)

## 🎯 Success Criteria

- [ ] Migration 028 applied successfully
- [ ] All 3 pages (home, about, contact) in database with real content
- [ ] 6 content blocks created and active
- [ ] Homepage fetches from CMS and displays correctly
- [ ] About page fetches from CMS and displays correctly
- [ ] Contact page fetches from CMS and displays correctly
- [ ] Green "✓ Using CMS" badge shows in development
- [ ] No errors in browser console
- [ ] Pages work with backend offline (fallback mode)
- [ ] SEO meta tags properly set from CMS

## 🔧 Troubleshooting

### Migration Fails

**Issue:** Migration fails with "relation already exists"
**Solution:** The migration uses `ON CONFLICT DO UPDATE` to handle existing records. Check if migration 017 (CMS schema) was applied first.

### CMS API Returns 404

**Issue:** `/api/v1/pages/home` returns 404
**Solution:**
1. Check if migration 028 was applied: `SELECT COUNT(*) FROM static_pages WHERE slug='home';`
2. Verify backend routes are registered in `cmd/server/main.go`
3. Check backend logs for errors

### Frontend Shows Fallback Content

**Issue:** Pages always show fallback content, never use CMS
**Solution:**
1. Check browser console for fetch errors
2. Verify backend is running: `http://localhost:8080/health`
3. Check network tab - is the CMS API being called?
4. Verify API response structure matches expected format

### Content Blocks Not Displaying

**Issue:** Content blocks exist in DB but don't show on page
**Solution:**
1. Check `content_blocks.is_active = true`
2. Verify JSONB structure matches expected format
3. Check `getContentBlock()` path parameter is correct
4. Inspect page object in browser DevTools

## 📖 Next Steps (Future Enhancements)

1. **Admin Panel UI**
   - Build PageList, PageEditor, ContentBlockEditor components
   - Add rich text editor (TinyMCE or Tiptap)
   - Add image upload and media library

2. **Additional Pages**
   - Migrate pricing FAQs to CMS
   - Migrate help/docs pages
   - Create blog CMS integration

3. **Advanced Features**
   - Page versioning/revisions UI
   - Content scheduling (publish/unpublish dates)
   - Multi-language support
   - Content approval workflow

4. **Performance Optimization**
   - Add caching layer (Redis) for frequently accessed pages
   - Implement ISR (Incremental Static Regeneration) in Next.js
   - Add CDN for static assets

## 📞 Support

If you encounter issues:
1. Check this guide's Troubleshooting section
2. Review backend logs: `docker logs urlshortener-backend`
3. Check database directly: `docker exec -it urlshortener-postgres psql -U urlshortener -d urlshortener_db`
4. Review migration files for correct SQL syntax

---

**Last Updated:** 2026-02-14
**Migration Version:** 028
**Status:** In Progress (Homepage ✅, About ⏳, Contact ⏳)
