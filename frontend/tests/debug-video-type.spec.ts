import { test, expect, chromium, Page } from '@playwright/test';
import * as path from 'path';
import * as os from 'os';
import * as fs from 'fs';

function createMockToken(user: any): string {
  const header = { alg: 'HS256', typ: 'JWT' };
  const payload = { sub: user.id, email: user.email, name: user.name, role: user.role, iat: Math.floor(Date.now()/1000), exp: Math.floor(Date.now()/1000)+3600 };
  return `${btoa(JSON.stringify(header))}.${btoa(JSON.stringify(payload))}.mock-signature`;
}

async function setAuthStorage(page: Page, userData: any): Promise<void> {
  const authState = { state: { user: userData, isLoading: false, _hasHydrated: true }, version: 0 };
  await page.evaluate((authStr) => localStorage.setItem('auth-storage', authStr), JSON.stringify(authState));
  await page.evaluate((t) => localStorage.setItem('token', t), createMockToken(userData));
}

function createTestVideoFile(): string {
  const tmpDir = os.tmpdir();
  const filePath = path.join(tmpDir, 'dbg-video-2.mp4');
  if (!fs.existsSync(filePath)) {
    const mp4Data = Buffer.alloc(200);
    mp4Data.writeUInt32BE(0x18, 0);
    mp4Data.write('ftyp', 4);
    mp4Data.write('isom', 8);
    mp4Data.writeUInt32BE(0x200, 12);
    mp4Data.write('isomiso2mp41', 16);
    mp4Data.writeUInt32BE(180, 32);
    mp4Data.write('mdat', 36);
    fs.writeFileSync(filePath, mp4Data);
  }
  return filePath;
}

async function setupMockAPIs(page: Page) {
  await page.route('**/api/v1/briefs/brief-test-001', async (route) => {
    if (route.request().method() === 'GET') {
      await route.fulfill({ status: 200, contentType: 'application/json',
        body: JSON.stringify({ id: 'brief-test-001', title: 'Test', description: '', status: 'published',
          bounty_budget: 500, submissions_limit: 10, current_submissions: 0, client_name: 'Test', tags: [], created_at: new Date().toISOString() }) });
    } else await route.continue();
  });
  await page.route('**/api/v1/briefs/*/raw-footage/upload-url', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json',
      body: JSON.stringify({ upload_url: 'http://mock-storage/upload', storj_key: 'mock-key' }) });
  });
  await page.route('http://mock-storage/upload', async (route) => {
    await route.fulfill({ status: 200, body: '' });
  });
  await page.route('**/api/v1/briefs/*/raw-footage/confirm', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: '{}' });
  });
  await page.route('**/api/v1/briefs/*/raw-footage/download-url', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json',
      body: JSON.stringify({ download_url: 'http://example.com/dl', expires_in: 3600 }) });
  });
  await page.route('**/api/v1/videos**', async (route) => {
    await route.fulfill({ status: 200, contentType: 'application/json', body: JSON.stringify([]) });
  });
}

test('diagnostic: check file.type and file.url after upload', async () => {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  try {
    await page.goto('http://localhost:3000/dashboard');
    await page.waitForTimeout(500);
    await setAuthStorage(page, { id: 'usr-test-001', email: 't@t.com', name: 'Test', role: 'client' });
    await setupMockAPIs(page);

    await page.evaluate(() => {
      Object.keys(localStorage).filter(k => k.startsWith('brief-files-')).forEach(k => localStorage.removeItem(k));
    });

    await page.goto('http://localhost:3000/dashboard/briefs/brief-test-001');
    await page.waitForSelector('text=Upload Raw Footage', { timeout: 15000 });

    // Upload
    const videoPath = createTestVideoFile();
    const fileInput = page.locator('input[type="file"]').first();
    await fileInput.setInputFiles(videoPath);
    
    await page.waitForTimeout(5000);

    // Check localStorage for saved files
    const storedFiles = await page.evaluate(() => {
      const raw = localStorage.getItem('brief-files-brief-test-001');
      return raw ? JSON.parse(raw) : null;
    });
    console.log('Stored files:', JSON.stringify(storedFiles, null, 2));

    // Check what type and url we have
    if (storedFiles && storedFiles.length > 0) {
      const file = storedFiles[0];
      console.log(`File type: '${file.type}'`);
      console.log(`File url: '${file.url}'`);
      console.log(`File url starts with 'blob:': ${file.url?.startsWith('blob:')}`);
      console.log(`File type starts with 'video/': ${file.type?.startsWith('video/')}`);
    } else {
      console.log('NO files stored!');
    }

    // Also check the actual file input to see what MIME type Playwright gave it
    const mimeType = await page.evaluate(() => {
      const input = document.querySelector('input[type="file"]') as HTMLInputElement;
      if (input && input.files && input.files.length > 0) {
        return input.files[0].type;
      }
      return null;
    });
    console.log(`File input MIME type (from File object): '${mimeType}'`);

    // Check the rendered file button's data attribute or text
    const fileButton = page.locator('[role="button"]').filter({ hasText: 'dbg-video-2.mp4' }).first();
    if (await fileButton.isVisible()) {
      console.log('File button is visible');
      
      // Try double-click with explicit dispatchEvent
      console.log('Trying dispatchEvent dblclick...');
      await fileButton.evaluate((el) => {
        const rect = el.getBoundingClientRect();
        const x = rect.left + rect.width / 2;
        const y = rect.top + rect.height / 2;
        
        // Full click sequence: pointerdown → pointerup → click × 2 → dblclick
        const events = ['pointerdown', 'pointerup', 'click', 'pointerdown', 'pointerup', 'click', 'dblclick'];
        events.forEach(type => {
          const EventClass = type.startsWith('pointer') ? PointerEvent : MouseEvent;
          el.dispatchEvent(new EventClass(type, { bubbles: true, cancelable: true, clientX: x, clientY: y }));
        });
      });
      
      await page.waitForTimeout(2000);
      
      // Check modal
      const modal = page.locator('.fixed.inset-0').first();
      const modalVisible = await modal.isVisible().catch(() => false);
      console.log(`After dispatchEvent, modal visible: ${modalVisible}`);
      
      if (modalVisible) {
        const video = page.locator('video').first();
        const videoVisible = await video.isVisible().catch(() => false);
        const src = videoVisible ? await video.getAttribute('src') : null;
        console.log(`Video visible: ${videoVisible}, src: ${src?.substring(0, 80)}`);
      }
    } else {
      console.log('File button NOT visible!');
    }
  } finally {
    await browser.close();
  }
});
