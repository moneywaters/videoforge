import { test, expect } from '@playwright/test';

// Test credentials - can be set via env vars or use test.skip conditionally
const TEST_EMAIL = process.env.TEST_EMAIL || '';
const TEST_PASSWORD = process.env.TEST_PASSWORD || '';

test.describe('Brief File Preview', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to dashboard
    await page.goto('/dashboard');

    // If redirected to sign-in, try to handle login
    if (page.url().includes('/auth/sign-in')) {
      // Skip if no credentials available
      if (!TEST_EMAIL || !TEST_PASSWORD) {
        test.skip(true, 'No test credentials available');
        return;
      }
      await completeLogin(page, TEST_EMAIL, TEST_PASSWORD);
    }
  });

  test('double-clicking image file opens preview modal', async ({ page }) => {
    // Navigate to briefs page
    await page.goto('/dashboard/briefs');

    // Wait for content
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);

    // Check for empty state
    const hasNoBriefs = await page.locator('text=No briefs yet').isVisible().catch(() => false);
    if (hasNoBriefs) {
      test.skip(true, 'No briefs available - need test data');
      return;
    }

    // Get first brief link
    const briefLink = page.locator('a[href^="/dashboard/briefs/"]').first();
    const briefCount = await briefLink.count();

    if (briefCount === 0) {
      test.skip(true, 'No briefs found');
      return;
    }

    // Click on first brief
    await briefLink.click();
    await page.waitForURL(/\/dashboard\/briefs\/.+/);
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(500);

    // Check for file explorer - look for the file items
    // File items have role="button" with tabIndex=0
    const fileItems = page.locator('[role="button"][tabIndex="0"]');
    const itemCount = await fileItems.count();

    if (itemCount === 0) {
      test.skip(true, 'No files in file explorer - need test data');
      return;
    }

    // Get the first item and check if it's a media file (has an image)
    const firstItem = fileItems.first();
    const hasThumbnail = await firstItem.locator('img').count() > 0;

    if (!hasThumbnail) {
      // Check if it's a media type by looking at the file type indicator
      // For now, we'll try double-clicking anyway - the component handles non-media gracefully
      test.skip(true, 'First file is not a media file (image/video)');
      return;
    }

    // Double-click to open preview
    await firstItem.dblclick();
    await page.waitForTimeout(300);

    // Verify modal opens - check for modal header text
    const modalHeader = page.locator('h3:has-text("Preview")');
    await expect(modalHeader).toBeVisible({ timeout: 5000 });

    // Verify modal content - should have either image or video
    const modalContent = page.locator('div.fixed.inset-0 img, div.fixed.inset-0 video');
    await expect(modalContent).toBeVisible({ timeout: 3000 });
  });

  test('closing preview modal via X button', async ({ page }) => {
    await page.goto('/dashboard/briefs');

    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);

    const hasNoBriefs = await page.locator('text=No briefs yet').isVisible().catch(() => false);
    if (hasNoBriefs) {
      test.skip(true, 'No briefs available');
      return;
    }

    const briefLink = page.locator('a[href^="/dashboard/briefs/"]').first();
    if (await briefLink.count() === 0) {
      test.skip(true, 'No briefs found');
      return;
    }

    await briefLink.click();
    await page.waitForURL(/\/dashboard\/briefs\/.+/);
    await page.waitForLoadState('domcontentloaded');

    const fileItems = page.locator('[role="button"][tabIndex="0"]');
    if (await fileItems.count() === 0) {
      test.skip(true, 'No files in file explorer');
      return;
    }

    const firstItem = fileItems.first();
    const hasThumbnail = await firstItem.locator('img').count() > 0;
    if (!hasThumbnail) {
      test.skip(true, 'First file is not a media file');
      return;
    }

    // Double-click to open preview
    await firstItem.dblclick();
    await page.waitForTimeout(300);

    // Verify modal is visible
    const modalHeader = page.locator('h3:has-text("Preview")');
    await expect(modalHeader).toBeVisible({ timeout: 5000 });

    // Click the close button (X icon)
    const closeButton = page.locator('button[aria-label="Close preview"]');
    await closeButton.click();
    await page.waitForTimeout(300);

    // Verify modal is closed
    await expect(modalHeader).not.toBeVisible({ timeout: 3000 });
  });

  test('single-click selects file but does not open preview', async ({ page }) => {
    await page.goto('/dashboard/briefs');

    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);

    const hasNoBriefs = await page.locator('text=No briefs yet').isVisible().catch(() => false);
    if (hasNoBriefs) {
      test.skip(true, 'No briefs available');
      return;
    }

    const briefLink = page.locator('a[href^="/dashboard/briefs/"]').first();
    if (await briefLink.count() === 0) {
      test.skip(true, 'No briefs found');
      return;
    }

    await briefLink.click();
    await page.waitForURL(/\/dashboard\/briefs\/.+/);
    await page.waitForLoadState('domcontentloaded');

    const fileItems = page.locator('[role="button"][tabIndex="0"]');
    if (await fileItems.count() === 0) {
      test.skip(true, 'No files in file explorer');
      return;
    }

    // Single click should NOT open modal
    await fileItems.first().click();
    await page.waitForTimeout(300);

    // Verify no modal is visible after single click
    const modalVisible = await page.locator('h3:has-text("Preview")').isVisible().catch(() => false);
    expect(modalVisible).toBe(false);
  });

  test('clicking outside modal closes it', async ({ page }) => {
    await page.goto('/dashboard/briefs');

    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);

    const hasNoBriefs = await page.locator('text=No briefs yet').isVisible().catch(() => false);
    if (hasNoBriefs) {
      test.skip(true, 'No briefs available');
      return;
    }

    const briefLink = page.locator('a[href^="/dashboard/briefs/"]').first();
    if (await briefLink.count() === 0) {
      test.skip(true, 'No briefs found');
      return;
    }

    await briefLink.click();
    await page.waitForURL(/\/dashboard\/briefs\/.+/);
    await page.waitForLoadState('domcontentloaded');

    const fileItems = page.locator('[role="button"][tabIndex="0"]');
    if (await fileItems.count() === 0) {
      test.skip(true, 'No files in file explorer');
      return;
    }

    const firstItem = fileItems.first();
    const hasThumbnail = await firstItem.locator('img').count() > 0;
    if (!hasThumbnail) {
      test.skip(true, 'First file is not a media file');
      return;
    }

    // Double-click to open preview
    await firstItem.dblclick();
    await page.waitForTimeout(300);

    // Verify modal is open
    const modal = page.locator('div.fixed.inset-0.bg-black\\/80');
    await expect(modal).toBeVisible({ timeout: 5000 });

    // Click outside the modal content (on the backdrop)
    // The backdrop is the div with bg-black/80
    await modal.click({ position: { x: 10, y: 10 } });
    await page.waitForTimeout(300);

    // Verify modal is closed
    await expect(modal).not.toBeVisible({ timeout: 3000 });
  });
});

async function completeLogin(page: any, email: string, password: string) {
  // Fill in login form
  const emailInput = page.locator('input[name="email"], input[type="email"], input[id="email"]');
  await emailInput.fill(email);

  const passwordInput = page.locator('input[name="password"], input[type="password"], input[id="password"]');
  await passwordInput.fill(password);

  // Click sign in button
  const submitButton = page.locator('button[type="submit"], button:has-text("Sign in"), button:has-text("Sign In")');
  await submitButton.click();

  // Wait for navigation
  await page.waitForTimeout(2000);
}