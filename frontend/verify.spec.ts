import { test, expect } from '@playwright/test';

test('final dashboard verification', async ({ page }) => {
  const results: string[] = [];
  const passed = (name: string) => results.push(`✅ ${name}`);
  const failed = (name: string, reason: string) => results.push(`❌ ${name} — ${reason}`);

  try {
    // Step 1: Navigate to http://localhost:3000/dashboard
    await page.goto('http://localhost:3000/dashboard');
    await page.waitForLoadState('networkidle');
    passed('Navigate to /dashboard');
  } catch (e) {
    failed('Navigate to /dashboard', (e as Error).message);
    return;
  }

  try {
    // Step 2: Inject mock user into localStorage
    await page.evaluate(() => {
      localStorage.setItem('auth-storage', JSON.stringify({
        state: {
          user: {
            id: 'test-user-id',
            email: 'test@example.com',
            name: 'Test User',
            role: 'client',
            createdAt: '2024-01-01T00:00:00.000Z',
            onboardingComplete: false
          }
        },
        version: 0
      }));
    });
    passed('Inject mock user into localStorage');
  } catch (e) {
    failed('Inject mock user into localStorage', (e as Error).message);
  }

  try {
    // Step 3: Reload the page
    await page.reload();
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    passed('Reload page for useEffect rehydration');
  } catch (e) {
    failed('Reload page for useEffect rehydration', (e as Error).message);
  }

  try {
    // Step 5: Take full-page screenshot
    await page.screenshot({ path: '/Users/asan/videoforge/frontend/playwright_final_dashboard.png', fullPage: true });
    passed('Take dashboard screenshot');
  } catch (e) {
    failed('Take dashboard screenshot', (e as Error).message);
  }

  try {
    // Step 6: Verify header with rounded-full button exists
    const header = await page.locator('header').first();
    const avatarButton = header.locator('button.rounded-full, button[class*="rounded-full"]').first();
    await expect(avatarButton).toBeVisible({ timeout: 5000 });
    passed('Verify header contains avatar trigger button');
  } catch (e) {
    failed('Verify header contains avatar trigger button', 'button.rounded-full not found in header');
  }

  try {
    // Step 7: Click the avatar trigger button
    await page.locator('header button.rounded-full, header button[class*="rounded-full"]').first().click();
    await page.waitForTimeout(500);
    passed('Click avatar trigger button');
  } catch (e) {
    failed('Click avatar trigger button', (e as Error).message);
  }

  try {
    // Step 8: Verify dropdown items
    const dropdownText = await page.textContent('body');
    const hasProfile = dropdownText?.includes('Profile');
    const hasAccount = dropdownText?.includes('Account');
    const hasBilling = dropdownText?.includes('Billing');
    const hasSettings = dropdownText?.includes('Settings');
    const hasSignOut = dropdownText?.includes('Sign out');

    if (hasProfile && hasAccount && hasBilling && hasSettings && hasSignOut) {
      passed('Verify dropdown contains all items');
    } else {
      const missing: string[] = [];
      if (!hasProfile) missing.push('Profile');
      if (!hasAccount) missing.push('Account');
      if (!hasBilling) missing.push('Billing');
      if (!hasSettings) missing.push('Settings');
      if (!hasSignOut) missing.push('Sign out');
      failed('Verify dropdown contains all items', `Missing: ${missing.join(', ')}`);
    }
  } catch (e) {
    failed('Verify dropdown contains all items', (e as Error).message);
  }

  try {
    // Step 9: Navigate to /dashboard/account
    await page.goto('http://localhost:3000/dashboard/account');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    passed('Navigate to /dashboard/account');
  } catch (e) {
    failed('Navigate to /dashboard/account', (e as Error).message);
  }

  try {
    // Step 11: Take account screenshot
    await page.screenshot({ path: '/Users/asan/videoforge/frontend/playwright_final_account.png', fullPage: true });
    passed('Take account screenshot');
  } catch (e) {
    failed('Take account screenshot', (e as Error).message);
  }

  try {
    // Step 12: Verify account page text
    const pageText = await page.textContent('body');
    const hasProfileInfo = pageText?.includes('Profile Information');
    const hasChangePassword = pageText?.includes('Change Password');
    const hasPreferences = pageText?.includes('Preferences');

    if (hasProfileInfo && hasChangePassword && hasPreferences) {
      passed('Verify account page contains required text');
    } else {
      const missing: string[] = [];
      if (!hasProfileInfo) missing.push('Profile Information');
      if (!hasChangePassword) missing.push('Change Password');
      if (!hasPreferences) missing.push('Preferences');
      failed('Verify account page contains required text', `Missing: ${missing.join(', ')}`);
    }
  } catch (e) {
    failed('Verify account page contains required text', (e as Error).message);
  }

  // Print results
  console.log('\n=== VERIFICATION REPORT ===\n');
  results.forEach(r => console.log(r));
  console.log('\n=== END REPORT ===');
});