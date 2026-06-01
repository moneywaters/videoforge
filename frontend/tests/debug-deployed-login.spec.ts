import { test, expect, chromium } from '@playwright/test';

test.describe('Deployed login test', () => {
  test('test login on cutthroatreels.com', async () => {
    const browser = await chromium.launch();
    const page = await browser.newPage();
    
    const consoleLogs: string[] = [];
    const networkLogs: string[] = [];
    
    page.on('console', msg => consoleLogs.push(`${msg.type()}: ${msg.text()}`));
    page.on('pageerror', err => consoleLogs.push(`PAGE_ERROR: ${err.message}`));
    page.on('request', req => networkLogs.push(`REQUEST: ${req.method()} ${req.url()}`));
    page.on('response', res => networkLogs.push(`RESPONSE: ${res.status()} ${res.url()}`));
    
    // Navigate to deployed frontend login
    await page.goto('https://cutthroatreels.com/auth/login');
    await page.waitForLoadState('networkidle');
    await page.screenshot({ path: 'test-results/deployed-login-page.png', fullPage: true });
    
    console.log('=== DEPLOYED PAGE URL ===');
    console.log('URL:', page.url());
    console.log('Title:', await page.title());
    
    // Find login form
    const emailInput = page.locator('input[type="email"], input[name="email"]').first();
    const passwordInput = page.locator('input[type="password"]').first();
    const submitButton = page.locator('button[type="submit"]').first();
    
    const hasEmail = await emailInput.isVisible().catch(() => false);
    const hasPassword = await passwordInput.isVisible().catch(() => false);
    const hasSubmit = await submitButton.isVisible().catch(() => false);
    
    console.log('Form elements:', { hasEmail, hasPassword, hasSubmit });
    
    if (hasEmail && hasPassword && hasSubmit) {
      await emailInput.fill('test@example.com');
      await passwordInput.fill('testpassword123');
      await submitButton.click();
      
      await page.waitForTimeout(5000);
      
      await page.screenshot({ path: 'test-results/deployed-login-after-submit.png', fullPage: true });
      
      console.log('URL after submit:', page.url());
    }
    
    console.log('=== CONSOLE LOGS ===');
    consoleLogs.forEach(log => console.log(log));
    
    console.log('=== NETWORK LOGS ===');
    networkLogs.filter(log => log.includes('/auth/') || log.includes('brief') || log.includes('error')).forEach(log => console.log(log));
    
    await browser.close();
  });
});