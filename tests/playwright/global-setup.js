const { chromium } = require('@playwright/test');

async function globalSetup(config) {
  const browser = await chromium.launch();
  const page = await browser.newPage();

  // Setup test database and environment
  console.log('Setting up test environment...');

  // You can add database migrations, seed data, etc. here
  try {
    // Example: Run database migrations for test environment
    // await page.goto(`${process.env.API_URL || 'http://localhost:8080'}/admin/migrate`);
    
    // Create admin user for tests if it doesn't exist
    const adminUser = {
      email: 'admin@test.com',
      password: 'testpassword123',
      firstName: 'Test',
      lastName: 'Admin',
      role: 'admin',
    };

    // Register admin user (this assumes your API supports user creation)
    try {
      await page.goto(`${process.env.BASE_URL || 'http://localhost:3000'}/register`);
      await page.fill('input[type="email"]', adminUser.email);
      await page.fill('input[type="password"]', adminUser.password);
      await page.fill('input[name="firstName"]', adminUser.firstName);
      await page.fill('input[name="lastName"]', adminUser.lastName);
      
      // Submit registration
      await page.click('button[type="submit"]');
      
      // Wait for success or handle existing user
      await page.waitForTimeout(2000);
      
      console.log('Admin user created or already exists');
    } catch (error) {
      console.log('Admin user might already exist or registration failed:', error.message);
    }

    // Create sample pages for testing
    const samplePages = [
      {
        title: 'Sample Page 1',
        slug: 'sample-page-1',
        content: '# Sample Page 1\n\nThis is a sample page for testing.',
        is_published: true,
      },
      {
        title: 'Draft Page',
        slug: 'draft-page',
        content: '# Draft Page\n\nThis is a draft page.',
        is_published: false,
      },
    ];

    // Login as admin to create sample data
    await page.goto(`${process.env.BASE_URL || 'http://localhost:3000'}/login`);
    await page.fill('input[type="email"]', adminUser.email);
    await page.fill('input[type="password"]', adminUser.password);
    await page.click('button[type="submit"]');
    
    // Wait for dashboard
    await page.waitForURL('**/dashboard', { timeout: 5000 });
    
    // Get auth token
    const authToken = await page.evaluate(() => localStorage.getItem('accessToken'));
    
    if (authToken) {
      // Create sample pages via API
      for (const samplePage of samplePages) {
        try {
          const response = await fetch(`${process.env.API_URL || 'http://localhost:8080'}/api/v1/admin/cms/pages`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${authToken}`,
            },
            body: JSON.stringify(samplePage),
          });
          
          if (response.ok) {
            console.log(`Created sample page: ${samplePage.title}`);
          }
        } catch (error) {
          console.log(`Failed to create sample page ${samplePage.title}:`, error.message);
        }
      }

      // Create sample global settings
      const sampleSettings = [
        {
          key: 'test_setting',
          value: 'test value',
          type: 'text',
          category: 'general',
          display_name: 'Test Setting',
          description: 'A test setting for Playwright tests',
          is_public: true,
          sort_order: 999,
        },
      ];

      for (const setting of sampleSettings) {
        try {
          const response = await fetch(`${process.env.API_URL || 'http://localhost:8080'}/api/v1/admin/settings`, {
            method: 'POST',
            headers: {
              'Content-Type': 'application/json',
              'Authorization': `Bearer ${authToken}`,
            },
            body: JSON.stringify(setting),
          });
          
          if (response.ok) {
            console.log(`Created sample setting: ${setting.key}`);
          }
        } catch (error) {
          console.log(`Failed to create sample setting ${setting.key}:`, error.message);
        }
      }
    }

  } catch (error) {
    console.log('Error during global setup:', error.message);
  }

  await browser.close();
  console.log('Global setup completed');
}

module.exports = globalSetup;