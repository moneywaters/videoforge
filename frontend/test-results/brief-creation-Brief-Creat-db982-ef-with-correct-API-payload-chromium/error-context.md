# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: brief-creation.spec.ts >> Brief Creation >> should create brief with correct API payload
- Location: tests/brief-creation.spec.ts:33:7

# Error details

```
Test timeout of 30000ms exceeded.
```

```
Error: page.waitForLoadState: Test timeout of 30000ms exceeded.
```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - region "Notifications alt+T"
  - generic [ref=e2]:
    - complementary [ref=e3]:
      - navigation [ref=e4]:
        - link "Dashboard" [ref=e5] [cursor=pointer]:
          - /url: /dashboard
        - link "Briefs" [ref=e6] [cursor=pointer]:
          - /url: /dashboard/briefs
        - link "Videos" [ref=e7] [cursor=pointer]:
          - /url: /dashboard/videos
        - link "Campaigns" [ref=e8] [cursor=pointer]:
          - /url: /dashboard/campaigns
        - link "Earnings" [ref=e9] [cursor=pointer]:
          - /url: /dashboard/earnings
        - link "Leaderboard" [ref=e10] [cursor=pointer]:
          - /url: /dashboard/leaderboard
    - generic [ref=e11]:
      - banner [ref=e12]:
        - heading "Dashboard" [level=1] [ref=e13]
      - main [ref=e14]:
        - generic [ref=e15]:
          - generic [ref=e17]:
            - heading "Briefs" [level=1] [ref=e18]
            - paragraph [ref=e19]: Manage your video briefs and campaigns
          - generic [ref=e20]:
            - button "All (0)" [ref=e21]
            - button "Open (0)" [ref=e22]
            - button "Closed (0)" [ref=e23]
            - button "Draft (0)" [ref=e24]
  - generic [ref=e92]:
    - img [ref=e94]
    - button "Open Tanstack query devtools" [ref=e142] [cursor=pointer]:
      - img [ref=e143]
  - button "Open Next.js Dev Tools" [ref=e196] [cursor=pointer]:
    - img [ref=e197]
  - alert [ref=e200]
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | // Mock data for user
  4   | const mockUser = {
  5   |   id: 'test-user-id',
  6   |   email: 'test@example.com',
  7   |   name: 'Test User',
  8   |   role: 'client',
  9   |   avatar_url: null,
  10  | };
  11  | 
  12  | // Mock data for created brief response
  13  | const mockBriefResponse = {
  14  |   id: 'test-brief-id',
  15  |   client_id: 'test-user-id',
  16  |   title: 'Test Brief',
  17  |   description: 'Test description',
  18  |   goals: '',
  19  |   target_audience: '',
  20  |   tone: '',
  21  |   style_preferences: '',
  22  |   cta: '',
  23  |   status: 'draft',
  24  |   bounty_budget: 500,
  25  |   bounty_deposited: false,
  26  |   submissions_limit: 5,
  27  |   is_blind: false,
  28  |   has_raw_footage: false,
  29  |   created_at: '2024-01-01T00:00:00Z',
  30  | };
  31  | 
  32  | test.describe('Brief Creation', () => {
  33  |   test('should create brief with correct API payload', async ({ page }) => {
  34  |     const timestamp = Date.now();
  35  |     const testTitle = `Test Brief ${timestamp}`;
  36  |     const testDescription = 'Test description';
  37  |     const testBountyBudget = 500;
  38  |     const testSubmissionsLimit = 5;
  39  | 
  40  |     // Track captured data
  41  |     let capturedRequestBody: any = null;
  42  |     let capturedResponseStatus: number | null = null;
  43  | 
  44  |     // Set up mock token for auth
  45  |     await page.addInitScript(() => {
  46  |       localStorage.setItem('token', 'mock-token-12345');
  47  |     });
  48  | 
  49  |     // Mock ALL API endpoints using page.route()
  50  |     // Note: The app calls the external videoforge-gateway.fly.dev API directly
  51  |     
  52  |     // GET /api/v1/auth/me → return mock user (client role)
  53  |     await page.route('**/videoforge-gateway.fly.dev/api/v1/auth/me', async (route) => {
  54  |       route.fulfill({
  55  |         status: 200,
  56  |         body: JSON.stringify(mockUser),
  57  |         headers: { 'Content-Type': 'application/json' },
  58  |       });
  59  |     });
  60  | 
  61  |     // GET /api/v1/briefs → return empty array (for the briefs list page)
  62  |     await page.route('**/videoforge-gateway.fly.dev/api/v1/briefs', async (route) => {
  63  |       if (route.request().method() === 'GET') {
  64  |         route.fulfill({
  65  |           status: 200,
  66  |           body: JSON.stringify([]),
  67  |           headers: { 'Content-Type': 'application/json' },
  68  |         });
  69  |       }
  70  |     });
  71  | 
  72  |     // GET /api/v1/briefs/{id} → return mock brief (for detail page after creation)
  73  |     await page.route('**/videoforge-gateway.fly.dev/api/v1/briefs/test-brief-id', async (route) => {
  74  |       route.fulfill({
  75  |         status: 200,
  76  |         body: JSON.stringify(mockBriefResponse),
  77  |         headers: { 'Content-Type': 'application/json' },
  78  |       });
  79  |     });
  80  | 
  81  |     // POST /api/v1/briefs → capture request body, return 201 with created brief
  82  |     await page.route('**/videoforge-gateway.fly.dev/api/v1/briefs', async (route) => {
  83  |       if (route.request().method() === 'POST') {
  84  |         try {
  85  |           capturedRequestBody = await route.request().postDataJSON();
  86  |         } catch (e: any) {
  87  |           console.log('Could not parse request body:', e.message);
  88  |         }
  89  | 
  90  |         capturedResponseStatus = 201;
  91  |         route.fulfill({
  92  |           status: 201,
  93  |           body: JSON.stringify(mockBriefResponse),
  94  |           headers: { 'Content-Type': 'application/json' },
  95  |         });
  96  |       }
  97  |     });
  98  | 
  99  |     // Navigate to /dashboard/briefs/new
  100 |     await page.goto('/dashboard/briefs/new');
> 101 |     await page.waitForLoadState('networkidle');
      |                ^ Error: page.waitForLoadState: Test timeout of 30000ms exceeded.
  102 |     await page.waitForTimeout(1500);
  103 | 
  104 |     // Fill the brief creation form with test data
  105 |     const titleInput = page.locator('input#title');
  106 |     await titleInput.waitFor({ state: 'visible', timeout: 10000 });
  107 |     await titleInput.fill(testTitle);
  108 | 
  109 |     // Description (optional field)
  110 |     const descInput = page.locator('textarea#description');
  111 |     if (await descInput.isVisible().catch(() => false)) {
  112 |       await descInput.fill(testDescription);
  113 |     }
  114 | 
  115 |     // Bounty Budget
  116 |     await page.locator('input#bountyBudget').fill(testBountyBudget.toString());
  117 | 
  118 |     // Submission Limit
  119 |     await page.locator('input#submissionLimit').fill(testSubmissionsLimit.toString());
  120 |     await page.waitForTimeout(300);
  121 | 
  122 |     // Click the submit/create button (Start AI Interview)
  123 |     await page.locator('button:has-text("Start AI Interview")').click();
  124 |     await page.waitForTimeout(3000);
  125 | 
  126 |     // Now in AI interview mode - answer all 5 questions
  127 |     for (let q = 0; q < 5; q++) {
  128 |       const answerInput = page.locator('input[placeholder*="Type your answer"]');
  129 |       await answerInput.waitFor({ state: 'visible', timeout: 5000 });
  130 |       await answerInput.fill(`Test answer ${q + 1}`);
  131 | 
  132 |       // Send button
  133 |       const sendBtn = page.locator('div.flex button').last();
  134 |       await sendBtn.click();
  135 |       await page.waitForTimeout(1200);
  136 |     }
  137 | 
  138 |     // Click Generate Brief button
  139 |     const generateBtn = page.getByRole('button', { name: /Generate Brief/i });
  140 |     if (await generateBtn.isVisible().catch(() => false)) {
  141 |       await generateBtn.click();
  142 |     }
  143 |     await page.waitForTimeout(3000);
  144 | 
  145 |     // Verify the POST request body contains all expected fields with correct types
  146 |     expect(capturedRequestBody).not.toBeNull();
  147 | 
  148 |     if (capturedRequestBody) {
  149 |       expect(typeof capturedRequestBody.title).toBe('string');
  150 |       expect(capturedRequestBody.title).toBe(testTitle);
  151 | 
  152 |       expect(typeof capturedRequestBody.description).toBe('string');
  153 |       expect(capturedRequestBody.description).toContain(testDescription);
  154 | 
  155 |       expect(typeof capturedRequestBody.bounty_budget).toBe('number');
  156 |       expect(capturedRequestBody.bounty_budget).toBe(testBountyBudget);
  157 | 
  158 |       expect(typeof capturedRequestBody.submissions_limit).toBe('number');
  159 |       expect(capturedRequestBody.submissions_limit).toBe(testSubmissionsLimit);
  160 |     }
  161 | 
  162 |     // Verify response status is 201
  163 |     expect(capturedResponseStatus).toBe(201);
  164 |   });
  165 | });
```