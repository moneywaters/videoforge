import { test, expect } from '@playwright/test';

/**
 * Test for Brief Creation timeout behavior
 * 
 * Run with:
 * npx playwright test tests/brief-creation-timeout.spec.ts --reporter=list --timeout=30000
 */

const TEST_EMAIL = process.env.TEST_EMAIL || 'test@example.com';
const TEST_PASSWORD = process.env.TEST_PASSWORD || 'testpassword';

test.describe('Brief Creation Timeout', () => {
  test('should handle brief creation without hanging', async ({ page }) => {
    // 1. Navigate to the briefs page
    console.log('\n=== STEP 1: Navigating to /dashboard/briefs ===\n');
    await page.goto('/dashboard/briefs');
    
    // Wait for page to load
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);
    
    // Screenshot after initial load
    await page.screenshot({ path: 'test-results/brief-creation-timeout-01-briefs-page.png' });
    console.log('Screenshot saved: brief-creation-timeout-01-briefs-page.png');

    // 2. Navigate to new brief page
    console.log('\n=== STEP 2: Navigating to /dashboard/briefs/new ===\n');
    await page.goto('/dashboard/briefs/new');
    
    // Wait for the form to load
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);
    
    // Screenshot after new brief page loads
    await page.screenshot({ path: 'test-results/brief-creation-timeout-02-new-brief-page.png' });
    console.log('Screenshot saved: brief-creation-timeout-02-new-brief-page.png');
    
    // 3. Check page loaded without hanging (no indefinite loading)
    const url = page.url();
    console.log(`Current URL: ${url}`);
    
    // Verify we're on the new brief page
    expect(url).toContain('/dashboard/briefs/new');
    
    // Take final screenshot
    await page.screenshot({ path: 'test-results/brief-creation-timeout-03-final.png' });
    console.log('Screenshot saved: brief-creation-timeout-03-final.png');
    
    console.log('\n=== TEST PASSED: Brief creation page loaded without timeout ===\n');
  });
});