import { test, expect } from '@playwright/test';

/**
 * Debug test for Brief Creation - captures console logs and network requests
 * 
 * Run with:
 * npx playwright test tests/brief-creation-debug.spec.ts --project=chromium
 */

const TEST_EMAIL = process.env.TEST_EMAIL || 'test@example.com';
const TEST_PASSWORD = process.env.TEST_PASSWORD || 'testpassword';

// Track pending requests to detect hangs
const pendingRequests = new Set<string>();
let requestTimeout: NodeJS.Timeout | null = null;

test.describe('Brief Creation Debug', () => {
  test('should debug brief creation stuck issue', async ({ page }) => {
    // Clear any previous timeout
    if (requestTimeout) clearTimeout(requestTimeout);

    // 1. Capture ALL console logs
    page.on('console', msg => {
      const type = msg.type();
      const text = msg.text();
      console.log(`[CONSOLE.${type.toUpperCase()}] ${text}`);
    });

    // 2. Capture all page errors
    page.on('pageerror', err => {
      console.log(`[PAGE ERROR] ${err.message}`);
    });

    // 3. Capture network requests - with timing to detect hangs
    page.on('request', req => {
      const url = req.url();
      const method = req.method();
      pendingRequests.add(url);
      console.log(`[REQUEST] ${method} ${url}`);
      
      // Set a timeout to detect hanging requests
      setTimeout(() => {
        if (pendingRequests.has(url)) {
          console.log(`[WARNING] Request still pending after 15s: ${method} ${url}`);
        }
      }, 15000);
    });

    // 4. Capture network responses
    page.on('response', res => {
      const url = res.url();
      const status = res.status();
      pendingRequests.delete(url);
      console.log(`[RESPONSE] ${status} ${res.statusText()} ${url}`);
    });

    // 5. Navigate to brief list first (to trigger the API call that hangs)
    console.log('\n=== STEP 1: Navigating to /dashboard/briefs ===\n');
    await page.goto('/dashboard/briefs');
    
    // Don't wait for networkidle - just DOM content loaded
    await page.waitForLoadState('domcontentloaded');
    console.log('DOM loaded, waiting 3s for API calls...');
    await page.waitForTimeout(3000);

    // Take screenshot showing loading state or empty state
    await page.screenshot({ path: 'test-results/debug-brief-01-briefs-page.png' });
    console.log('Screenshot saved: debug-brief-01-briefs-page.png');

    // Check what happened with the API call
    console.log(`\n=== Checking for hanging requests after 3s ===`);
    if (pendingRequests.size > 0) {
      console.log(`[CRITICAL] ${pendingRequests.size} requests still pending:`);
      pendingRequests.forEach(r => console.log(`  - ${r}`));
    }

    // 6. Now navigate to the NEW brief page
    console.log('\n=== STEP 2: Navigating to /dashboard/briefs/new ===\n');
    await page.goto('/dashboard/briefs/new');
    
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(2000);

    // Take screenshot
    await page.screenshot({ path: 'test-results/debug-brief-02-new-page.png' });
    console.log('Screenshot saved: debug-brief-02-new-page.png');

    // Check if we got stuck in loading spinner
    const loadingSpinner = page.locator('.animate-spin').first();
    const isLoading = await loadingSpinner.isVisible().catch(() => false);
    console.log(`\n=== Loading spinner visible: ${isLoading} ===\n`);

    if (isLoading) {
      console.log('\n!!! PAGE IS STUCK IN LOADING STATE !!!\n');
      
      // Capture what's in the page
      const pageContent = await page.content();
      console.log(`Page HTML length: ${pageContent.length} chars`);
      
      // Check for any error toasts
      const errorToast = page.locator('[role="alert"], .toast-error, [data-variant="error"]');
      const hasError = await errorToast.isVisible().catch(() => false);
      console.log(`Error toast visible: ${hasError}`);
      
      if (hasError) {
        const errorText = await errorToast.textContent().catch(() => 'N/A');
        console.log(`Error message: ${errorText}`);
      }
      
      await page.screenshot({ path: 'test-results/debug-brief-03-stuck-state.png' });
    }

    // 7. Fill in the form if NOT stuck
    if (!isLoading) {
      console.log('\n=== STEP 3: Filling brief form ===\n');
      
      const titleInput = page.locator('input#title');
      if (await titleInput.isVisible()) {
        await titleInput.fill('Test Brief Debug');
        console.log('Filled title: Test Brief Debug');
      }

      const descriptionInput = page.locator('textarea#description');
      if (await descriptionInput.isVisible()) {
        await descriptionInput.fill('This is a test brief created by Playwright debug');
        console.log('Filled description');
      }

      const budgetInput = page.locator('input#bountyBudget');
      if (await budgetInput.isVisible()) {
        await budgetInput.fill('1000');
        console.log('Filled budget: 1000');
      }

      // 8. Click "Start AI Interview" button
      console.log('\n=== STEP 4: Clicking Create Brief button ===\n');
      
      const createButton = page.locator('button:has-text("Start AI Interview"), button:has-text("Create Brief")').first();
      if (await createButton.isVisible()) {
        await createButton.click();
        console.log('Clicked Create button');
        
        // Wait for either success or failure
        console.log('\n=== Waiting for API response (15s timeout)... ===\n');
        
        try {
          await page.waitForURL(/\/dashboard\/briefs\/.+/, { timeout: 15000 });
          console.log('\n=== SUCCESS: Navigated to brief detail page ===\n');
        } catch (e) {
          console.log(`\n!!! TIMEOUT - Request likely hung: ${e} !!!\n`);
          console.log(`Current URL: ${page.url()}`);
          
          // Check for error toasts
          const toast = page.locator('[role="alert"]');
          if (await toast.isVisible().catch(() => false)) {
            console.log(`Error toast: ${await toast.textContent().catch(() => 'N/A')}`);
          }
        }
      }

      // Final screenshot
      await page.screenshot({ path: 'test-results/debug-brief-04-after-submit.png' });
    }

    // Final state summary
    console.log(`\n========================================`);
    console.log(`=== FINAL STATE ===`);
    console.log(`URL: ${page.url()}`);
    console.log(`Title: ${await page.title()}`);
    console.log(`Pending requests: ${pendingRequests.size}`);
    console.log(`========================================\n`);
  });
});