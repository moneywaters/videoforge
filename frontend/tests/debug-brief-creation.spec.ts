import { test, expect } from '@playwright/test';
import * as fs from 'fs';
import * as path from 'path';

/**
 * Debug test for Brief Creation Flow
 * 
 * Run with:
 * npx playwright test tests/debug-brief-creation.spec.ts --headed --timeout=30000
 */

// Helper to create a valid JWT
function createMockJWT(payload: {
  sub: string;
  email: string;
  role: string;
  name?: string;
}): string {
  const header = { alg: 'HS256', typ: 'JWT' };
  const now = Math.floor(Date.now() / 1000);
  const exp = now + 86400 * 7; // 7 days
  
  const fullPayload = {
    ...payload,
    iat: now,
    exp,
    iss: 'videoforge',
  };
  
  // Simple base64url encoding (not cryptographically valid, but sufficient for mock)
  const encodeBase64Url = (str: string): string => {
    return Buffer.from(str).toString('base64')
      .replace(/\+/g, '-')
      .replace(/\//g, '_')
      .replace(/=/g, '');
  };
  
  const base64Header = encodeBase64Url(JSON.stringify(header));
  const base64Payload = encodeBase64Url(JSON.stringify(fullPayload));
  
  return `${base64Header}.${base64Payload}.mock_signature`;
}

test.describe('Brief Creation Debug', () => {
  let consoleLogs: Array<{ type: string; text: string; timestamp: string }> = [];
  let networkLogs: Array<{ method: string; url: string; status?: number; body?: unknown; headers?: Record<string, string> }> = [];
  let screenshotPath = '';

  test.beforeEach(() => {
    consoleLogs = [];
    networkLogs = [];
    screenshotPath = '';
  });

  test('should debug brief creation flow', async ({ page }) => {
    // 1. Set up console log capture
    page.on('console', msg => {
      const log = {
        type: msg.type(),
        text: msg.text(),
        timestamp: new Date().toISOString(),
      };
      consoleLogs.push(log);
      console.log(`[CONSOLE.${msg.type().toUpperCase()}] ${msg.text()}`);
    });

    // 2. Capture page errors
    page.on('pageerror', err => {
      const log = {
        type: 'pageerror',
        text: err.message,
        timestamp: new Date().toISOString(),
      };
      consoleLogs.push(log);
      console.log(`[PAGE ERROR] ${err.message}`);
    });

    // 3. Capture network requests
    page.on('request', req => {
      const url = req.url();
      const method = req.method();
      console.log(`[REQUEST] ${method} ${url}`);
      
      // Capture /api/v1/briefs requests
      if (url.includes('/api/v1/briefs')) {
        networkLogs.push({
          method,
          url,
        });
      }
    });

    // 4. Capture network responses
    page.on('response', async res => {
      const url = res.url();
      const status = res.status();
      console.log(`[RESPONSE] ${status} ${res.statusText()} ${url}`);
      
      // Capture /api/v1/briefs response details
      if (url.includes('/api/v1/briefs')) {
        let body: unknown;
        try {
          const text = await res.text();
          try {
            body = JSON.parse(text);
          } catch {
            body = text;
          }
        } catch {
          body = '<unable to capture>';
        }
        
        // Get headers
        const headers: Record<string, string> = {};
        res.headers().forEach((value, key) => {
          headers[key] = value;
        });
        
        // Update the network log with response info
        const logEntry = networkLogs.find(l => l.url === url);
        if (logEntry) {
          logEntry.status = status;
          logEntry.body = body;
          logEntry.headers = headers;
        }
      }
    });

    // =====================================================
    // STEP 1: Navigate to /auth/login
    // =====================================================
    console.log('\n=== STEP 1: Navigating to /auth/login ===\n');
    await page.goto('/auth/login');
    await page.waitForLoadState('domcontentloaded');
    console.log('Login page loaded');

    // =====================================================
    // STEP 2: Set mock JWT in localStorage (properly for zustand)
    // =====================================================
    console.log('\n=== STEP 2: Setting mock JWT in localStorage ===\n');
    
    const mockToken = createMockJWT({
      sub: 'usr-test-client-001',
      email: 'test@example.com',
      role: 'client',
      name: 'Test Client',
    });

    // Set the JWT token first (this is what the OAuth flow does)
    await page.evaluate((token) => {
      localStorage.setItem('token', token);
    }, mockToken);
    
    // Set the auth-storage (zustand persist) - this is what gets read at hydration
    const userState = {
      state: {
        user: {
          id: 'usr-test-client-001',
          email: 'test@example.com',
          name: 'Test Client',
          role: 'client' as const,
          createdAt: new Date().toISOString(),
          onboardingComplete: true,
        },
        isLoading: false,
        _hasHydrated: true,
      },
      version: 0,
    };
    
    await page.evaluate((data) => {
      localStorage.setItem('auth-storage', JSON.stringify(data));
    }, userState);
    
    console.log('Mock JWT set with role: client');
    console.log('Token in localStorage:', await page.evaluate(() => localStorage.getItem('token')?.substring(0, 30)));
    console.log('Auth storage:', await page.evaluate(() => localStorage.getItem('auth-storage')));

    // Force a page reload to let zustand rehydrate
    console.log('\nReloading page to trigger auth rehydration...');
    await page.reload();
    await page.waitForLoadState('domcontentloaded');
    await page.waitForTimeout(1000);
    
    // Verify auth worked
    const userCheck = await page.evaluate(() => {
      const stored = localStorage.getItem('auth-storage');
      if (stored) {
        try {
          const parsed = JSON.parse(stored);
          return parsed.state?.user?.role || 'NO USER';
        } catch {
          return 'PARSE ERROR';
        }
      }
      return 'NO STORAGE';
    });
    console.log('Verified user role after reload:', userCheck);

    // =====================================================
    // STEP 3: Navigate to /dashboard/briefs/new
    // =====================================================
    console.log('\n=== STEP 3: Navigating to /dashboard/briefs/new ===\n');
    await page.goto('/dashboard/briefs/new');
    await page.waitForLoadState('domcontentloaded');
    
    // Wait a bit for any auth checks
    await page.waitForTimeout(2000);
    
    console.log('Brief creation page loaded');
    console.log('Current URL:', page.url());

    // Check if we're stuck in loading
    const loadingSpinner = page.locator('.animate-spin').first();
    const isLoading = await loadingSpinner.isVisible().catch(() => false);
    console.log('Loading spinner visible:', isLoading);

    if (isLoading) {
      console.log('\n!!! PAGE IS STUCK IN LOADING STATE !!!\n');
      await page.screenshot({ path: 'test-results/debug-brief-loading-stuck.png' });
      screenshotPath = 'test-results/debug-brief-loading-stuck.png';
    }

    // =====================================================
    // STEP 4: Fill in the form fields
    // =====================================================
    console.log('\n=== STEP 4: Filling brief form ===\n');

    // Fill Project Name / Brief Title
    const titleInput = page.locator('input#title');
    if (await titleInput.isVisible()) {
      await titleInput.fill('Test Brief');
      console.log('Filled title: Test Brief');
    }

    // Fill Description (optional but helps)
    const descriptionInput = page.locator('textarea#description');
    if (await descriptionInput.isVisible()) {
      await descriptionInput.fill('Test description for debugging');
      console.log('Filled description');
    }

    // Fill Budget
    const budgetInput = page.locator('input#bountyBudget');
    if (await budgetInput.isVisible()) {
      await budgetInput.fill('500');
      console.log('Filled budget: 500');
    }

    // Fill Submission Limit
    const submissionLimitInput = page.locator('input#submissionLimit');
    if (await submissionLimitInput.isVisible()) {
      await submissionLimitInput.fill('5');
      console.log('Filled submission limit: 5');
    }

    // =====================================================
    // STEP 5: Click "Start AI Interview" button
    // =====================================================
    console.log('\n=== STEP 5: Clicking Start AI Interview button ===\n');

    const startButton = page.locator('button:has-text("Start AI Interview")');
    if (await startButton.isVisible()) {
      console.log('Start AI Interview button found, clicking...');
      await startButton.click();
      console.log('Clicked Start AI Interview button');
      
      // Wait for the AI interview step to load
      await page.waitForTimeout(2000);
      
      console.log('\n=== In AI Interview mode ===\n');
      console.log('Current URL:', page.url());
      
      // =====================================================
      // STEP 5b: Answer AI questions to complete the flow
      // =====================================================
      console.log('\n=== STEP 5b: Answering AI questions ===\n');
      
      const questions = [
        'test', // target audience
        'test', // tone and style
        'test', // CTA
        'test', // features  
        '30s'   // length preference
      ];
      
      for (let i = 0; i < questions.length; i++) {
        console.log(`\nAnswering question ${i + 1}/${questions.length}: "${questions[i]}"`);
        
        const input = page.locator('input[placeholder*="Type your answer"]');
        if (await input.isVisible({ timeout: 3000 })) {
          await input.fill(questions[i]);
          console.log(`Filled answer: ${questions[i]}`);
          
          // Click send button
          const sendButton = page.locator('button:has-text("send"), button:has(svg)').last();
          if (await sendButton.isVisible()) {
            await sendButton.click();
            console.log('Sent answer');
            await page.waitForTimeout(1000);
          }
        } else {
          console.log('Input not found, waiting...');
          await page.waitForTimeout(500);
        }
      }
      
      console.log('\nAll questions answered, looking for Generate Brief button...');
      
      // Click Generate Brief button if visible
      const generateButton = page.locator('button:has-text("Generate Brief")');
      if (await generateButton.isVisible({ timeout: 5000 })) {
        console.log('Generate Brief button found, clicking...');
        await generateButton.click();
        console.log('Clicked Generate Brief button');
        await page.waitForTimeout(3000);
      } else {
        console.log('Generate Brief button NOT found - checking for other button...');
        const createButton = page.locator('button:has-text("Create Brief")');
        if (await createButton.isVisible()) {
          await createButton.click();
          console.log('Clicked Create Brief button');
          await page.waitForTimeout(3000);
        }
      }
      
      console.log('\nAfter brief creation attempt:');
      console.log('URL:', page.url());
    } else {
      console.log('Start AI Interview button NOT found');
      console.log('Page content preview:', await page.content().catch(() => 'N/A').then(c => c.substring(0, 500)));
    }

    // =====================================================
    // STEP 6: Wait 10 seconds and screenshot
    // =====================================================
    console.log('\n=== STEP 6: Waiting 10 seconds and taking screenshot ===\n');
    await page.waitForTimeout(10000);

    const ssPath = 'test-results/debug-brief-after-wait.png';
    await page.screenshot({ path: ssPath, fullPage: true });
    screenshotPath = ssPath;
    console.log('Screenshot saved:', ssPath);

    // =====================================================
    // STEP 7: Save all console logs to file
    // =====================================================
    console.log('\n=== STEP 7: Saving console logs to file ===\n');

    const consoleLogPath = 'test-results/debug-console-logs.json';
    fs.writeFileSync(consoleLogPath, JSON.stringify(consoleLogs, null, 2));
    console.log('Console logs saved to:', consoleLogPath);

    // =====================================================
    // STEP 8: Report network request details
    // =====================================================
    console.log('\n=== STEP 8: Network request details for /api/v1/briefs ===\n');

    const briefApiLogs = networkLogs.filter(l => l.url.includes('/api/v1/briefs'));
    if (briefApiLogs.length > 0) {
      const networkLogPath = 'test-results/debug-network-requests.json';
      fs.writeFileSync(networkLogPath, JSON.stringify(briefApiLogs, null, 2));
      console.log('Network logs saved to:', networkLogPath);
      
      briefApiLogs.forEach(log => {
        console.log(`\n--- Request: ${log.method} ${log.url} ---`);
        console.log('Status:', log.status || 'pending');
        const bodyStr = log.body ? JSON.stringify(log.body, null, 2) : 'undefined';
        console.log('Body:', bodyStr?.substring(0, 500) || bodyStr);
      });
    } else {
      console.log('No /api/v1/briefs requests captured');
    }

    // =====================================================
    // FINAL SUMMARY
    // =====================================================
    console.log('\n========================================');
    console.log('=== FINAL SUMMARY ===');
    console.log('========================================');
    console.log('URL:', page.url());
    console.log('Screenshot:', screenshotPath);
    console.log('Console log count:', consoleLogs.length);
    console.log('Network request count:', networkLogs.length);
    console.log('Brief API request count:', briefApiLogs.length);
    
    // Check if we successfully created the brief
    const url = page.url();
    const success = url.includes('/dashboard/briefs/') && !url.includes('/new');
    console.log('Brief created:', success);
    
    if (!success) {
      console.log('\n!!! BRIEF CREATION DID NOT SUCCEED !!!');
      console.log('Expected redirect to /dashboard/briefs/[id]');
      console.log('Actual URL:', url);
    }
    console.log('========================================\n');

    // Additional screenshot of final state
    await page.screenshot({ path: 'test-results/debug-brief-final-state.png', fullPage: true });
    console.log('Final state screenshot: test-results/debug-brief-final-state.png');
  });

  test.afterEach(async ({ }, testInfo) => {
    // Save test results summary
    const summary = {
      testName: testInfo.title,
      status: testInfo.status,
      screenshot: screenshotPath,
      consoleLogCount: consoleLogs.length,
      networkRequestCount: networkLogs.length,
    };
    
    const summaryPath = 'test-results/debug-test-summary.json';
    fs.writeFileSync(summaryPath, JSON.stringify(summary, null, 2));
  });
});