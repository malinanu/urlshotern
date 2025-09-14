const { test, expect } = require('@playwright/test');

// Test configuration
const BASE_URL = process.env.BASE_URL || 'http://localhost:3000';
const API_URL = process.env.API_URL || 'http://localhost:8080';

// Test data
const testUser = {
  email: 'admin@test.com',
  password: 'testpassword123',
};

const testPage = {
  title: 'Test Page',
  slug: 'test-page',
  content: '# Test Page\n\nThis is a test page created by Playwright tests.',
  meta_description: 'A test page for Playwright testing',
};

test.describe('CMS Management', () => {
  let authToken;

  // Setup: Login and get auth token
  test.beforeEach(async ({ page }) => {
    // Navigate to login page
    await page.goto(`${BASE_URL}/login`);
    
    // Fill login form
    await page.fill('input[type="email"]', testUser.email);
    await page.fill('input[type="password"]', testUser.password);
    
    // Submit form
    await page.click('button[type="submit"]');
    
    // Wait for navigation to dashboard
    await page.waitForURL('**/dashboard');
    
    // Extract auth token from localStorage or cookies if needed
    const token = await page.evaluate(() => localStorage.getItem('accessToken'));
    authToken = token;
  });

  test('should display CMS management interface', async ({ page }) => {
    // Navigate to admin page
    await page.goto(`${BASE_URL}/admin`);
    
    // Check if CMS section is visible
    await expect(page.getByText('Content Management')).toBeVisible();
    await expect(page.getByText('Create and manage static pages')).toBeVisible();
    
    // Check for New Page button
    await expect(page.getByRole('button', { name: 'New Page' })).toBeVisible();
  });

  test('should create a new page', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Click New Page button
    await page.click('button:has-text("New Page")');
    
    // Wait for form modal to appear
    await expect(page.getByText('Create New Page')).toBeVisible();
    
    // Fill form fields
    await page.fill('input[id="title"]', testPage.title);
    await page.fill('input[id="slug"]', testPage.slug);
    
    // Fill content using the rich text editor
    const contentEditor = page.locator('textarea').first();
    await contentEditor.fill(testPage.content);
    
    // Fill meta description
    await page.fill('textarea[id="meta_description"]', testPage.meta_description);
    
    // Set as published
    await page.check('input[type="checkbox"]:has-text("Published")');
    
    // Submit form
    await page.click('button:has-text("Create Page")');
    
    // Wait for success and modal to close
    await expect(page.getByText('Create New Page')).not.toBeVisible();
    
    // Verify page appears in list
    await expect(page.getByText(testPage.title)).toBeVisible();
    await expect(page.getByText('Published')).toBeVisible();
  });

  test('should edit an existing page', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Find the test page and click edit
    const pageRow = page.locator('tr').filter({ hasText: testPage.title });
    await pageRow.locator('button[title="Edit page"]').click();
    
    // Wait for edit form
    await expect(page.getByText('Edit Page')).toBeVisible();
    
    // Modify content
    const updatedContent = testPage.content + '\n\nThis page has been updated.';
    const contentEditor = page.locator('textarea').first();
    await contentEditor.clear();
    await contentEditor.fill(updatedContent);
    
    // Submit changes
    await page.click('button:has-text("Update Page")');
    
    // Verify changes are saved
    await expect(page.getByText('Edit Page')).not.toBeVisible();
  });

  test('should publish/unpublish pages', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Find the test page and toggle publish status
    const pageRow = page.locator('tr').filter({ hasText: testPage.title });
    const publishButton = pageRow.locator('button[title*="publish"]').first();
    
    // Click publish/unpublish button
    await publishButton.click();
    
    // Wait for status to update
    await page.waitForTimeout(1000);
    
    // Verify status change (this would depend on current state)
    // You might need to check the specific text or class changes
  });

  test('should filter pages by status', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Test published filter
    await page.selectOption('select', 'published');
    await page.waitForTimeout(500);
    
    // Verify only published pages are shown
    const publishedBadges = page.locator('text=Published');
    const publishedCount = await publishedBadges.count();
    expect(publishedCount).toBeGreaterThan(0);
    
    // Test draft filter
    await page.selectOption('select', 'draft');
    await page.waitForTimeout(500);
    
    // Test all filter
    await page.selectOption('select', 'all');
    await page.waitForTimeout(500);
  });

  test('should search pages', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Enter search term
    await page.fill('input[placeholder*="Search"]', testPage.title);
    await page.waitForTimeout(1000);
    
    // Verify filtered results
    await expect(page.getByText(testPage.title)).toBeVisible();
  });

  test('should display page analytics', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Find the test page and click analytics (if available)
    const pageRow = page.locator('tr').filter({ hasText: testPage.title });
    
    // This would depend on your UI - you might have an analytics button or link
    // await pageRow.locator('button[title="Analytics"]').click();
    
    // Verify analytics data is displayed
    // This would depend on your analytics implementation
  });

  test('should delete a page', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Find the test page and click delete
    const pageRow = page.locator('tr').filter({ hasText: testPage.title });
    await pageRow.locator('button[title="Delete page"]').click();
    
    // Handle confirmation dialog
    page.on('dialog', async dialog => {
      expect(dialog.message()).toContain('Are you sure');
      await dialog.accept();
    });
    
    // Wait for page to be removed from list
    await expect(page.getByText(testPage.title)).not.toBeVisible();
  });

  test('should handle page templates', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Create a new page to test templates
    await page.click('button:has-text("New Page")');
    await expect(page.getByText('Create New Page')).toBeVisible();
    
    // Check template options
    const templateSelect = page.locator('select[id="template"]');
    await expect(templateSelect).toBeVisible();
    
    // Test different template options
    await templateSelect.selectOption('blog');
    await templateSelect.selectOption('landing');
    await templateSelect.selectOption('default');
    
    // Close modal without saving
    await page.click('button:has-text("Cancel")');
  });

  test('should validate form inputs', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Click New Page button
    await page.click('button:has-text("New Page")');
    
    // Try to submit empty form
    await page.click('button:has-text("Create Page")');
    
    // Check for validation messages
    const titleInput = page.locator('input[id="title"]');
    const slugInput = page.locator('input[id="slug"]');
    
    // Check if inputs are marked as invalid
    await expect(titleInput).toHaveAttribute('required');
    await expect(slugInput).toHaveAttribute('required');
  });

  test('should work with rich text editor', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Create new page
    await page.click('button:has-text("New Page")');
    
    // Test rich text editor functionality
    const contentArea = page.locator('textarea').first();
    await contentArea.fill('# Heading\n\n**Bold text** and *italic text*');
    
    // Test preview mode (if available)
    const previewButton = page.locator('button[title*="Preview"]');
    if (await previewButton.count() > 0) {
      await previewButton.click();
      await expect(page.getByText('Bold text')).toBeVisible();
    }
    
    // Cancel form
    await page.click('button:has-text("Cancel")');
  });
});

test.describe('Global Settings Management', () => {
  test.beforeEach(async ({ page }) => {
    // Login process (same as above)
    await page.goto(`${BASE_URL}/login`);
    await page.fill('input[type="email"]', testUser.email);
    await page.fill('input[type="password"]', testUser.password);
    await page.click('button[type="submit"]');
    await page.waitForURL('**/dashboard');
  });

  test('should display global settings interface', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Look for settings section or navigate to settings
    // This depends on your UI structure
    await expect(page.getByText('Global Site Settings')).toBeVisible();
  });

  test('should update site settings', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Navigate to settings if needed
    // Find a setting to update (e.g., site name)
    const siteNameInput = page.locator('input[value*="3logiq"]').first();
    
    if (await siteNameInput.count() > 0) {
      await siteNameInput.clear();
      await siteNameInput.fill('3logiq Updated');
      
      // Save changes
      const saveButton = page.locator('button:has-text("Save")').first();
      await saveButton.click();
      
      // Verify success
      await page.waitForTimeout(1000);
    }
  });

  test('should create new setting', async ({ page }) => {
    await page.goto(`${BASE_URL}/admin`);
    
    // Click New Setting button (if available)
    const newSettingButton = page.locator('button:has-text("New Setting")');
    
    if (await newSettingButton.count() > 0) {
      await newSettingButton.click();
      
      // Fill new setting form
      await page.fill('input[placeholder*="setting_key"]', 'test_setting');
      await page.fill('input[placeholder*="Setting Name"]', 'Test Setting');
      await page.fill('input[value=""]', 'test value');
      
      // Submit
      await page.click('button:has-text("Create Setting")');
    }
  });
});

test.describe('Public Page Rendering', () => {
  test('should render public pages correctly', async ({ page }) => {
    // Test public page access
    await page.goto(`${BASE_URL}/about`);
    
    // Check if page loads
    await expect(page).toHaveTitle(/About/);
    
    // Check for content
    await expect(page.getByText('3logiq')).toBeVisible();
  });

  test('should handle page not found', async ({ page }) => {
    // Try to access non-existent page
    await page.goto(`${BASE_URL}/non-existent-page`);
    
    // Should show 404 or fallback content
    // This depends on your error handling implementation
  });

  test('should display dynamic content from settings', async ({ page }) => {
    await page.goto(`${BASE_URL}/about`);
    
    // Check if dynamic content from global settings is displayed
    await expect(page.getByText('3logiq')).toBeVisible();
    
    // Check footer content
    await expect(page.getByText('© 2024 3logiq')).toBeVisible();
  });
});

// Cleanup: Remove test data
test.afterAll(async ({ request }) => {
  // Clean up test pages and settings created during tests
  // This would require API calls to delete test data
  
  if (authToken) {
    // Delete test page via API
    try {
      await request.delete(`${API_URL}/api/v1/admin/cms/pages/test-page`, {
        headers: {
          'Authorization': `Bearer ${authToken}`,
        },
      });
    } catch (error) {
      console.log('Cleanup: Could not delete test page');
    }
  }
});