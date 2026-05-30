import { chromium } from 'playwright';

(async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  await page.goto('http://localhost:3000/dashboard', { waitUntil: 'networkidle' });

  // Inject user into localStorage
  await page.evaluate(() => {
    localStorage.setItem(
      'auth-storage',
      JSON.stringify({
        state: {
          user: {
            id: 'test-user-id',
            email: 'test@example.com',
            name: 'Test User',
            role: 'client',
            createdAt: new Date().toISOString(),
            onboardingComplete: false
          }
        },
        version: 0
      })
    );
  });

  // Reload to trigger rehydration via UserNav useEffect
  await page.reload({ waitUntil: 'networkidle' });
  await page.waitForTimeout(2000);

  // Screenshot raw page
  await page.screenshot({ path: '/Users/asan/videoforge/frontend/screenshot_dashboard_raw.png', fullPage: true });
  console.log('📸 Raw dashboard screenshot saved');

  // Check header content
  const headerHtml = await page.locator('header').first().innerHTML().catch(() => 'NO HEADER');
  console.log('Header HTML:', headerHtml.substring(0, 400));

  // Check for rounded-full button (avatar trigger)
  const roundedBtn = page.locator('header button[class*="rounded-full"]').first();
  if (await roundedBtn.isVisible().catch(() => false)) {
    console.log('✅ Avatar trigger (rounded-full button) found');
  } else {
    console.log('❌ Avatar trigger NOT found');
  }

  // Check all buttons in header
  const headerBtns = page.locator('header button');
  const count = await headerBtns.count();
  console.log(`Buttons in header: ${count}`);
  for (let i = 0; i < count; i++) {
    const cls = await headerBtns.nth(i).getAttribute('class');
    const text = await headerBtns.nth(i).textContent();
    console.log(`  Header button ${i}: class="${cls?.substring(0, 70)}" text="${text?.substring(0, 30)}"`);
  }

  // Click avatar and check dropdown
  const trigger = page.locator('header button[class*="rounded-full"]').first();
  if (await trigger.isVisible().catch(() => false)) {
    await trigger.click();
    await page.waitForTimeout(500);
    const dropdownItems = await page.locator('[data-slot="dropdown-menu-item"]').allTextContents();
    console.log('Dropdown items:', dropdownItems);
    const expected = ['Profile', 'Account', 'Billing', 'Settings', 'Sign out'];
    for (const item of expected) {
      if (dropdownItems.some((i) => i.includes(item))) {
        console.log(`✅ Dropdown item found: ${item}`);
      } else {
        console.log(`❌ Dropdown item missing: ${item}`);
      }
    }
  }

  // Navigate to account page
  await page.goto('http://localhost:3000/dashboard/account', { waitUntil: 'networkidle' });
  await page.waitForTimeout(500);
  const heading = await page.locator('h1:has-text("Account")').first().isVisible().catch(() => false);
  console.log(heading ? '✅ Account heading found' : '❌ Account heading NOT found');

  const sections = ['Profile Information', 'Change Password', 'Preferences'];
  for (const section of sections) {
    const found = await page.locator(`text=${section}`).first().isVisible().catch(() => false);
    console.log(found ? `✅ Account section found: ${section}` : `❌ Account section missing: ${section}`);
  }

  await page.screenshot({ path: '/Users/asan/videoforge/frontend/screenshot_account.png', fullPage: true });
  console.log('📸 Account page screenshot saved');

  await browser.close();
})();
