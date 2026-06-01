import { test, expect } from '@playwright/test';

test('debug login flow', async ({ page }) => {
  const consoleLogs: string[] = [];
  const networkLogs: string[] = [];

  page.on('console', msg => consoleLogs.push(`${msg.type()}: ${msg.text()}`));
  page.on('pageerror', err => consoleLogs.push(`PAGE_ERROR: ${err.message}`));
  page.on('request', req => networkLogs.push(`REQUEST: ${req.method()} ${req.url()}`));
  page.on('response', res => networkLogs.push(`RESPONSE: ${res.status()} ${res.url()}`));

  // Navigate to login
  await page.goto('/auth/login');
  await page.waitForLoadState('networkidle');

  // Screenshot login page
  await page.screenshot({ path: 'test-results/login-page.png', fullPage: true });

  // Try to find login form elements
  const emailInput = page.locator('input[type="email"], input[name="email"], input[placeholder*="email"]').first();
  const passwordInput = page.locator('input[type="password"], input[name="password"]').first();
  const submitButton = page.locator('button[type="submit"], button:has-text("Login"), button:has-text("Sign in")').first();

  const hasEmail = await emailInput.isVisible().catch(() => false);
  const hasPassword = await passwordInput.isVisible().catch(() => false);
  const hasSubmit = await submitButton.isVisible().catch(() => false);

  console.log('Form elements found:', { hasEmail, hasPassword, hasSubmit });

  if (hasEmail && hasPassword && hasSubmit) {
    await emailInput.fill('test@example.com');
    await passwordInput.fill('testpassword123');
    await submitButton.click();

    // Wait for response
    await page.waitForTimeout(5000);

    // Screenshot after submit
    await page.screenshot({ path: 'test-results/login-after-submit.png', fullPage: true });

    console.log('URL after submit:', page.url());
  } else {
    console.log('Could not find login form elements');
    // Dump page content
    const body = await page.content();
    console.log('Page content length:', body.length);
  }

  console.log('=== CONSOLE LOGS ===');
  consoleLogs.forEach(log => console.log(log));

  console.log('=== NETWORK LOGS ===');
  networkLogs.forEach(log => console.log(log));
});