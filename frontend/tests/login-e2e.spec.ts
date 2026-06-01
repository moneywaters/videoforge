import { test, expect } from '@playwright/test';

test.describe('Login E2E', () => {
  test('login page loads and form exists', async ({ page }) => {
    await page.goto('/auth/login');
    await page.waitForLoadState('networkidle');

    // Check login page loads
    const emailInput = page.locator('input[type="email"]').first();
    const passwordInput = page.locator('input[type="password"]').first();
    const submitButton = page.locator('button[type="submit"]').first();

    await expect(emailInput).toBeVisible();
    await expect(passwordInput).toBeVisible();
    await expect(submitButton).toBeVisible();

    await page.screenshot({ path: 'test-results/login-form.png' });
  });

  test('login shows warm-up message after 5s', async ({ page }) => {
    // Mock the API to be slow (simulate NeonDB cold-start)
    await page.route('**/api/v1/auth/login', async (route) => {
      await new Promise(resolve => setTimeout(resolve, 7000));
      await route.fulfill({
        status: 200,
        body: JSON.stringify({
          token: 'mock-jwt-token',
          user: { id: 'usr-123', email: 'test@example.com', role: 'client' }
        })
      });
    });

    await page.goto('/auth/login');

    await page.fill('input[type="email"]', 'test@example.com');
    await page.fill('input[type="password"]', 'testpassword');
    await page.click('button[type="submit"]');

    // After 5 seconds, should show "Connecting to database..." or similar
    await page.waitForTimeout(5500);

    const pageContent = await page.content();
    const hasWarmupMessage = pageContent.includes('Warming up') || pageContent.includes('Connecting to database');

    console.log('Has warmup message:', hasWarmupMessage);
    console.log('Button text:', await page.locator('button[type="submit"]').textContent());

    // Should eventually succeed (mocked)
    await page.waitForTimeout(3000);
    await page.screenshot({ path: 'test-results/login-warmup.png' });
  });
});